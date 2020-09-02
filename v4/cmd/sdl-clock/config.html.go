package main

const configHTML = `
<html>
<head>
	<title>Clock-8001 configuration</title>
</head>
	<body>
		{{if .Errors}}
			<div class="errors">
				<p>
					Following errors prevented the configuration from being saved:
					{{.Errors}}
				</p>
			</div>
		{{end}}
		<div class="config-form">
			<h1>Clock configuration editor</h1>
			<form action="/save" method="post">
				<fieldset>
					<legend>General settings</legend>
					<label for="Face">
						<span>Select the clock face to use</span>
						<select name="Face" id="Face">
							<option value="round" {{if eq .Face "round"}} selected {{end}}>Single round clock</option>
							<option value="dual-round" {{if eq .Face "dual-round"}} selected {{end}}>Dual round clocks</option>
							<option value="text" {{if eq .Face "text"}} selected {{end}}>Text clock</option>
							<option value="small" {{if eq .Face  "small"}} selected {{end}}>Small 192x192px round clock</option>
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

					<label for="NumberFont">
						<span>Font filename for text clock numbers</span>
						<input type="text" id="NumberFont" name="NumberFont" value="{{.NumberFont}}" />
					</label>

					<label for="LabelFont">
						<span>Font filename for text clock labels</span>
						<input type="text" id="LabelFont" name="LabelFont" value="{{.LabelFont}}" />
					</label>

					<label for="IconFont">
						<span>Font filename for text clock icons</span>
						<input type="text" id="IconFont" name="IconFont" value="{{.IconFont}}" />
					</label>

					<label for="Font">
						<span>Font filename for round clocks</span>
						<input type="text" id="Font" name="Font" value="{{.Font}}" />
					</label>

					<label for="Flash">
						<span>Flashing interval in milliseconds for ellapsed countdowns</span>
						<input type="number" min="0" id="Flash" name="Flash" value="{{.EngineOptions.Flash}}" />
					</label>

					<label for="Flash">
						<span>Timeout for clearing OSC text display messages, milliseconds</span>
						<input type="number" min="0" id="Timeout" name="Timeout" value="{{.EngineOptions.Timeout}}" />
					</label>

					<label for="Background">
						<span>Background image filename</span>
						<input type="text" id="Background" name="Background" value="{{.Background}}" />
					</label>
					<p>The image needs to be in the correct resolution and either png or jpeg file. Place the
					image in the fat partition and refer to it as /boot/imagename.png</p>

					<label for="BackgroundColor">
						<span>Background color, used if no background image is provided</span>
						<input type="color" id="BackgroundColor" name="BackgroundColor" value="{{.BackgroundColor}}" />
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
					<li>UDP protocol from Interspace / stage timer (not yet implemented)</li>
					<li>Associated timer if it is running</li>
					<li>Time of day in the selected time zone</li>
					<li>Blank display</li>
				</ol>

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

					<label for="source1-udp">
						<span>Enable UDP input on this source</span>
						<input type="checkbox" id="source1-udp" name="source1-udp" {{if .EngineOptions.Source1.UDP}} checked {{end}} />
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
						<input type="text" id="source1-timezone" name="source1-timezone" value="{{.EngineOptions.Source1.TimeZone}}" />
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

					<label for="source2-udp">
						<span>Enable UDP input on this source</span>
						<input type="checkbox" id="source2-udp" name="source2-udp" {{if .EngineOptions.Source2.UDP}} checked {{end}} />
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
						<input type="text" id="source2-timezone" name="source2-timezone" value="{{.EngineOptions.Source2.TimeZone}}" />
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

					<label for="source3-udp">
						<span>Enable UDP input on this source</span>
						<input type="checkbox" id="source3-udp" name="source3-udp" {{if .EngineOptions.Source3.UDP}} checked {{end}} />
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
						<input type="text" id="source3-timezone" name="source3-timezone" value="{{.EngineOptions.Source3.TimeZone}}" />
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

					<label for="source4-udp">
						<span>Enable UDP input on this source</span>
						<input type="checkbox" id="source4-udp" name="source4-udp" {{if .EngineOptions.Source4.UDP}} checked {{end}} />
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
						<input type="text" id="source4-timezone" name="source4-timezone" value="{{.EngineOptions.Source4.TimeZone}}" />
					</label>
				</fieldset>
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

					<label for="Row2Color">
						<span>Color timer row 2</span>
						<input type="color" id="Row2Color" name="Row2Color" value="{{.Row2Color}}" />
					</label>

					<label for="Row3Color">
						<span>Color timer row 3</span>
						<input type="color" id="Row3Color" name="Row3Color" value="{{.Row3Color}}" />
					</label>

					<label for="LabelColor">
						<span>Color labels</span>
						<input type="color" id="LabelColor" name="LabelColor" value="{{.LabelColor}}" />
					</label>

					<label for="DrawBoxes">
						<span>Draw background boxes for labels and timers</span>
						<input type="checkbox" id="DrawBoxes" name="DrawBoxes" {{if .DrawBoxes}} checked {{end}}/>
					</label>

					<label for="LabelBG">
						<span>Background color for labels</span>
						<input type="color" id="LabelBG" name="LabelBG" value="{{.LabelBG}}" />
					</label>

					<label for="TimerBG">
						<span>Background color for timers</span>
						<input type="color" id="TimerBG" name="TimerBG" value="{{.TimerBG}}" />
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
		.config-form{
			max-width: 800px;
			font-family: "Lucida Sans Unicode", "Lucida Grande", sans-serif;
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
