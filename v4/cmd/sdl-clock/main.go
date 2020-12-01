package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/veandco/go-sdl2/sdl"
	"gitlab.com/Depili/clock-8001/v4/clock"
	"gitlab.com/Depili/clock-8001/v4/debug"
	"gitlab.com/Depili/clock-8001/v4/util"
	"gitlab.com/Depili/go-rgb-led-matrix/bdf"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

var parser = flags.NewParser(&options, flags.Default)
var font *bdf.Bdf
var showBackground bool
var backgroundNumber int

func main() {
	var err error

	parseOptions()

	if options.Defaults {
		defaultSourceConfig()
	}

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

	if options.dualClock || options.textClock || options.countdown {
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
	if options.textClock {
		initTextClock()
	}

	if options.countdown {
		initCountdown()
	}

	// Trap SIGINT aka Ctrl-C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Intialize polling timers
	updateTicker := time.NewTicker(time.Millisecond * 30)
	eventTicker := time.NewTicker(time.Millisecond * 5)

	// Create main clock engine
	engine, err := clock.MakeEngine(options.EngineOptions)
	check(err)

	clockTextures = make([]*sdl.Texture, 2)

	clockTextures[0], err = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, 1080, 1080)
	check(err)
	clockTextures[1], err = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, 1080, 1080)
	check(err)

	err = clockTextures[0].SetBlendMode(sdl.BLENDMODE_BLEND)
	check(err)
	err = clockTextures[1].SetBlendMode(sdl.BLENDMODE_BLEND)
	check(err)

	loadBackground(options.Background)

	log.Printf("Entering main loop\n")
	var info string

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

			// startTime := time.Now()

			// Get the clock state snapshot
			state := engine.State()

			checkBackgroundUpdate(state)
			if options.textClock {
				drawTextClock(state)
			} else if options.countdown {
				drawCountdown()
			} else {
				drawRoundClocks(state)
			}

			if state.Info != "" {
				if state.Info != info {
					info = state.Info
					updateInfoScreen(info)
				}
				drawInfoScreen()
			}

			// Update the canvas
			renderer.Present()
			// debug.Printf("Frame time: %d ms\n", time.Now().Sub(startTime).Milliseconds())
		}
	}
}

func updateInfoScreen(info string) {
	var err error
	if infoTexture != nil {
		infoTexture.Destroy()
	}
	lines := strings.Split(info, "\n")

	height := infoFont.LineSkip() * len(lines)
	infoTexture, err = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, 1024, int32(height))
	infoTexture.SetBlendMode(sdl.BLENDMODE_BLEND)
	check(err)

	err = renderer.SetRenderTarget(infoTexture)
	check(err)
	renderer.SetDrawColor(0, 0, 0, 128)
	renderer.Clear()

	var rowTexture *sdl.Texture
	rowTexture.SetBlendMode(sdl.BLENDMODE_BLEND)
	h := int32(0)
	for _, l := range lines {
		rowTexture = renderText(l, infoFont, sdl.Color{R: 255, G: 255, B: 255, A: 128})
		rowTexture.SetBlendMode(sdl.BLENDMODE_BLEND)
		_, _, texW, texH, _ := rowTexture.Query()

		err := renderer.Copy(rowTexture, nil, &sdl.Rect{X: 20, Y: h, H: texH, W: texW})
		check(err)

		h += texH
		rowTexture.Destroy()
	}

	log.Printf("Updated info texture:\n%s", info)
}

func drawInfoScreen() {
	_, _, texW, texH, _ := infoTexture.Query()
	renderer.SetRenderTarget(nil)
	err := renderer.Copy(infoTexture, nil, &sdl.Rect{X: 0, Y: 0, H: texH, W: texW})
	check(err)
}

func checkBackgroundUpdate(state *clock.State) {
	// Check for background changes
	if backgroundNumber != state.Background {
		backgroundNumber = state.Background
		p := make([]string, 3)
		filemask := fmt.Sprintf("%d.*", state.Background)
		path := options.BackgroundPath
		p[0] = filepath.Join(path, filemask)
		p[1] = filepath.Join(path, "0"+filemask)
		p[2] = filepath.Join(path, "00"+filemask)
		for _, pattern := range p {
			files, _ := filepath.Glob(pattern)
			if files != nil {
				log.Printf("Loading background: %s", files[0])
				loadBackground(files[0])
				return
			}
		}
		log.Printf("Couldn't find background for number: %d", backgroundNumber)
	}
}

// parseOptions parses the command line options and provided ini file
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
			panic(err)
		}
	}

	switch options.Face {
	case "round":
	case "dual-round":
		options.dualClock = true
	case "small":
		options.small = true
	case "text":
		options.textClock = true
	case "countdown":
		options.countdown = true
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
	if options.small {
		secCircles = util.Points(center192, secondRadius192, 60)
		staticCircles = util.Points(center192, staticRadius192, 12)
	} else {
		secCircles = util.Points(center1080, secondRadius1080, 60)
		staticCircles = util.Points(center1080, staticRadius1080, 12)
	}
}

func defaultSourceConfig() {
	options.EngineOptions.Source1 = &clock.SourceOptions{
		Text:     "",
		LTC:      true,
		UDP:      true,
		Timer:    true,
		Counter:  0,
		Tod:      true,
		TimeZone: "Europe/Helsinki",
	}
	options.EngineOptions.Source2 = &clock.SourceOptions{
		Text:     "",
		LTC:      true,
		UDP:      true,
		Timer:    true,
		Counter:  1,
		Tod:      true,
		TimeZone: "Europe/Helsinki",
	}
	options.EngineOptions.Source3 = &clock.SourceOptions{
		Text:     "",
		LTC:      true,
		UDP:      true,
		Timer:    true,
		Counter:  2,
		Tod:      true,
		TimeZone: "Europe/Helsinki",
	}
	options.EngineOptions.Source4 = &clock.SourceOptions{
		Text:     "",
		LTC:      true,
		UDP:      true,
		Timer:    true,
		Counter:  3,
		Tod:      true,
		TimeZone: "Europe/Helsinki",
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
