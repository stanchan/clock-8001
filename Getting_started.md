# Getting started on Clock 8001 (hdmi version)


If you have some questions or would like to send  pictures of your clock setup please join
https://www.facebook.com/groups/clock8001/
Or email daniel.richert@svv.fi

## You need
Raspberry Pi,  micro sd memory card and reader for the card.

The images support raspberry pi 2B / 3B / 3B+ boards. They need at least 64Mb SD-cards. 

## Images
Download latest image from /images folder
https://kissa.depili.fi/clock-8001/

There are two versions.
clockworkadmin.img
Is a image with root logins (ssh and local) enabled. The root password is clockworkadmin. **Since this image uses a known hardcoded password it s
hould be for testing only and considered insecure.**

no_login.img 
This image has root logins disabled and is secure for production use.

Use any desired method of writing images on cards.
One recommended software is
Balena Etcher, cross-platform, free and easy to use.
https://www.balena.io/etcher/

## Internet connection
If you need the time of day (ToD)
8001 needs a connection to the internet, just for a short time to get the correct time from the ntp servers. When time is shown internet can be d
isconnected if needed.

Pi doesn't remember the time or any other parameter after a power cycle. This is because the program runs from RAM. The card is read on the start, 
but nothing is never written to it. This is for reliability. 1. The clock will start as new every time 2. The memory card does not corrupt if nothing is written to it.
You could even remove the memory card after the boot.
Because of this, you need to provide a connection to the internet on every power-up if you need ToD.

If you don't have wired internet, you can share your wifi from your mac.
Or use wifi bridges. We have tested the following models, but others should work too. tp-link RE200 and tp-link TL-WA850RE

[Internet sharing in OSX](OSX_internet_sharing.md)

**If you don't need ToD then the internet is not required.**

## Connecting to 8001
The 8001 tries to get a dhcp address on wired ethernet and also brings up a virtual interface eth0:1 with static ip (default 192.168.10.245 with 
255.255.255.0 netmask).

The 8001 sends feedback of the current time to both dhcp and virtual interface broadcast as a default. This can be changed.

You can send commands to the dhcp address or dhcp broadcast address x.x.x.255 or the virtual interface 192.168.10.245 and 192.168.10.255 (virtual
 interface broadcast)

## Settings
If you want to change timezone (default Europe/Helsinki),
Colours or network settings.
Go to http://www.svv.fi/clock
Chrome browser recommended.
There you can create and download setup files.

### clock_cmd.sh
For timezone, colours and flashing interval.
The flashing is global and will affect the flashing of the colon when 8001 is paused or when the timer reaches zero. If set to 0 the flashing will be disabled.
OSC feedback. You can redirect the feedback to desired address and port, leave this empty to get defaults.
Default feedback is 255.255.255.255:1245

### Interfaces
Contains ip address information for the virtual interface eth0:1

Both files can be edited in a text editor if access to http://svv.fi/clock is not possible.

Copy the files to the memory card and If you want to go back to defaults delete them from the card.


## Boot
8001 starts in ToD view XX:XX or versions after 3.1.1 will show the current version, The clock uses ntpd to synchronise its time from the internet. This happens on boot. After that ntpd will periodically check for clock drift and correct the time if it still has a working internet connection. If no connection is active, it will keep the time as long as it has power.

Companion Version 1.3.0 (Build 1277)  or later has Depili Clock module v.3.0.0 for control and monitoring.

There are button presets for 8001 in top right ```presets``` tab of Companion. Drag and drop them to your button grid and edit if needed.

Message function is not in presets and needs to be done manually.


# Basic functions
* Time of day (Tod)
* Count up
* Two separate countdown timers
* Modify
* Pause and resume
* Send text
* Display off

### Time of day 
Is displayed in hh:mm:ss and red led circle increases every second.

### Count up
Is displayed in mm:ss format until the value is over 60minutes after it will be hh:mm:ss. Maximum value 99:59:59. Led circles dot will increase every minute.
Pause/resume can be used for count up.

### Primary timer 
Displays hh:mm:ss if the value is over 60minutes. Everything under is shown in mm:ss format. Led circle represent the total time and is updated if the time is modified. It will do a countdown from longer times but everything over 99hours will be shown ++:++, and then when eventually it's down to 99h it will start displaying.

### Secondary timer
Will display secondary timer on the top segment. Arrow down symbol, two numerals and letter. Can be used in conjunction with the ToD or primary timer. 
The top segment has the following priorities:

Top priority ```video playback time``` upcoming feature to see remaining time of playback from Millumin or Mitti

Middle priority ```message.```
Four letter messages can be set. They stay 1sec and then fade away.

Low priority ```secondary timer.```
If nothing is displayed on the top segment secondary timer will be visible. 

If the secondary timer is running and the message is displayed, when message fades away secondary time will be visible again. The countdown is not disturbed by messages or other overlapping display content.

the last letter indicates:
s = seconds
m = minutes
h = hours
d = days

The maximum value for the secondary timer is 99 days.

### Modify
Commands are used when you want to add or take time from the two countdowns. This can be any amount of seconds, minutes and hours. Timers can be paused or running.

### Pause and resume

Will effect both timers and count up. You can start new timers when pause is active, but they will begin as paused. Modify works during the pause, so adding and subtracting works. Pause is not affected by Stop commands.

Pause indication is flashing the separator colon. 

The flashing time is universal and can be edited in the clock_cmd.sh file. if you prefer no flashing set the flashing to 0

### Send text
On the top segment, there is a four character space, and you can send messages to that space. All messages have 1 second default duration, and they will fade out if a new message is sent. For scrolling effect send multiple messages in 250ms intervals (relative delays active) For ```Hello``` you need eight Actions.
```
   H
  He
 Hel
Hell
ello
llo 
lo  
o   
```
### Display off
Will turn all timers and ToD off leaving only the outer dots and the first dot of the led circle on.

- - - -

### OSC commands
If you want to send osc and not use Companion or Universe (library coming soon), the commands are the following, send to port 1245

integrer <i>  int32 timer duration in seconds
```
/clock/countdown/start <i>
/clock/countdown2/start <i>
/clock/countdown/stop
/clock/countdown2/stop
/clock/countdown/modify <i>
/clock/countdown2/modify <i>

/clock/countup/start
/clock/kill
/clock/normal
```
Returns the clock to normal mode displaying time of day.
```
/clock/pause
/clock/resume
```
### Messages
```
/clock/display <f>  <f>  <f>  <s>
```
Float<f> String <s>

float32 Red component of the text colour
float32 Green component of the text colour
float32 Blue component of the text colour
string up to 4 characters to display


## OSC feedback
The clock sends it's state to the address specified with --osc-dest on _clock_status message. You can  modify the feedback address in web form : svv.fi/clock (Chrome browser recommended.) The information is saved in clock_cmd.sh default is 255.255.255.255:1245

The payload is:

* int32 clock display mode
* string: hours display
* string: minutes display
* string: seconds display
* string: "tally" text

