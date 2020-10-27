package main

import (
	"gitlab.com/Depili/clock-8001/v4/clock"
	"gitlab.com/Depili/clock-8001/v4/util"
	htmlTemplate "html/template"
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

type optionsColor struct {
	R uint8 `long:"red" description:"Red component of the color"`
	G uint8 `long:"green" description:"Green component of the color"`
	B uint8 `long:"blue" description:"Blue component of the color"`
}

type clockOptions struct {
	Config          func(s string) error `short:"C" long:"config" description:"read config from a file"`
	Face            string               `long:"face" description:"Select the clock face to use" default:"round" choice:"round" choice:"dual-round" choice:"small" choice:"text"`
	Debug           bool                 `long:"debug" description:"Enable debug output"`
	HTTPPort        string               `long:"http-port" description:"Port to listen on for the http configuration interface" default:":8080"`
	DisableHTTP     bool                 `long:"disable-http" description:"Disable the web configuration interface"`
	HTTPUser        string               `long:"http-user" description:"Username for web configuration" default:"admin"`
	HTTPPassword    string               `long:"http-password" description:"Password for web configuration interface" default:"clockwork"`
	DumpConfig      bool                 `long:"dump-config" description:"Write configuration to stdout and exit"`
	Defaults        bool                 `long:"defaults" description:"load defaults"`
	NoARCorrection  bool                 `long:"no-ar-correction" description:"Do not try to detect official raspberry pi display and correct it's aspect ratio"`
	Background      string               `long:"background" description:"Background image file location."`
	BackgroundPath  string               `long:"background-path" description:"path to load OSC backgrounds from" default:"/boot"`
	BackgroundColor string               `long:"background-color" description:"Background color, used if no background image is supplied" default:"#000000"`
	EngineOptions   *clock.EngineOptions

	// Round clock stuff
	Font           string `short:"F" long:"font" description:"Font for event name" default:"fonts/7x13.bdf"`
	TextColor      string `long:"text-color" description:"Color for round clock text" default:"#FF8000"`
	StaticColor    string `long:"static-color" description:"Color for round clock static circles" default:"#505000"`
	SecondColor    string `long:"second-color" description:"Color for round clock second circles" default:"#C80000"`
	CountdownColor string `long:"countdown-color" description:"Color for round clock second circles" default:"#FF0000"`

	// Text clock stuff
	NumberFont     string `long:"number-font" description:"Font for text clock face numbers" default:"Copse-Regular.ttf"`
	LabelFont      string `long:"label-font" description:"Font for text clock face labels" default:"RobotoMono-VariableFont_wght.ttf"`
	IconFont       string `long:"icon-font" description:"Font for text clock face icons" default:"MaterialIcons-Regular.ttf"`
	Row1Color      string `long:"row1-color" description:"Color for text clock row 1" default:"#FF8000"`
	Row2Color      string `long:"row2-color" description:"Color for text clock row 2" default:"#FF8000"`
	Row3Color      string `long:"row3-color" description:"Color for text clock row 3" default:"#FF8000"`
	Row1Hide       bool   `long:"row1-hide" description:"Hide timer row 1"`
	Row2Hide       bool   `long:"row2-hide" description:"Hide timer row 2"`
	Row3Hide       bool   `long:"row3-hide" description:"Hide timer row 3"`
	LabelColor     string `long:"label-color" description:"Color for text clock labels" default:"#FF8000"`
	TimerBG        string `long:"timer-bg-color" description:"Color for optional timer background box" default:"#202020"`
	LabelBG        string `long:"label-bg-color" description:"Color for optional label background box" default:"#202020"`
	Rows           int    `long:"text-rows" description:"Number of timer lines to display" default:"3" choice:"0" choice:"1" choice:"2" choice:"3"`
	DrawBoxes      bool   `long:"draw-boxes" description:"Draw the container boxes for timers"`
	NumberFontSize int    `long:"numbers-size" default:"200"`

	Raspberry bool   // Is the host a raspberry pi
	ConfigTxt string // /boot/config.txt contents

	configFile string
	small      bool
	dualClock  bool
	textClock  bool
	Errors     htmlTemplate.HTML // For passing errors to the html template
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
