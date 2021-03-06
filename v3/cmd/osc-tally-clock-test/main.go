package main

import (
	"fmt"
	"github.com/hypebeast/go-osc/osc"
	"github.com/jessevdk/go-flags"
	"github.com/stanchan/clock-8001/v3/clock"
	"log"
)

var options struct {
	ListenAddr string `long:"osc-listen"`
}

var parser = flags.NewParser(&options, flags.Default)

func listener(listenChan chan clock.Message) {
	for countMessage := range listenChan {
		fmt.Printf("%#v\n", countMessage)
	}
}

func run(oscServer *osc.Server) error {
	var clockServer = clock.MakeServer(oscServer)

	go listener(clockServer.Listen())

	log.Printf("osc server: listen %v", oscServer.Addr)

	return oscServer.ListenAndServe()
}

func main() {
	if _, err := parser.Parse(); err != nil {
		log.Fatalf("parse flags: %v", err)
	} else {
		log.Printf("options: %#v", options)
	}

	var oscServer = osc.Server{
		Addr: options.ListenAddr,
	}

	if err := run(&oscServer); err != nil {
		log.Fatalf("%v", err)
	}
}
