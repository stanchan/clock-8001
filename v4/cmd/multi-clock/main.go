package main

import (
	"bufio"
	"github.com/jessevdk/go-flags"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"gitlab.com/Depili/clock-8001/v4/clock"
	"gitlab.com/Depili/clock-8001/v4/debug"
	"gitlab.com/Depili/go-rgb-led-matrix/bdf"
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

var staticSDLColor = sdl.Color{R: 80, G: 80, B: 0, A: 255} // 12 static indicator circles
var secSDLColor = sdl.Color{R: 200, G: 0, B: 0, A: 255}
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

	_, err = sdl.ShowCursor(0) // Hide mouse cursor
	check(err)

	// Clear the screen
	err = renderer.Clear()
	check(err)
	defer renderer.Destroy()

	if err = ttf.Init(); err != nil {
		panic(err)
	}
	defer ttf.Quit()

	log.Printf("SDL init done\n")

	rendererInfo, _ := renderer.GetInfo()
	log.Printf("Renderer: %v\n", rendererInfo.Name)

	// Clock colors from flags
	textSDLColor = sdl.Color{R: options.TextRed, G: options.TextGreen, B: options.TextBlue, A: 255}
	staticSDLColor = sdl.Color{R: options.StaticRed, G: options.StaticGreen, B: options.StaticBlue, A: 255}
	secSDLColor = sdl.Color{R: options.SecRed, G: options.SecGreen, B: options.SecBlue, A: 255}
	// Default color for the OSC field (black)
	tallyColor := sdl.Color{R: 0, G: 0, B: 0, A: 255}

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

	err = renderer.SetRenderTarget(nil)
	check(err)
	textureSource = sdl.Rect{X: 0, Y: 0, W: textureSize, H: textureSize}

	// Trap SIGINT aka Ctrl-C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Cache the second ring texture
	ringTexture, err := renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, dx, dy)
	check(err)
	err = ringTexture.SetBlendMode(sdl.BLENDMODE_BLEND)
	check(err)

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
			err = renderer.SetDrawColor(0, 0, 0, 255) // Black
			check(err)
			err = renderer.Clear() // Clear screen
			check(err)
			if showTestPicture {
				// Show the test picture.
				err = renderer.Copy(testTexture, &sdl.Rect{X: 0, Y: 0, W: 1920, H: 1080}, &sdl.Rect{X: 0, Y: 0, W: 1920, H: 1080})
				check(err)
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

				err = renderer.SetRenderTarget(clockTexture)
				check(err)
				err = renderer.SetDrawColor(0, 0, 0, 255) // Black
				check(err)
				err = renderer.Clear() // Clear screen
				check(err)

				target := sdl.Rect{
					X: int32(dx * (i % 10)),
					Y: int32(dy * (i / 10)),
					H: dy,
					W: dx,
				}

				err = renderer.SetDrawColor(0, 0, 0, 255) // Black
				check(err)

				hourBitmap = font.TextBitmap(engine.Hours)
				minuteBitmap = font.TextBitmap(engine.Minutes)

				tallyColor = sdl.Color{R: feedBackEngine.TallyRed, G: feedBackEngine.TallyGreen, B: feedBackEngine.TallyBlue, A: 255}
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

					err = renderer.SetRenderTarget(ringTexture)
					check(err)
					err = renderer.SetDrawColor(0, 0, 0, 0) // Black
					check(err)
					err = renderer.Clear() // Clear screen
					check(err)

					drawStaticCircles()
					drawSecondCircles(seconds)
					if feedBackEngine.DisplaySeconds() {
						drawBitmask(secondBitmap, textSDLColor, 21, 8)
					}
				}

				err = renderer.SetRenderTarget(nil)
				check(err)
				err = renderer.Copy(clockTexture, &source, &target)
				check(err)
				err = renderer.Copy(ringTexture, &source, &target)
				check(err)
			}
			renderer.Present() // Update the screen

		}
	}
}

func drawSecondCircles(seconds int) {
	// Draw second circles
	for i := 0; i <= int(seconds); i++ {
		dest := sdl.Rect{X: secCircles[i][0] - 3, Y: secCircles[i][1] - 3, W: 5, H: 5}
		err := renderer.Copy(secTexture, &textureSource, &dest)
		check(err)
	}
}

func drawStaticCircles() {
	// Draw static indicator circles
	for _, p := range staticCircles {
		dest := sdl.Rect{X: p[0] - 3, Y: p[1] - 3, W: 5, H: 5}
		err := renderer.Copy(staticTexture, &textureSource, &dest)
		check(err)
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
	rect := sdl.Rect{X: x, Y: y, W: gridSize, H: gridSize}
	err := renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	check(err)
	err = renderer.FillRect(&rect)
	check(err)
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
	err := renderer.SetRenderTarget(testTexture)
	check(err)
	odd := true

	var font *ttf.Font
	var text *sdl.Surface
	var textTexture *sdl.Texture
	font, err = ttf.OpenFont("fonts/DejaVuSans.ttf", 120)
	check(err)

	defer font.Close()

	for y := 0; y < 4; y++ {
		for x := 0; x < 10; x++ {
			rect := sdl.Rect{X: int32(x * dx), Y: int32(y * dy), W: dx, H: dy}
			if odd {
				err = renderer.SetDrawColor(255, 0, 0, 255)
				check(err)
			} else {
				err = renderer.SetDrawColor(0, 0, 255, 255)
				check(err)
			}
			odd = !odd
			err = renderer.DrawRect(&rect)
			check(err)
			err = renderer.SetDrawColor(255, 255, 255, 255)
			check(err)

			text, err = font.RenderUTF8Blended(strconv.Itoa(1+x+(y*10)), sdl.Color{R: 255, G: 255, B: 255, A: 255})
			defer text.Free()
			check(err)

			textX := int32((x*dx)+(dx/2)) - (text.W / 2)
			textY := int32((y*dy)+(dy/2)) - (text.H / 2)

			textTexture, _ = renderer.CreateTextureFromSurface(text)
			defer textTexture.Destroy()
			err = renderer.Copy(textTexture, &sdl.Rect{X: 0, Y: 0, W: text.W, H: text.H}, &sdl.Rect{X: textX, Y: textY, W: text.W, H: text.H})
			check(err)
		}
		odd = !odd
	}
}

func setupSecTexture(textureSize int32) {
	var err error
	secTexture, err = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, textureSize, textureSize)
	check(err)

	err = secTexture.SetBlendMode(sdl.BLENDMODE_BLEND)
	check(err)

	err = renderer.SetRenderTarget(secTexture)
	check(err)

	err = renderer.SetDrawColor(secSDLColor.R, secSDLColor.G, secSDLColor.B, 255)
	check(err)

	for _, point := range circlePixels {
		if err := renderer.DrawPoint(point[0], point[1]); err != nil {
			panic(err)
		}
	}
}

func setupStaticTexture(textureSize int32) {
	var err error
	staticTexture, err = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, textureSize, textureSize)
	check(err)

	err = staticTexture.SetBlendMode(sdl.BLENDMODE_BLEND)
	check(err)

	err = renderer.SetRenderTarget(staticTexture)
	check(err)

	err = renderer.SetDrawColor(staticSDLColor.R, staticSDLColor.G, staticSDLColor.B, 255)
	check(err)

	for _, point := range circlePixels {
		if err := renderer.DrawPoint(point[0], point[1]); err != nil {
			panic(err)
		}
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
