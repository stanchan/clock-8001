package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"gitlab.com/Depili/clock-8001/v4/clock"
	"math"
	"strconv"
	"strings"
)

/*
 * Code for the original round clock faces
 */

func drawRoundClocks(state *clock.State) {
	var err error

	numClocks := 1

	if options.dualClock {
		numClocks = 2
	}

	for i := 0; i < numClocks; i++ {
		mainClock := state.Clocks[(2 * i)]
		auxClock := state.Clocks[(2*i)+1]
		tally := ""
		hours := ""
		minutes := ""
		seconds := ""
		leds := 0

		if mainClock.Text != "" {
			if mainClock.Mode == clock.LTC {
				parts := strings.Split(mainClock.Text, ":")
				tally = fmt.Sprintf(" %s", parts[0]) // Hours
				hours = parts[1]                     // Minutes
				minutes = parts[2]                   // Seconds
				seconds = parts[3]                   // Frames
				if options.EngineOptions.LTCSeconds {
					leds, _ = strconv.Atoi(minutes)
				} else {
					leds, _ = strconv.Atoi(seconds)
				}
				colors.tally = colors.text

			} else if !mainClock.Hidden {
				parts := strings.Split(mainClock.Text, ":")
				if len(parts) == 2 {
					// "00:00" view
					hours = "00"
					minutes = parts[0]
					seconds = parts[1]
					leds, _ = strconv.Atoi(minutes)
				} else if len(parts) == 3 {
					// "00:00:00" view
					hours = parts[0]
					minutes = parts[1]
					seconds = parts[2]
					leds, _ = strconv.Atoi(seconds)
				} else {
					// Unknown
					hours = "XX"
					minutes = "XX"
					seconds = "XX"
					leds = 0
				}

				if !state.DisplaySeconds && mainClock.Mode == clock.Normal {
					seconds = ""
				}

				if mainClock.Mode == clock.Countdown || mainClock.Mode == clock.Media {
					if mainClock.Expired {
						// TODO: Multiple different options of expired timers?
						seconds = ""
						leds = 59
						if state.Flash {
							hours = "00"
							minutes = "00"

						} else {
							hours = ""
							minutes = ""
						}
					} else {
						if hours == "00" {
							hours = minutes
							minutes = seconds
							seconds = ""
						}
						leds = int(math.Floor(mainClock.Progress * 59))
					}
				} else if mainClock.Mode == clock.Countup {
					if mainClock.Expired {
						hours = "00"
						minutes = "00"
						seconds = ""
					} else if hours == "00" {
						hours = minutes
						minutes = seconds
						seconds = ""
					}
					leds, _ = strconv.Atoi(minutes)
				}
			}
		}

		if mainClock.Mode != clock.LTC && !options.dualClock {
			if state.Tally != "" {
				tally = fmt.Sprintf("%-.4s", state.Tally)
				colors.tally = sdl.Color{R: state.TallyColor.R, G: state.TallyColor.G, B: state.TallyColor.B, A: 255}

			} else if auxClock.Mode != clock.Normal && !auxClock.Hidden {
				if auxClock.Expired {
					if state.Flash {
						tally = " 00"
					}
				} else {
					tally = auxClock.Compact
					colors.tally = colors.countdown
				}
			}
		}
		hourBitmap := font.TextBitmap(hours)
		minuteBitmap := font.TextBitmap(minutes)
		secondBitmap := font.TextBitmap(seconds)
		tallyBitmap := font.TextBitmap(tally)

		// Set renderer target to the corresponding clock texture
		err = renderer.SetRenderTarget(clockTextures[i])
		check(err)

		clearCanvas()

		// Dots between hours and minutes
		haveDisplay := (hours != "") && (minutes != "")
		if haveDisplay && (!mainClock.Paused || state.Flash) && (mainClock.Mode != clock.Off) {
			drawDots()
		}

		// Draw the text
		drawBitmask(hourBitmap, colors.text, 10, 0)
		drawBitmask(minuteBitmap, colors.text, 10, 17)
		drawBitmask(secondBitmap, colors.text, 21, 8)
		drawBitmask(tallyBitmap, colors.tally, 0, 2)

		drawStaticCircles()
		drawSecondCircles(leds)
	}

	composeRoundClocks(state)
}

func composeRoundClocks(state *clock.State) {
	err := renderer.SetRenderTarget(nil)
	check(err)

	// Clear output and setup background
	prepareCanvas()

	source := sdl.Rect{X: 0, Y: 0, W: 1080, H: 1080}

	// FIXME: the text positioning and size is just magic numbers

	if options.dualClock {
		// Render the dual clock displays
		dualText := font.TextBitmap(fmt.Sprintf("%-.8s", state.Tally))

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
						setPixel(y, x, colors.text, (1920-1064)/2, 800+50, 19, 16)
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
						setPixel(y, x, colors.text, (1080-1064)/2, 800+50, 19, 16)
					}
				}
			}

		}
	} else {
		// Single clock mode
		x, y, _ := renderer.GetOutputSize()
		var dest sdl.Rect

		if options.small {
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
}
