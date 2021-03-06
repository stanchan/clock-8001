package clock

import (
	"fmt"
	"github.com/stanchan/clock-8001/v4/debug"
	"image/color"
	"time"
)

/*
 * Counters representing generic countdowns / ups
 */

const (
	autoColorOff   = iota
	autoColorStart = iota
	autoColorWarn  = iota
	autoColorEnd   = iota
)

// Counter abstracts a generic counter counting up or down
type Counter struct {
	state          *counterState
	media          *mediaState
	slave          *slaveState
	active         bool // Is this counter active?
	countdown      bool // Count up / down from the target
	paused         bool // Is the counter paused?
	signalColor    color.RGBA
	autoColorState int
}

type slaveState struct {
	hours     int
	minutes   int
	seconds   int
	icon      string
	hideHours bool
}

type mediaState struct {
	paused    bool
	looping   bool
	hours     int32
	minutes   int32
	seconds   int32
	frames    int32
	progress  float64
	remaining time.Duration
}

type counterState struct {
	target   time.Time     // Target timestamp for main countdown
	duration time.Duration // Total duration of main countdown, used to scale the leds
	left     time.Duration // Duration left when paused
}

// CounterOutput the data structure returned by Counter.Output() and contains the static state of the counter at that time
type CounterOutput struct {
	Active      bool          // True if the counter is active
	Media       bool          // True if counter represents a playing media file
	Countdown   bool          // True if counting down, false if counting up
	Paused      bool          // True if counter has been paused
	Looping     bool          // True if the playing media is looping in the player
	Expired     bool          // Has the countdown timer expired?
	Hours       int           // Hour part of the timer
	Minutes     int           // Minutes of the timer, 0-60
	Seconds     int           // Seconds of the timer, 0-60
	Text        string        // HH:MM:SS string representation
	Icon        string        // Single unicode glyph to use as an icon for the timer
	Compact     string        // Compact 4-character output
	Progress    float64       // Percentage of total time elapsed of the countdown, 0-1
	Diff        time.Duration // raw difference
	HideHours   bool
	SignalColor color.RGBA
}

// Output generates the static output of the counter for use in clock displays
func (counter *Counter) Output(t time.Time) *CounterOutput {
	var out *CounterOutput
	if counter.slave != nil {
		out = counter.slaveOutput()
	} else if counter.media != nil {
		out = counter.mediaOutput()
	} else {
		out = counter.normalOutput(t)
	}
	out.SignalColor = counter.signalColor

	return out
}

func (counter *Counter) mediaOutput() *CounterOutput {
	debug.Printf("Mediaoutput")
	var icon string
	var seconds int64
	m := counter.media

	if m.paused {
		icon = "???"
	} else if m.looping {
		icon = "???"
	} else {
		icon = "???"
	}

	seconds = int64(m.hours) * 60
	seconds = (int64(m.minutes) + seconds) * 60
	seconds = seconds + int64(m.seconds)

	text := fmt.Sprintf("%02d:%02d:%02d", m.hours, m.minutes, m.seconds)
	compact := fmt.Sprintf("%s%s", icon, secsToCompact(seconds))

	out := &CounterOutput{
		Active:   true,
		Media:    true,
		Icon:     icon,
		Paused:   m.paused,
		Looping:  m.looping,
		Hours:    int(m.hours),
		Minutes:  int(m.minutes),
		Seconds:  int(m.seconds),
		Text:     text,
		Compact:  compact,
		Progress: counter.media.progress,
		Diff:     m.remaining,
	}

	return out
}

func (counter *Counter) normalOutput(t time.Time) *CounterOutput {
	if !counter.active {
		return &CounterOutput{Active: false}
	}

	var icon string
	diff := counter.Diff(t)

	hours := int(diff.Truncate(time.Hour).Hours())
	minutes := int(diff.Truncate(time.Minute).Minutes()) - (hours * 60)
	seconds := int(diff.Truncate(time.Second).Seconds()) - (((hours * 60) + minutes) * 60)

	progress := (float64(diff) / float64(counter.state.duration))
	expired := diff.Seconds() < 1

	if expired {
		hours = 0
		minutes = 0
		seconds = 0
		progress = 1
	}

	if progress >= 1 {
		progress = 1
	} else if progress < 0 {
		progress = 0
	}

	if counter.paused {
		icon = "???"
	} else if counter.countdown {
		icon = "???"
	} else {
		icon = "???"
	}

	rawSecs := int64((counter.Diff(t).Truncate(time.Second) + time.Second).Seconds())
	c := secsToCompact(rawSecs)

	out := &CounterOutput{
		Active:    counter.active,
		Countdown: counter.countdown,
		Paused:    counter.paused,
		Expired:   expired,
		Hours:     hours,
		Minutes:   minutes,
		Seconds:   seconds,
		Text:      fmt.Sprintf("%02d:%02d:%02d", abs(hours), abs(minutes), abs(seconds)),
		Compact:   fmt.Sprintf("%s%s", icon, c),
		Icon:      icon,
		Progress:  progress,
		Diff:      diff,
	}

	return out
}

