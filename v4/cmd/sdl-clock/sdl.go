package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"image/color"
	"log"
)

var colors struct {
	static     sdl.Color
	sec        sdl.Color
	text       sdl.Color
	countdown  sdl.Color
	tally      sdl.Color
	tallyBG    sdl.Color
	row        [3]sdl.Color
	rowBG      [3]sdl.Color
	icon       [3]sdl.Color
	signal     [4]sdl.Color
	label      sdl.Color
	labelBG    sdl.Color
	background sdl.Color
}

var window *sdl.Window

var renderer *sdl.Renderer

var textureSource sdl.Rect

var staticTexture *sdl.Texture
var secTexture *sdl.Texture
var backgroundTexture *sdl.Texture
var clockTextures []*sdl.Texture

var infoTexture *sdl.Texture
var infoFont *ttf.Font

// initSDL initializes the SDL library, creates a window and a hw accelerated renderer
func initSDL() {
	var err error
	if err = sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		log.Fatalf("Failed to initialize SDL: %s\n", err)
	}

	if window, err = sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight, sdl.WINDOW_OPENGL+sdl.WINDOW_SHOWN+sdl.WINDOW_RESIZABLE+sdl.WINDOW_ALLOW_HIGHDPI); err != nil {
		log.Fatalf("Failed to create window: %s\n", err)
	}

	if renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED); err != nil {
		log.Fatalf("Failed to create renderer: %s\n", err)
	}

	_, err = sdl.ShowCursor(0) // Hide mouse cursor
	check(err)

	err = renderer.Clear()
	check(err)

	err = ttf.Init()
	check(err)

	log.Printf("SDL init done\n")

	rendererInfo, err := renderer.GetInfo()
	check(err)
	log.Printf("Renderer: %v\n", rendererInfo.Name)

	infoFont, err = ttf.OpenFont(options.LabelFont, 50)
	check(err)
}

// initColors takes the color options from flags and translates them to sdl.Color variables
func initColors() {
	var err error

	colors.text, err = parseColor(options.TextColor)
	check(err)

	colors.static, err = parseColor(options.StaticColor)
	check(err)

	colors.sec, err = parseColor(options.SecondColor)
	check(err)

	colors.countdown, err = parseColor(options.CountdownColor)
	check(err)

	colors.row[0], err = parseColor(options.Row1Color)
	check(err)
	colors.row[0].A = options.Row1Alpha

	colors.row[1], err = parseColor(options.Row2Color)
	check(err)
	colors.row[1].A = options.Row1Alpha

	colors.row[2], err = parseColor(options.Row3Color)
	check(err)
	colors.row[2].A = options.Row1Alpha

	for i := 0; i < 3; i++ {
		colors.icon[i] = colors.row[i]
	}

	for i := 0; i < 4; i++ {
		colors.signal[i] = sdl.Color{R: 0, G: 0, B: 0, A: 0}
	}

	colors.label, err = parseColor(options.LabelColor)
	check(err)
	colors.label.A = options.LabelAlpha

	colors.labelBG, err = parseColor(options.LabelBG)
	check(err)
	colors.labelBG.A = options.LabelBGAlpha

	timerBG, err := parseColor(options.TimerBG)
	check(err)
	timerBG.A = options.TimerBGAlpha
	for i := 0; i < 3; i++ {
		colors.rowBG[i] = timerBG
	}

	colors.background, err = parseColor(options.BackgroundColor)
	check(err)

	colors.tally = sdl.Color{R: 0, G: 0, B: 0, A: 0}
}

