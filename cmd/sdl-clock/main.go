package main

import (
	"fmt"
	"github.com/SpComb/osc-tally/clock"
	"github.com/depili/go-rgb-led-matrix/bdf"
	// "github.com/depili/go-rgb-led-matrix/matrix"
	"github.com/desertbit/timer"
	"github.com/hypebeast/go-osc/osc"
	"github.com/jessevdk/go-flags"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"os"
	"os/signal"
	"time"
)

var staticSDLColor sdl.Color = sdl.Color{80, 80, 0, 255} // 12 static indicator circles
var secSDLColor sdl.Color = sdl.Color{200, 0, 0, 255}
var textSDLColor sdl.Color
var window *sdl.Window
var renderer *sdl.Renderer
var parser = flags.NewParser(&Options, flags.Default)

var textureSource sdl.Rect
var staticTexture *sdl.Texture
var secTexture *sdl.Texture

func runOSC(oscServer *osc.Server) {
	err := oscServer.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func main() {
	if _, err := parser.Parse(); err != nil {
		panic(err)
	}

	// Initialize osc listener
	var oscServer = osc.Server{
		Addr: Options.ListenAddr,
	}

	var clockServer = clock.MakeServer(&oscServer)
	log.Printf("osc server: listen %v", oscServer.Addr)

	go runOSC(&oscServer)

	// GPIO pin for toggling between timezones
	if err := embd.InitGPIO(); err != nil {
		panic(err)
	}

	timePin, err := embd.NewDigitalPin(Options.TimePin)
	if err != nil {
		panic(err)
	} else if err := timePin.SetDirection(embd.In); err != nil {
		panic(err)
	}

	fmt.Printf("GPIO initialized.\n")

	/*
		// Load timezones
		local, err := time.LoadLocation(Options.LocalTime)
		if err != nil {
			panic(err)
		}

		foreign, err := time.LoadLocation(Options.ForeignTime)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Timezones loaded.\n")
	*/

	// Parse font for clock text
	font, err := bdf.Parse(Options.Font)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Fonts loaded.\n")

	// Initialize SDL

	if err = sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize SDL: %s\n", err)
		return
	}
	defer sdl.Quit()

	if window, err = sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight, sdl.WINDOW_SHOWN); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return
	}
	defer window.Destroy()

	if renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		return // don't use os.Exit(3); otherwise, previous deferred calls will never run
	}

	sdl.ShowCursor(0) // Hide mouse cursor

	renderer.Clear()
	defer renderer.Destroy()

	fmt.Printf("SDL init done\n")

	rendererInfo, _ := renderer.GetInfo()
	fmt.Printf("Renderer: %v\n", rendererInfo.Name)

	// Clock colors from flags
	textSDLColor = sdl.Color{Options.TextRed, Options.TextGreen, Options.TextBlue, 255}
	staticSDLColor = sdl.Color{Options.StaticRed, Options.StaticGreen, Options.StaticBlue, 255}
	secSDLColor = sdl.Color{Options.SecRed, Options.SecGreen, Options.SecBlue, 255}
	// Default color for the OSC field (black)
	tallyColor := sdl.Color{0, 0, 0, 255}

	var textureSize int32 = 40
	var textureCoord int32 = 20
	var textureRadius int32 = 19

	// Create a texture for circles
	if Options.Small {
		textureSize = 8
		textureCoord = 3
		textureRadius = 3
		gridStartX = 24
		gridStartY = 24
		gridSize = 3
		gridSpacing = 4
		secCircles = smallSecCircles
		staticCircles = smallStaticCircles
	}
	staticTexture, _ = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, textureSize, textureSize)
	renderer.SetRenderTarget(staticTexture)
	gfx.FilledCircleColor(renderer, textureCoord, textureCoord, textureRadius, staticSDLColor)
	if !Options.Small {
		gfx.AACircleColor(renderer, textureCoord, textureCoord, textureRadius, staticSDLColor)
	}

	secTexture, _ = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, textureSize, textureSize)
	renderer.SetRenderTarget(secTexture)
	gfx.FilledCircleColor(renderer, textureCoord, textureCoord, textureRadius, secSDLColor)
	if !Options.Small {
		gfx.AACircleColor(renderer, textureCoord, textureCoord, textureRadius, secSDLColor)
	}

	renderer.SetRenderTarget(nil)
	textureSource = sdl.Rect{0, 0, textureSize, textureSize}

	// Trap SIGINT aka Ctrl-C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Create initial bitmaps for clock text
	hourBitmap := font.TextBitmap("15")
	minuteBitmap := font.TextBitmap("04")
	secondBitmap := font.TextBitmap("05")
	tallyBitmap := font.TextBitmap("  ")

	updateTicker := time.NewTicker(time.Millisecond * 30)
	oscChan := clockServer.Listen()

	timeout := timer.NewTimer(time.Duration(Options.Timeout) * time.Millisecond)

	engine, err := clock.MakeEngine(Options.LocalTime, time.Duration(Options.Flash)*time.Millisecond)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Entering main loop\n")

	for {
		select {
		case message := <-oscChan:
			// New OSC message received
			fmt.Printf("Got new osc data.\n")
			switch message.Type {
			case "count":
				msg := message.CountMessage
				tallyColor = sdl.Color{uint8(msg.ColorRed), uint8(msg.ColorGreen), uint8(msg.ColorBlue), 255}
				tallyBitmap = font.TextBitmap(fmt.Sprintf("%1s%02d%1s", msg.Symbol, msg.Count, msg.Unit))
				timeout.Reset(time.Duration(Options.Timeout) * time.Millisecond)
			case "countdownStart":
				msg := message.CountdownMessage
				engine.StartCountdown(time.Duration(msg.Seconds) * time.Second)
			case "countdownModify":
				msg := message.CountdownMessage
				engine.ModifyCountdown(time.Duration(msg.Seconds) * time.Second)
			case "countup":
				engine.StartCountup()
			case "kill":
				engine.Kill()
			case "normal":
				engine.Normal()
			}
		case <-timeout.C:
			// OSC message timeout
			tallyBitmap = font.TextBitmap("")
		case <-sigChan:
			// SIGINT received, shutdown gracefully
			os.Exit(1)
		case <-updateTicker.C:
			engine.Update()
			seconds := engine.Leds
			hourBitmap = font.TextBitmap(engine.Hours)
			minuteBitmap = font.TextBitmap(engine.Minutes)
			secondBitmap = font.TextBitmap(engine.Seconds)

			// Clear SDL canvas
			renderer.SetDrawColor(0, 0, 0, 255) // Black
			renderer.Clear()                    // Clear screen

			// renderer.SetDrawColor(255, 0, 0, 255) // Black
			// renderer.DrawRect(&sdl.Rect{0, 0, 176, 175})

			// Dots between hours and minutes
			if engine.Dots {
				drawDots()
			}

			// Draw the text
			drawBitmask(hourBitmap, textSDLColor, 10, 0)
			drawBitmask(minuteBitmap, textSDLColor, 10, 17)
			drawBitmask(secondBitmap, textSDLColor, 21, 8)
			drawBitmask(tallyBitmap, tallyColor, 0, 2)

			drawStaticCircles()
			drawSecondCircles(seconds)

			// Update the canvas
			renderer.Present()
		}
	}
}

