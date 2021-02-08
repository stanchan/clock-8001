package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"gitlab.com/Depili/clock-8001/v4/clock"
	"gitlab.com/Depili/clock-8001/v4/debug"
	"log"
	"strconv"
	"strings"
)

type outputLine struct {
	icon          string
	text          string
	label         string
	iconTex       *sdl.Texture
	textTex       *sdl.Texture
	labelTex      *sdl.Texture
	timeFragments [10]*sdl.Texture
	fragmentRect  sdl.Rect
	colonTex      *sdl.Texture
	colonRect     sdl.Rect
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
	ltcText    string
	ltcTex     [2]*sdl.Texture
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

	textClock.ltcText = "00:00:00:00"

	log.Printf("Precalcs!")
	for j := range textClock.r {
		for i := range textClock.r[j].timeFragments {
			text := fmt.Sprintf("%01d", i)
			textClock.r[j].timeFragments[i] = renderText(text, textClock.numberFont, colors.rows[j])
		}
		_, _, w, h, _ := textClock.r[j].timeFragments[0].Query()
		textClock.r[j].fragmentRect = sdl.Rect{X: 0, Y: 0, W: w, H: h}
		textClock.r[j].colonTex = renderText(":", textClock.numberFont, colors.rows[j])
		_, _, w, h, _ = textClock.r[j].colonTex.Query()
		textClock.r[j].colonRect = sdl.Rect{X: 0, Y: 0, W: w, H: h}
	}
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

			if parts := strings.Split(text, ":"); len(parts) > 1 {
				// Compose multipart texture
				l := int32(len(parts))
				texW := l * textClock.r[i].fragmentRect.W * 2
				texW += (l - 1) * textClock.r[i].colonRect.W
				texH := textClock.r[i].fragmentRect.H

				textClock.r[i].textTex, err = renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, texW, texH)
				textClock.r[i].textTex.SetBlendMode(sdl.BLENDMODE_BLEND)
				check(err)
				renderer.SetRenderTarget(textClock.r[i].textTex)
				renderer.SetDrawColor(0, 0, 0, 0)
				renderer.Clear()

				target := sdl.Rect{}
				for j := 0; j < len(parts)-1; j++ {
					frag, err := strconv.Atoi(parts[j])
					check(err)

					target.W = textClock.r[i].fragmentRect.W
					target.H = textClock.r[i].fragmentRect.H
					renderer.Copy(textClock.r[i].timeFragments[frag/10], nil, &target)
					target.X += target.W

					renderer.Copy(textClock.r[i].timeFragments[frag%10], nil, &target)
					target.X += target.W

					target.W = textClock.r[i].colonRect.W
					renderer.Copy(textClock.r[i].colonTex, nil, &target)
					target.X += target.W
				}
				frag, _ := strconv.Atoi(parts[len(parts)-1])
				target.W = textClock.r[i].fragmentRect.W
				target.H = textClock.r[i].fragmentRect.H
				renderer.Copy(textClock.r[i].timeFragments[frag/10], nil, &target)
				target.X += target.W

				renderer.Copy(textClock.r[i].timeFragments[frag%10], nil, &target)
				target.X += target.W
				renderer.SetRenderTarget(nil)

			} else {
				textClock.r[i].textTex = renderText(text, textClock.numberFont, colors.rows[i])
			}
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
				err = textClock.r[i].iconTex.SetAlphaMod(0)
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
	if err != nil {
		log.Printf("renderText RenderUTF8Blended error: %v")
		log.Printf("rendering error text")
		t, err = font.RenderUTF8Blended("INVALID TEXT", color)
		check(err)
	}

	tex, err := renderer.CreateTextureFromSurface(t)
	if err != nil {
		log.Printf("renderText CreateTextureFromSurface error: %v")
		log.Printf("rendering error text")
		t.Free()
		t, err = font.RenderUTF8Blended("INVALID TEXT", color)
		check(err)
		tex, err = renderer.CreateTextureFromSurface(t)
		check(err)
	}
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
