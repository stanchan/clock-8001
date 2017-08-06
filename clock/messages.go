package clock

import (
	"github.com/hypebeast/go-osc/osc"
)

// /qmsk/clock/count
type CountMessage struct {
	ColorRed   float32
	ColorGreen float32
	ColorBlue  float32
	Symbol     string
	Count      int32
}

func (message *CountMessage) UnmarshalOSC(msg *osc.Message) error {
	return msg.UnmarshalArguments(
		&message.ColorRed,
		&message.ColorGreen,
		&message.ColorBlue,
		&message.Symbol,
		&message.Count,
	)
}

func (message CountMessage) MarshalOSC(addr string) *osc.Message {
	return osc.NewMessage(addr,
		message.ColorRed,
		message.ColorGreen,
		message.ColorBlue,
		message.Symbol,
		message.Count,
	)
}
