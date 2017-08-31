package mitti

import (
	"github.com/hypebeast/go-osc/osc"
	"log"
)

func MakeListener(oscServer *osc.Server) *Listener {
	var listener = Listener{
		listeners: make(map[chan State]struct{}),
	}

	var state State
	state.Playing = false
	state.Paused = true
	state.Remaining = 0
	listener.state = &state

	listener.setup(oscServer)

	return &listener
}

type Listener struct {
	state     *State
	listeners map[chan State]struct{}
}

func (listener *Listener) Listen() chan State {
	var listenChan = make(chan State)

	listener.listeners[listenChan] = struct{}{}

	return listenChan
}

func (listener *Listener) update() {
	state := listener.state.Copy()

	for listenChan, _ := range listener.listeners {
		listenChan <- state
	}
	log.Printf("mitti state update: %v\n", listener.state)
}

func (listener *Listener) handleTogglePlay(msg *osc.Message) {
	var playing int32

	if err := msg.UnmarshalArgument(0, &playing); err != nil {
		log.Printf("mitti togglePlay unmarshal %v: %v\n", msg, err)
	}

	listener.state.TogglePlay(playing)
	listener.update()
}

func (listener *Listener) handleCueTimeLeft(msg *osc.Message) {
	var cueTimeLeft string

	if err := msg.UnmarshalArgument(0, &cueTimeLeft); err != nil {
		log.Printf("mitti cueTimeLeft unmarshal %v: %v\n", msg, err)
	}

	listener.state.CueTimeLeft(cueTimeLeft)
	listener.update()
}

func registerHandler(server *osc.Server, addr string, handler osc.HandlerFunc) {
	if err := server.Handle(addr, handler); err != nil {
		panic(err)
	}
}

func (listener *Listener) setup(server *osc.Server) {
	registerHandler(server, "/mitti/cueTimeLeft", listener.handleCueTimeLeft)
	registerHandler(server, "/mitti/togglePlay", listener.handleTogglePlay)
}
