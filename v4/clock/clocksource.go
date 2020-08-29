package clock

import (
	// "gitlab.com/Depili/clock-8001/v3/debug"
	// "image/color"
	"time"
)

const ()

type source struct {
	counter *Counter       // timer
	title   string         // title displayed on the screen if possible
	tz      *time.Location // timezone to use

	// Booleans controlling what might be displayed by this clock data source
	ltc   bool // LTC timecode decoded from sound input
	udp   bool // UDP time packets from the stagetimer protocol
	timer bool // Countdown / up timer
	tod   bool // Time of day, lowest priority
	off   bool // Master control to turn output off
}
