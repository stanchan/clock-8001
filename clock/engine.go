package clock

import (
	"fmt"
	"github.com/desertbit/timer"
	"github.com/hypebeast/go-osc/osc"
	"log"
	"math"
	"time"
)

const (
	Normal    = iota // Display current time
	Countdown = iota // Display countdown timer only
	Countup   = iota // Count time up
	Off       = iota // (Mostly) blank screen
)

type Engine struct {
	Timezone           *time.Location
	Mode               int
	countTarget        time.Time
	countdownDuration  time.Duration
	countdown          bool
	count2Target       time.Time
	countdown2Duration time.Duration
	countdown2         bool
	Hours              string
	Minutes            string
	Seconds            string
	Tally              string
	TallyRed           uint8
	TallyGreen         uint8
	TallyBlue          uint8
	Leds               int
	Dots               bool
	flashLeds          bool
	flasher            *time.Ticker
	clockServer        *Server
	oscServer          osc.Server
	timeout            time.Duration // Timeout for osc tally events
	oscTally           bool          // Tally text was from osc event
}

type EngineOptions struct {
	Flash      int    `long:"flash" description:"Flashing interval when countdown reached zero (ms)" default:"500"`
	Timezone   string `short:"t" long:"local-time" description:"Local timezone" default:"Europe/Helsinki"`
	ListenAddr string `long:"osc-listen" description:"Address to listen for incoming osc messages" default:"0.0.0.0:1245"`
	Timeout    int    `short:"d" long:"timeout" description:"Timeout for OSC message updates in milliseconds" default:"1000"`
}

// Create a clock engine
func MakeEngine(options *EngineOptions) (*Engine, error) {
	var engine = Engine{
		Mode:       Normal,
		Hours:      "",
		Minutes:    "",
		Seconds:    "",
		Leds:       0,
		flashLeds:  true,
		Dots:       true,
		oscTally:   false,
		countdown:  false,
		countdown2: false,
		timeout:    time.Duration(options.Timeout) * time.Millisecond,
	}

	// Setup the OSC listener
	engine.oscServer = osc.Server{
		Addr: options.ListenAddr,
	}

	engine.clockServer = MakeServer(&engine.oscServer)
	log.Printf("osc server: listen %v", engine.oscServer.Addr)

	go engine.runOSC()

	// Timezones
	tz, err := time.LoadLocation(options.Timezone)
	if err != nil {
		return nil, err
	}
	engine.Timezone = tz
	engine.flasher = time.NewTicker(time.Duration(options.Flash) * time.Millisecond)

	// Led flash cycle
	go engine.flash()

	// OSC listen
	go engine.listen()

	return &engine, nil
}

func (engine *Engine) runOSC() {
	err := engine.oscServer.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

// Listen for OSC messages
func (engine *Engine) listen() {
	oscChan := engine.clockServer.Listen()
	tallyTimer := timer.NewTimer(engine.timeout)

	for {
		select {
		case message := <-oscChan:
			// New OSC message received
			fmt.Printf("Got new osc data.\n")
			switch message.Type {
			case "count":
				msg := message.CountMessage
				engine.TallyRed = uint8(msg.ColorRed)
				engine.TallyGreen = uint8(msg.ColorGreen)
				engine.TallyBlue = uint8(msg.ColorBlue)
				engine.Tally = fmt.Sprintf("%1s%02d%1s", msg.Symbol, msg.Count, msg.Unit)
				engine.oscTally = true
				tallyTimer.Reset(engine.timeout)
			case "countdownStart":
				msg := message.CountdownMessage
				engine.StartCountdown(time.Duration(msg.Seconds) * time.Second)
			case "countdownStart2":
				msg := message.CountdownMessage
				engine.StartCountdown2(time.Duration(msg.Seconds) * time.Second)
			case "countdownModify":
				msg := message.CountdownMessage
				engine.ModifyCountdown(time.Duration(msg.Seconds) * time.Second)
			case "countdownModify2":
				msg := message.CountdownMessage
				engine.ModifyCountdown2(time.Duration(msg.Seconds) * time.Second)
			case "countdownStop":
				engine.StopCountdown()
			case "countdownStop2":
				engine.StopCountdown2()
			case "countup":
				engine.StartCountup()
			case "kill":
				engine.Kill()
			case "normal":
				engine.Normal()
			}
		case <-tallyTimer.C:
			// OSC message timeout
			engine.Tally = ""
			engine.oscTally = false
		}
	}
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
	engine.countdown = true
	engine.countTarget = time.Now().Add(timer)
	engine.countdownDuration = timer
}

// Start a countdown timer
func (engine *Engine) StartCountdown2(timer time.Duration) {
	engine.Mode = Countdown
	engine.countdown2 = true
	engine.count2Target = time.Now().Add(timer)
	engine.countdown2Duration = timer
}

func (engine *Engine) StartCountup() {
	engine.Mode = Countup
	engine.countTarget = time.Now()
}

func (engine *Engine) ModifyCountdown(delta time.Duration) {
	if engine.Mode == Countdown && engine.countdown {
		engine.countTarget = engine.countTarget.Add(delta)
		engine.countdownDuration += delta
	}
}

func (engine *Engine) ModifyCountdown2(delta time.Duration) {
	if engine.Mode == Countdown && engine.countdown2 {
		engine.count2Target = engine.count2Target.Add(delta)
		engine.countdown2Duration += delta
	}
}

func (engine *Engine) Normal() {
	engine.Mode = Normal
	engine.countdown = false
	engine.countdown2 = false
}

// Update Hours, Minutes and Seconds
func (engine *Engine) Update() {
	if engine.countdown2 && !engine.oscTally {
		engine.Tally = ""
	}
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
		engine.formatCount(diff)
		engine.Leds = display.Minute()
	} else {
		engine.Hours = "00"
		engine.Minutes = "00"
		engine.Leds = 59
	}
}

