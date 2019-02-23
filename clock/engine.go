package clock

import (
	"fmt"
	"github.com/desertbit/timer"
	"github.com/hypebeast/go-osc/osc"
	"log"
	"math"
	"net"
	"time"
)

// Version is the current clock engine version
const Version = "3.0.0"

// Will get overridden by ldflags in Makefile
var gitCommit string
var gitTag string

// EngineOptions contains all common options for clock.Engines
type EngineOptions struct {
	Flash          int    `long:"flash" description:"Flashing interval when countdown reached zero (ms), 0 disables" default:"500"`
	Timezone       string `short:"t" long:"local-time" description:"Local timezone" default:"Europe/Helsinki"`
	ListenAddr     string `long:"osc-listen" description:"Address to listen for incoming osc messages" default:"0.0.0.0:1245"`
	Timeout        int    `short:"d" long:"timeout" description:"Timeout for OSC message updates in milliseconds" default:"1000"`
	Connect        string `short:"o" long:"osc-dest" description:"Address to send OSC feedback to" default:"255.255.255.255:1245"`
	CountdownRed   uint8  `long:"cd-red" description:"Red component of secondary countdown color" default:"255"`
	CountdownGreen uint8  `long:"cd-green" description:"Green component of secondary countdown color" default:"0"`
	CountdownBlue  uint8  `long:"cd-blue" description:"Blue component of secondary countdown color" default:"0"`
}

// Clock engine state constants
const (
	Normal    = iota // Display current time
	Countdown = iota // Display countdown timer only
	Countup   = iota // Count time up
	Off       = iota // (Mostly) blank screen
	Paused    = iota // Paused countdown timer(s)
)

type countdownData struct {
	target   time.Time     // Target timestamp for main countdown
	duration time.Duration // Total duration of main countdown, used to scale the leds
	left     time.Duration // Duration left when paused
	active   bool
}

// Engine contains the state machine for clock-8001
type Engine struct {
	timeZone    *time.Location // Time zone, initialized from options
	mode        int            // Main display mode
	countdown   *countdownData
	countdown2  *countdownData
	paused      bool
	Hours       string
	Minutes     string
	Seconds     string
	Tally       string
	TallyRed    uint8
	TallyGreen  uint8
	TallyBlue   uint8
	cd2Red      uint8
	cd2Green    uint8
	cd2Blue     uint8
	Leds        int
	Dots        bool
	flashLeds   bool
	flasher     *time.Ticker
	clockServer *Server
	oscServer   osc.Server
	timeout     time.Duration // Timeout for osc tally events
	oscTally    bool          // Tally text was from osc event
	oscConn     *net.UDPConn  // Connection for sending feedback
}

