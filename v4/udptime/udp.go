package udptime

import (
	"errors"
	"fmt"
	"log"
	"net"
)

// This is an implementation of the 3-byte udp protocol used by irisdown and stage timer 2
// Byte 0: bits 0-3 - BCD "Minutes" most significant digit or negative sign (0xA)
// Byte 0: bits 4-7 - BCD "Minutes" least significant digit
// Byte 1: bits 0-3 - BCD "Seconds" most significant digit
// Byte 1: bits 4-7 - BCD "Seconds" least significant digit
// Byte 2: bit 0 - Green color flag
// Byte 2: bit 1 - Red color flag

// Colors
const (
	Off    = 0x00
	Green  = 0x01
	Red    = 0x02
	Orange = 0x03
)

// Message is a struct for upd time commands from stage timer 2 and irisdown
type Message struct {
	Minutes  int
	Seconds  int
	Green    bool
	Red      bool
	OverTime bool
}

func (msg *Message) String() string {
	return fmt.Sprintf("%02d:%02d Overtime: %v Green: %v Red %v", msg.OverTime, msg.Minutes, msg.Seconds, msg.Green, msg.Red)
}

// Listen for udptime messages on a port
func Listen(addr string) (chan *Message, error) {
	ch := make(chan *Message)
	pc, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, err
	}

	go server(pc, ch)
	return ch, nil
}

func server(pc net.PacketConn, ch chan *Message) {
	buffer := make([]byte, 3)
	for {
		n, _, err := pc.ReadFrom(buffer)
		if err != nil {
			log.Printf("Error listening for packets: %v\n", err)
			return
		}
		if n == 3 {
			msg, err := Decode(buffer)
			if err == nil {
				ch <- msg
			} else {
				log.Printf("Error decoding message: %v\n", err)
			}
		} else {
			log.Printf("Wrong number of bytes received, wanted 3, got %v\n", n)
		}
	}
}

// Decode takes the raw 3 bytes from the udp time message and decodes it into Message struct
func Decode(data []byte) (*Message, error) {
	if len(data) != 3 {
		return nil, errors.New("Data size is not 3 bytes")
	}
	msg := Message{
		Green: false,
		Red:   false,
	}

	if data[0]&0x0a == 0x0a {
		// Negative value, ignore the msd.
		msg.OverTime = true
		msg.Minutes = int(parseBCDByte(data[0] & 0xF0))
	} else {
		msg.Minutes = int(parseBCDByte(data[0]))
	}
	msg.Seconds = int(parseBCDByte(data[1]))
	if data[2]&Green == Green {
		msg.Green = true
	}
	if data[2]&Red == Red {
		msg.Red = true
	}
	return &msg, nil
}

// Encode takes a Message struct and encodes it into 3 byte message for the udptime protocol
func Encode(msg *Message) []byte {
	data := make([]byte, 3)
	minutes := msg.Minutes
	seconds := msg.Seconds

	// Clamp the max values
	if minutes > 99 {
		minutes = 99
	} else if minutes < -9 {
		minutes = -9
	}

	if minutes < 0 {
		data[0] = encodeBCD(-minutes)&0xF0 | 0x0A
	} else {
		data[0] = encodeBCD(minutes)
	}
	data[1] = encodeBCD(seconds)
	data[2] = 0
	if msg.Green {
		data[2] |= Green
	}
	if msg.Red {
		data[2] |= Red
	}
	return data
}

func encodeBCD(value int) byte {
	return byte((value % 10 << 4) + (value / 10))
}

func parseBCDByte(b byte) byte {
	value := (b & 0x0F) * 10
	value += (b >> 4)
	return value
}
