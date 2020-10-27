package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/veandco/go-sdl2/sdl"
	"gitlab.com/Depili/clock-8001/v4/clock"
	"gitlab.com/Depili/clock-8001/v4/debug"
	"gitlab.com/Depili/clock-8001/v4/util"
	"gitlab.com/Depili/go-rgb-led-matrix/bdf"
	"log"
	"os"
	"os/signal"
	"text/template"
	"time"
)

var parser = flags.NewParser(&options, flags.Default)
var font *bdf.Bdf
var showBackground bool

func main() {
	var err error

	parseOptions()

	if options.Defaults {
		dump := options.DumpConfig
		options = defaultConfig()
		options.DumpConfig = dump
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

	if options.dualClock || options.textClock {
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

			if options.textClock {
				drawTextClock(state)
			} else {
				drawRoundClocks(state)
			}

			// Update the canvas
			renderer.Present()
			// debug.Printf("Frame time: %d ms\n", time.Now().Sub(startTime).Milliseconds())
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

func defaultConfig() clockOptions {
	o := clockOptions{
		Face:            "round",
		Font:            "fonts/7x13.bdf",
		NumberFont:      "Copse-Regular.ttf",
		LabelFont:       "RobotoMono-VariableFont_wght.ttf",
		IconFont:        "MaterialIcons-Regular.ttf",
		TextColor:       "#FF8000",
		StaticColor:     "#505000",
		SecondColor:     "#C80000",
		CountdownColor:  "#FF0000",
		LabelColor:      "#FF8000",
		Row1Color:       "#FF8000",
		Row2Color:       "#FF8000",
		Row3Color:       "#FF8000",
		TimerBG:         "#202020",
		LabelBG:         "#202020",
		BackgroundColor: "#000000",
		HTTPPort:        ":80",
		HTTPUser:        "admin",
		HTTPPassword:    "clockwork",
		Background:      "",
		EngineOptions: &clock.EngineOptions{
			Flash:      500,
			ListenAddr: "0.0.0.0:1245",
			Timeout:    1000,
			Connect:    "255.255.255.255:1245",
			Ignore:     "ignore",
			Source1: &clock.SourceOptions{
				Text:     "",
				LTC:      true,
				UDP:      true,
				Timer:    true,
				Counter:  0,
				Tod:      true,
				TimeZone: "Europe/Helsinki",
			},
			Source2: &clock.SourceOptions{
				Text:     "",
				LTC:      true,
				UDP:      true,
				Timer:    true,
				Counter:  1,
				Tod:      true,
				TimeZone: "Europe/Helsinki",
			},
			Source3: &clock.SourceOptions{
				Text:     "",
				LTC:      true,
				UDP:      true,
				Timer:    true,
				Counter:  2,
				Tod:      true,
				TimeZone: "Europe/Helsinki",
			},
			Source4: &clock.SourceOptions{
				Text:     "",
				LTC:      true,
				UDP:      true,
				Timer:    true,
				Counter:  3,
				Tod:      true,
				TimeZone: "Europe/Helsinki",
			},
		},
	}
	return o
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