// MakeEngine creates a clock engine
func MakeEngine(options *EngineOptions) (*Engine, error) {
	var engine = Engine{
		mode:      Normal,
		Hours:     "",
		Minutes:   "",
		Seconds:   "",
		Leds:      0,
		flashLeds: true,
		Dots:      true,
		oscTally:  false,
		paused:    false,
		timeout:   time.Duration(options.Timeout) * time.Millisecond,
		cd2Red:    options.CountdownRed,
		cd2Green:  options.CountdownGreen,
		cd2Blue:   options.CountdownBlue,
	}

	log.Printf("Clock-8001 engine version %s git: %s\n", gitTag, gitCommit)

	countdown := countdownData{
		active: false,
	}
	engine.countdown = &countdown
	countdown2 := countdownData{
		active: false,
	}
	engine.countdown2 = &countdown2

	// Setup the OSC listener
	engine.oscServer = osc.Server{
		Addr: options.ListenAddr,
	}
	engine.clockServer = MakeServer(&engine.oscServer)
	log.Printf("osc server: listen %v", engine.oscServer.Addr)
	go engine.runOSC()

	// Time zones
	tz, err := time.LoadLocation(options.Timezone)
	if err != nil {
		return nil, err
	}
	engine.timeZone = tz

	// OSC feedback
	if udpAddr, err := net.ResolveUDPAddr("udp", options.Connect); err != nil {
		log.Printf("Failed to resolve OSC feedback address: %v", err)
	} else if udpConn, err := net.DialUDP("udp", nil, udpAddr); err != nil {
		log.Printf("Failed to open OSC feedback address: %v", err)
	} else {
		engine.oscConn = udpConn
	}

	// Led flash cycle
	// Setting the interval to 0 disables
	if options.Flash > 0 {
		engine.flasher = time.NewTicker(time.Duration(options.Flash) * time.Millisecond)
		go engine.flash()
	}

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
			log.Printf("Got new osc data.\n")
			switch message.Type {
			case "count":
				msg := message.CountMessage
				engine.TallyRed = uint8(msg.ColorRed)
				engine.TallyGreen = uint8(msg.ColorGreen)
				engine.TallyBlue = uint8(msg.ColorBlue)
				engine.Tally = fmt.Sprintf("%.1s%02d%.1s", msg.Symbol, msg.Count, msg.Unit)
				engine.oscTally = true
				tallyTimer.Reset(engine.timeout)
			case "display":
				msg := message.DisplayMessage
				engine.TallyRed = uint8(msg.ColorRed)
				engine.TallyGreen = uint8(msg.ColorGreen)
				engine.TallyBlue = uint8(msg.ColorBlue)
				engine.Tally = fmt.Sprintf("%-.4s", msg.Text)
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
			case "pause":
				engine.Pause()
			case "resume":
				engine.Resume()
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

// Send the clock state as /clock/state
func (engine *Engine) sendState() error {
	if engine.oscConn == nil {
		// No osc connection
		return nil
	}
	hours := engine.Hours
	minutes := engine.Minutes
	seconds := engine.Seconds

	if seconds == "" {
		hours = ""
		minutes = engine.Hours
		seconds = engine.Minutes
	}
	mode := engine.mode
	pause := int32(0)

	if engine.paused {
		pause = 1
	}

	packet := osc.NewMessage("/clock/state", int32(mode), hours, minutes, seconds, engine.Tally, pause)

	if data, err := packet.MarshalBinary(); err != nil {
		return err
	} else if _, err := engine.oscConn.Write(data); err != nil {
		return err
	} else {
		return nil
	}
}

func (engine *Engine) flash() {
	for range engine.flasher.C {
		engine.flashLeds = !engine.flashLeds
	}
}

/*
 * Display update requests
 */

// Update Hours, Minutes and Seconds
func (engine *Engine) Update() {
	switch engine.mode {
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

	engine.countdown2Update() // Update secondary countdown if needed

	if err := engine.sendState(); err != nil {
		log.Printf("Error sending osc state: %v", err)
	}
}

func (engine *Engine) normalUpdate() {
	t := time.Now().In(engine.timeZone)
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

func (engine *Engine) countupUpdate() {
	t := time.Now()
	diff := t.Sub(engine.countdown.target)
	display := time.Time{}.Add(diff)
	engine.Seconds = ""
	engine.Dots = true

	if diff > 0 {
		engine.formatCount(diff)
		engine.Leds = display.Minute()
	} else {
		engine.Hours = "00"
		engine.Minutes = "00"
		engine.Leds = 59
	}
}

func (engine *Engine) countdownUpdate() {
	var diff time.Duration
	if engine.paused {
		diff = engine.countdown.left
		if engine.flashLeds {
			engine.Dots = true
		} else {
			engine.Dots = false
		}
	} else {
		t := time.Now()
		diff = engine.countdown.target.Sub(t)
		engine.Dots = true
	}
	engine.Seconds = ""

	// Main countdown
	if diff > 0 {
		engine.formatCount(diff)
		progress := (float64(diff) / float64(engine.countdown.duration))
		if progress >= 1 {
			progress = 1
		} else if progress < 0 {
			progress = 0
		}
		engine.Leds = int(math.Floor(progress * 59))
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

// Secondary countdown, lower priority than Tally messages
func (engine *Engine) countdown2Update() {
	var diff2 time.Duration
	if engine.paused {
		diff2 = engine.countdown2.left
	} else {
		t := time.Now()
		diff2 = engine.countdown2.target.Sub(t).Truncate(time.Second)
	}
	if !engine.oscTally && !engine.countdown2.active {
		// Clear the countdown display on stop
		engine.Tally = ""
	} else if !engine.oscTally && engine.countdown2.active {
		if diff2 > 0 {
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

func (engine *Engine) formatCount(diff time.Duration) {
	hours := int32(diff.Truncate(time.Hour).Hours())
	minutes := int32(diff.Truncate(time.Minute).Minutes()) - (hours * 60)
	secs := int32(diff.Truncate(time.Second).Seconds()) - (((hours * 60) + minutes) * 60)

	if hours > 99 {
		engine.Hours = "++"
		engine.Minutes = "++"
	} else if hours != 0 {
		engine.Hours = fmt.Sprintf("%02d", hours)
		engine.Minutes = fmt.Sprintf("%02d", minutes)
		engine.Seconds = fmt.Sprintf("%02d", secs)
	} else {
		engine.Hours = fmt.Sprintf("%02d", minutes)
		engine.Minutes = fmt.Sprintf("%02d", secs)
	}
}

func (engine *Engine) formatCount2(diff time.Duration) {
	if !engine.oscTally {
		// osc tally messages take priority
		secs := int64(diff.Truncate(time.Second).Seconds())

		for _, unit := range clockUnits {
			if secs/int64(unit.seconds) >= 100 {
				continue
			}
			engine.TallyRed = engine.cd2Red
			engine.TallyGreen = engine.cd2Green
			engine.TallyBlue = engine.cd2Blue
			count := secs / int64(unit.seconds)
			if engine.paused {
				engine.Tally = fmt.Sprintf(" %02d%1s", count, unit.unit)
			} else {
				engine.Tally = fmt.Sprintf("↓%02d%1s", count, unit.unit)
			}
			return
		}
	}
}

/*
 * OSC Message handlers
 */

// StartCountdown starts a primary countdown timer
func (engine *Engine) StartCountdown(timer time.Duration) {
	cd := countdownData{
		target:   time.Now().Add(timer).Truncate(time.Second),
		duration: timer,
		left:     timer,
		active:   true,
	}
	engine.countdown = &cd
	engine.mode = Countdown
}

// StartCountdown2 starts a secondary countdown timer
func (engine *Engine) StartCountdown2(timer time.Duration) {
	cd := countdownData{
		target:   time.Now().Add(timer).Truncate(time.Second),
		duration: timer,
		left:     timer,
		active:   true,
	}
	engine.countdown2 = &cd
}

// StartCountup starts counting time up from this moment
func (engine *Engine) StartCountup() {
	cd := countdownData{
		target: time.Now().Truncate(time.Second),
		active: true,
	}
	engine.countdown = &cd
	engine.mode = Countup
}

// Normal returns main display to normal clock
func (engine *Engine) Normal() {
	engine.mode = Normal
	// engine.countdown2.active = false
}

// ModifyCountdown adds or removes time from primary countdown
func (engine *Engine) ModifyCountdown(delta time.Duration) {
	if engine.mode == Countdown {
		cd := countdownData{
			target:   engine.countdown.target.Add(delta),
			duration: engine.countdown.duration + delta,
			left:     engine.countdown.left + delta,
		}
		engine.countdown = &cd
	}
}

// ModifyCountdown2 adds or removes time from primary countdown
func (engine *Engine) ModifyCountdown2(delta time.Duration) {
	if engine.countdown2.active {
		cd := countdownData{
			target:   engine.countdown2.target.Add(delta),
			duration: engine.countdown2.duration + delta,
			left:     engine.countdown2.left + delta,
			active:   engine.countdown2.active,
		}
		engine.countdown2 = &cd
	}
}

// StopCountdown stops the primary countdown
func (engine *Engine) StopCountdown() {
	if engine.mode == Countdown {
		cd := countdownData{
			target:   time.Now(),
			duration: time.Millisecond,
			left:     time.Millisecond,
			active:   false,
		}
		engine.countdown = &cd
		engine.mode = Off
	}
}

// StopCountdown2 stops the secondary countdown
func (engine *Engine) StopCountdown2() {
	cd := countdownData{
		target:   time.Now(),
		duration: time.Millisecond,
		left:     time.Millisecond,
		active:   false,
	}
	engine.countdown2 = &cd
}

// Kill blanks the clock display
func (engine *Engine) Kill() {
	engine.mode = Off
	engine.countdown2.active = false
}

// Pause pauses both countdowns
func (engine *Engine) Pause() {
	if !engine.paused {
		t := time.Now()
		engine.countdown.left = engine.countdown.target.Sub(t).Truncate(time.Second)
		engine.countdown2.left = engine.countdown2.target.Sub(t).Truncate(time.Second)
		engine.paused = true
	}
}

// Resume resumes both countdowns if they have been paused
func (engine *Engine) Resume() {
	if engine.paused {
		t := time.Now()
		engine.countdown.target = t.Add(engine.countdown.left).Truncate(time.Second)
		engine.countdown2.target = t.Add(engine.countdown2.left).Truncate(time.Second)
		engine.paused = false
	}
}
