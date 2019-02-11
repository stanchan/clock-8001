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

func (server *Server) handleCountupStart(msg *osc.Message) {
	log.Printf("countup start: %#v", msg)
	message := ClockMessage{
		Type: "countup",
	}
	server.update(message)
}

func (server *Server) handleKill(msg *osc.Message) {
	log.Printf("kill: %#v", msg)
	message := ClockMessage{
		Type: "kill",
	}
	server.update(message)
}

func (server *Server) handleCountdownStart(msg *osc.Message) {
	server.sendCountdownMessage("countdownStart", msg)
}

func (server *Server) handleCountdownStart2(msg *osc.Message) {
	server.sendCountdownMessage("countdownStart2", msg)
}

func (server *Server) handleCountdownModify(msg *osc.Message) {
	server.sendCountdownMessage("countdownModify", msg)
}

func (server *Server) handleCountdownModify2(msg *osc.Message) {
	server.sendCountdownMessage("countdownModify2", msg)
}

func (server *Server) handleCountdownStop(msg *osc.Message) {
	log.Printf("countdownStop: %#v", msg)
	message := ClockMessage{
		Type: "countdownStop",
	}
	server.update(message)
}

func (server *Server) handleCountdownStop2(msg *osc.Message) {
	log.Printf("countdownStop2: %#v", msg)
	message := ClockMessage{
		Type: "countdownStop2",
	}
	server.update(message)
}

func (server *Server) sendCountdownMessage(cmd string, msg *osc.Message) {
	var message CountdownMessage

	if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		log.Printf("%s: %#v", cmd, message)
		msg := ClockMessage{
			Type:             cmd,
			CountdownMessage: &message,
		}
		server.update(msg)
	}
}

func (server *Server) handleNormal(msg *osc.Message) {
	log.Printf("normal: %#v", msg)
	message := ClockMessage{
		Type: "normal",
	}
	server.update(message)
}

func (server *Server) handlePause(msg *osc.Message) {
	log.Printf("normal: %#v", msg)
	message := ClockMessage{
		Type: "pause",
	}
	server.update(message)
}

func (server *Server) handleResume(msg *osc.Message) {
	log.Printf("normal: %#v", msg)
	message := ClockMessage{
		Type: "resume",
	}
	server.update(message)
}

func (server *Server) handleDisplay(msg *osc.Message) {
	var message DisplayMessage

	if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		log.Printf("display: %#v", message)
		msg := ClockMessage{
			Type:           "display",
			DisplayMessage: &message,
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
	registerHandler(oscServer, "/clock/tally", server.handleCount)
	registerHandler(oscServer, "/clock/display", server.handleDisplay)
	registerHandler(oscServer, "/clock/countdown/start", server.handleCountdownStart)
	registerHandler(oscServer, "/clock/countdown2/start", server.handleCountdownStart2)
	registerHandler(oscServer, "/clock/countdown/modify", server.handleCountdownModify)
	registerHandler(oscServer, "/clock/countdown2/modify", server.handleCountdownModify2)
	registerHandler(oscServer, "/clock/countdown/stop", server.handleCountdownStop)
	registerHandler(oscServer, "/clock/countdown2/stop", server.handleCountdownStop2)
	registerHandler(oscServer, "/clock/pause", server.handlePause)
	registerHandler(oscServer, "/clock/resume", server.handleResume)
	registerHandler(oscServer, "/clock/countup/start", server.handleCountupStart)
	registerHandler(oscServer, "/clock/kill", server.handleKill)
	registerHandler(oscServer, "/clock/normal", server.handleNormal)
}
