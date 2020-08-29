package main

const configHTML = `
<h1>Clock configuration editor</h1>
<form action="/save" method="post">

<select name="Face" id="Face">
	<option value="round" {{if eq .Face "round"}} selected {{end}}>Single round clock</option>
	<option value="dual-round" {{if eq .Face "dual-round"}} selected {{end}}>Dual round clocks</option>
	<option value="text" {{if eq .Face "text"}} selected {{end}}>Text clock</option>
	<option value="small" {{if eq .Face  "small"}} selected {{end}}>Small 192x192px round clock</option>
</select><br />
<label for="Face">Select the clock face to use.</label><br />

<input type="checkbox" id="Format12h" name="Format12h" {{if .EngineOptions.Format12h}} checked {{end}}/>
<label for="Format12h">Use 12 hour format for time-of-day display.</label><br />

<input type="checkbox" id="NoARCorrection" name="NoARCorrection" {{if .NoARCorrection}} checked {{end}}/>
<label for="NoARCorrection">Disable detection of official raspberry pi display for aspect ratio correction</label><br />

<input type="checkbox" id="Debug" name="Debug" {{if .Debug}} checked {{end}}/>
<label for="Debug">Output verbose debug information. This will impact performance.</label><br />

<input type="text" id="Font" name="Font" value="{{.Font}}" />
<label for="Font">Font filename for round clocks</label><br />

<input type="text" id="NumberFont" name="NumberFont" value="{{.NumberFont}}" />
<label for="NumberFont">Font filename for text clock numbers</label><br />

<input type="text" id="LabelFont" name="LabelFont" value="{{.LabelFont}}" />
<label for="LabelFont">Font filename for text clock labels</label><br />

<input type="text" id="IconFont" name="IconFont" value="{{.IconFont}}" />
<label for="IconFont">Font filename for text clock icons</label><br />

<input type="text" id="Font" name="Font" value="{{.Font}}" />
<label for="Font">Font filename for round clocks</label><br />

<input type="number" min="0" id="Flash" name="Flash" value="{{.EngineOptions.Flash}}" />
<label for="Flash">Flashing interval in milliseconds for ellapsed countdowns</label><br />

<input type="number" min="0" id="Timeout" name="Timeout" value="{{.EngineOptions.Timeout}}" />
<label for="Flash">Timeout for clearing OSC text display messages, milliseconds</label><br />

<input type="text" id="Background" name="Background" value="{{.Background}}" />
<label for="Background">Experimental background image support. Image filename</label><br />
<p>The image needs to be in the correct resolution and either png or jpeg file. Place the
image in the fat partition and refer to it as /boot/imagename.png</p>

<h2>Time sources</h2>

<p>The single round clock uses source 1 as the main display and source 2 as a secondary timer.
The dual round clock mode uses all four sources, with 1 and 2 in the left clock and 3 and 4 in the right clock.</p>

<p>The round clocks only support timers as the secondary display source, as others can't be compacted to 4 characters.</p>

<p>The sources choose their displayed time in the following priority if enabled:
<ol>
	<li>LTC</li>
	<li>UDP protocol from Interspace / stage timer (not yet implemented)</li>
	<li>Associated timer if it is running</li>
	<li>Time of day in the selected time zone</li>
	<li>Blank display</li>
</ol>

<h3>Source 1</h3>

<input type="text" id="source1-text" name="source1-text" value="{{.EngineOptions.Source1.Text}}" />
<label for="source1-text">Text label for time source</label><br />

<input type="checkbox" id="source1-ltc" name="source1-ltc" {{if .EngineOptions.Source1.LTC}} checked {{end}} />
<label for="source1-ltc">Enable LTC input on this source</label><br />

<input type="checkbox" id="source1-udp" name="source1-udp" {{if .EngineOptions.Source1.UDP}} checked {{end}} />
<label for="source1-udp">Enable UDP input on this source</label><br />

<input type="checkbox" id="source1-timer" name="source1-timer" {{if .EngineOptions.Source1.Timer}} checked {{end}} />
<label for="source1-timer">Enable input from the associated timer on this source</label><br />

<input type="number" min="0" max="9" id="source1-counter" name="source1-counter" value="{{.EngineOptions.Source1.Counter}}" />
<label for="source1-counter">Conter number to use as a timer (0-9)</label><br />

<input type="checkbox" id="source1-tod" name="source1-tod" {{if .EngineOptions.Source1.Tod}} checked {{end}} />
<label for="source1-tod">Enable time of day input on this source</label><br />

<input type="text" id="source1-timezone" name="source1-timezone" value="{{.EngineOptions.Source1.TimeZone}}" />
<label for="source1-timezone">Timezone for the time of day input</label><br />

<h3>Source 2</h3>

<input type="text" id="source2-text" name="source2-text" value="{{.EngineOptions.Source2.Text}}" />
<label for="source2-text">Text label for time source</label><br />

<input type="checkbox" id="source2-ltc" name="source2-ltc" {{if .EngineOptions.Source2.LTC}} checked {{end}} />
<label for="source2-ltc">Enable LTC input on this source</label><br />

<input type="checkbox" id="source2-udp" name="source2-udp" {{if .EngineOptions.Source2.UDP}} checked {{end}} />
<label for="source2-udp">Enable UDP input on this source</label><br />

<input type="checkbox" id="source2-timer" name="source2-timer" {{if .EngineOptions.Source2.Timer}} checked {{end}} />
<label for="source2-timer">Enable input from the associated timer on this source</label><br />

<input type="number" min="0" max="9" id="source2-counter" name="source2-counter" value="{{.EngineOptions.Source2.Counter}}" />
<label for="source2-counter">Conter number to use as a timer (0-9)</label><br />

<input type="checkbox" id="source2-tod" name="source2-tod" {{if .EngineOptions.Source2.Tod}} checked {{end}} />
<label for="source2-tod">Enable time of day input on this source</label><br />

<input type="text" id="source2-timezone" name="source2-timezone" value="{{.EngineOptions.Source2.TimeZone}}" />
<label for="source2-timezone">Timezone for the time of day input</label><br />

<h3>Source 3</h3>

<input type="text" id="source3-text" name="source3-text" value="{{.EngineOptions.Source3.Text}}" />
<label for="source3-text">Text label for time source</label><br />

<input type="checkbox" id="source3-ltc" name="source3-ltc" {{if .EngineOptions.Source3.LTC}} checked {{end}} />
<label for="source3-ltc">Enable LTC input on this source</label><br />

<input type="checkbox" id="source3-udp" name="source3-udp" {{if .EngineOptions.Source3.UDP}} checked {{end}} />
<label for="source3-udp">Enable UDP input on this source</label><br />

<input type="checkbox" id="source3-timer" name="source3-timer" {{if .EngineOptions.Source3.Timer}} checked {{end}} />
<label for="source3-timer">Enable input from the associated timer on this source</label><br />

<input type="number" min="0" max="9" id="source3-counter" name="source3-counter" value="{{.EngineOptions.Source3.Counter}}" />
<label for="source3-counter">Conter number to use as a timer (0-9)</label><br />

<input type="checkbox" id="source3-tod" name="source3-tod" {{if .EngineOptions.Source3.Tod}} checked {{end}} />
<label for="source3-tod">Enable time of day input on this source</label><br />

<input type="text" id="source3-timezone" name="source3-timezone" value="{{.EngineOptions.Source3.TimeZone}}" />
<label for="source3-timezone">Timezone for the time of day input</label><br />

<h3>Source 4</h3>

<input type="text" id="source4-text" name="source4-text" value="{{.EngineOptions.Source4.Text}}" />
<label for="source4-text">Text label for time source</label><br />

<input type="checkbox" id="source4-ltc" name="source4-ltc" {{if .EngineOptions.Source4.LTC}} checked {{end}} />
<label for="source4-ltc">Enable LTC input on this source</label><br />

<input type="checkbox" id="source4-udp" name="source4-udp" {{if .EngineOptions.Source4.UDP}} checked {{end}} />
<label for="source4-udp">Enable UDP input on this source</label><br />

<input type="checkbox" id="source4-timer" name="source4-timer" {{if .EngineOptions.Source4.Timer}} checked {{end}} />
<label for="source4-timer">Enable input from the associated timer on this source</label><br />

<input type="number" min="0" max="9" id="source4-counter" name="source4-counter" value="{{.EngineOptions.Source4.Counter}}" />
<label for="source4-counter">Conter number to use as a timer (0-9)</label><br />

<input type="checkbox" id="source4-tod" name="source4-tod" {{if .EngineOptions.Source4.Tod}} checked {{end}} />
<label for="source4-tod">Enable time of day input on this source</label><br />

<input type="text" id="source4-timezone" name="source4-timezone" value="{{.EngineOptions.Source4.TimeZone}}" />
<label for="source4-timezone">Time zone for the time of day input</label><br />

<h3>Mitti and Millumin</h3>

<input type="number" min="0" max="9" id="mitti" name="mitti" value="{{.EngineOptions.Mitti}}" />
<label for="mitti">Counter number for OSC feedback from Mitti.</label><br />

<input type="number" min="0" max="9" id="millumin" name="millumin" value="{{.EngineOptions.Millumin}}" />
<label for="millumin">Counter number for OSC feedback from Millumin.</label><br />

<input type="text" id="millumin-ignore" name="millumin-ignore" value="{{.EngineOptions.Ignore}}" />
<label for="millumin-ignore">Regeexp for ignoring media layers from the Millumin OSC feedback</label><br />


<h2>Colors</h2>

<h3>Round clock</h3>

<input type="color" id="TextColor" name="TextColor" value="{{.TextColor}}" />
<label for="TextColor">Color for text</label><br />

<input type="color" id="SecColor" name="SecColor" value="{{.SecondColor}}" />
<label for="SecColor">Color for second ring circles</label><br />

<input type="color" id="StaticColor" name="StaticColor" value="{{.StaticColor}}" />
<label for="StaticColor">Color for 12 static "hour" markers</label><br />

<input type="color" id="CountdownColor" name="CountdownColor" value="{{.CountdownColor}}" />
<label for="CountdownColor">Color for secondary countdown display</label><br />

<h3>Text clock</h3>

<input type="color" id="Row1Color" name="Row1Color" value="{{.Row1Color}}" />
<label for="Row1Color">Color timer row 1</label><br />

<input type="color" id="Row2Color" name="Row2Color" value="{{.Row2Color}}" />
<label for="Row2Color">Color timer row 2</label><br />

<input type="color" id="Row3Color" name="Row3Color" value="{{.Row3Color}}" />
<label for="Row3Color">Color timer row 3</label><br />

<input type="color" id="LabelColor" name="LabelColor" value="{{.LabelColor}}" />
<label for="LabelColor">Color labels</label><br />

<input type="checkbox" id="DrawBoxes" name="DrawBoxes" {{if .DrawBoxes}} checked {{end}}/>
<label for="DrawBoxes">Draw background boxes for labels and timers</label><br />

<input type="color" id="LabelBG" name="LabelBG" value="{{.LabelBG}}" />
<label for="LabelBG">Background color for labels</label><br />

<input type="color" id="TimerBG" name="TimerBG" value="{{.TimerBG}}" />
<label for="TimerBG">Background color for timers</label><br />


<h2>OSC</h2>

<input type="checkbox" id="DisableOSC" name="DisableOSC" {{if .EngineOptions.DisableOSC}} checked {{end}}/>
<label for="DisableOSC">Disable remote OSC commands</label><br />

<input type="checkbox" id="DisableFeedback" name="DisableFeedback" {{if .EngineOptions.DisableFeedback}} checked {{end}}/>
<label for="DisableOSC">Disable sending of OSC state feedback</label><br />

<input type="text" id="ListenAddr" name="ListenAddr" value="{{.EngineOptions.ListenAddr}}" />
<label for="ListenAddr">Address and port to listen for osc commands. 0.0.0.0 defaults to all network interfaces.</label><br />

<input type="text" id="Connect" name="Connect" value="{{.EngineOptions.Connect}}" />
<label for="Connect">Address and port to send OSC feedback to. 255.255.255.255 broadcasts to all network interfaces.</label><br />

<h2>Config interface</h2>

<input type="text" id="HTTPUser" name="HTTPUser" value="{{.HTTPUser}}" />
<label for="HTTPUser">Username for the web configuration interface.</label><br />

<input type="text" id="HTTPPassword" name="HTTPPassword" value="{{.HTTPPassword}}" />
<label for="HTTPUser">Password for the web configuration interface.</label><br />


<input type="checkbox" id="DisableHTTP" name="DisableHTTP" {{if .DisableHTTP}} checked {{end}}/>
<label for="DisableHTTP">Disable this web configuration interface. Undoing this needs access to the SD-card.</label><br />

<input type="text" id="HTTPPort" name="HTTPPort" value="{{.HTTPPort}}" />
<label for="HTTPPort">Port to listen for the web configuration. Needs to be in format of ":1234".</label><br />

<h2>LTC</h2>

<input type="checkbox" id="DisableLTC" name="DisableLTC" {{if .EngineOptions.DisableLTC}} checked {{end}}/>
<label for="DisableLTC">Disable LTC display.</label><br />

<input type="checkbox" id="LTCSeconds" name="LTCSeconds" {{if .EngineOptions.LTCSeconds}} checked {{end}}/>
<label for="LTCSeconds">Controls what is displayed on the clock ring in LTC mode, unchecked = frames, checked = seconds</label><br />

<input type="checkbox" id="LTCFollow" name="LTCFollow" {{if .EngineOptions.LTCFollow}} checked {{end}}/>
<label for="LTCFollow">Continue on internal clock if LTC signal is lost. If unset display will blank when signal is gone.</label><br />

{{if .Raspberry}}
	<h1>Raspberry pi configuration</h1>
	<textarea id="configtxt" name="configtxt" rows="20" cols="50">{{.ConfigTxt}}</textarea>
	<br />
	<label for="configtxt">Raspberry pi /boot/config.txt. Changing this will reboot the raspberry pi.</label><br />
{{end}}
<br />

<input type="submit" value="Save config and restart clock" />


</form>
`
