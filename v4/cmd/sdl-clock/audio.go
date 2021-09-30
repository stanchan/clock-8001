package main

import (
	_ "embed"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
	"gitlab.com/Depili/clock-8001/v4/clock"
)

//go:embed 1kHz_100ms.wav
var shortWav []byte

//go:embed 1kHz_500ms.wav
var longWav []byte

var shortBeep *mix.Chunk
var longBeep *mix.Chunk

var numAudioSources int
var lastBeep []int

func initAudio() {
	var err error

	if err = mix.OpenAudio(44100, mix.DEFAULT_FORMAT, 2, 4096); err != nil {
		panic(err)
	}

	shortBeep, err = mix.QuickLoadWAV(shortWav)
	if err != nil {
		panic(err)
	}

	longBeep, err = mix.QuickLoadWAV(longWav)
	if err != nil {
		panic(err)
	}
	lastBeep = make([]int, 4)
}

func checkBeep(s *clock.State, i int) {
	clk := s.Clocks[i]
	if clk.Mode == clock.Countdown {
		if clk.Hours == 0 && clk.Minutes == 0 {
			if clk.Seconds > 5 {
				lastBeep[i] = 6
			} else if clk.Seconds <= 5 && lastBeep[i] > clk.Seconds {
				if clk.Seconds == 0 {
					longBeep.Play(1, 0)
				} else {
					shortBeep.Play(1, 0)
				}
				lastBeep[i]--
			}
		}
	}
}

func play() {

	defer mix.CloseAudio()

	// Play 4 times
	shortBeep.Play(1, 0)
	sdl.Delay(900)
	shortBeep.Play(1, 0)
	sdl.Delay(900)
	shortBeep.Play(1, 0)
	sdl.Delay(900)
	shortBeep.Play(1, 0)
	sdl.Delay(900)
	longBeep.Play(1, 0)
	sdl.Delay(500)

	// Wait until it finishes playing
	for mix.Playing(-1) == 1 {
		sdl.Delay(16)
	}
}
