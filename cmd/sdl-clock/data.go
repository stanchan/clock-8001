package main

import (
	"gitlab.com/Depili/clock-8001/clock"
)

var winTitle = "SDL CLOCK"
var winWidth, winHeight int32 = 1920, 1080
var gridStartX int32 = 569
var gridStartY int32 = 149
var gridSize int32 = 20
var gridSpacing = 25

var options struct {
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
	EngineOptions *clock.EngineOptions
}

// 12 "Hour" static circles
var staticCircles = [12][2]int32{
	{1393, 790},
	{1210, 973},
	{960, 1040},
	{710, 973},
	{526, 790},
	{460, 540},
	{526, 290},
	{709, 106},
	{959, 40},
	{1210, 106},
	{1393, 289},
	{1460, 539},
}

// Second circles
var secCircles = [60][2]int32{
	{959, 90}, // 0
	{1007, 92},
	{1053, 99},
	{1099, 112},
	{1143, 128},
	{1185, 150}, // 5
	{1224, 175},
	{1261, 205},
	{1294, 238},
	{1324, 275},
	{1349, 315}, // 10
	{1371, 356},
	{1387, 400},
	{1400, 446},
	{1407, 492},
	{1410, 539}, // 15
	{1407, 587},
	{1400, 633},
	{1387, 679},
	{1371, 723},
	{1349, 765}, // 20
	{1324, 804},
	{1294, 841},
	{1261, 874},
	{1224, 904},
	{1185, 929}, // 25
	{1143, 951},
	{1099, 967},
	{1053, 980},
	{1007, 987},
	{960, 990}, // 30
	{912, 987},
	{866, 980},
	{820, 967},
	{776, 951},
	{735, 929}, // 35
	{695, 904},
	{658, 874},
	{625, 841},
	{595, 804},
	{570, 765}, // 40
	{548, 723},
	{532, 679},
	{519, 633},
	{512, 587},
	{510, 540}, // 45
	{512, 492},
	{519, 446},
	{532, 400},
	{548, 356},
	{570, 315}, // 50
	{595, 275},
	{625, 238},
	{658, 205},
	{695, 175},
	{734, 150}, // 55
	{776, 128},
	{820, 112},
	{866, 99},
	{912, 92},
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