func (engine *Engine) countdownUpdate() {
	t := time.Now()
	diff := engine.countTarget.Sub(t)
	diff2 := engine.count2Target.Sub(t)
	engine.Seconds = ""
	engine.Dots = true
	engine.normalUpdate() // Baseline

	// Main countdown
	if engine.countdown {
		engine.Seconds = ""
		if t.Before(engine.countTarget) {
			engine.formatCount(diff)
			progress := (float64(diff) / float64(engine.countdownDuration))
			if progress >= 1 {
				progress = 1
			} else if progress < 0 {
				progress = 0
			}
			engine.Leds = int(math.Floor(progress * 60))
		} else {
			engine.Leds = 59
			if engine.flashLeds {
				engine.Hours = "00"
				engine.Minutes = "00"
			} else {
				engine.Hours = ""
				engine.Minutes = ""
			}
		}
	}

	// Secundary countdown
	if engine.countdown2 {
		if t.Before(engine.count2Target) {
			engine.formatCount2(diff2)
		} else {
			if engine.flashLeds {
				engine.Tally = "↓00"
			} else {
				engine.Tally = ""
			}
		}
	}
}

func (engine *Engine) formatCount2(diff time.Duration) {
	// osc tally messages take priority
	if !engine.oscTally {
		secs := int64(diff.Round(time.Second).Seconds())

		for _, unit := range clockUnits {
			if secs/int64(unit.seconds) >= 100 {
				continue
			}
			engine.TallyRed = 255
			engine.TallyGreen = 0
			engine.TallyBlue = 0
			count := secs / int64(unit.seconds)
			engine.Tally = fmt.Sprintf("↓%02d%1s", count, unit.unit)
			return
		}
	}
}

func (engine *Engine) formatCount(diff time.Duration) {
	hours := int32(diff.Truncate(time.Hour).Hours())
	minutes := int32(diff.Truncate(time.Minute).Minutes()) - (hours * 60)
	secs := int32(diff.Truncate(time.Second).Seconds()) - (((hours * 60) + minutes) * 60)

	if hours > 99 {
		engine.Hours = "++"
		engine.Minutes = "++"
	} else if (hours*60 + minutes) > 99 {
		engine.Hours = fmt.Sprintf("%02d", hours)
		engine.Minutes = fmt.Sprintf("%02d", minutes)
	} else {
		engine.Hours = fmt.Sprintf("%02d", (hours*60)+minutes)
		engine.Minutes = fmt.Sprintf("%0d", secs)
	}
}

func (engine *Engine) StopCountdown() {
	engine.countdown = false
	if !engine.countdown2 {
		engine.Mode = Normal
	}
}

func (engine *Engine) StopCountdown2() {
	engine.countdown2 = false
	if !engine.countdown {
		engine.Mode = Normal
	}
}

func (engine *Engine) normalUpdate() {
	t := time.Now().In(engine.Timezone)
	engine.Dots = true

	if !engine.oscTally && !engine.countdown2 {
		// Clear the countdown display on stop
		engine.Tally = ""
	}

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
