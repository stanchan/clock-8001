package millumin

import (
	"github.com/hypebeast/go-osc/osc"
	"log"
	"time"
)

type MediaState struct {
	Playing  bool
	Info     MediaInfo
	Duration float32
	Paused   bool
	Time     float32
}

func (state *MediaState) mediaStarted(mediaStarted MediaStarted) {
	state.Playing = true
	state.Info = mediaStarted.MediaInfo
	state.Duration = mediaStarted.MediaInfo.Duration
	state.Paused = false
	state.Time = 0.0

	log.Printf("Media started: %#v", state)
}

func (state *MediaState) mediaPaused(mediaPaused MediaPaused) {
	state.Paused = true
	state.Info = mediaPaused.MediaInfo
	state.Duration = mediaPaused.MediaInfo.Duration

	log.Printf("Media paused: %#v", state)
}

func (state *MediaState) mediaStopped(mediaStopped MediaStopped) {
	state.Playing = false
	state.Info = mediaStopped.MediaInfo
	state.Duration = mediaStopped.MediaInfo.Duration

	log.Printf("Media stopped: %#v", state)
}

func (state *MediaState) mediaTime(mediaTime MediaTime) {
	state.Duration = mediaTime.Duration
	state.Time = mediaTime.Value

	log.Printf("Media time: %#v", state)
}

type Listener struct {
	mediaState MediaState
	updateTime time.Time
}

func (listener *Listener) update() {
	listener.updateTime = time.Now()
}

func (listener *Listener) handleMediaStarted(msg *osc.Message) {
	var message MediaStarted

	if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		listener.mediaState.mediaStarted(message)
		listener.update()
	}
}

func (listener *Listener) handleMediaTime(msg *osc.Message) {
	var message MediaTime

	if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		listener.mediaState.mediaTime(message)
		listener.update()
	}
}

func (listener *Listener) handleMediaPaused(msg *osc.Message) {
	var message MediaPaused

	if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		listener.mediaState.mediaPaused(message)
		listener.update()
	}
}

func (listener *Listener) handleMediaStopped(msg *osc.Message) {
	var message MediaStopped

	if err := message.UnmarshalOSC(msg); err != nil {
		log.Printf("Unmarshal %v: %v", msg, err)
	} else {
		listener.mediaState.mediaStopped(message)
		listener.update()
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
