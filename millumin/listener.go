package millumin

import (
	"github.com/hypebeast/go-osc/osc"
	"log"
)

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
