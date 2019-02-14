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

// radius 500 * 192 / 1080
var smallStaticCircles = [12][2]int32{
	{90, 169},
	{49, 158},
	{18, 128},
	{7, 87},
	{16, 47},
	{45, 17},
	{85, 7},
	{126, 17},
	{157, 47},
	{168, 88},
	{159, 128},
	{130, 158},
}

// Second circles
var secCircles = [60][2]int32{
	{959, 90},
	{1007, 92},
	{1053, 99},
	{1099, 112},
	{1143, 128},
	{1185, 150},
	{1224, 175},
	{1261, 205},
	{1294, 238},
	{1324, 275},
	{1349, 315},
	{1371, 356},
	{1387, 400},
	{1400, 446},
	{1407, 492},
	{1410, 539},
	{1407, 587},
	{1400, 633},
	{1387, 679},
	{1371, 723},
	{1349, 765},
	{1324, 804},
	{1294, 841},
	{1261, 874},
	{1224, 904},
	{1185, 929},
	{1143, 951},
	{1099, 967},
	{1053, 980},
	{1007, 987},
	{960, 990},
	{912, 987},
	{866, 980},
	{820, 967},
	{776, 951},
	{735, 929},
	{695, 904},
	{658, 874},
	{625, 841},
	{595, 804},
	{570, 765},
	{548, 723},
	{532, 679},
	{519, 633},
	{512, 587},
	{510, 540},
	{512, 492},
	{519, 446},
	{532, 400},
	{548, 356},
	{570, 315},
	{595, 275},
	{625, 238},
	{658, 205},
	{695, 175},
	{734, 150},
	{776, 128},
	{820, 112},
	{866, 99},
	{912, 92},
}

// n = 60, center = 192 / 2, radius = 450 * 192 / 1080
var smallSecCircles = [60][2]int32{
	{86, 15},
	{93, 15},
	{101, 16},
	{108, 18},
	{116, 21},
	{122, 24},
	{129, 28},
	{135, 33},
	{141, 39},
	{145, 45},
	{150, 51},
	{153, 58},
	{156, 65},
	{159, 72},
	{160, 80},
	{160, 88},
	{160, 95},
	{159, 103},
	{157, 110},
	{155, 117},
	{152, 124},
	{148, 130},
	{143, 136},
	{138, 142},
	{132, 147},
	{126, 151},
	{119, 154},
	{112, 157},
	{104, 159},
	{97, 160},
	{89, 161},
	{82, 160},
	{74, 159},
	{67, 157},
	{59, 154},
	{53, 151},
	{46, 147},
	{40, 142},
	{34, 136},
	{30, 130},
	{25, 124},
	{22, 117},
	{19, 110},
	{16, 103},
	{15, 95},
	{15, 87},
	{15, 80},
	{16, 72},
	{18, 65},
	{20, 58},
	{23, 51},
	{27, 45},
	{32, 39},
	{37, 33},
	{43, 28},
	{49, 24},
	{56, 21},
	{63, 18},
	{71, 16},
	{78, 15},
}
