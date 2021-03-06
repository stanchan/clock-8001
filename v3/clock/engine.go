package clock

import (
	"fmt"
	"github.com/desertbit/timer"
	"github.com/hypebeast/go-osc/osc"
	"github.com/stanchan/clock-8001/v3/debug"
	"image/color"
	"log"
	"math"
	"os"
	"os/exec"
	"regexp"
	db "runtime/debug"
	"strconv"
	"strings"
	"time"
)

// Version is the current clock engine version
const Version = "3.16.3"

// Will get overridden by ldflags in Makefile
var gitCommit = "Unknown"
var gitTag = "v3.16.3"

// EngineOptions contains all common options for clock.Engines
type EngineOptions struct {
	Flash           int    `long:"flash" description:"Flashing interval when countdown reached zero (ms), 0 disables" default:"500"`
	Timezone        string `short:"t" long:"local-time" description:"Local timezone" default:"Europe/Helsinki"`
	ListenAddr      string `long:"osc-listen" description:"Address to listen for incoming osc messages" default:"0.0.0.0:1245"`
	Timeout         int    `short:"d" long:"timeout" description:"Timeout for OSC message updates in milliseconds" default:"1000"`
	Connect         string `short:"o" long:"osc-dest" description:"Address to send OSC feedback to" default:"255.255.255.255:1245"`
	CountdownRed    uint8  `long:"cd-red" description:"Red component of secondary countdown color" default:"255"`
	CountdownGreen  uint8  `long:"cd-green" description:"Green component of secondary countdown color" default:"0"`
	CountdownBlue   uint8  `long:"cd-blue" description:"Blue component of secondary countdown color" default:"0"`
	DisableOSC      bool   `long:"disable-osc" description:"Disable OSC control and feedback"`
	DisableFeedback bool   `long:"disable-feedback" description:"Disable OSC feedback"`
	DisableLTC      bool   `long:"disable-ltc" description:"Disable LTC display mode"`
	LTCSeconds      bool   `long:"ltc-seconds" description:"Show seconds on the ring in LTC mode"`
	LTCFollow       bool   `long:"ltc-follow" description:"Continue on internal clock if LTC signal is lost. If unset display will blank when signal is gone."`
	Format12h       bool   `long:"format-12h" description:"Use 12 hour format for time-of-day display"`
}

// Clock engine state constants
const (
	Normal    = iota // Display current time
	Countdown = iota // Display countdown timer only
	Countup   = iota // Count time up
	Off       = iota // (Mostly) blank screen
	Paused    = iota // Paused countdown timer(s)
	LTC       = iota // LTC display
)

// Misc constants
const (
	numCounters      = 2 // Number of distinct counters to initialize
	numSources       = 2
	PrimaryCounter   = 0 // Main counter that replaces the ToD display on the round clock when active
	SecondaryCounter = 1 // Secondary counter that is displayed in the tally message space on the round clock
)

type ltcData struct {
	hours   int
	minutes int
	seconds int
	frames  int
	target  time.Time
	timeout bool
}

// Engine contains the state machine for clock-8001
type Engine struct {
	timeZone       *time.Location // Time zone, initialized from options
	mode           int            // Main display mode
	Counters       []*Counter     // Timer counters
	sources        []*source      // Time sources for 1-3 displays
	paused         bool
	Hours          string
	Minutes        string
	Seconds        string
	displaySeconds bool
	Tally          string // 4 character "tally" field, contains either the message from OSC or the secondary timer
	TallyRed       uint8  // Red component for the OSC tally message color 0-255
	TallyGreen     uint8  // Green component for the OSC tally message color 0-255
	TallyBlue      uint8  // Blue component for the OSC tally message color 0-255
	cd2Red         uint8
	cd2Green       uint8
	cd2Blue        uint8
	Leds           int
	Dots           bool
	flashLeds      bool
	flasher        *time.Ticker
	clockServer    *Server
	oscServer      osc.Server
	timeout        time.Duration        // Timeout for osc tally events
	oscTally       bool                 // Tally text was from osc event
	tallyFull      string               // Full tally message as received from OSC
	oscDests       *feedbackDestination // udp connections to send osc feedback to
	initialized    bool                 // Show version on startup until ntp synced or receiving OSC control
	ltc            *ltcData             // LTC time code status
	ltcShowSeconds bool                 // Toggles led display on LTC mode between seconds and frames
	ltcFollow      bool                 // Continue on internal timer if LTC signal is lost
	ltcEnabled     bool                 // Toggle LTC mode on or off
	ltcTimeout     bool                 // Set to true if LTC signal is lost by the ltc timer
	ltcActive      bool                 // Do we have a active LTC to display?
	udpActive      bool                 // Do we have a active UDP time to display?
	DualText       string               // Dual clock mode text message, 8 characters
	format12h      bool                 // Use 12 hour format for time-of-day
}

