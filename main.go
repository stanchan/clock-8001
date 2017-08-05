package main

import (
	"github.com/SpComb/osc-test/millumin"
	"github.com/hypebeast/go-osc/osc"
	"github.com/jessevdk/go-flags"
	"log"
)

var Options struct {
	ListenAddr string `long:"listen"`
	Debug      bool   `long:"debug"`
}

var parser = flags.NewParser(&Options, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		log.Fatalf("parse flags: %v", err)
	} else {
		log.Printf("options: %#v", Options)
	}

	var oscServer = osc.Server{
		Addr: Options.ListenAddr,
	}

	if Options.Debug {
		oscServer.Handle("*", func(msg *osc.Message) {
			osc.PrintMessage(msg)
		})
	}

	var milluminListener = millumin.MakeListener()

	milluminListener.Setup(&oscServer)

	log.Printf("osc server: listen %v", oscServer.Addr)

	if err := oscServer.ListenAndServe(); err != nil {
		log.Fatalf("osc server: %v", err)
	}
}
