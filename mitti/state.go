package mitti

import (
	"fmt"
	"log"
	"time"

	"github.com/stanchan/clock-8001/debug"
)

// State is the Mitti playback state from osc messages
type State struct {
	Remaining float32
	Playing   bool
	Paused    bool
	Loop      bool
	Updated   time.Time
}

func (state *State) String() string {
	return fmt.Sprintf("Mitti state updated %.2fs ago: playing=%v left=%f paused=%v loop=%v",
		time.Now().Sub(state.Updated).Seconds(),
		state.Playing,
		state.Remaining, state.Paused, state.Loop,
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

	min += hours * 60
	sec += min * 60
	cs += sec * 100

	state.Updated = time.Now()
	state.Remaining = float32(cs) / 100
}

// TogglePlay toggles the play/pause state
func (state *State) TogglePlay(i int32) {
	state.Updated = time.Now()
	if i == 0 {
		state.Playing = false
		state.Paused = true
	} else {
		state.Playing = true
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
		Playing:   state.Playing,
		Paused:    state.Paused,
		Updated:   state.Updated,
		Loop:      state.Loop,
	}

	return s
}
