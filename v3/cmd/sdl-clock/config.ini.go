package main

const configTemplate = `# Example configuration file for clock-8001
# Lines starting with '#' are comments and
# are ignored by clock-8001

# Username and password for the web configuration interface
HTTPUser={{.HTTPUser}}
HTTPPassword={{.HTTPPassword}}

# set to true for 192x192 clock
Small={{.Small}}

# Font to use
Font={{.Font}}

# Color for text
TextRed={{.TextRed}}
TextGreen={{.TextGreen}}
TextBlue={{.TextBlue}}

# Color for the 12 static "hour" markers
StaticRed={{.StaticRed}}
StaticGreen={{.StaticGreen}}
StaticBlue={{.StaticBlue}}

# Color for the second ring dots
SecRed={{.SecRed}}
SecGreen={{.SecGreen}}
SecBlue={{.SecBlue}}

# Color for the secondary countdown display
CountdownRed={{.EngineOptions.CountdownRed}}
CountdownGreen={{.EngineOptions.CountdownGreen}}
CountdownBlue={{.EngineOptions.CountdownBlue}}

# Set to true to output verbose debug information
Debug={{.Debug}}

# Flashing interval for ellapsed countdowns, in milliseconds
Flash={{.EngineOptions.Flash}}

# Timezone
Timezone={{.EngineOptions.Timezone}}

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
`
