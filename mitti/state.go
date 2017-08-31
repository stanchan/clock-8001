package mitti

import (
	"fmt"
	"log"
	"strconv"
	"strings"
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

func (state *State) CueTimeLeft(cueTimeLeft string) {
	s := strings.Split(strings.Trim(cueTimeLeft, "-"), ":")
	hours, err := strconv.Atoi(s[0])
	if err != nil {
		log.Printf("cueTimeLeft time conversion: %v\n", err)
	}
	min, err := strconv.Atoi(s[1])
	if err != nil {
		log.Printf("cueTimeLeft time conversion: %v\n", err)
	}
	sec, err := strconv.Atoi(s[2])
	if err != nil {
		log.Printf("cueTimeLeft time conversion: %v\n", err)
	}
	cs, err := strconv.Atoi(s[3])
	if err != nil {
		log.Printf("cueTimeLeft time conversion: %v\n", err)
	}

	min += hours * 60
	sec += min * 60
	cs += sec * 100

	state.Updated = time.Now()
	state.Remaining = float32(cs) / 100
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
