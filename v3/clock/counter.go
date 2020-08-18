package clock

import (
	"fmt"
	"time"
)

/*
 * Counters representing generic countdowns / ups
 */

// Counter abstracts a generic counter counting up or down
type Counter struct {
	state     *counterState
	active    bool // Is this counter active?
	countdown bool // Count up / down from the target
	paused    bool // Is the counter paused?
}

type counterState struct {
	target   time.Time     // Target timestamp for main countdown
	duration time.Duration // Total duration of main countdown, used to scale the leds
	left     time.Duration // Duration left when paused
}

// CounterOutput the data structure returned by Counter.Output() and contains the static state of the counter at that time
type CounterOutput struct {
	Active    bool          // True if the counter is active
	Countdown bool          // True if counting down, false if counting up
	Paused    bool          // True if counter has been paused
	Expired   bool          // Has the countdown timer expired?
	Hours     int           // Hour part of the timer
	Minutes   int           // Minutes of the timer, 0-60
	Seconds   int           // Seconds of the timer, 0-60
	Combined  string        // HH:MM:SS string representation
	Compact   string        // Compact 4-character output
	Progress  float64       // Percentage of total time elapsed of the countdown, 0-1
	Diff      time.Duration // raw difference
}

// Output generates the static output of the counter for use in clock displays
func (counter *Counter) Output(t time.Time) *CounterOutput {
	diff := counter.Diff(t)
	hours := int(diff.Truncate(time.Hour).Hours())
	minutes := int(diff.Truncate(time.Minute).Minutes()) - (hours * 60)
	seconds := int(diff.Truncate(time.Second).Seconds()) - (((hours * 60) + minutes) * 60)

	progress := (float64(diff) / float64(counter.state.duration))
	if progress >= 1 {
		progress = 1
	} else if progress < 0 {
		progress = 0
	}

	out := &CounterOutput{
		Active:    counter.active,
		Countdown: counter.countdown,
		Paused:    counter.paused,
		Expired:   diff < 0,
		Hours:     hours,
		Minutes:   minutes,
		Seconds:   seconds,
		Combined:  fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds),
		Compact:   counter.compactOutput(t),
		Progress:  progress,
		Diff:      diff,
	}
	return out
}

func (counter *Counter) compactOutput(t time.Time) string {
	rawSecs := int64(counter.Diff(t).Truncate(time.Second).Seconds())

	for _, unit := range clockUnits {
		if rawSecs/int64(unit.seconds) >= 100 {
			continue
		}
		count := rawSecs / int64(unit.seconds)
		if counter.paused {
			return fmt.Sprintf(" %02d%1s", count, unit.unit)
		} else if counter.countdown {
			return fmt.Sprintf("↓%02d%1s", count, unit.unit)
		} else {
			return fmt.Sprintf("↑%02d%1s", count, unit.unit)
		}
		return ""
	}
	return ""
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
	counter.active = true
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
		return counter.state.target.Sub(t).Truncate(time.Second)
	}
	return t.Sub(counter.state.target)
}