// Clock contains the state of a single component clock / timer
type Clock struct {
	Text     string  // Normal clock representation
	Compact  string  // 4 character condensed output
	Expired  bool    // true if asscociated timer is expired
	Mode     int     // Display type
	Paused   bool    // Is the clock/timer paused?
	Progress float64 // Progress of the total timer 0-1
}

// State is a snapshot of the clock representation on the time State() was called
type State struct {
	Initialized bool        // Does the clock have valid time or has it received an osc command?
	Clocks      []*Clock    // All configured clocks / timers
	Tally       string      // Tally message text
	TallyColor  *color.RGBA // Tally message color
}

// MakeEngine creates a clock engine
func MakeEngine(options *EngineOptions) (*Engine, error) {
	var engine = Engine{
		mode:           Normal,
		Hours:          "",
		Minutes:        "",
		Seconds:        "",
		displaySeconds: true,
		Leds:           0,
		flashLeds:      true,
		Dots:           true,
		oscTally:       false,
		paused:         false,
		timeout:        time.Duration(options.Timeout) * time.Millisecond,
		cd2Red:         options.CountdownRed,
		cd2Green:       options.CountdownGreen,
		cd2Blue:        options.CountdownBlue,
		initialized:    false,
		oscDests:       nil,
		DualText:       "",
		ltcShowSeconds: options.LTCSeconds,
		ltcFollow:      options.LTCFollow,
		ltcEnabled:     !options.DisableLTC,
		ltcActive:      false,
		udpActive:      false,
		format12h:      options.Format12h,
	}

	// Time zones
	tz, err := time.LoadLocation(options.Timezone)
	if err != nil {
		return nil, err
	}
	engine.timeZone = tz

	ltc := ltcData{hours: 0}
	engine.ltc = &ltc

	engine.printVersion()
	engine.initCounters()
	engine.initSources()
	engine.initOSC(options)

	// Led flash cycle
	// Setting the interval to 0 disables
	if options.Flash > 0 {
		engine.flasher = time.NewTicker(time.Duration(options.Flash) * time.Millisecond)
		go engine.flash()
	}

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
	tallyTimer.Stop()
	ltcTimer := timer.NewTimer(engine.timeout)
	ltcTimer.Stop() // Needed to prevent a timeout at the start

	for {
		select {
		case message := <-oscChan:
			// New OSC message received
			debug.Printf("Got new osc data: %v\n", message)
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
				engine.tallyFull = msg.Text
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
			case "countupModify":
				msg := message.CountdownMessage
				engine.ModifyCountup(time.Duration(msg.Seconds) * time.Second)
			case "kill":
				engine.Kill()
			case "normal":
				engine.Normal()
			case "secondsOff":
				engine.displaySeconds = false
			case "secondsOn":
				engine.displaySeconds = true
			case "setTime":
				engine.setTime(message.Data)
			case "LTC":
				if engine.ltcEnabled {
					engine.setLTC(message.Data)
					ltcTimer.Reset(engine.timeout)
				}
			case "dualText":
				engine.DualText = fmt.Sprintf("%-.8s", message.Data)
			}
			// We have received a osc command, so stop the version display
			engine.initialized = true
		case <-tallyTimer.C:
			// OSC message timeout
			engine.Tally = ""
			engine.oscTally = false
		case <-ltcTimer.C:
			// LTC message timeout
			engine.ltcTimeout = true
		}
	}
}

