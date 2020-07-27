package main

import (
	"gitlab.com/Depili/clock-8001/v3/clock"
	"gitlab.com/Depili/clock-8001/v3/util"
)

var winTitle = "SDL CLOCK"
var winWidth, winHeight int32 = 1920, 1080
var gridStartX int32 = 149
var gridStartY int32 = 149
var gridSize int32 = 20
var gridSpacing = 25

const center1080 = 1080 / 2
const center192 = 192 / 2
const staticRadius1080 = 500
const secondRadius1080 = 450
const staticRadius192 = staticRadius1080 * 192 / 1080
const secondRadius192 = secondRadius1080 * 192 / 1080

var secCircles []util.Point
var staticCircles []util.Point

type clockOptions struct {
	Config         func(s string) error `short:"C" long:"config" description:"read config from a file"`
	configFile     string
	Small          bool   `short:"s" description:"Scale to 192x192px"`
	Font           string `short:"F" long:"font" description:"Font for event name" default:"fonts/7x13.bdf"`
	TextRed        uint8  `short:"r" long:"red" description:"Red component of text color" default:"255"`
	TextGreen      uint8  `short:"g" long:"green" description:"Green component of text color" default:"128"`
	TextBlue       uint8  `short:"b" long:"blue" description:"Blue component of text color" default:"0"`
	StaticRed      uint8  `long:"static-red" description:"Red component of static color" default:"80"`
	StaticGreen    uint8  `long:"static-green" description:"Green component of static color" default:"80"`
	StaticBlue     uint8  `long:"static-blue" description:"Blue component of static color" default:"0"`
	SecRed         uint8  `long:"sec-red" description:"Red component of second color" default:"200"`
	SecGreen       uint8  `long:"sec-green" description:"Green component of second color" default:"0"`
	SecBlue        uint8  `long:"sec-blue" description:"Blue component of second color" default:"0"`
	TimePin        int    `short:"p" long:"time-pin" description:"Pin to select foreign timezone, active low" default:"15"`
	Debug          bool   `long:"debug" description:"Enable debug output"`
	HTTPPort       string `long:"http-port" description:"Port to listen on for the http configuration interface" default:":8080"`
	DisableHTTP    bool   `long:"disable-http" description:"Disable the web configuration interface"`
	HTTPUser       string `long:"http-user" description:"Username for web configuration" default:"admin"`
	HTTPPassword   string `long:"http-password" description:"Password for web configuration interface" default:"clockwork"`
	DualClock      bool   `long:"dual-clock" description:"Display two clock faces, with one of them being constant time of day display"`
	DumpConfig     bool   `long:"dump-config" description:"Write configuration to stdout and exit"`
	NoARCorrection bool   `long:"no-ar-correction" description:"Do not try to detect official raspberry pi display and correct it's aspect ratio"`
	Background     string `long:"background" description:"Background image file location."`
	EngineOptions  *clock.EngineOptions
}

var options clockOptions

// Pixel coordinates for the 192x192 pixel clock circles
var circlePixels = [][2]int32{
	{0, 1},
	{0, 2},
	{0, 3},
	{1, 0},
	{1, 1},
	{1, 2},
	{1, 3},
	{1, 4},
	{2, 0},
	{2, 1},
	{2, 2},
	{2, 3},
	{2, 4},
	{3, 0},
	{3, 1},
	{3, 2},
	{3, 3},
	{3, 4},
	{4, 1},
	{4, 2},
	{4, 3},
}
