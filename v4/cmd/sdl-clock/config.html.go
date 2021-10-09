package main

const configHTML = `
<html>
<head>
	<title>Clock-8001 configuration</title>
</head>
	<body>
		<h1>Clock configuration editor</h1>
		{{if .Errors}}
			<div class="errors">
				<p>
					Following errors prevented the configuration from being saved:
					{{.Errors}}
				</p>
			</div>
		{{end}}
		<div class="config-form">
			<form action="/import" method="post" enctype="multipart/form-data">
				<fieldset>
					<legend>Project links</legend>
					<ul>
						<li><a href="https://gitlab.com/depili/clock-8001/">View the project on gitlab</a></li>
						<li><a href="https://www.paypal.com/cgi-bin/webscr?cmd=_donations&business=XUMXUL5RX5MWJ&currency_code=EUR">Support development of clock-8001 via Paypal</a></li>
					</ul>
				</fieldset>

				<fieldset>
					<legend>Config Import / Export</legend>
					<p><a href="/export">Download current configuration.</a></p>
					<label for="import"><span>Import configurations file</span>
						<input type="file" id="import" name="import" />
					</label>
					<input type="submit" value="upload" />
				</fieldset>
			</form>

			<form action="/save" method="post">
				<fieldset>
					<legend>General settings</legend>
					<label for="Face">
						<span>Select the clock face to use</span>
						<select name="Face" id="Face">
							<option value="round" {{if eq .Face "round"}} selected {{end}}>Single round clock</option>
							<option value="dual-round" {{if eq .Face "dual-round"}} selected {{end}}>Dual round clocks</option>
							<option value="text" {{if eq .Face "text"}} selected {{end}}>Text clock with 3 timers</option>
							<option value="single" {{if eq .Face "single"}} selected {{end}}>Text clock with 1 timer</option>
							<option value="192" {{if eq .Face "192"}} selected {{end}}>Small 192x192px round clock</option>
							<option value="144" {{if eq .Face "144"}} selected {{end}}>Small 144x144px round clock</option>
							<option value="288x144" {{if eq .Face "288x144"}} selected {{end}}>Small 288x144px text clock</option>
						</select><br />
					</label>

					<label for="Format12h">
						<span>Use 12 hour format for time-of-day display</span>
						<input type="checkbox" id="Format12h" name="Format12h" {{if .EngineOptions.Format12h}} checked {{end}}/>
					</label>

					<label for="NoARCorrection">
						<span>Disable detection of official raspberry pi display for aspect ratio correction</span>
						<input type="checkbox" id="NoARCorrection" name="NoARCorrection" {{if .NoARCorrection}} checked {{end}}/>
					</label>

					<label for="Debug">
						<span>Output verbose debug information. This will impact performance</span>
						<input type="checkbox" id="Debug" name="Debug" {{if .Debug}} checked {{end}}/>
					</label>

					<label for="Font">
						<span>Font filename for round clocks</span>
						<input type="text" id="Font" name="Font" value="{{.Font}}" />
					</label>

					<datalist id="FontList">
						{{ range $font := .Fonts }}
							<option>{{$font}}</option>
						{{ end }}
					</datalist>

					<label for="NumberFont">
						<span>Font filename for text clock numbers</span>
						<input list="FontList" type="text" id="NumberFont" name="NumberFont" value="{{.NumberFont}}" />
					</label>

					<label for="LabelFont">
						<span>Font filename for text clock labels</span>
						<input list="FontList" type="text" id="LabelFont" name="LabelFont" value="{{.LabelFont}}" />
					</label>

					<label for="IconFont">
						<span>Font filename for text clock icons</span>
						<input list="FontList" type="text" id="IconFont" name="IconFont" value="{{.IconFont}}" />
					</label>

					<label for="Flash">
						<span>Flashing interval in milliseconds for ellapsed countdowns</span>
						<input type="number" min="0" id="Flash" name="Flash" value="{{.EngineOptions.Flash}}" />
					</label>

					<label for="Timeout">
						<span>Timeout for clearing OSC text display messages, milliseconds</span>
						<input type="number" min="0" id="Timeout" name="Timeout" value="{{.EngineOptions.Timeout}}" />
					</label>

					<label for="ShowInfo">
						<span>Time to show clock information on startup, seconds</span>
						<input type="number" min="0" id="ShowInfo" name="ShowInfo" value="{{.EngineOptions.ShowInfo}}" />
					</label>

					<label for="Background">
						<span>Background image filename</span>
						<input type="text" id="Background" name="Background" value="{{.Background}}" />
					</label>
					<p>The image needs to be in the correct resolution and either png or jpeg file. Place the
					image in the fat partition and refer to it as /boot/imagename.png</p>

					<label for="BackgroundPath">
						<span>Path for OSC selectable background images/span>
						<input type="text" id="BackgroundPath" name="BackgroundPath" value="{{.BackgroundPath}}" />
					</label>
					<p>The OSC command /clock/background/ can be used to select a numbered background image from this path.
					Files should be named with the number (eg 1.png or 01.jpeg). Supported filetypes are BMP, PNG and JPEG.</p>


					<label for="BackgroundColor">
						<span>Background color, used if no background image is provided</span>
						<input type="color" id="BackgroundColor" name="BackgroundColor" value="{{.BackgroundColor}}" />
					</label>

					<label for="AudioEnabled">
						<span>Enable audio cues for expiring countdown timers.</span>
						<input type="checkbox" id="AudioEnabled" name="AudioEnabled" {{if .AudioEnabled}} checked {{end}}/>
					</label>
					<label for="TODBeep">
						<span>Enable audio cues for time of day displays on each full hour.</span>
						<input type="checkbox" id="TODBeep" name="TODBeep" {{if .TODBeep}} checked {{end}}/>
					</label>

				</fieldset>
				<fieldset>
					<legend>Time sources</legend>

				<p>The single round clock uses source 1 as the main display and source 2 as a secondary timer.
				The dual round clock mode uses all four sources, with 1 and 2 in the left clock and 3 and 4 in the right clock.</p>

				<p>The round clocks only support timers as the secondary display source, as others can't be compacted to 4 characters.</p>

				<p>The sources choose their displayed time in the following priority if enabled:
					<ol>
						<li>LTC</li>
						<li>Associated timer if it is running</li>
						<li>Time of day in the selected time zone</li>
						<li>Blank display</li>
					</ol>
				</p>

				<fieldset>
					<legend>Source 1</legend>

					<label for="source1-text">
						<span>Text label for time source</span>
						<input type="text" id="source1-text" name="source1-text" value="{{.EngineOptions.Source1.Text}}" />
					</label>

					<label for="source1-ltc">
						<span>Enable LTC input on this source</span>
						<input type="checkbox" id="source1-ltc" name="source1-ltc" {{if .EngineOptions.Source1.LTC}} checked {{end}} />
					</label>

					<label for="source1-timer">
						<span>Enable input from the associated timer</span>
						<input type="checkbox" id="source1-timer" name="source1-timer" {{if .EngineOptions.Source1.Timer}} checked {{end}} />
					</label>

					<label for="source1-counter">
						<span>Timer number to use (0-9)</span>
						<input type="number" min="0" max="9" id="source1-counter" name="source1-counter" value="{{.EngineOptions.Source1.Counter}}" />
					</label>

					<label for="source1-tod">
						<span>Enable time of day input on this source</span>
						<input type="checkbox" id="source1-tod" name="source1-tod" {{if .EngineOptions.Source1.Tod}} checked {{end}} />
					</label>

					<label for="source1-timezone">
						<span>Timezone for the time of day input</span>
						{{$selected := .EngineOptions.Source1.TimeZone}}
						<select id="source1-timezone" name="source1-timezone" >
							{{ range $tz := .Timezones }}
								<option {{if eq $selected $tz}} selected {{end}}>{{$tz}}</option>
							{{ end }}
						</select>
					</label>

					<label for="source1-hidden">
						<span>Initially hide this source. Can be toggled by OSC on runtime.</span>
						<input type="checkbox" id="source1-hidden" name="source1-hidden" {{if .EngineOptions.Source1.Hidden}} checked {{end}} />
					</label>

					<label for="source1-overtime-color">
						<span>Background color for overtime countdowns</span>
						<input type="color" id="source1-overtime-color" name="source1-overtime-color" value="{{.EngineOptions.Source1.OvertimeColor}}" />
					</label>

				</fieldset>
				<fieldset>
					<legend>Source 2</legend>

					<label for="source2-text">
						<span>Text label for time source</span>
						<input type="text" id="source2-text" name="source2-text" value="{{.EngineOptions.Source2.Text}}" />
					</label>

					<label for="source2-ltc">
						<span>Enable LTC input on this source</span>
						<input type="checkbox" id="source2-ltc" name="source2-ltc" {{if .EngineOptions.Source2.LTC}} checked {{end}} />
					</label>

					<label for="source2-timer">
						<span>Enable input from the associated timer</span>
						<input type="checkbox" id="source2-timer" name="source2-timer" {{if .EngineOptions.Source2.Timer}} checked {{end}} />
					</label>

					<label for="source2-counter">
						<span>Timer number to use (0-9)</span>
						<input type="number" min="0" max="9" id="source2-counter" name="source2-counter" value="{{.EngineOptions.Source2.Counter}}" />
					</label>

					<label for="source2-tod">
						<span>Enable time of day input on this source</span>
						<input type="checkbox" id="source2-tod" name="source2-tod" {{if .EngineOptions.Source2.Tod}} checked {{end}} />
					</label>

					<label for="source2-timezone">
						<span>Timezone for the time of day input</span>
						{{$selected = .EngineOptions.Source2.TimeZone}}
						<select id="source2-timezone" name="source2-timezone" >
							{{ range $tz := .Timezones }}
								<option {{if eq $selected $tz}} selected {{end}}>{{$tz}}</option>
							{{ end }}
						</select>
					</label>

					<label for="source2-hidden">
						<span>Initially hide this source. Can be toggled by OSC on runtime.</span>
						<input type="checkbox" id="source2-hidden" name="source2-hidden" {{if .EngineOptions.Source2.Hidden}} checked {{end}} />
					</label>

					<label for="source2-overtime-color">
						<span>Background color for overtime countdowns</span>
						<input type="color" id="source2-overtime-color" name="source2-overtime-color" value="{{.EngineOptions.Source2.OvertimeColor}}" />
					</label>
				</fieldset>
				<fieldset>
					<legend>Source 3</legend>

					<label for="source3-text">
						<span>Text label for time source</span>
						<input type="text" id="source3-text" name="source3-text" value="{{.EngineOptions.Source3.Text}}" />
					</label>

					<label for="source3-ltc">
						<span>Enable LTC input on this source</span>
						<input type="checkbox" id="source3-ltc" name="source3-ltc" {{if .EngineOptions.Source3.LTC}} checked {{end}} />
					</label>

					<label for="source3-timer">
						<span>Enable input from the associated timer on this source</span>
						<input type="checkbox" id="source3-timer" name="source3-timer" {{if .EngineOptions.Source3.Timer}} checked {{end}} />
					</label>

					<label for="source3-counter">
						<span>Timer number to use (0-9)</span>
						<input type="number" min="0" max="9" id="source3-counter" name="source3-counter" value="{{.EngineOptions.Source3.Counter}}" />
					</label>

					<label for="source3-tod">
						<span>Enable time of day input on this source</span>
						<input type="checkbox" id="source3-tod" name="source3-tod" {{if .EngineOptions.Source3.Tod}} checked {{end}} />
					</label>

					<label for="source3-timezone">
						<span>Timezone for the time of day input</span>
						{{$selected = .EngineOptions.Source3.TimeZone}}
						<select id="source3-timezone" name="source3-timezone" >
							{{ range $tz := .Timezones }}
								<option {{if eq $selected $tz}} selected {{end}}>{{$tz}}</option>
							{{ end }}
						</select>
					</label>

					<label for="source3-hidden">
						<span>Initially hide this source. Can be toggled by OSC on runtime.</span>
						<input type="checkbox" id="source3-hidden" name="source3-hidden" {{if .EngineOptions.Source3.Hidden}} checked {{end}} />
					</label>

					<label for="source3-overtime-color">
						<span>Background color for overtime countdowns</span>
						<input type="color" id="source3-overtime-color" name="source3-overtime-color" value="{{.EngineOptions.Source3.OvertimeColor}}" />
					</label>
				</fieldset>
				<fieldset>
					<legend>Source 4</legend>

					<label for="source4-text">
						<span>Text label for time source</span>
						<input type="text" id="source4-text" name="source4-text" value="{{.EngineOptions.Source4.Text}}" />
					</label>

					<label for="source4-ltc">
						<span>Enable LTC input on this source</span>
						<input type="checkbox" id="source4-ltc" name="source4-ltc" {{if .EngineOptions.Source4.LTC}} checked {{end}} />
					</label>

					<label for="source4-timer">
						<span>Enable input from the associated timer on this source</span>
						<input type="checkbox" id="source4-timer" name="source4-timer" {{if .EngineOptions.Source4.Timer}} checked {{end}} />
					</label>

					<label for="source4-counter">
						<span>Timer number to use (0-9)</span>
						<input type="number" min="0" max="9" id="source4-counter" name="source4-counter" value="{{.EngineOptions.Source4.Counter}}" />
					</label>

					<label for="source4-tod">
						<span>Enable time of day input on this source</span>
						<input type="checkbox" id="source4-tod" name="source4-tod" {{if .EngineOptions.Source4.Tod}} checked {{end}} />
					</label>

					<label for="source4-timezone">
						<span>Time zone for the time of day input</span>
						{{$selected = .EngineOptions.Source4.TimeZone}}
						<select id="source4-timezone" name="source4-timezone" >
							{{ range $tz := .Timezones }}
								<option {{if eq $selected $tz}} selected {{end}}>{{$tz}}</option>
							{{ end }}
						</select>
					</label>

					<label for="source4-hidden">
						<span>Initially hide this source. Can be toggled by OSC on runtime.</span>
						<input type="checkbox" id="source4-hidden" name="source4-hidden" {{if .EngineOptions.Source4.Hidden}} checked {{end}} />
					</label>

					<label for="source4-overtime-color">
						<span>Background color for overtime countdowns</span>
						<input type="color" id="source4-overtime-color" name="source4-overtime-color" value="{{.EngineOptions.Source4.OvertimeColor}}" />
					</label>
				</fieldset>
			</fieldset>

			<fieldset>
				<legend>Overtime behaviour</legend>

				<label for="overtime-count-mode">
					<span>Countdown readout for overtime timers</span>
					<select name="overtime-count-mode" id="overtime-count-mode">
						<option value="zero" {{if eq .EngineOptions.OvertimeCountMode "zero"}} selected {{end}}>Show 00:00:00</option>
						<option value="blank" {{if eq .EngineOptions.OvertimeCountMode "blank"}} selected {{end}}>Blank display</option>
						<option value="continue" {{if eq .EngineOptions.OvertimeCountMode "continue"}} selected {{end}}>Continue counting up</option>
					</select>
				</label>

				<label for="overtime-visibility">
					<span>Extra visibility for overtime timers</span>
					<select name="overtime-visibility" id="overtime-visibility">
						<option value="blink" {{if eq .EngineOptions.OvertimeVisibility "blink"}} selected {{end}}>Blink readout</option>
						<option value="none" {{if eq .EngineOptions.OvertimeVisibility "none"}} selected {{end}}>No extra visibility</option>
						<option value="background" {{if eq .EngineOptions.OvertimeVisibility "background"}} selected {{end}}>Change background color</option>
						<option value="both" {{if eq .EngineOptions.OvertimeVisibility "both"}} selected {{end}}>Change background + blink</option>
					</select>
				</label>
			</fieldset>
			<fieldset>
				<legend>Timer signal colors</legend>
				<label for="auto-signals">
					<span>Automatically set signal color per timer state</span>
					<input type="checkbox" id="auto-signals" name="auto-signals" {{if .EngineOptions.AutoSignals}} checked {{end}} />
				</label>

				<label for="signal-start">
					<span>In automation mode, set a color on timer start</span>
					<input type="checkbox" id="signal-start" name="signal-start" {{if .EngineOptions.SignalStart}} checked {{end}} />
				</label>

				<label for="signal-color-start">
					<span>Start signal color</span>
					<input type="color" id="signal-color-start" name="signal-color-start" value="{{.EngineOptions.SignalColorStart}}" />
				</label>

				<label for="signal-threshold-warning">
					<span>Time threshold for warning color, in seconds. Set to 0 to disable.</span>
					<input type="number" min="0" id="signal-threshold-warning" name="signal-threshold-warning" value="{{.EngineOptions.SignalThresholdWarning}}" />
				</label>


				<label for="signal-color-warning">
					<span>Warning signal color</span>
					<input type="color" id="signal-color-warning" name="signal-color-warning" value="{{.EngineOptions.SignalColorWarning}}" />
				</label>

				<label for="signal-threshold-end">
					<span>Time threshold for end color, in seconds.</span>
					<input type="number" min="0" id="signal-threshold-end" name="signal-threshold-end" value="{{.EngineOptions.SignalThresholdEnd}}" />
				</label>

				<label for="signal-color-end">
					<span>End signal color</span>
					<input type="color" id="signal-color-end" name="signal-color-end" value="{{.EngineOptions.SignalColorEnd}}" />
				</label>

				<label for="signal-hw-type">
						<span>Signal hardware type</span>
						<select name="signal-hw-type" id="signal-hw-type">
						<option value="unicorn-hd" {{if eq .SignalType "unicorn-hd"}} selected {{end}}>Pimoroni Unicorn HD or Ubercorn</option>
						<option value="none" {{if eq .SignalType "none"}} selected {{end}}>None</option>
					</select><br />
				</label>

				<label for="signal-hw-group">
					<span>Hardware signal group</span>
					<input type="number" min="0" id="signal-hw-group" name="signal-hw-group" value="{{.EngineOptions.SignalHardware}}" />
				</label>

				<label for="signal-hw-brightness">
					<span>Hardware signal master brightness, 0 = off, 255 = maximum brightness</span>
					<input type="number" min="0" max="255" id="signal-hw-brightness" name="signal-hw-brightness" value="{{.SignalBrightness}}" />
				</label>


				<label for="signal-hw-follow">
					<span>Hardware signal follows source 1 color</span>
					<input type="checkbox" id="signal-hw-follow" name="signal-hw-follow" {{if .SignalFollow}} checked {{end}} />
				</label>


			</fieldset>


			<fieldset>
				<legend>Mitti and Millumin</legend>

				<label for="mitti">
					<span>Timer number for OSC feedback from Mitti</span>
					<input type="number" min="0" max="9" id="mitti" name="mitti" value="{{.EngineOptions.Mitti}}" />
				</label>

				<label for="millumin">
					<span>Timer number for OSC feedback from Millumin</span>
					<input type="number" min="0" max="9" id="millumin" name="millumin" value="{{.EngineOptions.Millumin}}" />
				</label>


				<label for="millumin-ignore">
					<span>Regexp for ignoring media layers from the Millumin OSC feedback</span>
					<input type="text" id="millumin-ignore" name="millumin-ignore" value="{{.EngineOptions.Ignore}}" />
				</label>
			</fieldset>

			<fieldset>
				<legend>InterSpace Industries Countdown2 UDP</legend>
				<p>StageTimer2 and Irisdown also support sending data with this protocol.</p>
				<label for="udp-time">
						<span>Countdown2 UDP mode</span>
						<select name="udp-time" id="udp-time">
							<option value="off" {{if eq .EngineOptions.UDPTime "off"}} selected {{end}}>Off</option>
							<option value="send" {{if eq .EngineOptions.UDPTime "send"}} selected {{end}}>Send timers</option>
							<option value="receive" {{if eq .EngineOptions.UDPTime "receive"}} selected {{end}}>Receive timers</option>
						</select>
				</label>

				<label for="upd-timer-1">
					<span>Timer number for StageTimer2 UDP timer 1 from port 36700</span>
					<input type="number" min="0" max="9" id="udp-timer-1" name="udp-timer-1" value="{{.EngineOptions.UDPTimer1}}" />
				</label>

				<label for="upd-timer-2">
					<span>Timer number for StageTimer2 UDP timer 2 from port 36701</span>
					<input type="number" min="0" max="9" id="udp-timer-2" name="udp-timer-2" value="{{.EngineOptions.UDPTimer2}}" />
				</label>
			</fieldset>
			<fieldset>
				<legend>Colors</legend>

				<fieldset>
					<legend>Round clocks</legend>

					<label for="TextColor">
						<span>Color for text</span>
						<input type="color" id="TextColor" name="TextColor" value="{{.TextColor}}" />
					</label>

					<label for="SecColor">
						<span>Color for second ring circles</span>
						<input type="color" id="SecColor" name="SecColor" value="{{.SecondColor}}" />
					</label>

					<label for="StaticColor">
						<span>Color for 12 static "hour" markers</span>
						<input type="color" id="StaticColor" name="StaticColor" value="{{.StaticColor}}" />
					</label>

					<label for="CountdownColor">
						<span>Color for secondary countdown display</span>
						<input type="color" id="CountdownColor" name="CountdownColor" value="{{.CountdownColor}}" />
					</label>
				</fieldset>
				<fieldset>
					<legend>Text clock</legend>

					<label for="Row1Color">
						<span>Color timer row 1</span>
						<input type="color" id="Row1Color" name="Row1Color" value="{{.Row1Color}}" />
					</label>

					<label for="row1-alpha">
						<span>Alpha for timer row 1</span>
						<input type="number" min="0" max="255" id="row1-alpha" name="row1-alpha" value="{{.Row1Alpha}}" />
					</label>

					<label for="Row2Color">
						<span>Color timer row 2</span>
						<input type="color" id="Row2Color" name="Row2Color" value="{{.Row2Color}}" />
					</label>

					<label for="row2-alpha">
						<span>Alpha for timer row 2</span>
						<input type="number" min="0" max="255" id="row2-alpha" name="row2-alpha" value="{{.Row2Alpha}}" />
					</label>

					<label for="Row3Color">
						<span>Color timer row 3</span>
						<input type="color" id="Row3Color" name="Row3Color" value="{{.Row3Color}}" />
					</label>

					<label for="row3-alpha">
						<span>Alpha for timer row 3</span>
						<input type="number" min="0" max="255" id="row3-alpha" name="row3-alpha" value="{{.Row3Alpha}}" />
					</label>

					<label for="LabelColor">
						<span>Color for timer titles</span>
						<input type="color" id="LabelColor" name="LabelColor" value="{{.LabelColor}}" />
					</label>

					<label for="label-alpha">
						<span>Alpha for timer titles</span>
						<input type="number" min="0" max="255" id="label-alpha" name="label-alpha" value="{{.LabelAlpha}}" />
					</label>

					<label for="DrawBoxes">
						<span>Draw background boxes for labels and timers</span>
						<input type="checkbox" id="DrawBoxes" name="DrawBoxes" {{if .DrawBoxes}} checked {{end}}/>
					</label>

					<label for="LabelBG">
						<span>Background color for timer titles</span>
						<input type="color" id="LabelBG" name="LabelBG" value="{{.LabelBG}}" />
					</label>

					<label for="label-bg-alpha">
						<span>Alpha for timer title backgrounds</span>
						<input type="number" min="0" max="255" id="label-bg-alpha" name="label-bg-alpha" value="{{.LabelBGAlpha}}" />
					</label>

					<label for="TimerBG">
						<span>Background color for timers</span>
						<input type="color" id="TimerBG" name="TimerBG" value="{{.TimerBG}}" />
					</label>

					<label for="timer-bg-alpha">
						<span>Alpha for timer backgrounds</span>
						<input type="number" min="0" max="255" id="timer-bg-alpha" name="timer-bg-alpha" value="{{.TimerBGAlpha}}" />
					</label>

					<label for="NumberFontSize">
						<span>Size used to render number text, higher results in smoother letters, but going too high will crash on the rpi</span>
						<input type="number" min="0" id="NumberFontSize" name="NumberFontSize" value="{{.NumberFontSize}}" />
					</label>

				</fieldset>
			</fieldset>
			<fieldset>
				<legend>OSC</legend>

				<label for="DisableOSC">
					<span>Disable remote OSC commands</span>
					<input type="checkbox" id="DisableOSC" name="DisableOSC" {{if .EngineOptions.DisableOSC}} checked {{end}}/>
				</label>

				<label for="DisableOSC">
					<span>Disable sending of OSC state feedback</span>
					<input type="checkbox" id="DisableFeedback" name="DisableFeedback" {{if .EngineOptions.DisableFeedback}} checked {{end}}/>
				</label>


				<label for="ListenAddr">
					<span>Address and port to listen for osc commands. 0.0.0.0 defaults to all network interfaces</span>
					<input type="text" id="ListenAddr" name="ListenAddr" value="{{.EngineOptions.ListenAddr}}" />
				</label>

				<label for="Connect">
					<span>Address and port to send OSC feedback to. 255.255.255.255 broadcasts to all network interfaces</span>
					<input type="text" id="Connect" name="Connect" value="{{.EngineOptions.Connect}}" />
				</label>
			</fieldset>
			<fieldset>
				<legend>Config interface</legend>

				<label for="HTTPUser">
					<span>Username for the web configuration interface</span>
					<input type="text" id="HTTPUser" name="HTTPUser" value="{{.HTTPUser}}" />
				</label>

				<label for="HTTPUser">
					<span>Password for the web configuration interface</span>
					<input type="text" id="HTTPPassword" name="HTTPPassword" value="{{.HTTPPassword}}" />
				</label>

				<label for="DisableHTTP">
					<span>Disable this web configuration interface. Undoing this needs access to the SD-card</span>
					<input type="checkbox" id="DisableHTTP" name="DisableHTTP" {{if .DisableHTTP}} checked {{end}}/>
				</label>

				<label for="HTTPPort">
					<span>Port to listen for the web configuration. Needs to be in format of ":1234"</span>
					<input type="text" id="HTTPPort" name="HTTPPort" value="{{.HTTPPort}}" />
				</label>
			</fieldset>
			<fieldset>
				<legend>LTC</legend>

				<label for="DisableLTC">
					<span>Disable LTC display</span>
					<input type="checkbox" id="DisableLTC" name="DisableLTC" {{if .EngineOptions.DisableLTC}} checked {{end}}/>
				</label>

				<label for="LTCSeconds">
					<span>Controls what is displayed on the clock ring in LTC mode, unchecked = frames, checked = seconds</span>
					<input type="checkbox" id="LTCSeconds" name="LTCSeconds" {{if .EngineOptions.LTCSeconds}} checked {{end}}/>
				</label>

				<label for="LTCFollow">
					<span>Continue on internal clock if LTC signal is lost. If unset display will blank when signal is gone</span>
					<input type="checkbox" id="LTCFollow" name="LTCFollow" {{if .EngineOptions.LTCFollow}} checked {{end}}/>
				</label>

			</fieldset>

			{{if .Raspberry}}
				<fieldset>
					<legend>Raspberry pi configuration</legend>

					<label for="configtxt">
						<span>Raspberry pi /boot/config.txt. Changing this will reboot the raspberry pi</span>
						<textarea id="configtxt" name="configtxt" rows="20" cols="50">{{.ConfigTxt}}</textarea>
					</label>
				</fieldset>
			{{end}}

			<input type="submit" value="Save config and restart clock" />
		</form>
	</div>


	<style type="text/css">
		h1 {
			color: F072A9;
			font-weight: bold;
			text-shadow: 1px 1px 1px #fff;
		}
		.errors {
			border-radius: 10px;
			-webkit-border-radius: 10px;
			-moz-border-radius: 10px;
			margin: 0px 0px 10px 0px;
			border: 1px solid red;
			padding: 20px;
			background: #FFF4F4;
			box-shadow: inset 0px 0px 15px #FFE5E5;
			-moz-box-shadow: inset 0px 0px 15px #FFE5E5;
			-webkit-box-shadow: inset 0px 0px 15px #FFE5E5;
			max-width: 760px;

		}
		.config-form{
			max-width: 800px;
			font-family: "Lucida Sans Unicode", "Lucida Grande", sans-serif;
		}
		p{
			color: #F072A9;
			font-weight: bold;
			font-size: 13px;
			text-shadow: 1px 1px 1px #fff;
		}
		.errors p{
			color: red;
			font-weight: bold;
			font-size: 13px;
			text-shadow: 1px 1px 1px #fff;
		}
		.config-form li{
			color: #F072A9;
			font-weight: bold;
			font-size: 13px;
			text-shadow: 1px 1px 1px #fff;
		}
		.errors li{
			color: red;
			font-weight: bold;
			font-size: 13px;
			text-shadow: 1px 1px 1px #fff;
		}
		.config-form label{
			display:block;
			margin-bottom: 10px;
			overflow: auto;
		}
		.config-form label > span{
			float: left;
			width: 300px;
			color: #F072A9;
			font-weight: bold;
			font-size: 13px;
			text-shadow: 1px 1px 1px #fff;
		}
		.config-form fieldset{
			border-radius: 10px;
			-webkit-border-radius: 10px;
			-moz-border-radius: 10px;
			margin: 0px 0px 10px 0px;
			border: 1px solid #FFD2D2;
			padding: 20px;
			background: #FFF4F4;
			box-shadow: inset 0px 0px 15px #FFE5E5;
			-moz-box-shadow: inset 0px 0px 15px #FFE5E5;
			-webkit-box-shadow: inset 0px 0px 15px #FFE5E5;
		}
		.config-form fieldset legend{
			color: #FFA0C9;
			border-top: 1px solid #FFD2D2;
			border-left: 1px solid #FFD2D2;
			border-right: 1px solid #FFD2D2;
			border-radius: 5px 5px 0px 0px;
			-webkit-border-radius: 5px 5px 0px 0px;
			-moz-border-radius: 5px 5px 0px 0px;
			background: #FFF4F4;
			padding: 0px 8px 3px 8px;
			box-shadow: -0px -1px 2px #F1F1F1;
			-moz-box-shadow:-0px -1px 2px #F1F1F1;
			-webkit-box-shadow:-0px -1px 2px #F1F1F1;
			font-weight: normal;
			font-size: 12px;
		}
		.config-form textarea{
			width:250px;
			height:100px;
		}
		.config-form input,
		.config-form select,
		.config-form textarea{
			border-radius: 3px;
			-webkit-border-radius: 3px;
			-moz-border-radius: 3px;
			border: 1px solid #FFC2DC;
			outline: none;
			color: #F072A9;
			padding: 5px 8px 5px 8px;
			box-shadow: inset 1px 1px 4px #FFD5E7;
			-moz-box-shadow: inset 1px 1px 4px #FFD5E7;
			-webkit-box-shadow: inset 1px 1px 4px #FFD5E7;
			background: #FFEFF6;
			width:50%;
		}
		.config-form  input[type=checkbox]{
			width:20px;
		}
		.config-form  input[type=submit],
		.config-form  input[type=button]{
			background: #EB3B88;
			border: 1px solid #C94A81;
			padding: 5px 15px 5px 15px;
			color: #FFCBE2;
			box-shadow: inset -1px -1px 3px #FF62A7;
			-moz-box-shadow: inset -1px -1px 3px #FF62A7;
			-webkit-box-shadow: inset -1px -1px 3px #FF62A7;
			border-radius: 3px;
			border-radius: 3px;
			-webkit-border-radius: 3px;
			-moz-border-radius: 3px;
			font-weight: bold;
			max-width: 800px;
			width: 100%;
		}
		.required{
			color:red;
			font-weight:normal;
		}
	</style>
</body>
</html>
`