// Send the clock state as /clock/state
func (engine *Engine) sendState() error {
	if engine.oscDests == nil {
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

	data, err := packet.MarshalBinary()
	if err != nil {
		return err
	}
	engine.oscDests.Write(data)
	return nil
}

func (engine *Engine) flash() {
	for range engine.flasher.C {
		engine.flashLeds = !engine.flashLeds
	}
}

// State creates a snapshot of the clock state for display on clock faces
func (engine *Engine) State() *State {
	t := time.Now()
	var clocks []*Clock
	for _, s := range engine.sources {
		if s.ltc && engine.ltcActive {
			c := Clock{
				Expired: engine.ltcTimeout,
				Compact: "",
				Mode:    LTC,
			}
			ltc := engine.ltc
			if !engine.ltcTimeout {
				// We have LTC time, so display it
				// engine.initialized = true
				c.Text = fmt.Sprintf("%02d:%02d:%02d:%02d", ltc.hours, ltc.minutes, ltc.seconds, ltc.frames)
			} else if engine.ltcFollow {
				// Follow the LTC time when signal is lost
				// Todo: must be easier way to print out the duration...
				t := time.Now()
				diff := t.Sub(engine.ltc.target)
				c.Text = fmt.Sprintf("%s:%02d", formatDuration(diff), 0)
			} else {
				// Timeout without follow mode
				c.Text = ""
			}
			clocks = append(clocks, &c)
		} else if s.udp && engine.udpActive {
			// UDP time reception

		} else if s.timer && s.counter.active {
			// Active timer
			out := s.counter.Output(t)
			c := Clock{
				Text:     out.Text,
				Compact:  out.Compact,
				Expired:  out.Expired,
				Paused:   out.Paused,
				Progress: out.Progress,
			}
			if out.Countdown {
				c.Mode = Countdown
			} else {
				c.Mode = Countup
			}
			clocks = append(clocks, &c)
		} else if s.tod {
			// Time of day
			c := Clock{
				Mode:    Normal,
				Compact: "",
				Expired: false,
			}
			if engine.format12h {
				c.Text = t.In(s.tz).Format("03:04:05")
			} else {
				c.Text = t.In(s.tz).Format("15:04:05")
			}
			clocks = append(clocks, &c)
		} else {
			// Nothing to display
			c := Clock{
				Text:    "",
				Compact: "",
				Mode:    Off,
			}
			clocks = append(clocks, &c)
		}
	}
	state := State{
		Initialized: engine.initialized,
		Clocks:      clocks,
	}

	if engine.oscTally {
		state.Tally = engine.tallyFull
		state.TallyColor = &color.RGBA{
			R: engine.TallyRed,
			G: engine.TallyGreen,
			B: engine.TallyBlue,
			A: 255,
		}
	}

	return &state
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
		engine.primaryCounterUpdate()
	case Countup:
		engine.primaryCounterUpdate()
	case LTC:
		engine.ltcUpdate()
	case Off:
		engine.Hours = ""
		engine.Minutes = ""
		engine.Seconds = ""
		engine.Leds = 0
		engine.Dots = false
	}

	engine.countdown2Update() // Update secondary countdown if needed

	if !engine.displaySeconds {
		// Clear the seconds field as requested
		engine.Seconds = ""
	}

	if err := engine.sendState(); err != nil {
		log.Printf("Error sending osc state: %v", err)
	}
}

func (engine *Engine) normalUpdate() {
	t := time.Now().In(engine.timeZone)
	engine.Dots = true

	// Check that the rpi has valid time
	if t.Year() > 2000 {
		// We have ntp synced time, so display it
		engine.initialized = true

		// Optional 12 hour format
		if engine.format12h {
			engine.Hours = t.Format("03")
		} else {
			engine.Hours = t.Format("15")
		}
		engine.Minutes = t.Format("04")
		engine.Seconds = t.Format("05")
		engine.Leds = t.Second()
	} else if !engine.initialized {
		// No valid time, display version number instead of Time of day
		var major, minor, bugfix int
		_, err := fmt.Sscanf(gitTag, "v%d.%d.%d", &major, &minor, &bugfix)
		if err != nil {
			panic(err)
		}
		engine.Tally = "ver"
		engine.TallyRed = 255
		engine.TallyGreen = 255
		engine.TallyBlue = 255
		engine.Hours = fmt.Sprintf("%2d", major)
		engine.Minutes = fmt.Sprintf("%2d", minor)
		engine.Seconds = fmt.Sprintf("%2d", bugfix)
		engine.Dots = false
		engine.Leds = t.Second()
	} else {
		// We have received a osc command but no valid time, black the fields
		engine.Hours = ""
		engine.Minutes = ""
		engine.Seconds = ""
		engine.Dots = false
		engine.Leds = 0
	}
}

