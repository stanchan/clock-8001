package main

const configHTML = `
<h1>Clock configuration editor</h1>
<form action="/save" method="post">

<input type="checkbox" id="Small" name="Small" {{if .Small}} checked {{end}}/>
<label for="Small">Render small 192x192 pixel clock</label><br />

<input type="checkbox" id="DualClock" name="DualClock" {{if .DualClock}} checked {{end}}/>
<label for="DualClock">Render two clock faces, one of them always displays time of day</label><br />

<input type="checkbox" id="Format12h" name="Format12h" {{if .EngineOptions.Format12h}} checked {{end}}/>
<label for="Format12h">Use 12 hour format for time-of-day display.</label><br />

<input type="checkbox" id="Debug" name="Debug" {{if .Debug}} checked {{end}}/>
<label for="Debug">Output verbose debug information. This will impact performance.</label><br />

<input type="text" id="Font" name="Font" value="{{.Font}}" />
<label for="Font">Font filename</label><br />

<input type="text" id="Timezone" name="Timezone" value="{{.EngineOptions.Timezone}}" />
<label for="Timezone">Timezone</label><br />

<input type="number" min="0" id="Flash" name="Flash" value="{{.EngineOptions.Flash}}" />
<label for="Flash">Flashing interval in milliseconds for ellapsed countdowns</label><br />

<input type="number" min="0" id="Timeout" name="Timeout" value="{{.EngineOptions.Timeout}}" />
<label for="Flash">Timeout for clearing OSC text display messages, milliseconds</label><br />


<h2>Colors</h2>

<input type="color" id="TextColor" name="TextColor" value="#{{printf "%0.2x%0.2x%0.2x" .TextRed .TextGreen .TextBlue }}" />
<label for="TextColor">Color for text</label><br />

<input type="color" id="SecColor" name="SecColor" value="#{{printf "%0.2x%0.2x%0.2x" .SecRed .SecGreen .SecBlue }}" />
<label for="SecColor">Color for second ring circles</label><br />

<input type="color" id="StaticColor" name="StaticColor" value="#{{printf "%0.2x%0.2x%0.2x" .StaticRed .StaticGreen .StaticBlue }}" />
<label for="StaticColor">Color for 12 static "hour" markers</label><br />

<input type="color" id="CountdownColor" name="CountdownColor" value="#{{printf "%0.2x%0.2x%0.2x" .EngineOptions.CountdownRed .EngineOptions.CountdownGreen .EngineOptions.CountdownBlue }}" />
<label for="CountdownColor">Color for secondary countdown display</label><br />

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

<br />
<input type="submit" value="Save config and restart clock" />


</form>
`
