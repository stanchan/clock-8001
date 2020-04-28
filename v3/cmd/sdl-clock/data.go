package main

import (
	"gitlab.com/Depili/clock-8001/v3/clock"
)

var winTitle = "SDL CLOCK"
var winWidth, winHeight int32 = 1080, 1080
var gridStartX int32 = 149
var gridStartY int32 = 149
var gridSize int32 = 20
var gridSpacing = 25

type clockOptions struct {
	Config        func(s string) error `short:"C" long:"config" description:"read config from a file"`
	configFile    string
	Small         bool   `short:"s" description:"Scale to 192x192px"`
	Font          string `short:"F" long:"font" description:"Font for event name" default:"fonts/7x13.bdf"`
	TextRed       uint8  `short:"r" long:"red" description:"Red component of text color" default:"255"`
	TextGreen     uint8  `short:"g" long:"green" description:"Green component of text color" default:"128"`
	TextBlue      uint8  `short:"b" long:"blue" description:"Blue component of text color" default:"0"`
	StaticRed     uint8  `long:"static-red" description:"Red component of static color" default:"80"`
	StaticGreen   uint8  `long:"static-green" description:"Green component of static color" default:"80"`
	StaticBlue    uint8  `long:"static-blue" description:"Blue component of static color" default:"0"`
	SecRed        uint8  `long:"sec-red" description:"Red component of second color" default:"200"`
	SecGreen      uint8  `long:"sec-green" description:"Green component of second color" default:"0"`
	SecBlue       uint8  `long:"sec-blue" description:"Blue component of second color" default:"0"`
	TimePin       int    `short:"p" long:"time-pin" description:"Pin to select foreign timezone, active low" default:"15"`
	Debug         bool   `long:"debug" description:"Enable debug output"`
	HTTPPort      string `long:"http-port" description:"Port to listen on for the http configuration interface" default:":8080"`
	DisableHTTP   bool   `long:"disable-http" description:"Disable the web configuration interface"`
	HTTPUser      string `long:"http-user" description:"Username for web configuration" default:"admin"`
	HTTPPassword  string `long:"http-password" description:"Password for web configuration interface" default:"clockwork"`
	EngineOptions *clock.EngineOptions
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

// Radius = 500
// 12 "Hour" static circles
var staticCircles = [12][2]int32{
	{540, 40},
	{790, 107},
	{973, 290},
	{1040, 540},
	{973, 790},
	{790, 973},
	{540, 1040},
	{290, 973},
	{107, 790},
	{40, 540},
	{107, 290},
	{290, 107},
}

// Radius = 450
// Second circles
var secCircles = [60][2]int32{
	{540, 90},
	{587, 92},
	{634, 100},
	{679, 112},
	{723, 129},
	{765, 150},
	{805, 176},
	{841, 206},
	{874, 239},
	{904, 275},
	{930, 315},
	{951, 357},
	{968, 401},
	{980, 446},
	{988, 493},
	{990, 540},
	{988, 587},
	{980, 634},
	{968, 679},
	{951, 723},
	{930, 765},
	{904, 805},
	{874, 841},
	{841, 874},
	{805, 904},
	{765, 930},
	{723, 951},
	{679, 968},
	{634, 980},
	{587, 988},
	{540, 990},
	{493, 988},
	{446, 980},
	{401, 968},
	{357, 951},
	{315, 930},
	{275, 904},
	{239, 874},
	{206, 841},
	{176, 805},
	{150, 765},
	{129, 723},
	{112, 679},
	{100, 634},
	{92, 587},
	{90, 540},
	{92, 493},
	{100, 446},
	{112, 401},
	{129, 357},
	{150, 315},
	{176, 275},
	{206, 239},
	{239, 206},
	{275, 176},
	{315, 150},
	{357, 129},
	{401, 112},
	{446, 100},
	{493, 92},
}

// radius 500 * 192 / 1080
var smallStaticCircles = [12][2]int32{
	{96, 8},
	{140, 20},
	{172, 52},
	{184, 96},
	{172, 140},
	{140, 172},
	{96, 184},
	{52, 172},
	{20, 140},
	{8, 96},
	{20, 52},
	{52, 20},
}

// n = 60, center = 192 / 2, radius = 450 * 192 / 1080
var smallSecCircles = [60][2]int32{
	{96, 16},
	{104, 16},
	{113, 18},
	{121, 20},
	{129, 23},
	{136, 27},
	{143, 31},
	{150, 37},
	{155, 42},
	{161, 49},
	{165, 56},
	{169, 63},
	{172, 71},
	{174, 79},
	{176, 88},
	{176, 96},
	{176, 104},
	{174, 113},
	{172, 121},
	{169, 129},
	{165, 136},
	{161, 143},
	{155, 150},
	{150, 155},
	{143, 161},
	{136, 165},
	{129, 169},
	{121, 172},
	{113, 174},
	{104, 176},
	{96, 176},
	{88, 176},
	{79, 174},
	{71, 172},
	{63, 169},
	{56, 165},
	{49, 161},
	{42, 155},
	{37, 150},
	{31, 143},
	{27, 136},
	{23, 129},
	{20, 121},
	{18, 113},
	{16, 104},
	{16, 96},
	{16, 88},
	{18, 79},
	{20, 71},
	{23, 63},
	{27, 56},
	{31, 49},
	{37, 42},
	{42, 37},
	{49, 31},
	{56, 27},
	{63, 23},
	{71, 20},
	{79, 18},
	{88, 16},
}
