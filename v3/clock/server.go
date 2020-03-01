package clock

import (
	"github.com/hypebeast/go-osc/osc"
	"gitlab.com/Depili/clock-8001/v3/debug"
	"log"
)

// MakeServer creates a clock.Server instance from osc.Server instance
func MakeServer(oscServer *osc.Server) *Server {
	var server = Server{
		listeners: make(map[chan Message]struct{}),
		Debug:     false,
	}

	server.setup(oscServer)

	return &server
}

// Server is a clock osc server and listens for incoming osc messages
type Server struct {
	listeners map[chan Message]struct{}
	Debug     bool
}

// Listen adds a new listener for the decoded incoming osc messages
func (server *Server) Listen() chan Message {
	var listenChan = make(chan Message)
	server.listeners[listenChan] = struct{}{}
	return listenChan
}

func (server *Server) update(message Message) {
	debug.Printf("update: %#v", message)

	for listenChan := range server.listeners {
		listenChan <- message
	}
}

func (server *Server) handleCount(msg *osc.Message) {
	var message CountMessage

	if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		msg := Message{
			Type:         "count",
			CountMessage: &message,
		}
		server.update(msg)
	}
}

func (server *Server) handleCountupStart(msg *osc.Message) {
	debug.Printf("countup start: %#v", msg)

	message := Message{
		Type: "countup",
	}
	server.update(message)
}

func (server *Server) handleKill(msg *osc.Message) {
	debug.Printf("kill: %#v", msg)

	message := Message{
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
	debug.Printf("countdownStop: %#v", msg)
	message := Message{
		Type: "countdownStop",
	}
	server.update(message)
}

func (server *Server) handleCountdownStop2(msg *osc.Message) {
	debug.Printf("countdownStop2: %#v", msg)
	message := Message{
		Type: "countdownStop2",
	}
	server.update(message)
}

func (server *Server) sendCountdownMessage(cmd string, msg *osc.Message) {
	var message CountdownMessage

	if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		debug.Printf("%s: %#v", cmd, message)
		msg := Message{
			Type:             cmd,
			CountdownMessage: &message,
		}
		server.update(msg)
	}
}

func (server *Server) handleNormal(msg *osc.Message) {
	debug.Printf("normal: %#v", msg)
	message := Message{
		Type: "normal",
	}
	server.update(message)
}

func (server *Server) handlePause(msg *osc.Message) {
	debug.Printf("pause: %#v", msg)
	message := Message{
		Type: "pause",
	}
	server.update(message)
}

func (server *Server) handleResume(msg *osc.Message) {
	debug.Printf("resume: %#v", msg)
	message := Message{
		Type: "resume",
	}
	server.update(message)
}

func (server *Server) handleDisplay(msg *osc.Message) {
	var message DisplayMessage

	if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		debug.Printf("display: %#v", message)
		msg := Message{
			Type:           "display",
			DisplayMessage: &message,
		}
		server.update(msg)
	}
}

func (server *Server) handleSecondsOff(msg *osc.Message) {
	debug.Printf("Second display off: %v\n", msg)
	message := Message{
		Type: "secondsOff",
	}
	server.update(message)
}

func (server *Server) handleSecondsOn(msg *osc.Message) {
	debug.Printf("Second display on: %v\n", msg)
	message := Message{
		Type: "secondsOn",
	}
	server.update(message)
}

func (server *Server) handleTimeSet(msg *osc.Message) {
	var message TimeMessage

	if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		debug.Printf("Set time: %v\n", message.Time)
		m := Message{
			Type: "setTime",
			Data: message.Time,
		}
		server.update(m)
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
	registerHandler(oscServer, "/clock/seconds/off", server.handleSecondsOff)
	registerHandler(oscServer, "/clock/seconds/on", server.handleSecondsOn)
	registerHandler(oscServer, "/clock/time/set", server.handleTimeSet)
}
