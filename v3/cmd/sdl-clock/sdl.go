package main

import (
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"log"
)

var staticSDLColor = sdl.Color{R: 80, G: 80, B: 0, A: 255} // 12 static indicator circles
var secSDLColor = sdl.Color{R: 200, G: 0, B: 0, A: 255}
var textSDLColor sdl.Color
var tallyColor sdl.Color

var window *sdl.Window

var renderer *sdl.Renderer

var textureSource sdl.Rect

var staticTexture *sdl.Texture
var secTexture *sdl.Texture
var backgroundTexture *sdl.Texture

// initSDL initializes the SDL library, creates a window and a hw accelerated renderer
func initSDL() {
	var err error
	if err = sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		log.Fatalf("Failed to initialize SDL: %s\n", err)
	}

	if window, err = sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight, sdl.WINDOW_OPENGL+sdl.WINDOW_SHOWN+sdl.WINDOW_RESIZABLE); err != nil {
		log.Fatalf("Failed to create window: %s\n", err)
	}

	if renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED); err != nil {
		log.Fatalf("Failed to create renderer: %s\n", err)
	}

	_, err = sdl.ShowCursor(0) // Hide mouse cursor
	check(err)

	err = renderer.Clear()
	check(err)

	log.Printf("SDL init done\n")

	rendererInfo, err := renderer.GetInfo()
	check(err)
	log.Printf("Renderer: %v\n", rendererInfo.Name)
}

// initColors takes the color options from flags and translates them to sdl.Color variables
func initColors() {
	textSDLColor = sdl.Color{R: options.TextRed, G: options.TextGreen, B: options.TextBlue, A: 255}
	staticSDLColor = sdl.Color{R: options.StaticRed, G: options.StaticGreen, B: options.StaticBlue, A: 255}
	secSDLColor = sdl.Color{R: options.SecRed, G: options.SecGreen, B: options.SecBlue, A: 255}
	// Default color for the OSC field (black)
	tallyColor = sdl.Color{R: 0, G: 0, B: 0, A: 255}
}

// initTextures initializes the circle textures for seconds and static "hour" markers
func initTextures() {
	var textureSize int32 = 40
	var textureCoord int32 = 20
	var textureRadius int32 = 19
	var err error

	// Constants for the small 192x192 px clock
	if options.Small {
		textureSize = 5
		textureCoord = 3
		textureRadius = 3
		gridStartX = 32
		gridStartY = 32
		gridSize = 3
		gridSpacing = 4
	}

	// Texture for 12 static circles
	staticTexture, err = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, textureSize, textureSize)
	check(err)
	err = staticTexture.SetBlendMode(sdl.BLENDMODE_NONE)
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

	// Texture for the second marker circles
	secTexture, err = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, textureSize, textureSize)
	check(err)
	err = secTexture.SetBlendMode(sdl.BLENDMODE_NONE)
	check(err)

	err = renderer.SetRenderTarget(secTexture)
	check(err)

	if !options.Small {
		gfx.FilledCircleColor(renderer, textureCoord, textureCoord, textureRadius, secSDLColor)
		// gfx.AACircleColor(renderer, textureCoord, textureCoord, textureRadius, secSDLColor)
	} else {
		err = renderer.SetDrawColor(secSDLColor.R, secSDLColor.G, secSDLColor.B, 255)
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
	// Draw second circles
	for i := 0; i <= int(seconds); i++ {
		dest := sdl.Rect{X: secCircles[i].X - 20, Y: secCircles[i].Y - 20, W: 40, H: 40}
		if options.Small {
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
		if options.Small {
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
func loadBackground() bool {
	var err error
	backgroundImage, err := img.Load(options.Background)
	if err == nil {
		// Create texture from surface
		backgroundTexture, err = renderer.CreateTextureFromSurface(backgroundImage)
		check(err)

		err = backgroundTexture.SetBlendMode(sdl.BLENDMODE_NONE)
		check(err)
		return true
	}

	// Failed to load background image, continue without it
	log.Printf("Error loading background image: %v %v\n", options.Background, err)
	log.Printf("Disabling background image.")
	return false
}

// clearCanvas fills the whole SDL window with black
func clearCanvas() {
	err := renderer.SetDrawColor(0, 0, 0, 0)
	check(err)

	err = renderer.Clear()
	check(err)
}
