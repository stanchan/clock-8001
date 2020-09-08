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
	Type             string
	Counter          int
	Countdown        bool
	Data             string
	CountMessage     *CountMessage
	CountdownMessage *CountdownMessage
	DisplayMessage   *DisplayMessage
}

// CountMessage is a legacy message format for /qmsk/clock/count
type CountMessage struct {
	ColorRed   float32
	ColorGreen float32
	ColorBlue  float32
	Symbol     string
	Count      int32
	Unit       string
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

// SetTimeRemaining sets the remaining time for CountMessage and formats the unit
func (message *CountMessage) SetTimeRemaining(seconds float32) {
	for _, unit := range clockUnits {
		if seconds/unit.seconds >= 100 {
			continue
		}

		message.Unit = unit.unit
		message.Count = int32(seconds / unit.seconds)
		return
	}

	message.Unit = "+"
	message.Count = 0
}

// UnmarshalOSC converts a osc.Message to CountMessage
func (message *CountMessage) UnmarshalOSC(msg *osc.Message) error {
	return msg.UnmarshalArguments(
		&message.ColorRed,
		&message.ColorGreen,
		&message.ColorBlue,
		&message.Symbol,
		&message.Count,
		&message.Unit,
	)
}

// MarshalOSC converts a CountMessage to osc.Message
func (message CountMessage) MarshalOSC(addr string) *osc.Message {
	return osc.NewMessage(addr,
		message.ColorRed,
		message.ColorGreen,
		message.ColorBlue,
		message.Symbol,
		message.Count,
		message.Unit,
	)
}
