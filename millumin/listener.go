package millumin

import (
	"fmt"
	"github.com/hypebeast/go-osc/osc"
	"log"
	"time"
)

type State map[string]LayerState

type LayerState struct {
	Layer    string
	Updated  time.Time
	Playing  bool
	Info     MediaInfo
	Duration float32
	Paused   bool
	Time     float32
}

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

	log.Printf("Media started: %v", state)
}

func (state *LayerState) mediaPaused(mediaPaused MediaPaused) {
	state.Updated = time.Now()
	state.Paused = true
	state.Info = mediaPaused.MediaInfo
	state.Duration = mediaPaused.MediaInfo.Duration

	log.Printf("Media paused: %v", state)
}

func (state *LayerState) mediaStopped(mediaStopped MediaStopped) {
	// millumin sends mediaStarted -> mediaStopped when switching medias
	// ignore the following mediaStopped if we have already started a different media
	if state.Info.Index > 0 && mediaStopped.MediaInfo.Index != state.Info.Index {
		log.Printf("Media stopped (ignore): %v", state)

		return
	}

	state.Updated = time.Now()
	state.Playing = false
	state.Info = mediaStopped.MediaInfo
	state.Duration = mediaStopped.MediaInfo.Duration

	log.Printf("Media stopped: %v", state)
}

func (state *LayerState) mediaTime(mediaTime MediaTime) {
	state.Updated = time.Now()
	state.Duration = mediaTime.Duration
	state.Time = mediaTime.Value

	log.Printf("Media time: %v", state)
}

func MakeListener(oscServer *osc.Server) *Listener {
	var listener = Listener{
		layers:    make(map[string]*LayerState),
		listeners: make(map[chan State]struct{}),
	}

	listener.setup(oscServer)

	return &listener
}

type Listener struct {
	layers    map[string]*LayerState
	listeners map[chan State]struct{}
}

func (listener *Listener) Listen() chan State {
	var listenChan = make(chan State)

	listener.listeners[listenChan] = struct{}{}

	return listenChan
}

func (listener *Listener) update() {
	var state = make(State)

	for layer, layerState := range listener.layers {
		state[layer] = *layerState
	}

	for listenChan, _ := range listener.listeners {
		listenChan <- state
	}
}

func (listener *Listener) layer(layer string) *LayerState {
	if state := listener.layers[layer]; state == nil {
		state = &LayerState{Layer: layer}

		listener.layers[layer] = state

		return state
	} else {
		return state
	}
}

func (listener *Listener) handleMediaStarted(msg *osc.Message) {
	var layer string
	var message MediaStarted

	if err := parseAddressLayer(msg, &layer); err != nil {
		log.Printf("Unmarshal %v address: %v", msg, err)
	} else if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		listener.layer(layer).mediaStarted(message)
		listener.update()
	}
}

func (listener *Listener) handleMediaTime(msg *osc.Message) {
	var layer string
	var message MediaTime

	if err := parseAddressLayer(msg, &layer); err != nil {
		log.Printf("Unmarshal %v address: %v", msg, err)
	} else if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		listener.layer(layer).mediaTime(message)
		listener.update()
	}
}

func (listener *Listener) handleMediaPaused(msg *osc.Message) {
	var layer string
	var message MediaPaused

	if err := parseAddressLayer(msg, &layer); err != nil {
		log.Printf("Unmarshal %v address: %v", msg, err)
	} else if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		listener.layer(layer).mediaPaused(message)
		listener.update()
	}
}

func (listener *Listener) handleMediaStopped(msg *osc.Message) {
	var layer string
	var message MediaStopped

	if err := parseAddressLayer(msg, &layer); err != nil {
		log.Printf("Unmarshal %v address: %v", msg, err)
	} else if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		listener.layer(layer).mediaStopped(message)
		listener.update()
	}
}

func registerHandler(server *osc.Server, addr string, handler osc.HandlerFunc) {
	if err := server.Handle(addr, handler); err != nil {
		panic(err)
	}
}

func (listener *Listener) setup(server *osc.Server) {
	registerHandler(server, "/millumin/layer:*/mediaStarted", listener.handleMediaStarted)
	registerHandler(server, "/millumin/layer:*/media/time", listener.handleMediaTime)
	registerHandler(server, "/millumin/layer:*/mediaPaused", listener.handleMediaPaused)
	registerHandler(server, "/millumin/layer:*/mediaStopped", listener.handleMediaStopped)
}
