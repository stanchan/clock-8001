# OSC controlled simple clock

This is a simplistic clock written in go that can be used either as a video out with SDL or as a dedicated clock buidt with a 32x32 pixel hub75 led matrix and a ring of 60 addressable leds.

The clock can be controlled with the depili-clock-8001 companion module: https://github.com/bitfocus/companion-module-depili-clock-8001

Developed in co-operation with [SVV](http://svv.fi/).

## SDL clock

You can build the clock binary with `go get gitlab.com/Depili/clock-8001/cmd/sdl_clock`. Compiling requires SDL 2 and SDL_GFX 2 libraries. On the raspberry pi the default libraries shipped with rasbian will only output data to X11 window, so for full screen dedicated clock you need to compile the SDL libraries from source. For compiling use `./configure --host=armv7l-raspberry-linux-gnueabihf --disable-pulseaudio --disable-esd --disable-video-mir --disable-video-wayland --disable-video-x11 --disable-video-opengl` for config flags.

### Ready made raspberry pi images

The images support raspberry pi 2B / 3B / 3B+ boards. They need at least 64Mb SD-cards. Write them to the card like any other raspberry pi sd-card image.

https://kissa.depili.fi/clock-8001/sdl-clock_v2.1-clockworkadmin.img Is a image with root logins (ssh and local) enabled. The root password is `clockworkadmin`. Since this image uses a known hard coded password it should be for testing only and considered insecure.

https://kissa.depili.fi/clock-8001/sdl-clock_v2.1-no_login.img This image has root logins disabled and is secure for production use.

#### Customizing the images

You can place the following files on the sd-card FAT partition to customize the installation:
* `hostname` to change the hostname used by the clock, it is available with "hostname.local" for bonjour / mDNS requests
* `sdl-clock` to update the clock binary with this file
* `clock_cmd.sh` is the command line for the clock, it should start with `/root/sdl-clock ` and be followed by any command line parameters you wish to use
* `interfaces` a replacement for /etc/network/interfaces for custom network configuration
* `ntp.conf` for custom ntp server configuration
* `config.sys` the normal raspberry pi boot configuration for changing video modes etc.

### Command line parameters
```
Usage:
  sdl-clock [OPTIONS]

Application Options:
  -s                  Scale to 192x192px
  -F, --font=         Font for event name (default: fonts/7x13.bdf)
  -r, --red=          Red component of text color (default: 255)
  -g, --green=        Green component of text color (default: 128)
  -b, --blue=         Blue component of text color (default: 0)
      --static-red=   Red component of static color (default: 80)
      --static-green= Green component of static color (default: 80)
      --static-blue=  Blue component of static color (default: 0)
      --sec-red=      Red component of second color (default: 200)
      --sec-green=    Green component of second color (default: 0)
      --sec-blue=     Blue component of second color (default: 0)
  -p, --time-pin=     Pin to select foreign timezone, active low (default: 15)
      --flash=        Flashing interval when countdown reached zero (ms), 0 disables (default: 500)
  -t, --local-time=   Local timezone (default: Europe/Helsinki)
      --osc-listen=   Address to listen for incoming osc messages (default: 0.0.0.0:1245)
  -d, --timeout=      Timeout for OSC message updates in milliseconds (default: 1000)
  -o, --osc-dest=     Address to send OSC feedback to (default: 255.255.255.255:1245)
      --cd-red=       Red component of secondary countdown color (default: 255)
      --cd-green=     Green component of secondary countdown color (default: 0)
      --cd-blue=      Blue component of secondary countdown color (default: 0)

Help Options:
  -h, --help          Show this help message
```

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

Payload: none

### /clock/countup/start

Starts counting up time.

Payload: none

### /clock/kill

(Almost) blanks the display. Only the 12 static leds and one led on the ring will be on.

Payload: none

### /clock/normal

Returns the clock to normal mode displaying current time.

Payload: none

### /clock/pause

Pauses countdown timer(s).

Payload: none

### /clock/resume

Resumes countdown timers()

Payload: none

## OSC feedback

The clock sends it's state to the address specified with --osc-dest on `/clock/status` message. The payload is:
1. int32 clock display mode
2. string: hours display
3. string: minutes display
4. string: seconds display
5. string: "tally" text