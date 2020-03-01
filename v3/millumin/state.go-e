package millumin

import (
	"fmt"
	"gitlab.com/Depili/clock-8001/debug"
	"time"
)

// State is complete Millumin playback state for all layers
type State map[string]LayerState

// LayerState is the state of individual Millumin playback layer
type LayerState struct {
	Layer    string
	Updated  time.Time
	Playing  bool
	Info     MediaInfo
	Duration float32
	Paused   bool
	Time     float32
}

// Remaining returns the time remaining on the given layer
func (state LayerState) Remaining() float32 {
	// TODO: sanity-check
	return state.Duration - state.Time
}

// String prints the layer information
func (state *LayerState) String() string {
	return fmt.Sprintf("Layer %s updated %.2fs ago: playing=%v info={index=%v name=%v duration=%f} duration=%f paused=%v time=%f",
		state.Layer, time.Now().Sub(state.Updated).Seconds(),
		state.Playing,
		state.Info.Index, state.Info.Name, state.Info.Duration,
		state.Duration, state.Paused, state.Time,
	)
}

func (state *LayerState) mediaStarted(mediaStarted MediaStarted) {
	state.Updated = time.Now()
	state.Playing = true
	state.Info = mediaStarted.MediaInfo
	state.Duration = mediaStarted.MediaInfo.Duration
	state.Paused = false
	state.Time = 0.0

	debug.Printf("Media started: %v", state)
}

func (state *LayerState) mediaPaused(mediaPaused MediaPaused) {
	state.Updated = time.Now()
	state.Paused = true
	state.Info = mediaPaused.MediaInfo
	state.Duration = mediaPaused.MediaInfo.Duration

	debug.Printf("Media paused: %v", state)
}

func (state *LayerState) mediaStopped(mediaStopped MediaStopped) {
	// millumin sends mediaStarted -> mediaStopped when switching medias
	// ignore the following mediaStopped if we have already started a different media
	if state.Info.Index > 0 && mediaStopped.MediaInfo.Index != state.Info.Index {
		debug.Printf("Media stopped (ignore): %v", state)

		return
	}

	state.Updated = time.Now()
	state.Playing = false
	state.Info = mediaStopped.MediaInfo
	state.Duration = mediaStopped.MediaInfo.Duration

	debug.Printf("Media stopped: %v", state)
}

func (state *LayerState) mediaTime(mediaTime MediaTime) {
	state.Updated = time.Now()
	state.Playing = true
	state.Duration = mediaTime.Duration
	state.Time = mediaTime.Value

	debug.Printf("Media time: %v", state)
}
