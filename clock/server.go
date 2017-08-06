package clock

import (
	"github.com/hypebeast/go-osc/osc"
	"log"
)

func MakeServer(oscServer *osc.Server) *Server {
	var server = Server{
		listeners: make(map[chan CountMessage]struct{}),
	}

	server.setup(oscServer)

	return &server
}

type Server struct {
	listeners map[chan CountMessage]struct{}
}

func (server *Server) Listen() chan CountMessage {
	var listenChan = make(chan CountMessage)

	server.listeners[listenChan] = struct{}{}

	return listenChan
}

func (server *Server) update(message CountMessage) {
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
		server.update(message)
	}
}

func registerHandler(server *osc.Server, addr string, handler osc.HandlerFunc) {
	if err := server.Handle(addr, handler); err != nil {
		panic(err)
	}
}

func (server *Server) setup(oscServer *osc.Server) {
	registerHandler(oscServer, "/qmsk/clock/count", server.handleCount)
}
