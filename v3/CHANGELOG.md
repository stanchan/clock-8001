## Version 3.8.0
* multi-clock: Implement new display of up to 40 clocks
* config file support: use -C to read config flags from a file
* osx support: Remove unused rpi gpio stuff from sdl-clock and handle sdl2 events

## Version 3.7.1
* clock-bridge:
  * BUGFIX: Millumin osc tally and layer renaming. Renaming an active layer caused the internal state to become corrupt. The clock-bridge now filters layer states that haven't been updated in 1 second to remove ghost layers from the renames.

## Version 3.7.0
* BUGFIX: Correct the monitor pixel aspect ratio on the official raspberry pi display. The clock is now round instead of oval.
* FEATURE: Detect if a monitor is rotated 90 or 270 degrees and move the clock to the top portion of the screen.

## Version 3.6.2
* BUGFIX: Make /clock/time/set timezone aware.

## Version 3.6.1
* BUGFIX: /clock/time/set OSC command and busybox date command.

## Version 3.6.0
* Adds /clock/time/set osc command for setting system time, see README.md for details

## Version 3.5.1
* BUFIX: build issue on 3.5.0

## Version 3.5.0
* OSC commands for hiding/showing the numerical seconds field
  * /clock/seconds/off - hides the second display
  * /clock/seconds/on - shows the second display
* Add --debug command line parameter to enable verbose logging

## Version 3.4.1
* Greatly reduced the amount of logging that gets printed on the console. This caused slowdowns on raspberry pi

## Version 3.4.0
* Automatically rescan network interfaces and addresses every 5 seconds
  * This removes the race condition with dhcp lease on startup
  * Now if feedback address is 255.255.255.255 the feedback is sent to broadcast on all configured ipv4 interfaces
* Added --disable-feedback command line switch
  * This disables the osc feedback but the clock will still listen for osc commands

## Version 3.3.0
* Add --disable-osc command line parameter

## Version 3.2.0
* Implemented resolution scaling.
  * This adds support for the official raspberry pi 7" display
* Fixes to the 192x192 small mode

## Version 3.1.1rc1
* CI environment now builds raspberry pi sd card images

## Version 3.1.0
* Clock now shows its version number before acquiring the correct time

## Version 3.0.2
* BUGIX: "tally" text formatting for /qmsk/clock/count osc messages

## Version 3.0.1
* Implement CI environment for testing and automated builds

## Version 3.0.0
* Breaking change to OSC feedback, clocks now send paused/running state
* Countdowns can now be paused and resumed with /clock/pause and /clock/resume

## Version 2.1.0
* Disable flashing when countdown reaches zero with --flash=0 parameter
* Print version to console when started

## Version 2.0.0
* Added countdown and count up timer modes
* Added OSC commands for timers
* Added OSC feedback of clock mode and display date