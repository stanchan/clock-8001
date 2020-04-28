package main

import (
	"gitlab.com/Depili/clock-8001/v3/clock"
	"gitlab.com/Depili/go-rgb-led-matrix/bdf"
	// "github.com/depili/go-rgb-led-matrix/matrix"
	"github.com/jessevdk/go-flags"
	// "github.com/kidoman/embd"
	// _ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"gitlab.com/Depili/clock-8001/v3/debug"
	"log"
	"os"
	"os/signal"
	"time"
)

var staticSDLColor = sdl.Color{R: 80, G: 80, B: 0, A: 255} // 12 static indicator circles
var secSDLColor = sdl.Color{R: 200, G: 0, B: 0, A: 255}
var textSDLColor sdl.Color
var window *sdl.Window
var renderer *sdl.Renderer
var parser = flags.NewParser(&options, flags.Default)

var textureSource sdl.Rect
var staticTexture *sdl.Texture
var secTexture *sdl.Texture

func main() {
	options.Config = func(s string) error {
		ini := flags.NewIniParser(parser)
		options.configFile = s
		return ini.ParseFile(s)
	}

	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	if !options.DisableHTTP {
		go runHTTP()
	}

	if options.Debug {
		debug.Enabled = true
	}

	/*
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
	*/
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

	_, err = sdl.ShowCursor(0) // Hide mouse cursor
	check(err)

	err = renderer.Clear()
	defer renderer.Destroy()
	check(err)

	log.Printf("SDL init done\n")

	rendererInfo, err := renderer.GetInfo()
	check(err)
	log.Printf("Renderer: %v\n", rendererInfo.Name)

	// Clock colors from flags
	textSDLColor = sdl.Color{R: options.TextRed, G: options.TextGreen, B: options.TextBlue, A: 255}
	staticSDLColor = sdl.Color{R: options.StaticRed, G: options.StaticGreen, B: options.StaticBlue, A: 255}
	secSDLColor = sdl.Color{R: options.SecRed, G: options.SecGreen, B: options.SecBlue, A: 255}
	// Default color for the OSC field (black)
	tallyColor := sdl.Color{R: 0, G: 0, B: 0, A: 255}

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
	} else if !options.DualClock {
		// Scale down if needed
		err = renderer.SetLogicalSize(1080, 1080)
		check(err)

		// the official raspberry pi display has weird pixels
		// We detect it by the unusual 800 x 480 resolution
		// We will eventually support rotated displays also
		x, y, _ := renderer.GetOutputSize()
		log.Printf("SDL renderer size: %v x %v", x, y)
		scaleX, scaleY := renderer.GetScale()
		log.Printf("Scaling: x: %v, y: %v\n", scaleX, scaleY)

		if (x == 800) && (y == 480) {
			// Official display, rotated 0 or 180 degrees
			// Scale for Y is 480 / 1080 = 0.44444445
			// Scale for X is 0.44444445 * ((9.0*800) / (16*480)) = 0.416666671875
			err = renderer.SetScale(0.416666671875, 0.44444445)
			check(err)
			log.Printf("Detected official raspberry pi display, correcting aspect ratio\n")
		} else if (y == 800) && (x == 480) {
			// Official display rotated 90 or 270 degrees
			err = renderer.SetScale(0.44444445, 0.416666671875)
			check(err)
			log.Printf("Detected official raspberry pi display (rotated 90 or 270 deg), correcting aspect ratio.\n")
			log.Printf("Moving clock to top corner of the display.\n")
		}

		if y > x {
			log.Printf("Display rotated 90 or 270 degrees, moving clock to top corner.\n")
			viewport := renderer.GetViewport()
			log.Printf("Renderer viewport: %v\n", viewport)
			viewport = sdl.Rect{X: 0, Y: 0, W: 1080, H: 1080}
			err = renderer.SetViewport(&viewport)
			check(err)
		}
	} else {
		// Dual clock
		x, y, _ := renderer.GetOutputSize()

		if x > y {
			err = renderer.SetLogicalSize(1920, 1080)
			check(err)
		} else {
			// rotated display
			err = renderer.SetLogicalSize(1080, 1920)
			check(err)
		}
	}

	staticTexture, err = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, textureSize, textureSize)
	check(err)

	err = renderer.SetRenderTarget(staticTexture)
	check(err)

	if !options.Small {
		gfx.FilledCircleColor(renderer, textureCoord, textureCoord, textureRadius, staticSDLColor)
		// gfx.AACircleColor(renderer, textureCoord, textureCoord, textureRadius, staticSDLColor)
	} else {
		err = renderer.SetDrawColor(staticSDLColor.R, staticSDLColor.G, staticSDLColor.B, 255)
		check(err)

		for _, point := range circlePixels {
			if err := renderer.DrawPoint(point[0], point[1]); err != nil {
				panic(err)
			}
		}
	}

	secTexture, _ = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, textureSize, textureSize)
	err = renderer.SetRenderTarget(secTexture)
	check(err)

	if !options.Small {
		gfx.FilledCircleColor(renderer, textureCoord, textureCoord, textureRadius, secSDLColor)
		// gfx.AACircleColor(renderer, textureCoord, textureCoord, textureRadius, secSDLColor)
	} else {
		err = renderer.SetDrawColor(secSDLColor.R, secSDLColor.G, secSDLColor.B, 255)
		check(err)

		for _, point := range circlePixels {
			if err := renderer.DrawPoint(point[0], point[1]); err != nil {
				panic(err)
			}
		}
	}

	err = renderer.SetRenderTarget(nil)
	check(err)

	textureSource = sdl.Rect{X: 0, Y: 0, W: textureSize, H: textureSize}

	// Trap SIGINT aka Ctrl-C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Create initial bitmaps for clock text
	hourBitmap := font.TextBitmap("15")
	minuteBitmap := font.TextBitmap("04")
	secondBitmap := font.TextBitmap("05")
	tallyBitmap := font.TextBitmap("  ")

	updateTicker := time.NewTicker(time.Millisecond * 30)
	eventTicker := time.NewTicker(time.Millisecond * 5)

	engine, err := clock.MakeEngine(options.EngineOptions)
	check(err)

	clockTextures := make([]*sdl.Texture, 2)

	clockTextures[0], _ = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, 1080, 1080)
	clockTextures[1], _ = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, 1080, 1080)

	// Second clock engine for constant time of day display with dual clock mode
	var todOptions = clock.EngineOptions{
		Timezone:        options.EngineOptions.Timezone,
		Flash:           options.EngineOptions.Flash,
		DisableOSC:      true,
		DisableFeedback: true,
	}

	todEngine, err := clock.MakeEngine(&todOptions)
	check(err)

	var engines []*clock.Engine

	if options.DualClock {
		engines = make([]*clock.Engine, 2)
		engines[0] = todEngine
		engines[1] = engine
	} else {
		engines = make([]*clock.Engine, 1)
		engines[0] = engine
	}

	log.Printf("Entering main loop\n")
	for {
		select {
		case <-sigChan:
			// SIGINT received, shutdown gracefully
			os.Exit(1)
		case <-eventTicker.C:
			e := sdl.PollEvent()
			switch e.(type) {
			case *sdl.QuitEvent:
				os.Exit(0)
			}
		case <-updateTicker.C:
			for i, eng := range engines {

				eng.Update()
				seconds := eng.Leds
				hourBitmap = font.TextBitmap(eng.Hours)
				minuteBitmap = font.TextBitmap(eng.Minutes)
				secondBitmap = font.TextBitmap(eng.Seconds)

				if eng.LtcActive() {
					tallyColor = textSDLColor
				} else {
					tallyColor = sdl.Color{R: eng.TallyRed, G: eng.TallyGreen, B: eng.TallyBlue, A: 255}
				}
				tallyBitmap = font.TextBitmap(eng.Tally)

				// Renderer target
				err = renderer.SetRenderTarget(clockTextures[i])
				check(err)

				// Clear SDL canvas
				err = renderer.SetDrawColor(0, 0, 0, 255) // Black
				check(err)

				err = renderer.Clear() // Clear screen
				check(err)

				// Dots between hours and minutes
				if eng.Dots {
					drawDots()
				}

				// Draw the text
				drawBitmask(hourBitmap, textSDLColor, 10, 0)
				drawBitmask(minuteBitmap, textSDLColor, 10, 17)
				drawBitmask(secondBitmap, textSDLColor, 21, 8)
				drawBitmask(tallyBitmap, tallyColor, 0, 2)

				drawStaticCircles()
				drawSecondCircles(seconds)
			}

			err = renderer.SetRenderTarget(nil)
			check(err)

			// Clear SDL canvas
			err = renderer.SetDrawColor(0, 0, 0, 255) // Black
			check(err)

			err = renderer.Clear() // Clear screen
			check(err)

			source := sdl.Rect{X: 0, Y: 0, W: 1080, H: 1080}

			// FIXME: the text positioning and size is just magic numbers

			if options.DualClock {
				dualText := font.TextBitmap(engine.DualText)

				x, y, _ := renderer.GetOutputSize()
				if x > y {
					dest := sdl.Rect{X: 0, Y: 0, W: 800, H: 800}
					err := renderer.Copy(clockTextures[0], &source, &dest)
					check(err)
					dest = sdl.Rect{X: 1920 - 800, Y: 0, W: 800, H: 800}
					err = renderer.Copy(clockTextures[1], &source, &dest)
					check(err)

					for y, row := range dualText {
						for x, b := range row {
							if b {
								setPixel(y, x, textSDLColor, (1920-1064)/2, 800+50, 19, 16)
							}
						}
					}
				} else {
					// Rotated
					dest := sdl.Rect{X: (1080 - 800) / 2, Y: 0, W: 800, H: 800}
					err := renderer.Copy(clockTextures[0], &source, &dest)
					check(err)
					dest = sdl.Rect{X: (1080 - 800) / 2, Y: 1920 - 800, W: 800, H: 800}
					err = renderer.Copy(clockTextures[1], &source, &dest)
					check(err)
					for y, row := range dualText {
						for x, b := range row {
							if b {
								setPixel(y, x, textSDLColor, (1080-1064)/2, 800+50, 19, 16)
							}
						}
					}

				}
			} else {
				dest := sdl.Rect{X: 0, Y: 0, W: 1080, H: 1080}
				err := renderer.Copy(clockTextures[0], &source, &dest)
				check(err)
			}

			// Update the canvas
			renderer.Present()
		}
	}
}