func (engine *Engine) ltcUpdate() {
	engine.Dots = true
	if !engine.ltcTimeout {
		// We have LTC time, so display it
		engine.initialized = true
		engine.Tally = fmt.Sprintf(" %02d", engine.ltc.hours)
		engine.Hours = fmt.Sprintf("%02d", engine.ltc.minutes)
		engine.Minutes = fmt.Sprintf("%02d", engine.ltc.seconds)
		engine.Seconds = fmt.Sprintf("%02d", engine.ltc.frames)
		if engine.ltcShowSeconds {
			engine.Leds = engine.ltc.seconds
		} else {
			engine.Leds = engine.ltc.frames
		}
	} else if engine.ltcFollow {
		// Follow the LTC time when signal is lost
		t := time.Now()
		diff := t.Sub(engine.ltc.target)
		hours := int32(diff.Truncate(time.Hour).Hours())
		minutes := int32(diff.Truncate(time.Minute).Minutes()) - (hours * 60)
		secs := int32(diff.Truncate(time.Second).Seconds()) - (((hours * 60) + minutes) * 60)

		engine.Tally = fmt.Sprintf(" %02d", hours)
		engine.Hours = fmt.Sprintf("%02d", minutes)
		engine.Minutes = fmt.Sprintf("%02d", secs)
		engine.Seconds = ""
		if engine.ltcShowSeconds {
			engine.Leds = int(secs)
		} else {
			engine.Leds = 0
		}
	} else {
		// Timeout without follow mode
		engine.Tally = ""
		engine.Hours = ""
		engine.Minutes = ""
		engine.Seconds = ""
		engine.Leds = 0
		engine.Dots = false
	}
}

// TODO display ToD if nothing else is active?
func (engine *Engine) primaryCounterUpdate() {
	t := time.Now()
	output := engine.Counters[PrimaryCounter].Output(t)
	engine.Dots = true
	engine.Seconds = ""

	if !output.Expired {
		if output.Hours > 99 {
			engine.Hours = "++"
			engine.Minutes = "++"
		} else if output.Hours != 0 {
			engine.Hours = fmt.Sprintf("%02d", output.Hours)
			engine.Minutes = fmt.Sprintf("%02d", output.Minutes)
			engine.Seconds = fmt.Sprintf("%02d", output.Seconds)
		} else {
			engine.Hours = fmt.Sprintf("%02d", output.Minutes)
			engine.Minutes = fmt.Sprintf("%02d", output.Seconds)
		}

		if output.Countdown {
			engine.Leds = int(math.Floor(output.Progress * 59))
		} else {
			engine.Leds = output.Minutes
		}
	} else {
		if output.Countdown {
			engine.expiredCountdown()
		} else {
			// TODO: wrapping count up?
			engine.Hours = "00"
			engine.Minutes = "00"
			engine.Leds = 59
		}
	}

	// Flash the separator leds if counter is paused
	if output.Paused {
		if engine.flashLeds {
			engine.Dots = true
		} else {
			engine.Dots = false
		}
	}
}

// TODO: different behaviours on expired timers
func (engine *Engine) expiredCountdown() {
	engine.Leds = 0
	if engine.flashLeds {
		engine.Hours = "00"
		engine.Minutes = "00"
	} else {
		engine.Hours = ""
		engine.Minutes = ""

	}
}

// Secondary countdown, lower priority than Tally messages
func (engine *Engine) countdown2Update() {
	if !engine.initialized {
		// No valid time and no osc messages received
		return
	}

	output := engine.Counters[SecondaryCounter].Output(time.Now())
	if !engine.oscTally && !engine.Counters[SecondaryCounter].active {
		// Clear the countdown display on stop
		engine.Tally = ""
	} else if !engine.oscTally && output.Active {
		engine.TallyRed = engine.cd2Red
		engine.TallyGreen = engine.cd2Green
		engine.TallyBlue = engine.cd2Blue
		if !output.Expired {
			engine.Tally = output.Compact
		} else {
			if engine.flashLeds {
				engine.Tally = " 00"
			} else {
				engine.Tally = ""
			}
		}
	}
}

/*
 * OSC Message handlers
 */