func drawSecondCircles(seconds int) {
	// Draw second circles
	for i := 0; i <= int(seconds); i++ {
		// gfx.FilledCircleColor(renderer, secCircles[i][0], secCircles[i][1], 20, secSDLColor)
		dest := sdl.Rect{secCircles[i][0] - 20, secCircles[i][1] - 20, 40, 40}
		if Options.Small {
			dest = sdl.Rect{secCircles[i][0] - 4, secCircles[i][1] - 4, 8, 8}
		}
		renderer.Copy(secTexture, &textureSource, &dest)
	}
}

func drawStaticCircles() {
	// Draw static indicator circles
	for _, p := range staticCircles {
		if Options.Small {
			dest := sdl.Rect{p[0] - 4, p[1] - 4, 8, 8}
			renderer.Copy(staticTexture, &textureSource, &dest)
		} else {
			dest := sdl.Rect{p[0] - 20, p[1] - 20, 40, 40}
			renderer.Copy(staticTexture, &textureSource, &dest)
		}
	}
}

func drawDots() {
	// Draw the dots between hours and minutes
	setPixel(14, 15, textSDLColor)
	setPixel(14, 16, textSDLColor)
	setPixel(15, 15, textSDLColor)
	setPixel(15, 16, textSDLColor)

	setPixel(18, 15, textSDLColor)
	setPixel(18, 16, textSDLColor)
	setPixel(19, 15, textSDLColor)
	setPixel(19, 16, textSDLColor)
}

// Set "led matrix" pixel
func setPixel(cy, cx int, color sdl.Color) {
	x := gridStartX + int32(cx*gridSpacing)
	y := gridStartY + int32(cy*gridSpacing)
	rect := sdl.Rect{x, y, gridSize, gridSize}
	renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	renderer.FillRect(&rect)
}

func drawBitmask(bitmask [][]bool, color sdl.Color, r int, c int) {
	for y, row := range bitmask {
		for x, b := range row {
			if b {
				setPixel(r+y, c+x, color)
			}
		}
	}
}
