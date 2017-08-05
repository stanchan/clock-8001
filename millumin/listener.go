package millumin

import (
	"github.com/hypebeast/go-osc/osc"
	"log"
)

type Listener struct {
}

func (listener *Listener) handleMediaStarted(msg *osc.Message) {
	var message MediaStarted

	if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		log.Printf("Media started: %#v", message)
	}
}

func (listener *Listener) handleMediaTime(msg *osc.Message) {
	var message MediaTime

	if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		log.Printf("Media time: %#v", message)
	}
}

func (listener *Listener) handleMediaPaused(msg *osc.Message) {
	var message MediaPaused

	if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		log.Printf("Media paused: %#v", message)
	}
}

func (listener *Listener) handleMediaStopped(msg *osc.Message) {
	var message MediaStopped

	if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		log.Printf("Media stopped: %#v", message)
	}
}

func registerHandler(server *osc.Server, addr string, handler osc.HandlerFunc) {
	if err := server.Handle(addr, handler); err != nil {
		panic(err)
	}
}

func (listener *Listener) Setup(server *osc.Server) {
	registerHandler(server, "/millumin/layer:*/mediaStarted", listener.handleMediaStarted)
	registerHandler(server, "/millumin/layer:*/media/time", listener.handleMediaTime)
	registerHandler(server, "/millumin/layer:*/mediaPaused", listener.handleMediaPaused)
	registerHandler(server, "/millumin/layer:*/mediaStopped", listener.handleMediaStopped)
}
