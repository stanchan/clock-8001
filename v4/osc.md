# OSC API commands

## Feedback messages

The clock sends feedback with `/clock/source/*/state` and `/clock/timer/*/state` messages. The messages are sent as one OSC bundle

### `/clock/source/*/state`

1. string; Clock UUID
2. bool; is the source hidden
3. string; source output string, generally HH:MM:SS or HH:MM:SS:FF for timecode
4. string; source compact output
5. string; icon for source mode
6. float; source timer progress, 0-1
7. boolean; is the source timer expired
8. boolean; is the source timer paused
9. string; title for the source
10. int; source mode


### `/clock/timer/*/state`

1. string; Clock UUID
2. bool; is the timer active
3. string; timer output, generally HH:MM:SS, but HH:MM:SS:FF for timecode
4. string; timer compact output
5. string; icon for the current timer mode
6. float; timer progress 0-1
7. boolean; is the timer expired
8. boolean; is the timer paused

## Timers

In the following command addresses `*` denotes the timer number, in range of 0 - 9.

### `/clock/timer/*/countdown`

Starts the timer as a countdown.

Parameters:
1. integer; countdown duration in seconds

### `/clock/timer/*/countdown/target`

Starts a countdown timer targeting a given time of day. The countdown duration will be calculated to hit the givet time during the next 24 hours.

Parameters:
1. string; The time of day for the countdown in the format of `HH:MM:SS`

### `/clock/timer/*/countup`

Starts counting time up from the current time

### `/clock/timer/*/countup/target`

Starts a counting timer from a given time of day. The count will be based on the given time from the previous 24 hours.

Parameters:
1. string; The time of day for the count in the format of `HH:MM:SS`

### `/clock/timer/*/modify`

Add or substract time from the active timer.

Parameters:
1. integer; Seconds to add or to substract from the timer

### `/clock/timer/*/pause`

Pauses the given timer.

### `/clock/timer/*/resume`

Resumes a paused timer.

### `/clock/timer/*/stop`

Stops a given timer.

### `/clock/timer/*/signal`

Sets the signal color for the given timer

Parameters:
1. integer; Red component of the signal color
2. integer; Green component of the signal color
3. integer; Blue component of the signal color

### `/clock/pause`

Pauses all timers.

### `/clock/resume`

Resumes all timers.

## Time sources

In the following command addresses `*` will denote the time source number, in range of 1-4

### `/clock/source/*/hide`

Hide the time source.

### `/clock/source/*/show`

Show the time source.

### `/clock/source/*/title`

Set the title text for the given source.

Parameters:
1. string; the label text in utf8 encoding.

### `/clock/source/*/colors`

Set the text and background colors for the given source output.

Parameters:
1. int; Red component of text color, 0-255
2. int; Green component of text color, 0-255
3. int; Blue component of text color, 0-255
4. int; Alpha for text, 0-255
5. int; Red component for text background, 0-255
6. int; Green component for text background, 0-255
7. int; Blue component for text background, 0-255
8. int; Alpha for text background, 0-255

### `/clock/hide`

Hide all time sources.

### `/clock/show`

Show all time sources.

## Misc commands

### `/clock/info`

Shows the information overlay.

Parameters:
1. integer; duration in seconds for the info overlay

### `/clock/background`

Changes the background image to a file with the given number on the background directory in clock configuration.

Parameters:
1. integer: the background number

### `/clock/text`

Shows a text message on the clock face. Specify duration of 0 for infinite duration.

Parameters:
1. integer; the red color component for the text color, 0-255
2. integer; the green color component for the text color, 0-255
3. integer; the blue color component for the text color, 0-255
4. integer; the alpha for the text, 0-255
5. integer; the red color component for the text background color, 0-255
6. integer; the green color component for the text background color, 0-255
7. integer; the blue color component for the text background color, 0-255
8. integer; the alpha for text background color, 0-255
9. integer; the duration in seconds to show the text for
10. string; the text to display, in utf8 encoding

### `/clock/titlecolors`

Set the text and background colors for source titles.

Parameters:
1. int; Red component of text color, 0-255
2. int; Green component of text color, 0-255
3. int; Blue component of text color, 0-255
4. int; Alpha for text, 0-255
5. int; Red component for text background, 0-255
6. int; Green component for text background, 0-255
7. int; Blue component for text background, 0-255
8. int; Alpha for text background, 0-255


### `/clock/seconds/off`

Hide the second display from the ring in the round clocks.

### `/clock/seconds/on`

Show the seconds in the ring on the round clocks.

### `/clock/time/set`

Sets the local time of the clock.

Parameters:
1. string; time of day in format `01:02:03` where 01 is the hours in 24 hour format, 02 the minutes and 03 the seconds.

### `/clock/flash`

Flashes the screen full white for 200ms

## Internal commands

These are used to bridge different state information accross multiple clocks.

### `/clock/ltc`

Relays the LTC time to the clock.

Parameters:
1. string; `HH:MM:SS:FF` where HH = hours, MM = minutes, SS = seconds, FF = frames

### `/clock/media/*`

Where `*` is either `mitti` or `millumin`

### `/clock/resetmedia/*`

Where `*` is either `mitti` or `millumin`

## Deprecated commands

These are implemented for compatibility with the version 3 clocks. They shouldn't be used for any new implementations and will be removed in future.

### `/clock/dual/text`

Use `/clock/text` instead.

### `/clock/kill`

Alias for `/clock/hide`

### `/clock/normal`

Alias for `/clock/show`

### `/clock/display`

Use `/clock/text` instead.

### `/clock/countdown/start`

Alias for `/clock/timer/1/countdown`

### `/clock/countdown2/start`

Alias for `/clock/timer/2/countdown`

### `/clock/countdown/modify`

Alias for `/clock/timer/1/modify`

### `/clock/countdown2/modify`

Alias for `/clock/timer/2/modify`

### `/clock/countdown/stop`

Alias for `/clock/timer/1/stop`

### `/clock/countdown2/stop`

Alias for `/clock/timer/2/stop`

### `/clock/countup/start`

Alias for `/clock/timer/1/countup`

### `/clock/countup/modify`

Alias for `/clock/timer/1/modify`
