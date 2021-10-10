package mitti

import (
	"fmt"
	"github.com/stanchan/clock-8001/v4/debug"
	"log"
	"time"
)

// State is the Mitti playback state from osc messages
type State struct {
	Remaining int
	Elapsed   int
	Hours     int
	Minutes   int
	Seconds   int
	Frames    int
	Progress  float64
	Paused    bool
	Loop      bool
	Updated   time.Time
}

func (state *State) String() string {
	return fmt.Sprintf("Mitti state updated %.2fs ago: left=%d elapsed=%d paused=%v loop=%v",
		time.Now().Sub(state.Updated).Seconds(),
		state.Remaining,
		state.Elapsed,
		state.Paused,
		state.Loop,
	)
}

// CueTimeLeft gets the time left on current Mitti cue
func (state *State) CueTimeLeft(cueTimeLeft string) {
	var hours, min, sec, cs int

	n, err := fmt.Sscanf(cueTimeLeft, "-%2d:%2d:%2d:%2d", &hours, &min, &sec, &cs)
	if err != nil || n != 4 {
		log.Printf("Error parsing cueTimeLeft string: %v, error: %v", cueTimeLeft, err)
		return
	}

	state.Hours = hours
	state.Minutes = min
	state.Seconds = sec
	state.Updated = time.Now()

	min += hours * 60
	sec += min * 60
	cs += sec * 100

	state.Remaining = sec
}

// CueTimeElapsed gets the elapsed time on current Mitti cue
func (state *State) CueTimeElapsed(cueTimeElapsed string) {
	var hours, min, sec, cs int

	n, err := fmt.Sscanf(cueTimeElapsed, "%2d:%2d:%2d:%2d", &hours, &min, &sec, &cs)
	if err != nil || n != 4 {
		log.Printf("Error parsing cueTimeElapsed string: \"%v\", error: %v", cueTimeElapsed, err)
		return
	}

	min += hours * 60
	sec += min * 60
	cs += sec * 100

	state.Updated = time.Now()
	state.Elapsed = sec
	debug.Printf("Mitti: elpased: %f\n", state.Elapsed)
}

// TogglePlay toggles the play/pause state
func (state *State) TogglePlay(i int32) {
	state.Updated = time.Now()
	if i == 0 {
		state.Paused = true
	} else {
		state.Paused = false
	}

	debug.Printf("Mitti: togglePlay: %d", i)
}

// ToggleLoop toggles the loop state
func (state *State) ToggleLoop(i int32) {
	state.Updated = time.Now()
	if i == 0 {
		state.Loop = false
	} else {
		state.Loop = true
	}

	debug.Printf("Mitti: toggleLoop: %d", i)
}

// Copy creates a new copy of the Mitti state
func (state *State) Copy() State {
	s := State{
		Remaining: state.Remaining,
		Elapsed:   state.Elapsed,
		Hours:     state.Hours,
		Minutes:   state.Minutes,
		Seconds:   state.Seconds,
		Frames:    state.Frames,
		Progress:  state.Progress,
		Paused:    state.Paused,
		Updated:   state.Updated,
		Loop:      state.Loop,
	}

	return s
}
