package main

import (
	"github.com/stanchan/ubercorn/unicorn"
	"log"
)

type unicornSignal struct{}

func (us *unicornSignal) Init() {
	unicorn, err := unicorn.Init()
	if err != nil {
		log.Printf("Error initializing unicorn support: %v", err)
		return
	}
	log.Printf("Initialized Pimoroni Unicorn-HD support")
	signalHardware = unicorn
}

func init() {
	signalHardwareList["unicorn-hd"] = &unicornSignal{}
}
