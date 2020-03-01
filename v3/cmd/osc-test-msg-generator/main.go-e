package main

import (
	"fmt"
	"github.com/hypebeast/go-osc/osc"
	"github.com/jessevdk/go-flags"
	"gitlab.com/Depili/clock-8001/clock"
	"gitlab.com/Depili/clock-8001/millumin"
	"log"
	"math/rand"
	"time"
)

var options struct {
	ClockClientOptions      clock.ClientOptions `group:"qmsk/osc-tally clock client"`
	ClockRemainingThreshold float32             `long:"clock-remaining-threshold" default:"20"`

	ListenAddr string `long:"osc-listen"`
	Debug      bool   `long:"osc-debug"`
}

var parser = flags.NewParser(&options, flags.Default)

func updateClock(clockClient *clock.Client, state millumin.State) error {
	// default is empty state when nothing is playing
	var clockCount = clock.CountMessage{}

	// XXX: select named layer, not first playing?
	for _, layerState := range state {
		if !layerState.Playing {
			continue
		}

		if layerState.Paused {
			clockCount = clock.CountMessage{
				ColorRed:   0,
				ColorGreen: 0,
				ColorBlue:  255,
				Symbol:     "~",
			}
		} else if layerState.Remaining() > options.ClockRemainingThreshold {
			clockCount = clock.CountMessage{
				ColorRed:   0,
				ColorGreen: 255,
				ColorBlue:  0,
				Symbol:     " ",
			}
		} else {
			clockCount = clock.CountMessage{
				ColorRed:   255,
				ColorGreen: 0,
				ColorBlue:  0,
				Symbol:     " ",
			}
		}

		clockCount.SetTimeRemaining(layerState.Remaining())

		break
	}

	return clockClient.SendCount(clockCount)
}

func runClockClient(clockClient *clock.Client, listenChan chan millumin.State) {
	t := time.Tick(1 * time.Second)
	for range t {
		var display = clock.DisplayMessage{
			ColorRed:   (rand.Float32() * 255),
			ColorGreen: (rand.Float32() * 255),
			ColorBlue:  (rand.Float32() * 255),
			Text:       "stop",
		}
		if err := clockClient.SendDisplay(display); err != nil {
			log.Fatalf("update clock: %v", err)
		} else {
			log.Printf("update clock")
		}

		/*
			var clockCount = clock.CountMessage{}
			clockCount = clock.CountMessage{
				ColorRed:   (rand.Float32() * 255),
				ColorGreen: (rand.Float32() * 255),
				ColorBlue:  (rand.Float32() * 255),
				Symbol:     "▶",
			}
			clockCount.SetTimeRemaining(rand.Float32() * 99 * 100)
			if err := clockClient.SendCount(clockCount); err != nil {
				log.Fatalf("update clock: %v", err)
			} else {
				log.Printf("update clock")
			}
		*/
		/*
			start := clock.CountdownMessage{
				Seconds: 100,
			}
			if err := clockClient.SendStart(start); err != nil {
				log.Fatalf("update clock: %v", err)
			} else {
				log.Printf("update clock")
			}
		*/
	}

	for state := range listenChan {
		// TODO: also refresh on tick
		if err := updateClock(clockClient, state); err != nil {
			log.Fatalf("update clock: %v", err)
		} else {
			log.Printf("update clock")
		}
	}
}

func startClockClient(milluminListener *millumin.Listener) error {
	client, err := options.ClockClientOptions.MakeClient()
	if err != nil {
		return err
	}

	go runClockClient(client, milluminListener.Listen())
	return nil
}

func run(oscServer *osc.Server) error {
	if options.Debug {
		oscServer.Handle("*", func(msg *osc.Message) {
			osc.PrintMessage(msg)
		})
	}

	var milluminListener = millumin.MakeListener(oscServer)

	if options.ClockClientOptions.Connect == "" {

	} else if err := startClockClient(milluminListener); err != nil {
		return fmt.Errorf("start clock client: %v", err)
	}

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
