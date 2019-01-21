package clock

import (
	"time"
)

const (
	Normal    = iota
	Countdown = iota
	Countup   = iota
	Off       = iota
)

type Engine struct {
	Timezone        *time.Location
	Mode            int
	countdownTarget time.Time
	Hours           string
	Minutes         string
	Seconds         string
	Leds            int
	countdownLeds   bool
	flasher         *time.Ticker
}

// Create a clock engine
func MakeEngine(timezone string, flash time.Duration) (*Engine, error) {
	var engine = Engine{
		Mode:          Normal,
		Hours:         "",
		Minutes:       "",
		Seconds:       "",
		Leds:          0,
		countdownLeds: true,
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

func (engine *Engine) flash() {
	for range engine.flasher.C {
		engine.countdownLeds = !engine.countdownLeds
	}
}

// Check if the countdown timer has elapsed
func (engine *Engine) CountDownDone() bool {
	return time.Now().After(engine.countdownTarget)
}

// Start a countdown timer
func (engine *Engine) StartCountdown(timer time.Duration) {
	engine.Mode = Countdown
	engine.countdownTarget = time.Now().Add(timer)
}

// Update Hours, Minutes and Seconds
func (engine *Engine) Update() {
	switch engine.Mode {
	case Normal:
		engine.normalUpdate()
	case Countdown:
		engine.countdownUpdate()
	case Countup:
	case Off:
		engine.Hours = ""
		engine.Minutes = ""
		engine.Seconds = ""
		engine.Leds = 0
	}
}

func (engine *Engine) countdownUpdate() {
	t := time.Now()
	diff := engine.countdownTarget.Sub(t)
	display := time.Time{}.Add(diff)
	engine.Seconds = ""

	if t.Before(engine.countdownTarget) {
		engine.Hours = display.Format("04")
		engine.Minutes = display.Format("05")
		if display.Minute() > 0 {
			engine.Leds = 59
		} else {
			engine.Leds = display.Second()
		}
	} else {
		if engine.countdownLeds {
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

func (engine *Engine) normalUpdate() {
	t := time.Now().In(engine.Timezone)

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
