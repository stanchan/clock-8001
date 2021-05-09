package clock

import (
	"gitlab.com/Depili/clock-8001/v4/debug"
	"gitlab.com/Depili/go-osc/osc"
	"image/color"
	"log"
	"regexp"
	"strconv"
	"time"
)

const (
	timerPattern  = `/clock/timer/(\d)/`
	sourcePattern = `/clock/source/([1-4])/`
)

// MakeServer creates a clock.Server instance from osc.Server instance
func MakeServer(oscServer *osc.Server, uuid string) *Server {
	var server = Server{
		listeners:    make(map[chan Message]struct{}),
		Debug:        false,
		timerRegexp:  regexp.MustCompile(timerPattern),
		sourceRegexp: regexp.MustCompile(sourcePattern),
		uuid:         uuid,
	}

	server.setup(oscServer)

	return &server
}

// Server is a clock osc server and listens for incoming osc messages
type Server struct {
	listeners    map[chan Message]struct{}
	Debug        bool
	timerRegexp  *regexp.Regexp
	sourceRegexp *regexp.Regexp
	lastMedia    time.Time
	uuid         string
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

/*
 * Timer related handlers
 */

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

	server.sendTimerCommand("timerStop", msg)
}

func (server *Server) handleTimerPause(msg *osc.Message) {
	debug.Printf("handleTimerPause: %v", msg)
	server.sendTimerCommand("timerPause", msg)
}

func (server *Server) handleTimerResume(msg *osc.Message) {
	debug.Printf("handlTimerResume: %v", msg)
	server.sendTimerCommand("timerResume", msg)
}

func (server *Server) handleCountdownTarget(msg *osc.Message) {
	server.sendTargetMessage(msg, true)
}

func (server *Server) handleCountupTarget(msg *osc.Message) {
	server.sendTargetMessage(msg, false)
}

func (server *Server) sendTargetMessage(msg *osc.Message, countdown bool) {
	debug.Printf("sendTargetMessage: %v %v", countdown, msg)
	if matches := server.timerRegexp.FindStringSubmatch(msg.Address); len(matches) == 2 {
		counter, _ := strconv.Atoi(matches[1])
		var target string
		err := msg.UnmarshalArguments(&target)
		if err != nil {
			log.Printf("handleTimerTarget error: %v", err)
			return
		}
		m := Message{
			Type:      "timerTarget",
			Countdown: countdown,
			Counter:   counter,
			Data:      target,
		}
		server.update(m)
	}
}

