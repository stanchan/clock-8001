package mitti

import (
	"github.com/hypebeast/go-osc/osc"
	"log"
	"strconv"
	"strings"
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
	var remaining float32
	var cueTimeLeft string

	if err := msg.UnmarshalArgument(0, &cueTimeLeft); err != nil {
		log.Printf("mitti cueTimeLeft unmarshal %v: %v\n", msg, err)
	}
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

	remaining = float32(cs) / 100

	listener.state.Remaining = remaining
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
