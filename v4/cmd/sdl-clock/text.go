package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"gitlab.com/Depili/clock-8001/v4/clock"
	"gitlab.com/Depili/clock-8001/v4/debug"
	"log"
)

type outputLine struct {
	icon     string
	text     string
	label    string
	iconTex  *sdl.Texture
	textTex  *sdl.Texture
	labelTex *sdl.Texture
}

var textClock struct {
	main       *sdl.Texture
	numberFont *ttf.Font
	labelFont  *ttf.Font
	iconFont   *ttf.Font
	todColor   sdl.Color
	todBG      sdl.Color
	labelColor sdl.Color
	labelBG    sdl.Color
	r          [3]outputLine
}

// Font sizes. Rpi <4 is limited to 2048x2048 texture size.
const (
	numberSize = 100
	labelSize  = 200
	iconSize   = 200
)

func initTextClock() {
	var f *ttf.Font
	var err error

	if f, err = ttf.OpenFont(options.NumberFont, options.NumberFontSize); err != nil {
		panic(err)
	}
	textClock.numberFont = f

	if f, err = ttf.OpenFont(options.LabelFont, labelSize); err != nil {
		panic(err)
	}
	textClock.labelFont = f

	if f, err = ttf.OpenFont(options.IconFont, iconSize); err != nil {
		panic(err)
	}
	textClock.iconFont = f

	log.Printf("Text clock face intialized.")
}

func drawTextClock(state *clock.State) {
	var err error
	var x, y int32

	for i := range textClock.r {
		clk := state.Clocks[i]
		if clk.Hidden {
			continue
		}
		// Icon would need better unicode support from the fonts

		// text := fmt.Sprintf("%s %s", state.Clocks[i].Icon, state.Clocks[i].Text)

		text := clk.Text
		if clk.Expired && clk.Mode == clock.Countdown {
			if !state.Flash {
				text = " "
			} else {
				text = "00:00:00"
			}
		}

		if clk.Expired && clk.Mode == clock.Countup {
			text = "00:00:00"
		}

		if textClock.r[i].text != text {
			textClock.r[i].text = text
			if textClock.r[i].textTex != nil {
				textClock.r[i].textTex.Destroy()
			}

			textClock.r[i].textTex = renderText(text, textClock.numberFont, colors.rows[i])
		}

		label := fmt.Sprintf("%.10s", clk.Label)
		if textClock.r[i].label != label {
			textClock.r[i].label = label
			if textClock.r[i].labelTex != nil {
				textClock.r[i].labelTex.Destroy()
			}

			textClock.r[i].labelTex = renderText(label, textClock.labelFont, colors.label)
		}

		icon := materialIcon(clk.Icon)
		if textClock.r[i].icon != icon {
			textClock.r[i].icon = icon
			if textClock.r[i].iconTex != nil {
				textClock.r[i].iconTex.Destroy()
			}
			if icon != "" {
				textClock.r[i].iconTex = renderText(icon, textClock.iconFont, colors.rows[i])
			} else {
				renderer.SetDrawColor(0, 0, 0, 0)
				textClock.r[i].iconTex, err = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, 1, 1)
				check(err)
				err = textClock.r[i].iconTex.SetBlendMode(sdl.BLENDMODE_BLEND)
				check(err)
			}
		}
	}

	// Clear output and setup background
	prepareCanvas()

	if options.SingleLine && !state.Clocks[0].Hidden {
		labelR := sdl.Rect{X: 25, Y: 115, H: 150, W: 900}
		// 25px margin bellow label
		numberBox := sdl.Rect{X: 25, Y: 290, H: 440, W: 1920 - 50}
		iconR := sdl.Rect{X: 25, Y: 290, H: 440, W: 300}
		textR := sdl.Rect{X: 375, Y: 290, H: 440, W: 1920 - 425}

		if options.DrawBoxes {
			// Draw the placeholder boxes for timers and labels
			renderer.SetDrawColor(colors.timerBG.R, colors.timerBG.G, colors.timerBG.B, colors.timerBG.A)
			renderer.FillRect(&numberBox)

			renderer.SetDrawColor(colors.labelBG.R, colors.labelBG.G, colors.labelBG.B, colors.labelBG.A)
			renderer.FillRect(&labelR)
		}

		copyIntoRect(textClock.r[0].labelTex, labelR)
		if state.Clocks[0].Mode != clock.LTC {
			// Clock time

			copyIntoRect(textClock.r[0].textTex, textR)
			if textClock.r[0].iconTex != nil {
				copyIntoRect(textClock.r[0].iconTex, iconR)
			} else {
				debug.Printf("Nil icon texture!")
			}
		} else {
			// LTC
			// Maintain little spacing with the box borders
			numberBox.Y = numberBox.Y + 10
			numberBox.W = numberBox.W - 20

			copyIntoRect(textClock.r[0].textTex, numberBox)
		}
	} else if options.SingleLine && state.Clocks[0].Hidden {
		// Nothing to do
	} else {
		// 3 rows
		for i := range textClock.r {
			if state.Clocks[i].Hidden {
				// Row is hidden
				continue
			}
			y = 25 + (365 * int32(i))
			x = 530
			numberBox := sdl.Rect{X: x, Y: y, W: 1380, H: 300}
			textR := sdl.Rect{X: x + 300, Y: y, W: 1380 - 300, H: 300}
			iconR := sdl.Rect{X: x, Y: y, W: 300, H: 300}
			x = 10
			labelR := sdl.Rect{X: x, Y: y, W: 500, H: 100}

			if options.DrawBoxes {
				// Draw the placeholder boxes for timers and labels
				renderer.SetDrawColor(colors.timerBG.R, colors.timerBG.G, colors.timerBG.B, colors.timerBG.A)
				renderer.FillRect(&numberBox)

				renderer.SetDrawColor(colors.labelBG.R, colors.labelBG.G, colors.labelBG.B, colors.labelBG.A)
				renderer.FillRect(&labelR)
			}

			copyIntoRect(textClock.r[i].labelTex, labelR)
			if state.Clocks[i].Mode != clock.LTC {
				// Clock time

				copyIntoRect(textClock.r[i].textTex, textR)
				if textClock.r[i].iconTex != nil {
					copyIntoRect(textClock.r[i].iconTex, iconR)
				} else {
					debug.Printf("Nil icon texture!")
				}

			} else {
				// LTC

				// Maintain little spacing with the box borders
				numberBox.Y = numberBox.Y + 10
				numberBox.W = numberBox.W - 20

				copyIntoRect(textClock.r[i].textTex, numberBox)
			}
		}
	}

	// Draw possible OSC text message
	if state.Tally != "" {
		tallyColor := sdl.Color{
			R: state.TallyColor.R,
			G: state.TallyColor.G,
			B: state.TallyColor.B,
			A: state.TallyColor.A,
		}
		bgColor := sdl.Color{
			R: state.TallyBG.R,
			G: state.TallyBG.G,
			B: state.TallyBG.B,
			A: state.TallyBG.A,
		}
		tallyTexture := renderText(state.Tally, textClock.labelFont, tallyColor)
		tallyTexture.SetBlendMode(sdl.BLENDMODE_BLEND)
		tallyTexture.SetAlphaMod(tallyColor.A)

		tallyRect := sdl.Rect{X: 10, Y: 25 + (365 * 2), W: 1920 - 20, H: 300}
		if options.SingleLine {
			tallyRect.X = 25
			tallyRect.W = 1920 - 50
		}

		x1 := tallyRect.X
		y1 := tallyRect.Y

		x2 := x1 + tallyRect.W
		y2 := y1 + tallyRect.H

		gfx.BoxColor(renderer, x1, y1, x2, y2, bgColor)

		// renderer.FillRect(&tallyRect)
		copyIntoRect(tallyTexture, tallyRect)

		tallyTexture.Destroy()
	}
}