func (counter *Counter) slaveOutput() *CounterOutput {
	hours := counter.slave.hours
	minutes := counter.slave.minutes
	seconds := counter.slave.seconds

	text := fmt.Sprintf("%02d:%02d:%02d", hours, abs(minutes), abs(seconds))
	if counter.slave.hideHours {
		text = text[3:8]
	}

	out := &CounterOutput{
		Active:    true,
		Countdown: true,
		Paused:    false,
		Expired:   false,
		Hours:     hours,
		Minutes:   minutes,
		Seconds:   seconds,
		Text:      text,
		Compact:   "",
		Icon:      counter.slave.icon,
		Progress:  0,
		Diff:      0,
		HideHours: counter.slave.hideHours,
	}

	return out
}

func secsToCompact(rawSecs int64) string {
	for _, unit := range clockUnits {
		if rawSecs/int64(unit.seconds) >= 100 {
			continue
		}
		count := rawSecs / int64(unit.seconds)
		return fmt.Sprintf("%02d%1s", count, unit.unit)
	}
	return "+++"
}

// Start begins counting time up or down
func (counter *Counter) Start(countdown bool, timer time.Duration) {
	s := counterState{
		target:   time.Now().Add(timer).Truncate(time.Second),
		duration: timer,
		left:     timer,
	}
	counter.state = &s

	counter.countdown = countdown

	t := time.Now()

	if counter.countdown {
		counter.state.left = counter.state.target.Sub(t).Truncate(time.Second)
	} else {
		counter.state.left = t.Sub(counter.state.target).Truncate(time.Second)
	}

	counter.active = true
}

// Target sets the target date and time for a counter
func (counter *Counter) Target(target time.Time) {
	timer := target.Sub(time.Now())
	fmt.Printf("target: %v timer: %v\n", target, timer)
	if timer < 0 {
		counter.Start(false, timer)
	} else {
		counter.Start(true, timer)
	}
}

// SetSlave sets the counter state as a slave from external source
func (counter *Counter) SetSlave(hours, minutes, seconds int, hideHours bool, icon string) {
	s := &slaveState{
		hours:     hours,
		minutes:   minutes,
		seconds:   seconds,
		hideHours: hideHours,
		icon:      icon,
	}
	counter.slave = s
	counter.active = true
}

// ResetSlave removes slave state from external source
func (counter *Counter) ResetSlave() {
	counter.slave = nil
}

// SetMedia sets the counter state from a playing media file
func (counter *Counter) SetMedia(hours, minutes, seconds, frames int32, remaining time.Duration, progress float64, paused bool, looping bool) {
	// FIXME: .truncate(time.Second) and mitti timers cause blinking on second changes!
	m := mediaState{
		hours:     hours,
		minutes:   minutes,
		seconds:   seconds,
		frames:    frames,
		paused:    paused,
		looping:   looping,
		progress:  progress,
		remaining: remaining,
	}
	counter.media = &m
	counter.active = true
}

// ResetMedia removes the media state from a counter
func (counter *Counter) ResetMedia() {
	if counter.media != nil {
		counter.active = false
		counter.media = nil
	}
}

// Modify alters the counter target on a running counter
func (counter *Counter) Modify(delta time.Duration) {
	if !counter.active {
		return
	}
	if !counter.countdown {
		// Invert delta if counting up
		delta = -delta
	}

	s := counterState{
		target:   counter.state.target.Add(delta),
		duration: counter.state.duration + delta,
		left:     counter.state.left + delta,
	}
	counter.state = &s
}

// Stop stops and deactivates the counter
func (counter *Counter) Stop() {
	counter.active = false
	counter.paused = false

	s := counterState{
		target:   time.Now(),
		duration: time.Millisecond,
		left:     time.Millisecond,
	}
	counter.state = &s
}

// Pause pauses a running counter
func (counter *Counter) Pause() {
	// TODO: atomic replace
	if counter.paused {
		return
	}
	t := time.Now()
	if counter.countdown {
		counter.state.left = counter.state.target.Sub(t).Truncate(time.Second)
	} else {
		counter.state.left = t.Sub(counter.state.target).Truncate(time.Second)
	}
	counter.paused = true
}

// Resume resumes a paused counter
func (counter *Counter) Resume() {
	// TODO: atomic replace
	if !counter.paused {
		return
	}
	t := time.Now()
	if counter.countdown {
		counter.state.target = t.Add(counter.state.left).Truncate(time.Second)
	} else {
		counter.state.target = t.Add(-counter.state.left).Truncate(time.Second)
	}
	counter.paused = false
}

// Diff gives a time difference to current time that can be used to format clock output strings
func (counter *Counter) Diff(t time.Time) time.Duration {
	if counter.paused {
		return counter.state.left
	}
	if counter.countdown {
		return counter.state.target.Sub(t)
	}
	return t.Sub(counter.state.target)
}

func (counter *Counter) setAutoColor(c color.RGBA, state int) {
	if counter.autoColorState != state {
		counter.autoColorState = state
		counter.signalColor = c
	}
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}