// StartCounter starts a counter
func (engine *Engine) StartCounter(counter int, countdown bool, timer time.Duration) {
	if counter < 0 || counter >= numCounters {
		log.Printf("engine.StartCounter: illegal counter number %d (have %d counters)\n", counter, numCounters)
	}

	engine.Counters[counter].Start(countdown, timer)

	// Handle main display mode changes
	if counter == PrimaryCounter {
		if countdown {
			engine.mode = Countdown
		} else {
			engine.mode = Countup
		}
	}
}

// ModifyCounter adds or removes time from a counter
func (engine *Engine) ModifyCounter(counter int, delta time.Duration) {
	if counter < 0 || counter >= numCounters {
		log.Printf("engine.StartCounter: illegal counter number %d (have %d counters)\n", counter, numCounters)
	}

	engine.Counters[counter].Modify(delta)
}

// StopCounter stops a given counter
func (engine *Engine) StopCounter(counter int) {
	if counter < 0 || counter >= numCounters {
		log.Printf("engine.StartCounter: illegal counter number %d (have %d counters)\n", counter, numCounters)
	}

	engine.Counters[counter].Stop()
	if counter == PrimaryCounter {
		engine.mode = Off
	}
}

/* Legacy handlers */

// StartCountdown starts a primary countdown timer
func (engine *Engine) StartCountdown(timer time.Duration) {
	engine.StartCounter(PrimaryCounter, true, timer)
}

// StartCountdown2 starts a secondary countdown timer
func (engine *Engine) StartCountdown2(timer time.Duration) {
	engine.StartCounter(SecondaryCounter, true, timer)
}

// StartCountup starts counting time up from this moment
func (engine *Engine) StartCountup() {
	engine.StartCounter(PrimaryCounter, false, 0*time.Second)
}

// ModifyCountdown adds or removes time from primary countdown
func (engine *Engine) ModifyCountdown(delta time.Duration) {
	engine.ModifyCounter(PrimaryCounter, delta)
}

// ModifyCountdown2 adds or removes time from primary countdown
func (engine *Engine) ModifyCountdown2(delta time.Duration) {
	engine.ModifyCounter(SecondaryCounter, delta)
}

// ModifyCountup adds or removes time from countup timer
func (engine *Engine) ModifyCountup(delta time.Duration) {
	engine.ModifyCounter(PrimaryCounter, delta)
}

// StopCountdown stops the primary countdown
func (engine *Engine) StopCountdown() {
	engine.StopCounter(PrimaryCounter)
}

// StopCountdown2 stops the secondary countdown
func (engine *Engine) StopCountdown2() {
	engine.StopCounter(SecondaryCounter)
}

// Normal returns main display to normal clock
func (engine *Engine) Normal() {
	engine.mode = Normal
	// engine.countdown2.active = false
}

// Kill blanks the clock display
func (engine *Engine) Kill() {
	engine.mode = Off
	for _, c := range engine.Counters {
		c.Stop()
	}
}

// Pause pauses both countdowns
func (engine *Engine) Pause() {
	for _, c := range engine.Counters {
		c.Pause()
	}
	engine.paused = true
	for i, c := range engine.Counters {
		fmt.Printf("Counter %d: %v\n", i, c)
	}
}

// Resume resumes both countdowns if they have been paused
func (engine *Engine) Resume() {
	for _, c := range engine.Counters {
		c.Resume()
	}
	engine.paused = false
}

// DisplaySeconds returns true if the clock should display seconds
func (engine *Engine) DisplaySeconds() bool {
	return engine.displaySeconds
}

// TimeOfDay returns true if the clock is showing time of day, false otherwise
func (engine *Engine) TimeOfDay() bool {
	return engine.mode == Normal
}

func (engine *Engine) setTime(time string) {
	debug.Printf("Set time: %#v", time)
	_, lookErr := exec.LookPath("date")
	if lookErr != nil {
		debug.Printf("Date binary not found, cannot set system date: %s\n", lookErr.Error())
		return
	}
	// Validate the received time
	match, _ := regexp.MatchString("^(2[0-3]|[01][0-9]):([0-5][0-9]):([0-5][0-9])$", time)
	if match {
		// Set the system time
		dateString := fmt.Sprintf("2019-01-01 %s", time)
		tzString := fmt.Sprintf("TZ=%s", engine.timeZone.String())
		debug.Printf("Setting system date to: %s\n", dateString)
		args := []string{"--set", dateString}
		cmd := exec.Command("date", args...) // #nosec we have strict validation in place
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, tzString)
		if err := cmd.Run(); err != nil {
			panic(err)
		}
	} else {
		debug.Printf("Invalid time provided: %v\n", time)
	}
}

