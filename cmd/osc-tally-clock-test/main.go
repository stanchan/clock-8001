package main

import (
	"fmt"
	"github.com/SpComb/osc-tally/clock"
	"github.com/hypebeast/go-osc/osc"
	"github.com/jessevdk/go-flags"
	"log"
)

var Options struct {
	ListenAddr string `long:"osc-listen"`
}

var parser = flags.NewParser(&Options, flags.Default)

func listener(listenChan chan clock.CountMessage) {
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
		log.Printf("options: %#v", Options)
	}

	var oscServer = osc.Server{
		Addr: Options.ListenAddr,
	}

	if err := run(&oscServer); err != nil {
		log.Fatalf("%v", err)
	}
}
