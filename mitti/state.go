package mitti

import (
	"fmt"
	"log"
	"time"
)

type State struct {
	Remaining float32
	Playing   bool
	Paused    bool
	Updated   time.Time
}

func (state *State) String() string {
	return fmt.Sprintf("Mitti state updated %.2fs ago: playing=%v left=%f paused=%v",
		time.Now().Sub(state.Updated).Seconds(),
		state.Playing,
		state.Remaining, state.Paused,
	)
}

func (state *State) TogglePlay(i int32) {
	state.Updated = time.Now()
	if i == 0 {
		state.Playing = false
		state.Paused = true
	} else {
		state.Playing = true
		state.Paused = false
	}

	log.Printf("togglePlay: %d", i)
}

func (state *State) Copy() State {
	s := State{
		Remaining: state.Remaining,
		Playing:   state.Playing,
		Paused:    state.Paused,
		Updated:   state.Updated,
	}

	return s
}
