package main

import (
	"fmt"
	"github.com/SpComb/osc-tally/clock"
	"github.com/SpComb/osc-tally/millumin"
	"github.com/SpComb/osc-tally/mitti"
	"github.com/hypebeast/go-osc/osc"
	"github.com/jessevdk/go-flags"
	"log"
	"regexp"
)

var Options struct {
	ClockClientOptions      clock.ClientOptions `group:"qmsk/osc-tally clock client"`
	ClockRemainingThreshold float32             `long:"clock-remaining-threshold" default:"20"`
	Ignore                  string              `long:"millumin-ignore-layer" value-name:"REGEXP" description:"Ignore matching millumin layers (case-insensitive regexp)" default:"ignore"`
	ignoreRegexp            *regexp.Regexp
	ListenAddr              string `long:"osc-listen"`
	Debug                   bool   `long:"osc-debug"`
}

var parser = flags.NewParser(&Options, flags.Default)

func sendClockMessage(clockClient *clock.Client, remaining float32, playing bool) error {
	var clockCount = clock.CountMessage{}

	if !playing {
		clockCount = clock.CountMessage{
			ColorRed:   0,
			ColorGreen: 0,
			ColorBlue:  255,
			Symbol:     "Ⅱ",
		}
	} else if remaining > Options.ClockRemainingThreshold {
		clockCount = clock.CountMessage{
			ColorRed:   0,
			ColorGreen: 255,
			ColorBlue:  0,
			Symbol:     "▶",
		}
	} else {
		clockCount = clock.CountMessage{
			ColorRed:   255,
			ColorGreen: 255,
			ColorBlue:  255,
			Symbol:     "▶",
		}
	}
	clockCount.SetTimeRemaining(remaining)
	return clockClient.SendCount(clockCount)
}

func updateMilluminClock(clockClient *clock.Client, state millumin.State) error {
	var err error

	// XXX: select named layer, not first playing?
	for _, layerState := range state {
		if !layerState.Playing {
			continue
		} else if Options.ignoreRegexp.MatchString(layerState.Layer) {
			continue
		}

		err = sendClockMessage(clockClient, layerState.Remaining(), !layerState.Paused)
		break
	}
	return err
}

func runMilluminClockClient(clockClient *clock.Client, listenChan chan millumin.State) {
	for state := range listenChan {
		// TODO: also refresh on tick
		if err := updateMilluminClock(clockClient, state); err != nil {
			log.Fatalf("update clock: %v", err)
		} else {
			log.Printf("update clock")
		}
	}
}

func updateMittiClock(clockClient *clock.Client, state mitti.State) error {
	return sendClockMessage(clockClient, state.Remaining, state.Playing)
}

func runMittiClockClient(clockClient *clock.Client, listenChan chan mitti.State) {
	for state := range listenChan {
		// TODO: also refresh on tick
		if err := updateMittiClock(clockClient, state); err != nil {
			log.Fatalf("update clock: %v", err)
		} else {
			log.Printf("update clock")
		}
	}
}

func startClockClient(milluminListener *millumin.Listener, mittiListener *mitti.Listener) error {
	client, err := Options.ClockClientOptions.MakeClient()
	if err != nil {
		return err
	} else {

	}

	go runMilluminClockClient(client, milluminListener.Listen())
	go runMittiClockClient(client, mittiListener.Listen())

	return nil
}

func run(oscServer *osc.Server) error {
	if Options.Debug {
		oscServer.Handle("*", func(msg *osc.Message) {
			osc.PrintMessage(msg)
		})
	}

	var milluminListener = millumin.MakeListener(oscServer)
	var mittiListener = mitti.MakeListener(oscServer)

	if Options.ClockClientOptions.Connect == "" {

	} else if err := startClockClient(milluminListener, mittiListener); err != nil {
		return fmt.Errorf("start clock client: %v", err)
	}

	log.Printf("osc server: listen %v", oscServer.Addr)

	return oscServer.ListenAndServe()
}

func main() {
	if _, err := parser.Parse(); err != nil {
		log.Fatalf("parse flags: %v", err)
	} else {
		log.Printf("options: %#v", Options)
	}

	regexp, err := regexp.Compile("(?i)" + Options.Ignore)
	if err != nil {
		log.Fatalf("Invalid --millumin-ignore-layer=%v: %v", Options.Ignore, err)
	}
	Options.ignoreRegexp = regexp

	var oscServer = osc.Server{
		Addr: Options.ListenAddr,
	}

	if err := run(&oscServer); err != nil {
		log.Fatalf("%v", err)
	}
}
