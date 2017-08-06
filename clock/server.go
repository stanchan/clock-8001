package clock

import (
	"github.com/hypebeast/go-osc/osc"
	"log"
)

func MakeServer(oscServer *osc.Server) *Server {
	var server = Server{}

	server.setup(oscServer)

	return &server
}

type Server struct {
}

func (server *Server) handleCount(msg *osc.Message) {
	var message CountMessage

	if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		log.Printf("count: %#v", message)
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
