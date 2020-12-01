package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	// "gitlab.com/Depili/clock-8001/v4/clock"
	"gitlab.com/Depili/clock-8001/v4/debug"
	"log"
	"math"
	"time"
)

var countdown struct {
	largeFont *ttf.Font
	smallFont *ttf.Font
	bgColor   sdl.Color
	color     sdl.Color
	target    time.Time
	loc       *time.Location
}

func initCountdown() {
	log.Printf("Initializing countdown display...")

	var f *ttf.Font
	var err error

	if f, err = ttf.OpenFont(options.NumberFont, 250); err != nil {
		panic(err)
	}
	countdown.smallFont = f

	if f, err = ttf.OpenFont(options.NumberFont, 350); err != nil {
		panic(err)
	}
	countdown.largeFont = f

	countdown.loc, err = time.LoadLocation(options.EngineOptions.Source1.TimeZone)
	check(err)

	countdown.target, err = time.ParseInLocation("2006-01-02 15:04:05", options.CountdownTarget, countdown.loc)
	check(err)
	log.Printf("Target: %v", countdown.target)

	countdown.color = sdl.Color{R: 255, G: 255, B: 255, A: 255}
	countdown.bgColor = sdl.Color{R: 0, G: 0, B: 0, A: 255}

	log.Printf("Countdown face intialized.")
}

func drawCountdown() {
	debug.Printf("drawCountdown")

	t := countdown.target.Sub(time.Now().In(countdown.loc))

	hours := t.Truncate(time.Hour).Hours()
	minutes := t.Truncate(time.Minute).Minutes() - (hours * 60)
	seconds := t.Truncate(time.Second).Seconds() - (hours * 60 * 60) - (minutes * 60)
	days := hours / 24
	hours -= math.Floor(days) * 24

	if t.Seconds() < 0 {
		days = 0.0
		hours = 0.0
		minutes = 0.0
		seconds = 0.0
	}

	dayTex := renderText(fmt.Sprintf("%.0f", days), countdown.largeFont, countdown.color)
	defer dayTex.Destroy()

	lineTex := renderText(fmt.Sprintf("%02.0f:%02.0f:%02.0f", hours, minutes, seconds), countdown.smallFont, countdown.color)
	defer lineTex.Destroy()

	prepareCanvas()

	dayRect := sdl.Rect{Y: (1080 / 2) - 200, H: 400, X: (1920 / 2) - 300, W: 600}
	lineRect := sdl.Rect{X: 10, Y: 755, H: 300, W: 1920 - 20}
	copyIntoRect(dayTex, dayRect)
	copyIntoRect(lineTex, lineRect)
}
