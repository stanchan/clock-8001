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

// Generic clock message
type ClockMessage struct {
	Type             string
	CountMessage     *CountMessage
	CountdownMessage *CountdownMessage
	DisplayMessage   *DisplayMessage
}

// /qmsk/clock/count
type CountMessage struct {
	ColorRed   float32
	ColorGreen float32
	ColorBlue  float32
	Symbol     string
	Count      int32
	Unit       string
}

// /clock/countdown/start
type CountdownMessage struct {
	Seconds int32
}

// /clock/display
type DisplayMessage struct {
	ColorRed   float32
	ColorGreen float32
	ColorBlue  float32
	Text       string
}

func (message *DisplayMessage) UnmarshalOSC(msg *osc.Message) error {
	return msg.UnmarshalArguments(
		&message.ColorRed,
		&message.ColorGreen,
		&message.ColorBlue,
		&message.Text,
	)
}

func (message DisplayMessage) MarshalOSC(addr string) *osc.Message {
	return osc.NewMessage(addr,
		message.ColorRed,
		message.ColorGreen,
		message.ColorBlue,
		message.Text,
	)
}

func (message *CountdownMessage) UnmarshalOSC(msg *osc.Message) error {
	return msg.UnmarshalArguments(
		&message.Seconds,
	)
}

func (message CountdownMessage) MarshalOSC(addr string) *osc.Message {
	return osc.NewMessage(addr,
		message.Seconds,
	)
}

func (message *CountMessage) SetTimeRemaining(seconds float32) {
	for _, unit := range clockUnits {
		if seconds/unit.seconds >= 100 {
			continue
		}

		message.Unit = unit.unit
		message.Count = int32(seconds/unit.seconds + 0.5) // round

		return
	}

	message.Unit = "+"
	message.Count = 0
}

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
