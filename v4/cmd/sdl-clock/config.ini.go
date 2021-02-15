package main

const configTemplate = `# Example configuration file for clock-8001
# Lines starting with '#' are comments and
# are ignored by clock-8001

# Clock face to use. (round, dual-round, text, single or small)
Face={{.Face}}

# Username and password for the web configuration interface
HTTPUser={{.HTTPUser}}
HTTPPassword={{.HTTPPassword}}

# Set to true to use 12 hour format for time-of-day display.
Format12h={{.EngineOptions.Format12h}}

# Set to true to disable detection of official raspberry pi display for aspect ratio correction
NoARCorrection={{.NoARCorrection}}

# Background image support. You need to provide the background in
# the correct resolution as a png or jpeg image.
Background={{.Background}}

# Background image path for OSC background selection
BackgroundPath={{.BackgroundPath}}

# Background color, used if no image is provided
BackgroundColor={{.BackgroundColor}}

# Truetype fonts for the text clock face

# Font for numbers
NumberFont={{.NumberFont}}

# Font for label texts
LabelFont={{.LabelFont}}

# Font for icons
IconFont={{.IconFont}}

# Show clock info for X seconds at startup
info-timer={{.EngineOptions.ShowInfo}}

# InterSpace Industries Countdown2 UDP timer protocol
# (also supported by Stagetimer2 and Irisdown):
# set to off, send or receive
udp-time={{.EngineOptions.UDPTime}}

# Timer to send / receive as UDP timer 1 (port 36700)
udp-timer-1={{.EngineOptions.UDPTimer1}}

# Timer to send / receive as UDP timer 1 (port 36700)
udp-timer-2={{.EngineOptions.UDPTimer2}}

# Time sources
#
# The single round clock uses source 1 as the main display and source 2 as a secondary timer
# The dual round clock mode uses all four sources, with 1 and 2 in the left clock and 3 and 4 in the right
#
# The round clocks only support timers as the secondary display source, as others can't be compacted to 4 characters
#
# The sources choose their displayed time in the following priority if enabled:
# 1. LTC
# 2. Associated timer if running
# 3. Time of day
# 4. Blank display

# Text label for time source
source1.text={{.EngineOptions.Source1.Text}}
# Set to true to enable LTC input on this source
source1.ltc={{.EngineOptions.Source1.LTC}}
# Set to true for countdown / count up timer input on this source
source1.timer={{.EngineOptions.Source1.Timer}}
# Counter number for timer support (0-9)
source1.counter={{.EngineOptions.Source1.Counter}}
# Set to true to enable time of day input on this source
source1.tod={{.EngineOptions.Source1.Tod}}
# Time zone for the time of day input
source1.timezone={{.EngineOptions.Source1.TimeZone}}
# Initially hide this source, can be toggled via OSC
source1.hidden={{.EngineOptions.Source1.Hidden}}

source2.text={{.EngineOptions.Source2.Text}}
source2.ltc={{.EngineOptions.Source2.LTC}}
source2.timer={{.EngineOptions.Source2.Timer}}
source2.counter={{.EngineOptions.Source2.Counter}}
source2.tod={{.EngineOptions.Source2.Tod}}
source2.timezone={{.EngineOptions.Source2.TimeZone}}
source2.hidden={{.EngineOptions.Source2.Hidden}}

source3.text={{.EngineOptions.Source3.Text}}
source3.ltc={{.EngineOptions.Source3.LTC}}
source3.timer={{.EngineOptions.Source3.Timer}}
source3.counter={{.EngineOptions.Source3.Counter}}
source3.tod={{.EngineOptions.Source3.Tod}}
source3.timezone={{.EngineOptions.Source3.TimeZone}}
source3.hidden={{.EngineOptions.Source3.Hidden}}

source4.text={{.EngineOptions.Source4.Text}}
source4.ltc={{.EngineOptions.Source4.LTC}}
source4.timer={{.EngineOptions.Source4.Timer}}
source4.counter={{.EngineOptions.Source4.Counter}}
source4.tod={{.EngineOptions.Source4.Tod}}
source4.timezone={{.EngineOptions.Source4.TimeZone}}
source4.hidden={{.EngineOptions.Source4.Hidden}}

# Counter number for Mitti OSC feedback
mitti={{.EngineOptions.Mitti}}

# Counter number for Millumin OSC feedback
millumin={{.EngineOptions.Millumin}}

# Millumin layer ignore regexp
millumin-ignore-layer={{.EngineOptions.Ignore}}

# Font to use
Font={{.Font}}

# Colors, in hex format, #XXX or #XXXXXX

# Round clocks

# Color for text
text-color={{.TextColor}}

# Color for the 12 static "hour" markers
static-color={{.StaticColor}}

# Color for the second ring dots
second-color={{.SecondColor}}

# Color for the secondary countdown display
countdown-color={{.CountdownColor}}

# Text clock

# Color for labels
label-color={{.LabelColor}}
# Alpha for labels
label-alpha={{.LabelAlpha}}

# Timer row 1 color
row1-color={{.Row1Color}}
# Timer row 1 alpha
row1-alpha={{.Row1Alpha}}

# Timer row 2 color
row2-color={{.Row2Color}}
# Timer row 2 alpha
row2-alpha={{.Row2Alpha}}

# Timer row 3 color
row3-color={{.Row3Color}}
# Timer row 3 alpha
row3-alpha={{.Row3Alpha}}

# Draw background boxes for timers and labels
draw-boxes={{.DrawBoxes}}

# Timer background color
timer-bg-color={{.TimerBG}}
# Timer background alpha
timer-bg-alpha={{.TimerBGAlpha}}

# Label background color
label-bg-color={{.LabelBG}}
# Label background alpa
label-bg-alpha={{.LabelBGAlpha}}

# Numbers font size
numbers-size={{.NumberFontSize}}

# Engine internals

# Set to true to output verbose debug information
Debug={{.Debug}}

# Flashing interval for ellapsed countdowns, in milliseconds
Flash={{.EngineOptions.Flash}}

# Set to true to disable remote osc commands
DisableOSC={{.EngineOptions.DisableOSC}}

# Address to listen for osc commands. 0.0.0.0 defaults to all network interfaces
ListenAddr={{.EngineOptions.ListenAddr}}

# Timeout for clearing OSC text display messages
Timeout={{.EngineOptions.Timeout}}

# Set to true to disable sending of the OSC feedback messages
DisableFeedback={{.EngineOptions.DisableFeedback}}

# Address to send OSC feedback to. 255.255.255.255 broadcasts to all network interfaces
Connect={{.EngineOptions.Connect}}

# Set to true to disable the web configuration interface
DisableHTTP={{.DisableHTTP}}

# Port to listen for the web configuration. Needs to be in format of ":1234".
HTTPPort={{.HTTPPort}}

# Set to true to disable LTC timecode display mode
DisableLTC={{.EngineOptions.DisableLTC}}

# Controls what is displayed on the clock ring in LTC mode, false = frames, true = seconds
LTCSeconds={{.EngineOptions.LTCSeconds}}

# Continue on internal clock if LTC signal is lost. If unset display will blank when signal is gone.
LTCFollow={{.EngineOptions.LTCFollow}}
`