func destroyTextures(textures []*sdl.Texture) {
	for i := range textures {
		if textures[i] != nil {
			textures[i].Destroy()
		}
	}
}

func copyIntoRect(t *sdl.Texture, r sdl.Rect) {
	_, _, w, h, err := t.Query()
	if err != nil {
		debug.Printf("copyIntoRect: %v", err)
		return
	}
	dest := centerRect(w, h, r)
	renderer.Copy(t, nil, &dest)
}

func renderText(text string, font *ttf.Font, color sdl.Color) *sdl.Texture {
	if text == "" {
		text = " "
	}

	t, err := font.RenderUTF8Blended(text, color)
	check(err)

	tex, err := renderer.CreateTextureFromSurface(t)
	check(err)
	t.Free()
	t = nil
	return tex
}

func centerRect(w, h int32, r sdl.Rect) sdl.Rect {
	dest := sdl.Rect{}
	rSource := float64(w) / float64(h)
	rDest := float64(r.W) / float64(r.H)
	if rSource < rDest {
		dest.W = w * r.H / h
		dest.H = r.H
	} else {
		dest.W = r.W
		dest.H = h * r.W / w
	}
	dest.X = r.X + ((r.W - dest.W) / 2)
	dest.Y = r.Y + ((r.H - dest.H) / 2)
	return dest
}

// Substitute unicode glyphs used for icons to material design icon font private glyphs
func materialIcon(icon string) string {
	switch icon {
	case "Ⅱ":
		return "\ue034"
	case "↓":
		return "\ue5db"
	case "↑":
		return "\ue5d8"
	case "⇄":
		return "\ue040"
	case "▶":
		return "\ue037"
	}
	return ""
}

func bounded(in, max int32) int32 {
	if in > max {
		return max
	}
	return in
}
