# OSC controlled simple clock

This is a simplistic clock written in go that can be used either as a video out with SDL or as a dedicated clock buidt with a 32x32 pixel hub75 led matrix and a ring of 60 addressable leds.

## SDL clock

You can build the clock binary with `go get gitlab.com/Depili/clock-8001/cmd/sdl_clock`. Compiling requires SDL 2 and SDL_GFX 2 libraries. On the raspberry pi the default libraries shipped with rasbian will only output data to X11 window, so for full screen dedicated clock you need to compile the SDL libraries from source. For compiling use `./configure --host=armv7l-raspberry-linux-gnueabihf --disable-pulseaudio --disable-esd --disable-video-mir --disable-video-wayland --disable-video-x11 --disable-video-opengl` for config flags.

## Dedicated led matrix clock

Bill of materials:
* Raspberry pi
* 32x32 pixel 4mm pixel pitch led matrix
* Led ring with 60 ws2812b leds
* Arduino (nano recommend)
* Adapter hat for the raspberry pi to connect to the led matrix
* 5V 3A power supply
* 12 leds of your choice for the static "hour" markers and current limiting resistors

You need to compile https://gitlab.com/Depili/rpi-matrix for a small program that will listen on udp socket for the led matrix data and handle driving the led matrix.

Compile the led matrix clock binary with `go get gitlab.com/Depili/clock-8001/cmd/clock`

## Mitti and Millumin osc-converter

To convert timecodes and video information sent by Mitti or Millumin to commands understood by the clock use `go get gitlab.com/Depili/clock-8001`. This can also be used to bridge the osc traffic across different networks.

## OSC commands understood by the clock

### /clock/tally and /qmsk/clock/count (legacy)

Payload:
1. float32 red component of the text color
2. float32 green component of the text color
3. float32 blue component of the text color
4. string single character symbol to display before the time
5. int32 the time, 0-99, will be displayed as two characters
6. string single character for the time unit (h, m, s)

### /clock/display

Displays up to 4 characters above the main time display for the time specified on the clock -d command line parameter (default 1000ms).

Payload:
1. float32 Red component of the text color
2. float32 Green component of the text color
3. float32 Blue component of the text color
4. string up to 4 characters to display

### /clock/countdown/start

Starts a countdown timer with the duration from the payload.

Payload:
1. int32 timer duration in seconds

### /clock/countdown2/start

Starts a secondary countdown above the main clock display. This is the same area as is used by the "tally" display. The countdown has lower priority than the tally.

Payload:
1. int32 timer duration in seconds

### /clock/countdown/modify & /clock/countdown2/modify

Modifies the duration of running countdown.

Payload:
1. int32 time in seconds to add or substract from the running timer

### /clock/countdown/stop & /clock/countdown2/stop

Stops the countdown. The killed countdown will vanish from the clock display. To restore time display issue /clock/normal command.

### /clock/countup/start

Starts counting up time.

Payload: none

### /clock/kill

(Almost) blanks the display. Only the 12 static leds and one led on the ring will be on.

### /clock/normal

Returns the clock to normal mode displaying current time.