// initTextures initializes the circle textures for seconds and static "hour" markers
func initTextures() {
	var textureSize int32 = 40
	var textureCoord int32 = 20
	var textureRadius int32 = 19
	var err error

	// Constants for the small 192x192 px clock
	if options.small {
		textureSize = 5
		textureCoord = 3
		textureRadius = 3
		if options.Face == "144" {
			gridStartX = 24
			gridStartY = 24
			gridSize = 2
			gridSpacing = 3
		} else {
			// 192x192px clock
			gridStartX = 32
			gridStartY = 32
			gridSize = 3
			gridSpacing = 4
		}
	}

	// Texture for 12 static circles
	renderer.SetDrawColor(0, 0, 0, 0)
	staticTexture, err = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, textureSize, textureSize)
	check(err)
	err = staticTexture.SetBlendMode(sdl.BLENDMODE_BLEND)
	check(err)

	err = renderer.SetRenderTarget(staticTexture)
	check(err)
	renderer.Clear()

	if !options.small {
		gfx.FilledCircleColor(renderer, textureCoord, textureCoord, textureRadius, colors.static)
		// gfx.AACircleColor(renderer, textureCoord, textureCoord, textureRadius, staticSDLColor)
	} else {
		err = renderer.SetDrawColor(colors.static.R, colors.static.G, colors.static.B, 255)
		check(err)

		for _, point := range circlePixels {
			if err := renderer.DrawPoint(point[0], point[1]); err != nil {
				panic(err)
			}
		}
	}

	// Texture for the second marker circles
	renderer.SetDrawColor(0, 0, 0, 0)
	secTexture, err = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, textureSize, textureSize)
	check(err)
	err = secTexture.SetBlendMode(sdl.BLENDMODE_BLEND)
	check(err)

	err = renderer.SetRenderTarget(secTexture)
	check(err)
	renderer.Clear()

	if !options.small {
		gfx.FilledCircleColor(renderer, textureCoord, textureCoord, textureRadius, colors.sec)
		// gfx.AACircleColor(renderer, textureCoord, textureCoord, textureRadius, secSDLColor)
	} else {
		err = renderer.SetDrawColor(colors.sec.R, colors.sec.G, colors.sec.B, 255)
		check(err)

		for _, point := range circlePixels {
			if err = renderer.DrawPoint(point[0], point[1]); err != nil {
				panic(err)
			}
		}
	}

	err = renderer.SetRenderTarget(nil)
	check(err)

	textureSource = sdl.Rect{X: 0, Y: 0, W: textureSize, H: textureSize}
}

func setupScaling() {
	if options.dualClock || options.textClock || options.countdown {
		// FIXME: rpi display scaling fix
		// Dual clock
		x, y, _ := renderer.GetOutputSize()
		log.Printf("SDL2 output size: %v, %v", x, y)
		if x > y {
			err := renderer.SetLogicalSize(1920, 1080)
			check(err)
		} else {
			// rotated display
			err := renderer.SetLogicalSize(1080, 1920)
			check(err)
		}
	} else if !options.NoARCorrection {
		rpiDisplayCorrection()
	}

}

// rpiDisplayCorrection detects the official 7" rpi display and applies aspect ratio correction.
// The official display has non-square pixels...
func rpiDisplayCorrection() {
	// the official raspberry pi display has weird pixels
	// We detect it by the unusual 800 x 480 resolution
	// We will eventually support rotated displays also
	x, y, _ := renderer.GetOutputSize()
	log.Printf("SDL renderer size: %v x %v", x, y)
	scaleX, scaleY := renderer.GetScale()
	log.Printf("Scaling: x: %v, y: %v\n", scaleX, scaleY)

	if (x == 800) && (y == 480) {
		// Official display, rotated 0 or 180 degrees
		// The display has non-square pixels and needs correction:
		// Y scale = 1
		// Scale for x is ((9*800) / (16*480)) = 0.9375
		err := renderer.SetScale(0.9375, 1)
		check(err)
		log.Printf("Detected official raspberry pi display, correcting aspect ratio\n")
		check(err)
	} else if (y == 800) && (x == 480) {
		// Official display rotated 90 or 270 degrees
		err := renderer.SetScale(1, 0.9375)
		check(err)
		log.Printf("Detected official raspberry pi display (rotated 90 or 270 deg), correcting aspect ratio.\n")
		log.Printf("Moving clock to top corner of the display.\n")
	}
}

// drawSecondCircles draws the requested amount of the second marker circles on the ring
func drawSecondCircles(seconds int) {
	// Clamp the array index
	if seconds > 59 {
		seconds = 59
	} else if seconds < 0 {
		seconds = 0
	}
	// Draw second circles
	for i := 0; i <= int(seconds); i++ {
		dest := sdl.Rect{X: secCircles[i].X - 20, Y: secCircles[i].Y - 20, W: 40, H: 40}
		if options.small {
			dest = sdl.Rect{X: secCircles[i].X - 3, Y: secCircles[i].Y - 3, W: 5, H: 5}
		}
		err := renderer.Copy(secTexture, &textureSource, &dest)
		check(err)
	}
}

