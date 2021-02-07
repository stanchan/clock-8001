package clock

import (
	"github.com/hypebeast/go-osc/osc"
)

var clockUnits = []struct {
	unit    string
	seconds float32
}{
	{"s", 1},
	{"m", 60},
	{"h", 60 * 60},
	{"d", 24 * 60 * 60},
}

// Message is a generic clock message for decoded osc data
type Message struct {
	Type               string
	Counter            int
	Countdown          bool
	Data               string
	CountdownMessage   *CountdownMessage
	DisplayMessage     *DisplayMessage
	MediaMessage       *MediaMessage
	DisplayTextMessage *displayTextMessage
}

// MediaMessage contains data from media players
type MediaMessage struct {
	hours     int32
	minutes   int32
	seconds   int32
	frames    int32
	remaining int32
	progress  float64
	paused    bool
	looping   bool
	timeStamp osc.Timetag
}

// UnmarshalOSC converts a osc.Message to MediaMessage
func (message *MediaMessage) UnmarshalOSC(msg *osc.Message) error {
	return msg.UnmarshalArguments(
		&message.hours,
		&message.minutes,
		&message.seconds,
		&message.frames,
		&message.remaining,
		&message.progress,
		&message.paused,
		&message.looping,
		&message.timeStamp,
	)
}

// MarshalOSC converts a MediaMessage to osc.Message
func (message MediaMessage) MarshalOSC(addr string) *osc.Message {
	return osc.NewMessage(addr,
		message.hours,
		message.minutes,
		message.seconds,
		message.frames,
		message.remaining,
		message.progress,
		message.paused,
		message.looping,
	)
}

// CountdownMessage is for /clock/countdown/start
type CountdownMessage struct {
	Seconds int32
}

// DisplayMessage is for /clock/display
type DisplayMessage struct {
	ColorRed   float32
	ColorGreen float32
	ColorBlue  float32
	Text       string
}

// TimeMessage is for /clock/settime and /clock/ltc
type TimeMessage struct {
	Time string
}

type displayTextMessage struct {
	r    int32
	g    int32
	b    int32
	a    int32
	bgR  int32
	bgG  int32
	bgB  int32
	bgA  int32
	time int32
	text string
}

func (message *displayTextMessage) UnmarshalOSC(msg *osc.Message) error {
	return msg.UnmarshalArguments(
		&message.r,
		&message.g,
		&message.b,
		&message.a,
		&message.bgR,
		&message.bgG,
		&message.bgB,
		&message.bgA,
		&message.time,
		&message.text,
	)
}

// TextMessage is for text only messages like /clock/dual/text
type TextMessage struct {
	Text string
}

// UnmarshalOSC converts a osc.Message to TextMessage
func (message *TextMessage) UnmarshalOSC(msg *osc.Message) error {
	return msg.UnmarshalArguments(
		&message.Text,
	)
}

// MarshalOSC converts a TextMessage to osc.Message
func (message TextMessage) MarshalOSC(addr string) *osc.Message {
	return osc.NewMessage(addr,
		message.Text,
	)
}

// UnmarshalOSC converts a osc.Message to TimeMessage
func (message *TimeMessage) UnmarshalOSC(msg *osc.Message) error {
	return msg.UnmarshalArguments(
		&message.Time,
	)
}

// MarshalOSC converts a TimeMessage to osc.Message
func (message TimeMessage) MarshalOSC(addr string) *osc.Message {
	return osc.NewMessage(addr,
		message.Time,
	)
}

// UnmarshalOSC converts a osc.Message to DisplayMessage
func (message *DisplayMessage) UnmarshalOSC(msg *osc.Message) error {
	return msg.UnmarshalArguments(
		&message.ColorRed,
		&message.ColorGreen,
		&message.ColorBlue,
		&message.Text,
	)
}

// MarshalOSC converts a DisplayMessage to osc.Message
func (message DisplayMessage) MarshalOSC(addr string) *osc.Message {
	return osc.NewMessage(addr,
		message.ColorRed,
		message.ColorGreen,
		message.ColorBlue,
		message.Text,
	)
}

// UnmarshalOSC converts a osc.Message to CountdownMessage
func (message *CountdownMessage) UnmarshalOSC(msg *osc.Message) error {
	return msg.UnmarshalArguments(
		&message.Seconds,
	)
}

// MarshalOSC converts a CountdownMessage to osc.Message
func (message CountdownMessage) MarshalOSC(addr string) *osc.Message {
	return osc.NewMessage(addr,
		message.Seconds,
	)
}
