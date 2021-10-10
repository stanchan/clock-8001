package main

import (
	"fmt"
	"github.com/stanchan/go-osc/osc"
	"time"
)

func main() {
	client := osc.NewClient("255.255.255.255", 1245)
	frames := 0
	for {
		msg := osc.NewMessage("/clock/ltc")
		t := time.Now().Format("03:04:05")
		text := fmt.Sprintf("%s:%02d", t, frames)
		msg.Append(text)
		client.Send(msg)
		frames++
		frames = frames % 60
		time.Sleep(time.Second / 30)
	}
}
