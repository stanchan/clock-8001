package main

import (
	"gitlab.com/Depili/clock-8001/v3/clock"
	"gitlab.com/Depili/go-rgb-led-matrix/bdf"
	// "github.com/depili/go-rgb-led-matrix/matrix"
	"github.com/jessevdk/go-flags"
	// "github.com/kidoman/embd"
	// _ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	// "github.com/veandco/go-sdl2/gfx"
	"bufio"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"gitlab.com/Depili/clock-8001/debug"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

// Clock size
const dy = 267
const dx = 192

var staticSDLColor = sdl.Color{80, 80, 0, 255} // 12 static indicator circles
var secSDLColor = sdl.Color{200, 0, 0, 255}
var textSDLColor sdl.Color
var window *sdl.Window
var renderer *sdl.Renderer
var parser = flags.NewParser(&options, flags.Default)

var textureSource sdl.Rect
var staticTexture *sdl.Texture
var secTexture *sdl.Texture
var testTexture *sdl.Texture

func main() {
	options.Config = func(s string) error {
		ini := flags.NewIniParser(parser)
		return ini.ParseFile(s)
	}

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

	log.Printf("Reading timezones for clocks from %s\n", options.Timezones)
	file, err := os.Open(options.Timezones)
	if err != nil {
		panic(err)
	}

	var tzList [][]string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		log.Println(line)
		if strings.HasPrefix(line, "#") {
			// skip comments
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			log.Fatalf("Error parsing line\n")
		}
		tzList = append(tzList, parts)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	err = file.Close()
	if err != nil {
		panic(err)
	}

	if len(tzList) > 40 {
		log.Fatalf("Too many timezones: %d\n", len(tzList))
	}

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

	if err = ttf.Init(); err != nil {
		return
	}
	defer ttf.Quit()

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

	textureSize = 5
	gridStartX = 32
	gridStartY = 32
	gridSize = 3
	gridSpacing = 4
	secCircles = smallSecCircles
	staticCircles = smallStaticCircles

	setupStaticTexture(textureSize)
	setupSecTexture(textureSize)

	setupTestTexture()

	renderer.SetRenderTarget(nil)
	textureSource = sdl.Rect{0, 0, textureSize, textureSize}

	// Trap SIGINT aka Ctrl-C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Cache the second ring texture
	ringTexture, _ := renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, dx, dy)
	ringTexture.SetBlendMode(sdl.BLENDMODE_BLEND)

	// Destination texture for the sub-clocks
	clockTexture, _ := renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, dx, dy)
	source := sdl.Rect{
		X: 0,
		Y: 0,
		H: dy,
		W: dx,
	}

	// Create initial bitmaps for clock text
	hourBitmap := font.TextBitmap("15")
	minuteBitmap := font.TextBitmap("04")
	secondBitmap := font.TextBitmap("05")
	tallyBitmap := font.TextBitmap("  ")
	tzBitmap := font.TextBitmap("ABCD")

	updateTicker := time.NewTicker(time.Millisecond * 30)
	eventTicker := time.NewTicker(time.Millisecond * 5)

	feedBackEngine, err := clock.MakeEngine(options.EngineOptions)
	if err != nil {
		panic(err)
	}

	var engOpt = clock.EngineOptions{
		Flash:           500,
		ListenAddr:      "0.0.0.0:1245",
		Timeout:         1000,
		Connect:         "255.255.255.255:1245",
		CountdownRed:    255,
		CountdownGreen:  0,
		CountdownBlue:   0,
		DisableOSC:      true,
		DisableFeedback: true,
	}

	engines := make([]*clock.Engine, len(tzList))

	for i, tz := range tzList {
		engOpt.Timezone = tz[0]
		engines[i], err = clock.MakeEngine(&engOpt)
		if err != nil {
			panic(err)
		}
	}

	showTestPicture := false

	log.Printf("Entering main loop\n")
	for {
		select {
		case <-sigChan:
			// SIGINT received, shutdown gracefully
			os.Exit(1)
		case <-eventTicker.C:
			e := sdl.PollEvent()
			switch t := e.(type) {
			case *sdl.QuitEvent:
				os.Exit(0)
			case *sdl.KeyboardEvent:
				if t.State == sdl.PRESSED {
					showTestPicture = !showTestPicture
				}
			}
		case <-updateTicker.C:
			// Clear SDL canvas
			renderer.SetDrawColor(0, 0, 0, 255) // Black
			renderer.Clear()                    // Clear screen
			if showTestPicture {
				// Show the test picture.
				renderer.Copy(testTexture, &sdl.Rect{0, 0, 1920, 1080}, &sdl.Rect{0, 0, 1920, 1080})
				renderer.Present()
				continue
			}

			feedBackEngine.Update()

			for i, eng := range engines {
				var engine *clock.Engine
				if feedBackEngine.TimeOfDay() {
					engine = eng
					engine.Update()
				} else {
					engine = feedBackEngine
				}

				renderer.SetRenderTarget(clockTexture)
				renderer.SetDrawColor(0, 0, 0, 255) // Black
				renderer.Clear()                    // Clear screen

				target := sdl.Rect{
					X: int32(dx * (i % 10)),
					Y: int32(dy * (i / 10)),
					H: dy,
					W: dx,
				}

				renderer.SetDrawColor(0, 0, 0, 255) // Black

				hourBitmap = font.TextBitmap(engine.Hours)
				minuteBitmap = font.TextBitmap(engine.Minutes)

				tallyColor = sdl.Color{feedBackEngine.TallyRed, feedBackEngine.TallyGreen, feedBackEngine.TallyBlue, 255}
				tallyBitmap = font.TextBitmap(feedBackEngine.Tally)

				tzBitmap = font.TextBitmap(tzList[i][1])

				// Dots between hours and minutes
				if engine.Dots {
					drawDots()
				}

				// Draw the text
				drawBitmask(hourBitmap, textSDLColor, 10, 0)
				drawBitmask(minuteBitmap, textSDLColor, 10, 17)
				drawBitmask(tallyBitmap, tallyColor, 0, 2)
				tzCoord := 15 - (len(tzBitmap[0]) / 2)
				drawBitmask(tzBitmap, textSDLColor, 43, tzCoord)

				if i == 0 {
					seconds := engine.Leds
					secondBitmap = font.TextBitmap(engine.Seconds)

					renderer.SetRenderTarget(ringTexture)
					renderer.SetDrawColor(0, 0, 0, 0) // Black
					renderer.Clear()                  // Clear screen
					drawStaticCircles()
					drawSecondCircles(seconds)
					if feedBackEngine.DisplaySeconds() {
						drawBitmask(secondBitmap, textSDLColor, 21, 8)
					}
				}

				renderer.SetRenderTarget(nil)
				renderer.Copy(clockTexture, &source, &target)
				renderer.Copy(ringTexture, &source, &target)
				// Update the canvas
			}
			renderer.Present()

		}
	}
}