func (server *Server) sendTimerCommand(cmd string, msg *osc.Message) {
	if matches := server.timerRegexp.FindStringSubmatch(msg.Address); len(matches) == 2 {
		counter, _ := strconv.Atoi(matches[1])

		msg := Message{
			Type:    cmd,
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

func (server *Server) handleTimerSignal(msg *osc.Message) {
	debug.Printf("handleTimerSignal: %v", msg)
	if matches := server.timerRegexp.FindStringSubmatch(msg.Address); len(matches) == 2 {
		counter, _ := strconv.Atoi(matches[1])
		var r, g, b, a int32
		err := msg.UnmarshalArguments(&r, &g, &b, &a)
		if err != nil {
			log.Printf("handleTimerSignal: %v %v", err, msg)
			return
		}

		colors := make([]color.RGBA, 1)
		colors[0] = color.RGBA{
			R: uint8(r),
			G: uint8(g),
			B: uint8(b),
			A: uint8(a),
		}
		message := Message{
			Type:    "timerSignal",
			Counter: counter,
			Colors:  colors,
		}

		server.update(message)
	} else {
		log.Printf("handleTimerSignal: Invalid message: %v", msg)
	}
}

/*
 * Source related handlers
 */
func (server *Server) parseSourceMsg(msg *osc.Message, cmd string) {
	if matches := server.sourceRegexp.FindStringSubmatch(msg.Address); len(matches) == 2 {
		counter, _ := strconv.Atoi(matches[1])

		msg := Message{
			Type:    cmd,
			Counter: counter - 1,
		}
		server.update(msg)
	} else {
		log.Printf("matches: %v", matches)
		log.Printf("invalid source message: %v\n", msg)
	}
}

func (server *Server) handleHideAll(msg *osc.Message) {
	debug.Printf("handleHide: %#v", msg)

	message := Message{
		Type: "hideAll",
	}
	server.update(message)
}

func (server *Server) handleShowAll(msg *osc.Message) {
	debug.Printf("handleShowAll: %#v", msg)
	message := Message{
		Type: "showAll",
	}
	server.update(message)
}

func (server *Server) handleHide(msg *osc.Message) {
	debug.Printf("handleHide: %v", msg)
	server.parseSourceMsg(msg, "sourceHide")
}

func (server *Server) handleShow(msg *osc.Message) {
	debug.Printf("handleShow: %v", msg)
	server.parseSourceMsg(msg, "sourceShow")
}

func (server *Server) handleSourceTitle(msg *osc.Message) {
	debug.Printf("handleSourceTitle: %v", msg)

	if matches := server.sourceRegexp.FindStringSubmatch(msg.Address); len(matches) == 2 {
		counter, _ := strconv.Atoi(matches[1])

		var label string
		err := msg.UnmarshalArguments(&label)
		if err != nil {
			log.Printf("handleSourceTitle error: %v", err)
			return
		}

		msg := Message{
			Type:    "sourceTitle",
			Counter: counter - 1,
			Data:    label,
		}
		server.update(msg)
	} else {
		log.Printf("matches: %v", matches)
		log.Printf("invalid source message: %v\n", msg)
	}
}

func (server *Server) handleSourceColor(msg *osc.Message) {
	debug.Printf("handleSourceColor: %v", msg)
	if matches := server.sourceRegexp.FindStringSubmatch(msg.Address); len(matches) == 2 {
		counter, _ := strconv.Atoi(matches[1])

		cm := ColorMessage{}
		err := cm.UnmarshalOSC(msg)
		if err != nil {
			log.Printf("colors unmarshal: %v - %v", err, msg)
			return
		}

		m := Message{
			Type:    "sourceColors",
			Counter: counter - 1,
			Colors:  cm.ToRGBA(),
		}
		server.update(m)
	}
}

func (server *Server) handleTitleColors(msg *osc.Message) {
	debug.Printf("handleTitleColor: %v", msg)
	cm := ColorMessage{}
	err := cm.UnmarshalOSC(msg)
	if err != nil {
		log.Printf("colors unmarshal: %v - %v", err, msg)
		return
	}

	m := Message{
		Type:   "titleColors",
		Colors: cm.ToRGBA(),
	}
	server.update(m)
}

/*
 * Clock sync handlers
 */

func (server *Server) handleMedia(msg *osc.Message) {
	debug.Printf("handleMedia: %v", msg)
	message := Message{}
	if msg.Address == "/clock/media/mitti" {
		message.Type = "mitti"
	} else if msg.Address == "/clock/media/millumin" {
		message.Type = "millumin"
	} else {
		log.Printf("Unknown media message: %v", msg)
		return
	}
	mm := MediaMessage{}
	err := mm.UnmarshalOSC(msg)
	if err != nil {
		log.Printf("error unmarshaling media message: %v", err)
		return
	}

	if mm.uuid == server.uuid {
		// Our own message, ignore
		return
	}

	if server.lastMedia.Before(mm.timeStamp.Time()) {
		server.lastMedia = mm.timeStamp.Time()
		message.MediaMessage = &mm
		server.update(message)
	}
}

func (server *Server) handleResetMedia(msg *osc.Message) {
	var uuid string
	var timeStamp *osc.Timetag

	debug.Printf("handleResetMedia: %v", msg)
	message := Message{}
	if msg.Address == "/clock/resetmedia/mitti" {
		message.Type = "mittiReset"
	} else if msg.Address == "/clock/resetmedia/millumin" {
		message.Type = "milluminReset"
	} else {
		log.Printf("Unknown resetMedia message: %v", msg)
		return
	}

	err := msg.UnmarshalArguments(&timeStamp, &uuid)
	if err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
		return
	}

	if uuid == server.uuid {
		return
	}

	if server.lastMedia.Before(timeStamp.Time()) {
		server.lastMedia = timeStamp.Time()
		server.update(message)
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

/*
 * Misc commands
 */

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

func (server *Server) handleBackground(msg *osc.Message) {
	debug.Printf("background: %v", msg)
	var bg int32
	err := msg.UnmarshalArguments(&bg)
	if err != nil {
		log.Printf("Background msg error: %v", err)
		return
	}
	m := Message{
		Type:    "background",
		Counter: int(bg),
	}
	server.update(m)
}

func (server *Server) handleInfo(msg *osc.Message) {
	debug.Printf("handleInfo")
	var bg int32
	err := msg.UnmarshalArguments(&bg)
	if err != nil {
		log.Printf("Info msg error: %v", err)
		return
	}
	m := Message{
		Type:    "showInfo",
		Counter: int(bg),
	}
	server.update(m)
}

func (server *Server) handleFlash(msg *osc.Message) {
	debug.Printf("handleFlash")
	m := Message{
		Type: "screenFlash",
	}
	server.update(m)
}

/*
 * Deprecated message handlers awaiting removal
 */

func (server *Server) handleDisplayText(msg *osc.Message) {
	debug.Printf("handleText")
	var message displayTextMessage
	if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("handleText unmarshal: %v: %v", msg, err)
	} else {
		m := Message{
			Type:               "displayText",
			DisplayTextMessage: &message,
		}
		server.update(m)
	}
}

func (server *Server) handleCountupStart(msg *osc.Message) {
	debug.Printf("countup start: %#v", msg)

	if len(msg.Arguments) != 0 {
		log.Printf("handleCountupStart: too many arguments")
		return
	}

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

// Le huge registerHandler block
func (server *Server) setup(oscServer *osc.Server) {
	// Sync messages
	registerHandler(oscServer, "^/clock/media/*", server.handleMedia)
	registerHandler(oscServer, "^/clock/resetmedia/*", server.handleResetMedia)
	registerHandler(oscServer, "^/clock/ltc", server.handleLTC)

	// Timer related
	registerHandler(oscServer, "^/clock/timer/*/countdown/target", server.handleCountdownTarget)
	registerHandler(oscServer, "^/clock/timer/*/countdown$", server.handleCountdownStart)
	registerHandler(oscServer, "^/clock/timer/*/countup/target", server.handleCountupTarget)
	registerHandler(oscServer, "^/clock/timer/*/countup$", server.handleCountupStart)
	registerHandler(oscServer, "^/clock/timer/*/modify", server.handleTimerModify)
	registerHandler(oscServer, "^/clock/timer/*/signal", server.handleTimerSignal)
	registerHandler(oscServer, "^/clock/timer/*/stop", server.handleTimerStop)
	registerHandler(oscServer, "^/clock/timer/*/pause", server.handleTimerPause)
	registerHandler(oscServer, "^/clock/timer/*/resume", server.handleTimerResume)
	registerHandler(oscServer, "^/clock/pause", server.handlePause)
	registerHandler(oscServer, "^/clock/resume", server.handleResume)

	// Source related
	registerHandler(oscServer, "^/clock/source/*/hide", server.handleHide)
	registerHandler(oscServer, "^/clock/source/*/show", server.handleShow)
	registerHandler(oscServer, "^/clock/source/*/title", server.handleSourceTitle)
	registerHandler(oscServer, "^/clock/source/*/color", server.handleSourceColor)
	registerHandler(oscServer, "^/clock/hide", server.handleHideAll)
	registerHandler(oscServer, "^/clock/show", server.handleShowAll)

	// Misc commands
	registerHandler(oscServer, "^/clock/background", server.handleBackground)
	registerHandler(oscServer, "^/clock/info", server.handleInfo)
	registerHandler(oscServer, "^/clock/text", server.handleDisplayText)
	registerHandler(oscServer, "^/clock/titlecolors", server.handleTitleColors)
	registerHandler(oscServer, "^/clock/seconds/off", server.handleSecondsOff)
	registerHandler(oscServer, "^/clock/seconds/on", server.handleSecondsOn)
	registerHandler(oscServer, "^/clock/time/set", server.handleTimeSet)
	registerHandler(oscServer, "^/clock/flash", server.handleFlash)

	// Deprecated
	registerHandler(oscServer, "^/clock/dual/text", server.handleDualText)
	registerHandler(oscServer, "^/clock/kill", server.handleHideAll)
	registerHandler(oscServer, "^/clock/normal", server.handleShowAll)
	registerHandler(oscServer, "^/clock/countup/start", server.handleCountupStart)
	registerHandler(oscServer, "^/clock/countup/modify", server.handleTimerModify)
	registerHandler(oscServer, "^/clock/display", server.handleDisplay)
	registerHandler(oscServer, "^/clock/countdown/start", server.handleCountdownStart)
	registerHandler(oscServer, "^/clock/countdown2/start", server.handleCountdownStart)
	registerHandler(oscServer, "^/clock/countdown/modify", server.handleTimerModify)
	registerHandler(oscServer, "^/clock/countdown2/modify", server.handleTimerModify)
	registerHandler(oscServer, "^/clock/countdown/stop", server.handleTimerStop)
	registerHandler(oscServer, "^/clock/countdown2/stop", server.handleTimerStop)
}

func registerHandler(server *osc.Server, addr string, handler osc.HandlerFunc) {
	if err := server.Handle(addr, handler); err != nil {
		panic(err)
	}
}
