package clock

import (
	"github.com/hypebeast/go-osc/osc"
	"log"
)

func MakeServer(oscServer *osc.Server) *Server {
	var server = Server{
		listeners: make(map[chan ClockMessage]struct{}),
	}

	server.setup(oscServer)

	return &server
}

type Server struct {
	listeners map[chan ClockMessage]struct{}
}

func (server *Server) Listen() chan ClockMessage {
	var listenChan = make(chan ClockMessage)

	server.listeners[listenChan] = struct{}{}

	return listenChan
}

func (server *Server) update(message ClockMessage) {
	log.Printf("update: %#v", message)

	for listenChan, _ := range server.listeners {
		listenChan <- message
	}
}

func (server *Server) handleCount(msg *osc.Message) {
	var message CountMessage

	if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		msg := ClockMessage{
			Type:         "count",
			CountMessage: &message,
		}
		server.update(msg)
	}
}

func (server *Server) handleStart(msg *osc.Message) {
	var message StartMessage

	if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		log.Printf("countdown start: %#v", message)
		msg := ClockMessage{
			Type:         "start",
			StartMessage: &message,
		}
		server.update(msg)
	}

}

func registerHandler(server *osc.Server, addr string, handler osc.HandlerFunc) {
	if err := server.Handle(addr, handler); err != nil {
		panic(err)
	}
}

func (server *Server) setup(oscServer *osc.Server) {
	registerHandler(oscServer, "/qmsk/clock/count", server.handleCount)
	registerHandler(oscServer, "/clock/countdown/start", server.handleStart)
}
