package main

import (
	"gitlab.com/Depili/clock-8001/clock"
	"gitlab.com/Depili/go-rgb-led-matrix/bdf"
	// "github.com/depili/go-rgb-led-matrix/matrix"
	"github.com/jessevdk/go-flags"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"gitlab.com/Depili/clock-8001/debug"
	"log"
	"os"
	"os/signal"
	"time"
)

var staticSDLColor = sdl.Color{80, 80, 0, 255} // 12 static indicator circles
var secSDLColor = sdl.Color{200, 0, 0, 255}
var textSDLColor sdl.Color
var window *sdl.Window
var renderer *sdl.Renderer
var parser = flags.NewParser(&options, flags.Default)

var textureSource sdl.Rect
var staticTexture *sdl.Texture
var secTexture *sdl.Texture

func main() {
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	if options.Debug {
		debug.Enabled = true
	}

	// GPIO pin for toggling between timezones
	if err := embd.InitGPIO(); err != nil {
		panic(err)
	}

	timePin, err := embd.NewDigitalPin(options.TimePin)
	if err != nil {
		panic(err)
	} else if err := timePin.SetDirection(embd.In); err != nil {
		panic(err)
	}

	log.Printf("GPIO initialized.\n")

	/*
		// Load timezones
		local, err := time.LoadLocation(options.LocalTime)
		if err != nil {
			panic(err)
		}

		foreign, err := time.LoadLocation(options.ForeignTime)
		if err != nil {
			panic(err)
		}
		log.Printf("Timezones loaded.\n")
	*/

	// Parse font for clock text
	font, err := bdf.Parse(options.Font)
	if err != nil {
		panic(err)
	}

	log.Printf("Fonts loaded.\n")

	// Initialize SDL

	if err = sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		log.Fatalf("Failed to initialize SDL: %s\n", err)
	}
	defer sdl.Quit()

	if window, err = sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight, sdl.WINDOW_SHOWN); err != nil {
		log.Fatalf("Failed to create window: %s\n", err)
	}
	defer window.Destroy()

	if renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED); err != nil {
		log.Fatalf("Failed to create renderer: %s\n", err)
	}

	sdl.ShowCursor(0) // Hide mouse cursor

	renderer.Clear()
	defer renderer.Destroy()

	log.Printf("SDL init done\n")

	rendererInfo, _ := renderer.GetInfo()
	log.Printf("Renderer: %v\n", rendererInfo.Name)

	// Clock colors from flags
	textSDLColor = sdl.Color{options.TextRed, options.TextGreen, options.TextBlue, 255}
	staticSDLColor = sdl.Color{options.StaticRed, options.StaticGreen, options.StaticBlue, 255}
	secSDLColor = sdl.Color{options.SecRed, options.SecGreen, options.SecBlue, 255}
	// Default color for the OSC field (black)
	tallyColor := sdl.Color{0, 0, 0, 255}

	var textureSize int32 = 40
	var textureCoord int32 = 20
	var textureRadius int32 = 19

	// Create a texture for circles
	if options.Small {
		textureSize = 5
		textureCoord = 3
		textureRadius = 3
		gridStartX = 32
		gridStartY = 32
		gridSize = 3
		gridSpacing = 4
		secCircles = smallSecCircles
		staticCircles = smallStaticCircles
	} else {
		// Scale down if needed
		renderer.SetLogicalSize(1080, 1080)
	}

	staticTexture, _ = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, textureSize, textureSize)
	renderer.SetRenderTarget(staticTexture)
	if !options.Small {
		gfx.FilledCircleColor(renderer, textureCoord, textureCoord, textureRadius, staticSDLColor)
		// gfx.AACircleColor(renderer, textureCoord, textureCoord, textureRadius, staticSDLColor)
	} else {
		renderer.SetDrawColor(staticSDLColor.R, staticSDLColor.G, staticSDLColor.B, 255)
		renderer.DrawPoint(0, 1)
		renderer.DrawPoint(0, 2)
		renderer.DrawPoint(0, 3)
		renderer.DrawPoint(1, 0)
		renderer.DrawPoint(1, 1)
		renderer.DrawPoint(1, 2)
		renderer.DrawPoint(1, 3)
		renderer.DrawPoint(1, 4)
		renderer.DrawPoint(2, 0)
		renderer.DrawPoint(2, 1)
		renderer.DrawPoint(2, 2)
		renderer.DrawPoint(2, 3)
		renderer.DrawPoint(2, 4)
		renderer.DrawPoint(3, 0)
		renderer.DrawPoint(3, 1)
		renderer.DrawPoint(3, 2)
		renderer.DrawPoint(3, 3)
		renderer.DrawPoint(3, 4)
		renderer.DrawPoint(4, 1)
		renderer.DrawPoint(4, 2)
		renderer.DrawPoint(4, 3)
	}

	secTexture, _ = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, textureSize, textureSize)
	renderer.SetRenderTarget(secTexture)
	if !options.Small {
		gfx.FilledCircleColor(renderer, textureCoord, textureCoord, textureRadius, secSDLColor)
		// gfx.AACircleColor(renderer, textureCoord, textureCoord, textureRadius, secSDLColor)
	} else {
		renderer.SetDrawColor(secSDLColor.R, secSDLColor.G, secSDLColor.B, 255)
		renderer.DrawPoint(0, 1)
		renderer.DrawPoint(0, 2)
		renderer.DrawPoint(0, 3)
		renderer.DrawPoint(1, 0)
		renderer.DrawPoint(1, 1)
		renderer.DrawPoint(1, 2)
		renderer.DrawPoint(1, 3)
		renderer.DrawPoint(1, 4)
		renderer.DrawPoint(2, 0)
		renderer.DrawPoint(2, 1)
		renderer.DrawPoint(2, 2)
		renderer.DrawPoint(2, 3)
		renderer.DrawPoint(2, 4)
		renderer.DrawPoint(3, 0)
		renderer.DrawPoint(3, 1)
		renderer.DrawPoint(3, 2)
		renderer.DrawPoint(3, 3)
		renderer.DrawPoint(3, 4)
		renderer.DrawPoint(4, 1)
		renderer.DrawPoint(4, 2)
		renderer.DrawPoint(4, 3)
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

	engine, err := clock.MakeEngine(options.EngineOptions)
	if err != nil {
		panic(err)
	}

	log.Printf("Entering main loop\n")
	for {
		select {
		case <-sigChan:
			// SIGINT received, shutdown gracefully
			os.Exit(1)
		case <-updateTicker.C:
			engine.Update()
			seconds := engine.Leds
			hourBitmap = font.TextBitmap(engine.Hours)
			minuteBitmap = font.TextBitmap(engine.Minutes)
			secondBitmap = font.TextBitmap(engine.Seconds)

			tallyColor = sdl.Color{engine.TallyRed, engine.TallyGreen, engine.TallyBlue, 255}
			tallyBitmap = font.TextBitmap(engine.Tally)

			// Clear SDL canvas
			renderer.SetDrawColor(0, 0, 0, 255) // Black
			renderer.Clear()                    // Clear screen

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
		dest := sdl.Rect{secCircles[i][0] - 20, secCircles[i][1] - 20, 40, 40}
		if options.Small {
			dest = sdl.Rect{secCircles[i][0] - 3, secCircles[i][1] - 3, 5, 5}
		}
		renderer.Copy(secTexture, &textureSource, &dest)
	}
}

func drawStaticCircles() {
	// Draw static indicator circles
	for _, p := range staticCircles {
		if options.Small {
			dest := sdl.Rect{p[0] - 3, p[1] - 3, 5, 5}
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