// drawStaticCircles draws the 12 static "hour" marker circles
func drawStaticCircles() {
	// Draw static indicator circles
	for _, p := range staticCircles {
		if options.small {
			dest := sdl.Rect{X: p.X - 3, Y: p.Y - 3, W: 5, H: 5}
			err := renderer.Copy(staticTexture, &textureSource, &dest)
			check(err)
		} else {
			dest := sdl.Rect{X: p.X - 20, Y: p.Y - 20, W: 40, H: 40}
			err := renderer.Copy(staticTexture, &textureSource, &dest)
			check(err)
		}
	}
}

// drawDots draws the two dots between hours and minutes on the clock
func drawDots(y int, x int, c sdl.Color) {
	// Draw the dots between hours and minutes
	setMatrix(y, x, c)
	setMatrix(y, x+1, c)
	setMatrix(y+1, x, c)
	setMatrix(y+1, x+1, c)

	setMatrix(y+4, x, c)
	setMatrix(y+4, x+1, c)
	setMatrix(y+5, x, c)
	setMatrix(y+5, x+1, c)
}

// Fills the screen with white
func drawWhiteScreen() {
	err := renderer.SetRenderTarget(nil)
	check(err)

	err = renderer.SetDrawColor(255, 255, 255, 255)
	check(err)

	err = renderer.Clear()
	check(err)
}

// setMatrix draws a "led matrix" pixel
func setMatrix(cy, cx int, color sdl.Color) {
	x := gridStartX + int32(cx*gridSpacing)
	y := gridStartY + int32(cy*gridSpacing)
	rect := sdl.Rect{X: x, Y: y, W: gridSize, H: gridSize}
	err := renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	check(err)

	err = renderer.FillRect(&rect)
	check(err)
}

// setPixel sets a generic "pixel" on a grid
func setPixel(cy, cx int, color sdl.Color, startX, startY, spacing, pixelSize int32) {
	x := startX + int32(cx)*spacing
	y := startY + int32(cy)*spacing
	rect := sdl.Rect{X: x, Y: y, W: pixelSize, H: pixelSize}
	err := renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	check(err)

	err = renderer.FillRect(&rect)
	check(err)
}

// drawBitmask draws a 2d boolean array
func drawBitmask(bitmask [][]bool, color sdl.Color, r int, c int) {
	for y, row := range bitmask {
		for x, b := range row {
			if b {
				setMatrix(r+y, c+x, color)
			}
		}
	}
}

// loadBackground loads and processes the background image into sdl.Texture
func loadBackground(file string) {
	var err error
	backgroundImage, err := img.Load(file)
	if err == nil {
		if backgroundTexture != nil {
			backgroundTexture.Destroy()
		}
		// Create texture from surface
		backgroundTexture, err = renderer.CreateTextureFromSurface(backgroundImage)
		backgroundImage.Free()
		check(err)

		err = backgroundTexture.SetBlendMode(sdl.BLENDMODE_NONE)
		check(err)
		showBackground = true
		return
	}
	backgroundImage.Free()
	// Failed to load background image, continue without it
	log.Printf("Error loading background image: %v %v\n", options.Background, err)
	log.Printf("Disabling background image.")
	showBackground = false
}

// clearCanvas fills the whole SDL window with black
func clearCanvas() {
	err := renderer.SetDrawColor(0, 0, 0, 0)
	check(err)

	err = renderer.Clear()
	check(err)
}

// prepare the main window canvas with the background
func prepareCanvas() {
	err := renderer.SetRenderTarget(nil)
	check(err)

	err = renderer.SetDrawColor(colors.background.R, colors.background.G, colors.background.B, 255)
	check(err)

	err = renderer.Clear()
	check(err)

	// Copy the background image as needed
	if showBackground {
		renderer.Copy(backgroundTexture, nil, nil)
	}
}

// parseColor parses a string "#XXX or #XXXXXX to a sdl.Color"
func parseColor(s string) (c sdl.Color, err error) {
	c.A = 0xff
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("parseColor(): invalid length, must be 7 or 4: %v", s)
	}
	return
}

func toSDLColor(in color.RGBA) sdl.Color {
	return sdl.Color{
		R: in.R,
		G: in.G,
		B: in.B,
		A: in.A,
	}
}

func toRGBA(in sdl.Color) color.RGBA {
	return color.RGBA{
		R: in.R,
		G: in.G,
		B: in.B,
		A: in.A,
	}
}
