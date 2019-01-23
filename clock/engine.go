package clock

import (
	"fmt"
	"math"
	"time"
)

const (
	Normal    = iota
	Countdown = iota
	Countup   = iota
	Off       = iota
)

type Engine struct {
	Timezone          *time.Location
	Mode              int
	countTarget       time.Time
	countdownDuration time.Duration
	Hours             string
	Minutes           string
	Seconds           string
	Leds              int
	Dots              bool
	flashLeds         bool
	flasher           *time.Ticker
}

// Create a clock engine
func MakeEngine(timezone string, flash time.Duration) (*Engine, error) {
	var engine = Engine{
		Mode:      Normal,
		Hours:     "",
		Minutes:   "",
		Seconds:   "",
		Leds:      0,
		flashLeds: true,
		Dots:      true,
	}
	tz, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, err
	}
	engine.Timezone = tz
	engine.flasher = time.NewTicker(flash)
	go engine.flash()
	return &engine, nil
}

func (engine *Engine) Kill() {
	engine.Mode = Off
}

func (engine *Engine) flash() {
	for range engine.flasher.C {
		engine.flashLeds = !engine.flashLeds
	}
}

// Check if the countdown timer has elapsed
func (engine *Engine) CountDownDone() bool {
	return time.Now().After(engine.countTarget)
}

// Start a countdown timer
func (engine *Engine) StartCountdown(timer time.Duration) {
	engine.Mode = Countdown
	engine.countTarget = time.Now().Add(timer)
	engine.countdownDuration = timer
}

func (engine *Engine) StartCountup() {
	engine.Mode = Countup
	engine.countTarget = time.Now()
}

func (engine *Engine) ModifyCountdown(delta time.Duration) {
	if engine.Mode == Countdown {
		engine.countTarget = engine.countTarget.Add(delta)
		engine.countdownDuration += delta
	}
}

func (engine *Engine) Normal() {
	engine.Mode = Normal
}

// Update Hours, Minutes and Seconds
func (engine *Engine) Update() {
	switch engine.Mode {
	case Normal:
		engine.normalUpdate()
	case Countdown:
		engine.countdownUpdate()
	case Countup:
		engine.countupUpdate()
	case Off:
		engine.Hours = ""
		engine.Minutes = ""
		engine.Seconds = ""
		engine.Leds = 0
		engine.Dots = false
	}
}

func (engine *Engine) countupUpdate() {
	t := time.Now()
	diff := t.Sub(engine.countTarget)
	display := time.Time{}.Add(diff)
	engine.Seconds = ""
	engine.Dots = true

	if t.After(engine.countTarget) {
		engine.formatCount(display)
		engine.Leds = display.Second()
	} else {
		engine.Hours = "00"
		engine.Minutes = "00"
		engine.Leds = 59
	}
}

func (engine *Engine) countdownUpdate() {
	t := time.Now()
	diff := engine.countTarget.Sub(t)
	display := time.Time{}.Add(diff)
	engine.Seconds = ""
	engine.Dots = true

	if t.Before(engine.countTarget) {
		engine.formatCount(display)
		progress := (float64(diff) / float64(engine.countdownDuration))
		engine.Leds = int(math.Floor(progress * 60))
	} else {
		if engine.flashLeds {
			engine.Hours = "00"
			engine.Minutes = "00"
			engine.Leds = 59
		} else {
			engine.Hours = ""
			engine.Minutes = ""
			engine.Leds = 59
		}
	}
}

func (engine *Engine) formatCount(display time.Time) {
	if display.Hour() == 0 {
		engine.Hours = display.Format("04")
		engine.Minutes = display.Format("05")
	} else if display.Hour() == 1 && display.Minute() < 40 {
		min := 60 + display.Minute()
		engine.Hours = fmt.Sprintf("%d", min)
		engine.Minutes = display.Format("05")
	} else {
		// More than 99 minutes to display
		engine.Hours = "99"
		engine.Minutes = "99"
	}
}

func (engine *Engine) normalUpdate() {
	t := time.Now().In(engine.Timezone)
	engine.Dots = true

	// Check that the rpi has valid time
	if t.Year() > 2000 {
		engine.Hours = t.Format("15")
		engine.Minutes = t.Format("04")
		engine.Seconds = t.Format("05")
	} else {
		// No valid time, indicate it with "XX" as the time
		engine.Hours = "XX"
		engine.Minutes = "XX"
		engine.Seconds = ""
		engine.Leds = 59
	}
	engine.Leds = t.Second()
}
