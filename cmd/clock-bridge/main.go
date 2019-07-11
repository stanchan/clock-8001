package main

import (
	"fmt"
	"github.com/hypebeast/go-osc/osc"
	"github.com/jessevdk/go-flags"
	"gitlab.com/Depili/clock-8001/clock"
	"gitlab.com/Depili/clock-8001/debug"
	"gitlab.com/Depili/clock-8001/millumin"
	"gitlab.com/Depili/clock-8001/mitti"
	"log"
	"os"
	"regexp"
	"time"
)

var options struct {
	ClockClientOptions      clock.ClientOptions `group:"qmsk/osc-tally clock client"`
	ClockRemainingThreshold float32             `long:"clock-remaining-threshold" description:"Remaining time highlight threshold" default:"20"`
	Ignore                  string              `long:"millumin-ignore-layer" value-name:"REGEXP" description:"Ignore matching millumin layers (case-insensitive regexp)" default:"ignore"`
	ignoreRegexp            *regexp.Regexp
	ListenAddr              string `long:"osc-listen" description:"Address to listen for mitti/millumin osc messages" default:"0.0.0.0:1234"`
	Debug                   bool   `long:"osc-debug"`
}

var parser = flags.NewParser(&options, flags.Default)

func sendClockMessage(clockClient *clock.Client, remaining float32, playing bool) error {
	var clockCount = clock.CountMessage{}

	if !playing {
		clockCount = clock.CountMessage{
			ColorRed:   0,
			ColorGreen: 0,
			ColorBlue:  255,
			Symbol:     "Ⅱ",
		}
	} else if remaining > options.ClockRemainingThreshold {
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
		} else if options.ignoreRegexp.MatchString(layerState.Layer) {
			debug.Printf("Ignored layer update\n")
			continue
		} else if layerState.Updated.Before(time.Now().Add(-1*time.Second)) == true {
			debug.Printf("Layer information stale, ignored")
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
			log.Fatalf("Millumin: update clock: %v", err)
		} else if options.Debug {
			log.Printf("Millumin: update clock: %v\n", state)
		}
	}
}

func updateMittiClock(clockClient *clock.Client, state mitti.State) error {
	if !state.Loop {
		return sendClockMessage(clockClient, state.Remaining, state.Playing)
	}
	clockCount := clock.CountMessage{
		ColorRed:   0,
		ColorGreen: 255,
		ColorBlue:  0,
		Symbol:     "⇄",
	}
	clockCount.SetTimeRemaining(state.Remaining)
	return clockClient.SendCount(clockCount)
}

func runMittiClockClient(clockClient *clock.Client, listenChan chan mitti.State) {
	for state := range listenChan {
		// TODO: also refresh on tick
		if err := updateMittiClock(clockClient, state); err != nil {
			log.Fatalf("Mitti: update clock: %v", err)
		} else if options.Debug {
			log.Printf("Mitti: update clock: %v\n", state)
		}
	}
}

func startClockClient(milluminListener *millumin.Listener, mittiListener *mitti.Listener) error {
	client, err := options.ClockClientOptions.MakeClient()
	if err != nil {
		return err
	}

	go runMilluminClockClient(client, milluminListener.Listen())
	go runMittiClockClient(client, mittiListener.Listen())

	return nil
}

func run(oscServer *osc.Server) error {
	if options.Debug {
		oscServer.Handle("*", func(msg *osc.Message) {
			osc.PrintMessage(msg)
		})
	}

	var milluminListener = millumin.MakeListener(oscServer)
	var mittiListener = mitti.MakeListener(oscServer)

	if options.ClockClientOptions.Connect == "" {

	} else if err := startClockClient(milluminListener, mittiListener); err != nil {
		return fmt.Errorf("start clock client: %v", err)
	}

	log.Printf("Clock bridge: listening on %v", oscServer.Addr)

	return oscServer.ListenAndServe()
}

func main() {
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			panic(err)
		}
	}

	if options.Debug {
		debug.Enabled = true
	}

	regexp, err := regexp.Compile("(?i)" + options.Ignore)
	if err != nil {
		log.Fatalf("Invalid --millumin-ignore-layer=%v: %v", options.Ignore, err)
	}
	options.ignoreRegexp = regexp

	var oscServer = osc.Server{
		Addr: options.ListenAddr,
	}

	if err := run(&oscServer); err != nil {
		log.Fatalf("%v", err)
	}
}
