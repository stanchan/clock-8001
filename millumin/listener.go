package millumin

import (
	"github.com/hypebeast/go-osc/osc"
	"log"
	"time"
)

type LayerState struct {
	Layer    string
	Updated  time.Time
	Playing  bool
	Info     MediaInfo
	Duration float32
	Paused   bool
	Time     float32
}

func (state *LayerState) mediaStarted(mediaStarted MediaStarted) {
	state.Updated = time.Now()
	state.Playing = true
	state.Info = mediaStarted.MediaInfo
	state.Duration = mediaStarted.MediaInfo.Duration
	state.Paused = false
	state.Time = 0.0

	log.Printf("Media started: %#v", state)
}

func (state *LayerState) mediaPaused(mediaPaused MediaPaused) {
	state.Updated = time.Now()
	state.Paused = true
	state.Info = mediaPaused.MediaInfo
	state.Duration = mediaPaused.MediaInfo.Duration

	log.Printf("Media paused: %#v", state)
}

func (state *LayerState) mediaStopped(mediaStopped MediaStopped) {
	state.Updated = time.Now()
	state.Playing = false
	state.Info = mediaStopped.MediaInfo
	state.Duration = mediaStopped.MediaInfo.Duration

	log.Printf("Media stopped: %#v", state)
}

func (state *LayerState) mediaTime(mediaTime MediaTime) {
	state.Updated = time.Now()
	state.Duration = mediaTime.Duration
	state.Time = mediaTime.Value

	log.Printf("Media time: %#v", state)
}

func MakeListener() *Listener {
	var listener = Listener{
		layers: make(map[string]*LayerState),
	}

	return &listener
}

type Listener struct {
	layers map[string]*LayerState
}

func (listener *Listener) updateLayer(layer string) *LayerState {
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
		listener.updateLayer(layer).mediaStarted(message)
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
		listener.updateLayer(layer).mediaTime(message)
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
		listener.updateLayer(layer).mediaPaused(message)
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
		listener.updateLayer(layer).mediaStopped(message)
	}
}

func registerHandler(server *osc.Server, addr string, handler osc.HandlerFunc) {
	if err := server.Handle(addr, handler); err != nil {
		panic(err)
	}
}

func (listener *Listener) Setup(server *osc.Server) {
	registerHandler(server, "/millumin/layer:*/mediaStarted", listener.handleMediaStarted)
	registerHandler(server, "/millumin/layer:*/media/time", listener.handleMediaTime)
	registerHandler(server, "/millumin/layer:*/mediaPaused", listener.handleMediaPaused)
	registerHandler(server, "/millumin/layer:*/mediaStopped", listener.handleMediaStopped)
}
