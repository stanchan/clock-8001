package debug

import (
	"log"
)

// Enabled controls if debug messages are printed out to stdout or not
var Enabled = false // Set to true to enable debug output

// Printf prints a debug message via fmt.Printf if debug messages are enabled
func Printf(format string, v ...interface{}) {
	if Enabled {
		log.Printf(format, v...)
	}
}
