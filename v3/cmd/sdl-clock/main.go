package main

import (
	"gitlab.com/Depili/clock-8001/v3/clock"
	"gitlab.com/Depili/go-rgb-led-matrix/bdf"
	// "github.com/depili/go-rgb-led-matrix/matrix"
	"github.com/jessevdk/go-flags"
	// "github.com/kidoman/embd"
	// _ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"github.com/veandco/go-sdl2/sdl"
	"gitlab.com/Depili/clock-8001/v3/debug"
	"gitlab.com/Depili/clock-8001/v3/util"
	"log"
	"os"
	"os/signal"
	"text/template"
	"time"
)

var parser = flags.NewParser(&options, flags.Default)
var font *bdf.Bdf

func main() {
	var err error

	parseOptions()

	// Dump the current config to stdout
	if options.DumpConfig {
		dumpConfig()
	}

	if !options.DisableHTTP {
		go runHTTP()
	}

	if options.Debug {
		debug.Enabled = true
	}

	// Parse font for clock text
	font, err = bdf.Parse(options.Font)
	if err != nil {
		panic(err)
	}

	log.Printf("Fonts loaded.\n")

	// Initialize SDL
	initSDL()
	defer sdl.Quit()
	defer window.Destroy()
	defer renderer.Destroy()

	initColors()

	createRings()

	if options.DualClock {
		// FIXME: rpi display scaling fix
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
	} else if !options.NoARCorrection {
		rpiDisplayCorrection()
	}

	initTextures()

	// Trap SIGINT aka Ctrl-C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Create initial bitmaps for clock text
	hourBitmap := font.TextBitmap("15")
	minuteBitmap := font.TextBitmap("04")
	secondBitmap := font.TextBitmap("05")
	tallyBitmap := font.TextBitmap("  ")

	// Intialize polling timers
	updateTicker := time.NewTicker(time.Millisecond * 30)
	eventTicker := time.NewTicker(time.Millisecond * 5)

	// Create main clock engine
	engine, err := clock.MakeEngine(options.EngineOptions)
	check(err)

	clockTextures := make([]*sdl.Texture, 2)

	clockTextures[0], err = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, 1080, 1080)
	check(err)
	clockTextures[1], err = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, 1080, 1080)
	check(err)

	err = clockTextures[0].SetBlendMode(sdl.BLENDMODE_BLEND)
	check(err)
	err = clockTextures[1].SetBlendMode(sdl.BLENDMODE_BLEND)
	check(err)

	showBackground := loadBackground()

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

	// Main clock event loop

	log.Printf("Entering main loop\n")
	for {
		select {
		case <-sigChan:
			// SIGINT received, shutdown gracefully
			os.Exit(1)
		case <-eventTicker.C:
			// SDL event polling
			e := sdl.PollEvent()
			switch e.(type) {
			case *sdl.QuitEvent:
				os.Exit(0)
			}
		case <-updateTicker.C:
			// Display update

			startTime := time.Now()
			for i, eng := range engines {

				// Update clock display fields
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

				// Set renderer target to the corresponding clock texture
				err = renderer.SetRenderTarget(clockTextures[i])
				check(err)

				clearCanvas()

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

			clearCanvas()

			// Copy the background image as needed
			if showBackground {
				renderer.Copy(backgroundTexture, nil, nil)
			}

			source := sdl.Rect{X: 0, Y: 0, W: 1080, H: 1080}

			// FIXME: the text positioning and size is just magic numbers

			if options.DualClock {
				// Render the dual clock displays
				dualText := font.TextBitmap(engine.DualText)

				x, y, _ := renderer.GetOutputSize()
				if x > y {
					// Normal horizontal view with the clocks side by side
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
					// Rotated view with the clocks on top of each other
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
				// Single clock mode
				x, y, _ := renderer.GetOutputSize()
				var dest sdl.Rect

				if options.Small {
					// Do not scale the small 192x192 px clock

					err := renderer.Copy(clockTextures[0], nil, nil)
					check(err)
				} else if x > y {
					dest = sdl.Rect{
						X: (x - y) / 2,
						Y: 0,
						W: y,
						H: y,
					}
				} else {
					// Rotated display
					// FIXME this centers the clock
					dest = sdl.Rect{
						X: 0,
						Y: (y - x) / 2,
						W: x,
						H: x,
					}
				}
				err := renderer.Copy(clockTextures[0], &source, &dest)
				check(err)
			}

			// Update the canvas
			renderer.Present()
			debug.Printf("Frame time: %d ms\n", time.Now().Sub(startTime).Milliseconds())
		}
	}
}

// paRseOptions parses the command line options and provided ini file
func parseOptions() {
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
}

// dumpConfig dumps the clock configuration ini file to stdout
func dumpConfig() {
	tmpl, err := template.New("config.ini").Parse(configTemplate)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, options)
	os.Exit(0)
}

// createRings creates the coordinate rings for the static and second rings
func createRings() {
	if options.Small {
		secCircles = util.Points(center192, secondRadius192, 60)
		staticCircles = util.Points(center192, staticRadius192, 12)
	} else {
		secCircles = util.Points(center1080, secondRadius1080, 60)
		staticCircles = util.Points(center1080, staticRadius1080, 12)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
