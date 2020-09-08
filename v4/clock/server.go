package clock

import (
	"github.com/hypebeast/go-osc/osc"
	"gitlab.com/Depili/clock-8001/v4/debug"
	"log"
	"regexp"
	"strconv"
)

const (
	timerPattern = `/clock/timer/(\d)/`
)

// MakeServer creates a clock.Server instance from osc.Server instance
func MakeServer(oscServer *osc.Server) *Server {
	var server = Server{
		listeners:   make(map[chan Message]struct{}),
		Debug:       false,
		timerRegexp: regexp.MustCompile(timerPattern),
	}

	server.setup(oscServer)

	return &server
}

// Server is a clock osc server and listens for incoming osc messages
type Server struct {
	listeners   map[chan Message]struct{}
	Debug       bool
	timerRegexp *regexp.Regexp
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

func (server *Server) handleKill(msg *osc.Message) {
	debug.Printf("kill: %#v", msg)

	message := Message{
		Type: "kill",
	}
	server.update(message)
}

func (server *Server) handleCountupStart(msg *osc.Message) {
	log.Printf("countup start: %#v", msg)
	if msg.Address == "/clock/countup/start" {
		msg.Address = "/clock/timer/0/countup"
	}

	if matches := server.timerRegexp.FindStringSubmatch(msg.Address); len(matches) == 2 {
		counter, _ := strconv.Atoi(matches[1])

		msg := Message{
			Type:             "timerStart",
			Counter:          counter,
			Countdown:        false,
			CountdownMessage: &CountdownMessage{Seconds: 0},
		}
		server.update(msg)

	} else {
		log.Printf("matches: %v", matches)
		log.Printf("invalid timer message: %v\n", msg)
	}
}

func (server *Server) handleCountdownStart(msg *osc.Message) {
	log.Printf("handleCountdownStart: %v", msg)
	if msg.Address == "/clock/countdown/start" {
		msg.Address = "/clock/timer/0/start"
	} else if msg.Address == "/clock/countdown2/start" {
		msg.Address = "/clock/timer/1/start"
	}
	server.sendTimerMessage("timerStart", true, msg)
}

func (server *Server) handleTimerModify(msg *osc.Message) {
	if msg.Address == "/clock/countdown/modify" {
		msg.Address = "/clock/timer/0/modify"
	} else if msg.Address == "/clock/countdown2/modify" {
		msg.Address = "/clock/timer/1/modify"
	} else if msg.Address == "/clock/countup/modify" {
		msg.Address = "/clock/timer/0/modify"
	}
	server.sendTimerMessage("timerModify", false, msg)
}

func (server *Server) handleTimerStop(msg *osc.Message) {
	debug.Printf("countdownStop: %#v", msg)
	if msg.Address == "/clock/countdown/stop" {
		msg.Address = "/clock/timer/0/stop"
	} else if msg.Address == "/clock/countdown2/stop" {
		msg.Address = "/clock/timer/1/stop"
	}

	if matches := server.timerRegexp.FindStringSubmatch(msg.Address); len(matches) == 2 {
		counter, _ := strconv.Atoi(matches[1])

		msg := Message{
			Type:    "timerStop",
			Counter: counter,
		}
		server.update(msg)

	} else {
		log.Printf("matches: %v", matches)
		log.Printf("invalid timer message: %v\n", msg)
	}
}

func (server *Server) sendTimerMessage(cmd string, countdown bool, msg *osc.Message) {
	if matches := server.timerRegexp.FindStringSubmatch(msg.Address); len(matches) == 2 {
		counter, _ := strconv.Atoi(matches[1])
		var message CountdownMessage

		if err := message.UnmarshalOSC(msg); err != nil {
			log.Printf("Unmarshal %v: %v", msg, err)
		} else {
			debug.Printf("%s: %#v", cmd, message)
			msg := Message{
				Type:             cmd,
				Counter:          counter,
				Countdown:        countdown,
				CountdownMessage: &message,
			}
			server.update(msg)
		}
	} else {
		log.Printf("matches: %v", matches)
		log.Printf("invalid timer message: %v\n", msg)
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

func (server *Server) handleLTC(msg *osc.Message) {
	var message TimeMessage

	if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		debug.Printf("LTC: %v\n", message.Time)
		m := Message{
			Type: "LTC",
			Data: message.Time,
		}
		server.update(m)
	}
}

func (server *Server) handleDualText(msg *osc.Message) {
	var message TextMessage
	if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		debug.Printf("Dual clock text: %v\n", message.Text)
		m := Message{
			Type: "dualText",
			Data: message.Text,
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
	registerHandler(oscServer, "/clock/timer/*/countdown", server.handleCountdownStart)
	registerHandler(oscServer, "/clock/timer/*/countup", server.handleCountupStart)
	registerHandler(oscServer, "/clock/timer/*/modify", server.handleTimerModify)
	registerHandler(oscServer, "/clock/timer/*/stop", server.handleTimerStop)

	// Old OSC Api from V3
	registerHandler(oscServer, "/qmsk/clock/count", server.handleCount)
	registerHandler(oscServer, "/clock/tally", server.handleCount)
	registerHandler(oscServer, "/clock/display", server.handleDisplay)
	registerHandler(oscServer, "/clock/countdown/start", server.handleCountdownStart)
	registerHandler(oscServer, "/clock/countdown2/start", server.handleCountdownStart)
	registerHandler(oscServer, "/clock/countdown/modify", server.handleTimerModify)
	registerHandler(oscServer, "/clock/countdown2/modify", server.handleTimerModify)
	registerHandler(oscServer, "/clock/countdown/stop", server.handleTimerStop)
	registerHandler(oscServer, "/clock/countdown2/stop", server.handleTimerStop)
	registerHandler(oscServer, "/clock/pause", server.handlePause)
	registerHandler(oscServer, "/clock/resume", server.handleResume)
	registerHandler(oscServer, "/clock/countup/start", server.handleCountupStart)
	registerHandler(oscServer, "/clock/countup/modify", server.handleTimerModify)
	registerHandler(oscServer, "/clock/kill", server.handleKill)
	registerHandler(oscServer, "/clock/normal", server.handleNormal)
	registerHandler(oscServer, "/clock/seconds/off", server.handleSecondsOff)
	registerHandler(oscServer, "/clock/seconds/on", server.handleSecondsOn)
	registerHandler(oscServer, "/clock/time/set", server.handleTimeSet)
	registerHandler(oscServer, "/clock/ltc", server.handleLTC)
	registerHandler(oscServer, "/clock/dual/text", server.handleDualText)
}
