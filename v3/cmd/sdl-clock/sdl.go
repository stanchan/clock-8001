package main

import (
	"github.com/veandco/go-sdl2/gfx"
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

func initColors() {
	textSDLColor = sdl.Color{R: options.TextRed, G: options.TextGreen, B: options.TextBlue, A: 255}
	staticSDLColor = sdl.Color{R: options.StaticRed, G: options.StaticGreen, B: options.StaticBlue, A: 255}
	secSDLColor = sdl.Color{R: options.SecRed, G: options.SecGreen, B: options.SecBlue, A: 255}
	// Default color for the OSC field (black)
	tallyColor = sdl.Color{R: 0, G: 0, B: 0, A: 255}
}

func initTextures() {
	var textureSize int32 = 40
	var textureCoord int32 = 20
	var textureRadius int32 = 19
	var err error

	// Create a texture for circles
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
