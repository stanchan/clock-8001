## Version 4.6.0
* Features:
  * 144x144px round clock face
  * 288x144px text clock face
  * Improved signal color handling, You can now override the automation set colors via osc commands
    * If enabled, automation does one color change when timer thresholds are reached, but doesn't prevent manually setting the signal colors.


## Version 4.5.2
* Bugfix: /clock/signal/* OSC command was expecting 4 parameters

## Version 4.5.1
* Bugfixes:
  * Config generation for overtime colors
  * Round clock and overtime visibility modes

## Version 4.5.0
* New: Customizable overtime behaviour
  * Handling of overtime countdowns can now be customized
  * Readout can be chosen to be either: show zeros, show blank, continue counting up
  * The extra visibility can be chosen to be: none, blink, change background color or change both background and blink

## Version 4.4.1
* Bugfix: Load the SPI kernel modules in the clock image for unicorn hats

## Version 4.4.0
* New: Initial hardware signal light support
  * Currently only pimoroni Unicorn HD and Ubercorn hats are supported
  * You can control the color via OSC by groups, or the light color can follow source 1

## Version 4.3.1
* Bugfix for the signal colors

## Version 4.3.0
* New: Color signals for timers
  * `/clock/timer/*/signal` OSC command
  * Optional automation for the signal colors based on timer thresholds

## Version 4.2.0
* New:
  * /clock/flash command, it flashes the screen white for 200ms

## Version 4.1.1
* Fix version information stamping in build automation

## Version 4.1.0
* New:
  * Support for hyperpixel4 square displays
  * Documentation for the displays in README.md

## Version 4.0.3
* Bugfixes:
  * Unhide sources when starting a targeted count

## Version 4.0.2
* Bugfixes:
  * Hide behaviour of the different osc commands to be uniform
  * Starting a timer to given target while clock is paused

## Version 4.0.1
* Fixed hyperpixel4-square configuration

## Version 4.0.0
### New feature highlights

* Support for different clock faces
  * Single round clock
  * Dual round clocks
  * Text clocks
    * 3 and 1 clocks per screen
* New concept of sources and timers
  * A clock face displays a clock source, which can be associated to a timer and different sources for the time, eg. LTC
  * Timers can be used on multiple clock setups
* HTTP configuration editor with input validation
  * import / export for the whole configuration file
* New OSC commands, see osc.md
  * New companion module version supporting the V4 commands and feedback
  * Background selection
  * Timers targeting a certain time-of-day
  * Sending of text messages with custom text and background colors
  * Setting of clock source labels
* Info overlay shown on startup containing version, ip and port information
* Support of LTC input via usb-soundcard / hifiberry dac+ adc pro hat
* Mitti and millumin message processing completely built in
  * You can choose which timers are the destinations of the playback state
* Support for sending and receiving timers as Interspace Industries Countdown2 UDP messages for display units
  * Also supported by StageTimer2 and Irisdown

### Breaking changes

* clock configuration file contents have been changed.
* OSC commands have been changed, compatibility with v3 commands has been kept as close as possible
  * V3 commands will be dropped at a later date
* OSC feedback has been changed
* Command line arguments have been changed
* New dependencies; sdl-ttf, sdl-image, ttf fonts for text rendering

## Version 3.16.3
* BUGFIX: Do not unpause counters when started

## Version 3.16.2
* BUGFIX: Timer pause, resume and turning clock off
* BUGFIX: /clock/countup/modify direction flip to be more natural

## Version 3.16.1
* Fix a linter error preventing automatic builds

## Version 3.16.0
* Start refactoring internal clock engine to implement new clock faces
* Add /clock/countup/modify OSC command
* Add /boot/config.txt editing to the web configuration

## Version 3.15.0
* Add experimental support for Interspace Industries USB AiO soundcard for LTC to the generated images

## Version 3.14.1
* BUGFIX: Clock scaling on the official 7" display

## Version 3.14.0
* Experimental background image support

## Version 3.13.0
* Add an option to disable detection and aspect ratio correction of official raspberry pi displays
* BUGFIX: setting 12h format from the http interface didn't work
* Internal refactoring and code cleanup

## Version 3.12.1
* BUGFIX: missing colon from default clock.ini for port 80

## Version 3.12.0
* Add support for 12 hour format on time-of-day display

## Version 3.11.1
* LTC Bugfix, we had a wrong audio device in the default configuration

## Version 3.11.0
* LTC added options:
  * Toggle between frames and seconds on the clock ring
  * Toggle for loss of signal handling; either blank the clock or continue on internal time
  * Toggle for disabling the LTC handling
* Images
  * Build raspberry pi 2 compatible images
  * Use config file generated by the build automation

## Version 3.10.2
* LTC reception & Hifiberry ADC+ DAC Pro
  * Bugfix on sdl-clock to display the hour part of LTC timestamp
  * Added configuration for the hifiberry to the sd-card image

## Version 3.10.1
* Little more black space between dual clock text display "pixels"


## Version 3.10.0
* Add /clock/dual/text OSC message for setting a text field on dual clock mode

## Version 3.9.0
* Build automation:
  * Build one unified image
* Clock:
  * Web configuration interface
  * Dual-clock mode

## Version 3.8.2
* Build automation improvements
  * Read version information from runtime/debug.ReadBuildInfo()
  * Include pimoroni hyperpixel4 square display support

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