func drawSecondCircles(seconds int) {
	// Draw second circles
	for i := 0; i <= int(seconds); i++ {
		dest := sdl.Rect{X: secCircles[i][0] - 20, Y: secCircles[i][1] - 20, W: 40, H: 40}
		if options.Small {
			dest = sdl.Rect{X: secCircles[i][0] - 3, Y: secCircles[i][1] - 3, W: 5, H: 5}
		}
		err := renderer.Copy(secTexture, &textureSource, &dest)
		check(err)
	}
}

func drawStaticCircles() {
	// Draw static indicator circles
	for _, p := range staticCircles {
		if options.Small {
			dest := sdl.Rect{X: p[0] - 3, Y: p[1] - 3, W: 5, H: 5}
			err := renderer.Copy(staticTexture, &textureSource, &dest)
			check(err)
		} else {
			dest := sdl.Rect{X: p[0] - 20, Y: p[1] - 20, W: 40, H: 40}
			err := renderer.Copy(staticTexture, &textureSource, &dest)
			check(err)
		}
	}
}

func drawDots() {
	// Draw the dots between hours and minutes
	setMatrix(14, 15, textSDLColor)
	setMatrix(14, 16, textSDLColor)
	setMatrix(15, 15, textSDLColor)
	setMatrix(15, 16, textSDLColor)

	setMatrix(18, 15, textSDLColor)
	setMatrix(18, 16, textSDLColor)
	setMatrix(19, 15, textSDLColor)
	setMatrix(19, 16, textSDLColor)
}

// Set "led matrix" pixel
func setMatrix(cy, cx int, color sdl.Color) {
	x := gridStartX + int32(cx*gridSpacing)
	y := gridStartY + int32(cy*gridSpacing)
	rect := sdl.Rect{X: x, Y: y, W: gridSize, H: gridSize}
	err := renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	check(err)

	err = renderer.FillRect(&rect)
	check(err)
}

func setPixel(cy, cx int, color sdl.Color, startX, startY, spacing, pixelSize int32) {
	x := startX + int32(cx)*spacing
	y := startY + int32(cy)*spacing
	rect := sdl.Rect{X: x, Y: y, W: pixelSize, H: pixelSize}
	err := renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	check(err)

	err = renderer.FillRect(&rect)
	check(err)
}

func drawBitmask(bitmask [][]bool, color sdl.Color, r int, c int) {
	for y, row := range bitmask {
		for x, b := range row {
			if b {
				setMatrix(r+y, c+x, color)
			}
		}
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
