package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/veandco/go-sdl2/sdl"
	"gitlab.com/Depili/clock-8001/v4/clock"
	"gitlab.com/Depili/clock-8001/v4/debug"
	"gitlab.com/Depili/clock-8001/v4/util"
	"image/color"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	_ "time/tzdata"
)

var parser = flags.NewParser(&options, flags.Default)
var showBackground bool
var backgroundNumber int

const updateTime = time.Second / 30

func main() {
	var err error
	var info string

	parseOptions()

	if !options.DisableHTTP {
		go runHTTP()
	}

	if hw, ok := signalHardwareList[options.SignalType]; ok {
		hw.Init()
	}

	// Initialize SDL
	initSDL()
	defer sdl.Quit()
	defer window.Destroy()
	defer renderer.Destroy()

	setupScaling()
	initColors()
	initTextures()

	if options.textClock {
		initTextClock()
	} else if options.Face == "288x144" {
		initSmallTextClock()
	} else if options.countdown {
		initCountdown()
	} else {
		initRoundClock()
	}

	// Trap SIGINT aka Ctrl-C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Intialize polling timers
	updateTicker := time.NewTicker(updateTime)
	eventTicker := time.NewTicker(time.Millisecond * 5)

	// Create main clock engine
	engine, err := clock.MakeEngine(options.EngineOptions)
	check(err)
	for i := 0; i < 3; i++ {
		engine.SetSourceColors(i, toRGBA(colors.row[i]), toRGBA(colors.rowBG[i]))
	}
	engine.SetTitleColors(toRGBA(colors.label), toRGBA(colors.labelBG))

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
			switch t := e.(type) {
			case *sdl.QuitEvent:
				os.Exit(0)
			case *sdl.KeyboardEvent:
				key := t.Keysym.Sym
				if key == sdl.K_f {
					window.SetFullscreen(sdl.WINDOW_FULLSCREEN_DESKTOP)
				} else if key == sdl.K_ESCAPE {
					window.SetFullscreen(0)
				}
			}
		case <-updateTicker.C:
			// Display update

			// startTime := time.Now()

			// Get the clock state snapshot
			state := engine.State()
			if state.ScreenFlash {
				drawWhiteScreen()
			} else {
				checkBackgroundUpdate(state)
				if options.textClock {
					drawTextClock(state)
				} else if options.Face == "288x144" {
					drawSmallTextClock(state)
				} else if options.countdown {
					drawCountdown()
				} else {
					drawRoundClocks(state)
				}
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
			if signalHardware != nil {
				c := state.HardwareSignalColor
				if options.SignalFollow {
					c = state.Clocks[0].SignalColor
				}
				signalHardware.Fill(signalBrightness(c))
				signalHardware.Update()
			}
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
		showBackground = false
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
		// Legacy option
		options.small = true
		options.Face = "192"
	case "144":
		options.small = true
	case "192":
		options.small = true
	case "text":
		options.textClock = true
	case "single":
		options.textClock = true
		options.singleLine = true
	case "countdown":
		options.countdown = true
	}

	if options.Defaults {
		defaultSourceConfig()
	}

	// Dump the current config to stdout
	if options.DumpConfig {
		dumpConfig()
	}

	if options.Debug {
		debug.Enabled = true
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
	if options.Face == "144" {
		secCircles = util.Points(center144, secondRadius144, 60)
		staticCircles = util.Points(center144, staticRadius144, 12)
	} else if options.Face == "small" || options.Face == "192" {
		secCircles = util.Points(center192, secondRadius192, 60)
		staticCircles = util.Points(center192, staticRadius192, 12)
	} else {
		secCircles = util.Points(center1080, secondRadius1080, 60)
		staticCircles = util.Points(center1080, staticRadius1080, 12)
	}
}

func defaultSourceConfig() {
	options.EngineOptions.Source1 = &clock.SourceOptions{
		Text:          "",
		LTC:           true,
		Timer:         true,
		Counter:       1,
		Tod:           true,
		TimeZone:      "Europe/Helsinki",
		OvertimeColor: "#FF0000",
	}
	options.EngineOptions.Source2 = &clock.SourceOptions{
		Text:          "",
		LTC:           true,
		Timer:         true,
		Counter:       2,
		Tod:           true,
		TimeZone:      "Europe/Helsinki",
		OvertimeColor: "#FF0000",
	}
	options.EngineOptions.Source3 = &clock.SourceOptions{
		Text:          "",
		LTC:           true,
		Timer:         true,
		Counter:       3,
		Tod:           true,
		TimeZone:      "Europe/Helsinki",
		OvertimeColor: "#FF0000",
	}
	options.EngineOptions.Source4 = &clock.SourceOptions{
		Text:          "",
		LTC:           true,
		Timer:         true,
		Counter:       4,
		Tod:           true,
		TimeZone:      "Europe/Helsinki",
		OvertimeColor: "#FF0000",
	}
}

func signalBrightness(c color.RGBA) color.RGBA {
	return color.RGBA{
		R: uint8(int(c.R) * options.SignalBrightness / 255),
		G: uint8(int(c.G) * options.SignalBrightness / 255),
		B: uint8(int(c.B) * options.SignalBrightness / 255),
		A: 255,
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