func drawSecondCircles(seconds int) {
	// Draw second circles
	for i := 0; i <= int(seconds); i++ {
		dest := sdl.Rect{secCircles[i][0] - 3, secCircles[i][1] - 3, 5, 5}
		renderer.Copy(secTexture, &textureSource, &dest)
	}
}

func drawStaticCircles() {
	// Draw static indicator circles
	for _, p := range staticCircles {
		dest := sdl.Rect{p[0] - 3, p[1] - 3, 5, 5}
		renderer.Copy(staticTexture, &textureSource, &dest)
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

func setupTestTexture() {
	testTexture, _ = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, 1920, 1080)
	renderer.SetRenderTarget(testTexture)
	odd := true

	var font *ttf.Font
	var text *sdl.Surface
	var textTexture *sdl.Texture
	font, _ = ttf.OpenFont("fonts/DejaVuSans.ttf", 120)

	defer font.Close()

	for y := 0; y < 4; y++ {
		for x := 0; x < 10; x++ {
			rect := sdl.Rect{int32(x * dx), int32(y * dy), dx, dy}
			if odd {
				renderer.SetDrawColor(255, 0, 0, 255)
			} else {
				renderer.SetDrawColor(0, 0, 255, 255)
			}
			odd = !odd
			renderer.DrawRect(&rect)
			renderer.SetDrawColor(255, 255, 255, 255)

			text, _ = font.RenderUTF8Blended(strconv.Itoa(1+x+(y*10)), sdl.Color{R: 255, G: 255, B: 255, A: 255})
			defer text.Free()

			textX := int32((x*dx)+(dx/2)) - (text.W / 2)
			textY := int32((y*dy)+(dy/2)) - (text.H / 2)

			textTexture, _ = renderer.CreateTextureFromSurface(text)
			defer textTexture.Destroy()
			renderer.Copy(textTexture, &sdl.Rect{0, 0, text.W, text.H}, &sdl.Rect{textX, textY, text.W, text.H})
		}
		odd = !odd
	}
}

func setupSecTexture(textureSize int32) {
	secTexture, _ = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, textureSize, textureSize)
	secTexture.SetBlendMode(sdl.BLENDMODE_BLEND)
	renderer.SetRenderTarget(secTexture)
	renderer.SetDrawColor(secSDLColor.R, secSDLColor.G, secSDLColor.B, 255)
	for _, point := range circlePixels {
		if err := renderer.DrawPoint(point[0], point[1]); err != nil {
			panic(err)
		}
	}
}

func setupStaticTexture(textureSize int32) {
	staticTexture, _ = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, textureSize, textureSize)
	staticTexture.SetBlendMode(sdl.BLENDMODE_BLEND)
	renderer.SetRenderTarget(staticTexture)

	renderer.SetDrawColor(staticSDLColor.R, staticSDLColor.G, staticSDLColor.B, 255)
	for _, point := range circlePixels {
		if err := renderer.DrawPoint(point[0], point[1]); err != nil {
			panic(err)
		}
	}
}
