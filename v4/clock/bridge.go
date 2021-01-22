package clock

import (
	"fmt"
	"github.com/desertbit/timer"
	"github.com/hypebeast/go-osc/osc"
	"gitlab.com/Depili/clock-8001/v4/debug"
	"gitlab.com/Depili/clock-8001/v4/millumin"
	"gitlab.com/Depili/clock-8001/v4/mitti"
	"log"
	"time"
)

const (
	clockRemainingThreshold = 20
	updateTimeout           = 1000 * time.Millisecond
)

func (engine *Engine) oscBridge() error {
	var milluminListener = millumin.MakeListener(&engine.oscServer)
	var mittiListener = mitti.MakeListener(&engine.oscServer)

	if err := engine.startClockClient(milluminListener, mittiListener); err != nil {
		return fmt.Errorf("start clock client: %v", err)
	}

	log.Printf("Clock bridge: listening on %v", engine.oscServer.Addr)
	return nil
}

func (engine *Engine) updateMilluminClock(state millumin.State) error {
	var err error

	// XXX: select named layer, not first playing?
	for _, layerState := range state {
		if !layerState.Playing {
			continue
		} else if engine.ignoreRegexp.MatchString(layerState.Layer) {
			debug.Printf("Ignored layer update\n")
			continue
		} else if layerState.Updated.Before(time.Now().Add(-1*time.Second)) == true {
			debug.Printf("Layer information stale, ignored")
			continue
		}

		// Fix one second offset with millumin time remaining...
		remaining := (time.Duration(layerState.Remaining()) + 1) * time.Second
		total := time.Duration(layerState.Duration) * time.Second

		hours := int32(remaining.Truncate(time.Hour).Hours())
		minutes := int32(remaining.Truncate(time.Minute).Minutes()) - (hours * 60)
		seconds := int32(remaining.Truncate(time.Second).Seconds()) - (((hours * 60) + minutes) * 60)

		progress := float64(remaining) / float64(total)

		engine.milluminCounter.SetMedia(hours, minutes, seconds, 0, remaining, progress, layerState.Paused, false)
		engine.sendMedia("millumin", hours, minutes, seconds, 0, int32(layerState.Remaining()+1), progress, layerState.Paused, false)

		return nil
	}
	// No playing media found
	engine.milluminCounter.ResetMedia()

	// FIXME: send no media to others
	engine.sendResetMedia("millumin")

	return err
}

func (engine *Engine) runMilluminClockClient(listenChan chan millumin.State) {
	for state := range listenChan {
		// TODO: also refresh on tick
		if err := engine.updateMilluminClock(state); err != nil {
			log.Fatalf("Millumin: update clock: %v", err)
		} else {
			debug.Printf("Millumin: update clock: %v\n", state)
		}
	}
}

func (engine *Engine) updateMittiClock(state mitti.State) error {
	// FIXME: need to fudge this by one second to get the displays to agree?
	remaining := time.Duration(state.Remaining) * time.Second
	total := remaining + (time.Duration(state.Elapsed) * time.Second)
	progress := float64(remaining) / float64(total)

	hours := int32(state.Hours)
	minutes := int32(state.Minutes)
	seconds := int32(state.Seconds)
	frames := int32(state.Frames)

	debug.Printf("Mitti update, remaining: %v total: %v\n", remaining.Seconds(), total.Seconds())

	debug.Printf(" -> update state: %02d:%02d:%02d", state.Hours, state.Minutes, state.Seconds)
	engine.mittiCounter.SetMedia(hours, minutes, seconds, frames, remaining, progress, state.Paused, state.Loop)
	engine.sendMedia("mitti", hours, minutes, seconds, frames, int32(state.Remaining), progress, state.Paused, state.Loop)

	/* TODO: loop?
	clockCount := CountMessage{
		ColorRed:   0,
		ColorGreen: 255,
		ColorBlue:  0,
		Symbol:     "⇄",
	}
	clockCount.SetTimeRemaining(state.Remaining)
	*/

	return nil
}

func (engine *Engine) runMittiClockClient(listenChan chan mitti.State) {
	timeout := timer.NewTimer(updateTimeout)
	for {
		select {
		case state := <-listenChan:
			timeout.Reset(updateTimeout)
			// TODO: also refresh on tick
			if err := engine.updateMittiClock(state); err != nil {
				log.Fatalf("Mitti: update clock: %v", err)
			} else {
				debug.Printf("Mitti: update clock: %v\n", state)
			}
		case <-timeout.C:
			engine.mittiCounter.ResetMedia()
		}
	}
}

func (engine *Engine) startClockClient(milluminListener *millumin.Listener, mittiListener *mitti.Listener) error {
	go engine.runMilluminClockClient(milluminListener.Listen())
	go engine.runMittiClockClient(mittiListener.Listen())

	return nil
}

func (engine *Engine) sendMedia(player string, hours, minutes, seconds, frames, remaining int32, progress float64, paused, looping bool) error {
	if engine.oscDests == nil {
		// No osc connection
		return nil
	}

	address := fmt.Sprintf("/clock/media/%s", player)

	packet := osc.NewMessage(address, hours, minutes, seconds, frames, remaining, progress, paused, looping)

	data, err := packet.MarshalBinary()
	if err != nil {
		return err
	}
	engine.oscDests.Write(data)

	return nil
}

func (engine *Engine) sendResetMedia(player string) error {
	if engine.oscDests == nil {
		// No osc connection
		return nil
	}

	address := fmt.Sprintf("/clock/resetmedia/%s", player)

	packet := osc.NewMessage(address)

	data, err := packet.MarshalBinary()
	if err != nil {
		return err
	}
	engine.oscDests.Write(data)

	return nil

}