// LtcActive returns true if the clock is displaying LTC timecode
func (engine *Engine) LtcActive() bool {
	return engine.mode == LTC
}

func (engine *Engine) setLTC(timestamp string) {
	match, _ := regexp.MatchString("^([0-9][0-9]):([0-5][0-9]):([0-5][0-9]):([0-9][0-9])$", timestamp)
	if match {
		parts := strings.Split(timestamp, ":")
		hours, _ := strconv.Atoi(parts[0])
		minutes, _ := strconv.Atoi(parts[1])
		seconds, _ := strconv.Atoi(parts[2])
		frames, _ := strconv.Atoi(parts[3])
		var ltcTarget time.Time
		if frames == 0 {
			// Update the LTC countdown target
			ltcDuration := time.Duration(hours) * time.Hour
			ltcDuration += time.Duration(minutes) * time.Minute
			ltcDuration += time.Duration(seconds) * time.Second
			ltcTarget = time.Now().Add(-ltcDuration)
		} else {
			ltcTarget = engine.ltc.target
		}
		engine.mode = LTC
		engine.ltcActive = true
		engine.oscTally = true
		engine.ltcTimeout = false
		engine.ltc = &ltcData{
			hours:   hours,
			minutes: minutes,
			seconds: seconds,
			frames:  frames + 1,
			target:  ltcTarget,
		}
	}
}

// printVersion prints to stdout the clock version and dependency versions
func (engine *Engine) printVersion() {
	clockModule, ok := db.ReadBuildInfo()
	if ok {
		for _, mod := range clockModule.Deps {
			log.Printf("Dep: %s: version %s", mod.Path, mod.Version)
			if mod.Path == "github.com/stanchan/clock-8001" {
				gitTag = mod.Version
			}
		}
	} else {
		log.Printf("Error reading BuildInfo, version data unavailable")
	}
	log.Printf("Clock-8001 engine version %s git: %s\n", gitTag, gitCommit)
}

// initCounters initializes the countdown and count up timers
func (engine *Engine) initCounters() {
	engine.Counters = make([]*Counter, numCounters)
	for i := 0; i < numCounters; i++ {
		engine.Counters[i] = &Counter{
			active:    false,
			paused:    false,
			countdown: false,
			state:     &counterState{},
		}
	}
	log.Printf("aff %v", engine.Counters)
}

func (engine *Engine) initSources() {
	// Todo get the map from options...
	engine.sources = make([]*source, numSources)
	engine.sources[0] = &source{
		counter: engine.Counters[0],
		tod:     true,
		timer:   true,
		ltc:     engine.ltcEnabled,
		udp:     true,
		tz:      engine.timeZone,
	}

	engine.sources[1] = &source{
		counter: engine.Counters[1],
		tod:     false,
		timer:   true,
		ltc:     false,
		udp:     false,
		tz:      engine.timeZone,
	}

	log.Printf("fsdf %v", engine.sources)
}

// initOSC Sets up the OSC listener and feedback
func (engine *Engine) initOSC(options *EngineOptions) {
	if !options.DisableOSC {
		engine.oscServer = osc.Server{
			Addr: options.ListenAddr,
		}
		engine.clockServer = MakeServer(&engine.oscServer)
		log.Printf("OSC control: listening on %v", engine.oscServer.Addr)
		go engine.runOSC()

		// process osc commands
		go engine.listen()

		if options.DisableFeedback {
			engine.oscDests = nil
			log.Printf("OSC feedback disabled")
		} else {
			// OSC feedback
			engine.oscDests = initFeedback(options.Connect)
		}
	} else {
		log.Printf("OSC control and feedback disabled.\n")
	}
}

func formatDuration(diff time.Duration) string {
	hours := int32(diff.Truncate(time.Hour).Hours())
	minutes := int32(diff.Truncate(time.Minute).Minutes()) - (hours * 60)
	seconds := int32(diff.Truncate(time.Second).Seconds()) - (((hours * 60) + minutes) * 60)
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}
