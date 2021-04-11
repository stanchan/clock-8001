# OSC controlled simple clock
* Clock binary builds: [![pipeline status](https://gitlab.com/Depili/clock-8001/badges/master/pipeline.svg)](https://gitlab.com/Depili/clock-8001/commits/master)
* Clock image builds: [![pipeline status](https://gitlab.com/Depili/buildroot-clock-8001/badges/master/pipeline.svg)](https://gitlab.com/Depili/buildroot-clock-8001/commits/master)

Support clock-8001 development by paypal: [![](https://www.paypalobjects.com/en_US/i/btn/btn_donate_LG.gif)](https://www.paypal.com/cgi-bin/webscr?cmd=_donations&business=XUMXUL5RX5MWJ&currency_code=EUR)


This is a simplistic clock written in go that can be used either as a video out with SDL or as a dedicated clock buidt with a 32x32 pixel hub75 led matrix and a ring of 60 addressable leds.

The README has been written for the version 4.x.x clocks, ie. the gitlab.com/Depili/clock-8001/v4 go module.

The clock can be controlled with the depili-clock-8001 companion module: https://github.com/bitfocus/companion-module-depili-clock-8001 Module version 5.0.0 is the first to implement V4 clock API.

Features and configuration in greated detail can be found in the [getting started guide in wiki](https://gitlab.com/Depili/clock-8001/-/wikis/Getting-started-on-clock-8001-version-4).

Developed in co-operation with Daniel Richert and with a grant from FUUG - Finnish Unix User Group.

## Ready made raspberry pi images

SD-card images for raspberry pi can be found at https://kissa.depili.fi/clock-8001/images

Clock-8001 no longer has multiple different images, they have all been consolidated to one unified image which uses `enable_` files to activate various parts of the clock system as desired. The new images are named as `clock-800-unified-<version>.img`

The images support raspberry pi 2B / 3B / 3B+ boards. They need at least 64Mb SD-cards. Write them to the card like any other raspberry pi sd-card image.

The image tries to get a dhcp address on wired ethernet and also brings up a virtual interface eth0:1 with static ip (default 192.168.10.245 with 255.255.255.0 netmask).

### Customizing the images

You can place the following files on the sd-card FAT partition to customize the installation:
* `clock.ini` main clock configuration file that is used by the default `clock_cmd.sh`
* `hostname` to change the hostname used by the clock, it is available with "hostname.local" for bonjour / mDNS requests
* `interfaces` a replacement for /etc/network/interfaces for custom network configuration
* `ntp.conf` for custom ntp server configuration
* `config.sys` the normal raspberry pi boot configuration for changing video modes etc.
* `sdl-clock` to update the clock binary with this file
* `clock_cmd.sh` is the command line for the clock, it should start with `/root/sdl-clock ` and be followed by any command line parameters you wish to use for the clock.
* `clock_bridge` to update the clock bridge binary file
* `clock_bridge_cmd.sh` to update the clock bridge command line. It should start with `/root/clock-bridge` and be followed by any command line paramaters for the bridge.
* `enable_clock` delete this file and the main clock will not be active
* `enable_bridge` delete this file and the mitti / millumin osc bridge will not be active
* `enable_ssh` delete this file and remote ssh logins to the raspberry pi will not be allowed
* `enable_ltc` delete this file and the LTC audio -> OSC functionality will not be active

### Hyperpixel4 displays

The image supports both the retangular and square hyperpixel4 displays from Pimoroni. To use them you need to rename either `config.txt.hp4_square` or `config.txt.hp4_rect` to `config.txt`

For the rectangular display you need to also modify `clock_cmd.sh` to contain:
```
/hyperpixel4_rect/hyperpixel4-init
/root/sdl-clock -C /boot/clock.ini
```
So that the display is initialized on boot.

### Web configuration interface

The new unified images have a web configuration interface for the clock settings. You can access this interface by pointing your browser to the address of the clock. The default username is `admin` and the default password is `clockwork`. You should change them from the interface or the clock.ini file.

## sdl-clock - Output the clock to hdmi on the raspberry pi

You can build the clock binary with `go get gitlab.com/Depili/clock-8001/cmd/sdl_clock`. Compiling requires SDL 2 and SDL_GFX 2 libraries. On the raspberry pi the default libraries shipped with rasbian will only output data to X11 window, so for full screen dedicated clock you need to compile the SDL libraries from source. For compiling use `./configure --host=armv7l-raspberry-linux-gnueabihf --disable-pulseaudio --disable-esd --disable-video-mir --disable-video-wayland --disable-video-x11 --disable-video-opengl` for config flags.

### Precompiled binaries

* Latest from git master: [sdl-clock](https://gitlab.com/Depili/clock-8001/-/jobs/artifacts/master/raw/sdl-clock?job=build)
* Testing builds: https://kissa.depili.fi/clock-8001/testing/
* Tagged releases: https://kissa.depili.fi/clock-8001/releases/

### Command line parameters
```
Usage:
  sdl-clock [OPTIONS]

Application Options:
  -C, --config=                                                read config from a file
      --face=[round|dual-round|small|text|single|countdown]    Select the clock face to use (default: round)
      --debug                                                  Enable debug output
      --http-port=                                             Port to listen on for the http configuration interface (default: :8080)
      --disable-http                                           Disable the web configuration interface
      --http-user=                                             Username for web configuration (default: admin)
      --http-password=                                         Password for web configuration interface (default: clockwork)
      --dump-config                                            Write configuration to stdout and exit
      --defaults                                               load defaults
      --no-ar-correction                                       Do not try to detect official raspberry pi display and correct it's aspect
                                                               ratio
      --background=                                            Background image file location.
      --background-path=                                       path to load OSC backgrounds from (default: /boot)
      --background-color=                                      Background color, used if no background image is supplied (default: #000000)
      --flash=                                                 Flashing interval when countdown reached zero (ms), 0 disables (default: 500)
      --osc-listen=                                            Address to listen for incoming osc messages (default: 0.0.0.0:1245)
  -d, --timeout=                                               Timeout for OSC message updates in milliseconds (default: 1000)
  -o, --osc-dest=                                              Address to send OSC feedback to (default: 255.255.255.255:1245)
      --disable-osc                                            Disable OSC control and feedback
      --disable-feedback                                       Disable OSC feedback
      --disable-ltc                                            Disable LTC display mode
      --ltc-seconds                                            Show seconds on the ring in LTC mode
      --udp-time=[off|send|receive]                            Stagetimer2 UDP protocol support (default: receive)
      --udp-timer-1=                                           Timer to send as UDP timer 1 (port 36700) (default: 1)
      --udp-timer-2=                                           Timer to send as UDP timer 2 (port 36701) (default: 2)
      --ltc-follow                                             Continue on internal clock if LTC signal is lost. If unset display will blank
                                                               when signal is gone.
      --format-12h                                             Use 12 hour format for time-of-day display
      --mitti=                                                 Counter number for Mitti OSC feedback (default: 8)
      --millumin=                                              Counter number for Millumin OSC feedback (default: 9)
      --millumin-ignore-layer=REGEXP                           Ignore matching millumin layers (case-insensitive regexp) (default: ignore)
      --info-timer=                                            Show clock status for x seconds on startup (default: 30)
  -F, --font=                                                  Font for event name (default: fonts/7x13.bdf)
      --text-color=                                            Color for round clock text (default: #FF8000)
      --static-color=                                          Color for round clock static circles (default: #505000)
      --second-color=                                          Color for round clock second circles (default: #C80000)
      --countdown-color=                                       Color for round clock second circles (default: #FF0000)
      --number-font=                                           Font for text clock face numbers (default: Copse-Regular.ttf)
      --label-font=                                            Font for text clock face labels (default: RobotoMono-VariableFont_wght.ttf)
      --icon-font=                                             Font for text clock face icons (default: MaterialIcons-Regular.ttf)
      --row1-color=                                            Color for text clock row 1 (default: #FF8000)
      --row1-alpha=                                            Alpha channel for text clock row 1 (default: 255)
      --row2-color=                                            Color for text clock row 2 (default: #FF8000)
      --row2-alpha=                                            Alpha channel for text clock row 2 (default: 255)
      --row3-color=                                            Color for text clock row 3 (default: #FF8000)
      --row3-alpha=                                            Alpha channel for text clock row 3 (default: 255)
      --label-color=                                           Color for text clock labels (default: #FF8000)
      --label-alpha=                                           Alpha channel for label text color (default: 255)
      --timer-bg-color=                                        Color for optional timer background box (default: #202020)
      --timer-bg-alpha=                                        Alpha channel for timer background boxes (default: 255)
      --label-bg-color=                                        Color for optional label background box (default: #202020)
      --label-bg-alpha=                                        Alpha channel for label bacground boxes (default: 255)
      --draw-boxes                                             Draw the container boxes for timers
      --numbers-size=
      --font-path=                                             Path for loading font choices into web config (default: .)
      --countdown-target=

1st clock display source:
      --source1.text=                                          Title text for the time source
      --source1.counter=                                       Counter number to associate with this source, leave empty to disable it as a
                                                               suorce (default: 0)
      --source1.ltc                                            Enable LTC as a source
      --source1.timer                                          Enable timer counter as a source
      --source1.tod                                            Enable time-of-day as a source
      --source1.timezone=                                      Time zone to use for ToD display (default: Europe/Helsinki)
      --source1.hidden                                         Hide this time source

2nd clock display source:
      --source2.text=                                          Title text for the time source
      --source2.counter=                                       Counter number to associate with this source, leave empty to disable it as a
                                                               suorce (default: 0)
      --source2.ltc                                            Enable LTC as a source
      --source2.timer                                          Enable timer counter as a source
      --source2.tod                                            Enable time-of-day as a source
      --source2.timezone=                                      Time zone to use for ToD display (default: Europe/Helsinki)
      --source2.hidden                                         Hide this time source

3rd clock display source:
      --source3.text=                                          Title text for the time source
      --source3.counter=                                       Counter number to associate with this source, leave empty to disable it as a
                                                               suorce (default: 0)
      --source3.ltc                                            Enable LTC as a source
      --source3.timer                                          Enable timer counter as a source
      --source3.tod                                            Enable time-of-day as a source
      --source3.timezone=                                      Time zone to use for ToD display (default: Europe/Helsinki)
      --source3.hidden                                         Hide this time source

4th clock display source:
      --source4.text=                                          Title text for the time source
      --source4.counter=                                       Counter number to associate with this source, leave empty to disable it as a
                                                               suorce (default: 0)
      --source4.ltc                                            Enable LTC as a source
      --source4.timer                                          Enable timer counter as a source
      --source4.tod                                            Enable time-of-day as a source
      --source4.timezone=                                      Time zone to use for ToD display (default: Europe/Helsinki)
      --source4.hidden                                         Hide this time source

Help Options:
  -h, --help                                                   Show this help message
```

### LTC timecode support

The clock can be used to display a SMPTE LTC time. This requires a Hifiberry ADC+ DAC Pro hat: https://www.hifiberry.com/shop/boards/hifiberry-dac-adc-pro/ or Interpace Industries USB AiO interface. Currently only these two options for audio input are supported.

For Hifiberry it is recommended to use the pin headers for balanced audio input and to wire the incoming mono LTC signal to both left and right channels.

Other usb audio interfaces might work, but they are completely untested and no support is offered for them.

## matrix-clock - Dedicated led matrix clock

**The matrix clock v4 isn't currently available. You can still use the version 3.**

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

### Precompiled binaries

* Latest from git master: [matrix-clock](https://gitlab.com/Depili/clock-8001/-/jobs/artifacts/master/raw/matrix-clock?job=build)

## OSC commands understood by the clock

See https://gitlab.com/Depili/clock-8001/-/blob/master/v4/osc.md
