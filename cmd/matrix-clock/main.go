package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"github.com/stanchan/clock-8001/clock"
	"github.com/stanchan/go-rgb-led-matrix/bdf"
	"github.com/stanchan/go-rgb-led-matrix/matrix"
	"github.com/tarm/serial"
	"os"
	"os/signal"
	"time"
)

var options struct {
	Font          string `short:"F" long:"font" description:"Font for event name" default:"fonts/7x13.bdf"`
	Matrix        string `short:"m" long:"matrix" description:"Matrix to connect to" required:"true"`
	SerialName    string `long:"serial-name" description:"Serial device for arduino" default:"/dev/ttyUSB0"`
	SerialBaud    int    `long:"serial-baud" value-name:"BAUD" default:"57600"`
	TextRed       int    `short:"r" long:"red" description:"Red component of text color" default:"128"`
	TextGreen     int    `short:"g" long:"green" description:"Green component of text color" default:"128"`
	TextBlue      int    `short:"b" long:"blue" description:"Blue component of text color" default:"0"`
	TimePin       int    `short:"p" long:"time-pin" description:"Pin to select foreign timezone, active low" default:"15"`
	EngineOptions *clock.EngineOptions
}

var parser = flags.NewParser(&options, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		panic(err)
	}

	// Serial connection for the led ring
	serialConfig := serial.Config{
		Name: options.SerialName,
		Baud: options.SerialBaud,
		// ReadTimeout:    options.SerialTimeout,
	}

	serial, err := serial.OpenPort(&serialConfig)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Serial open.\n")

	// GPIO pin for toggling between timezones
	if err := embd.InitGPIO(); err != nil {
		panic(err)
	}

	timePin, err := embd.NewDigitalPin(options.TimePin)
	if err != nil {
		panic(err)
	} else if err := timePin.SetDirection(embd.In); err != nil {
		panic(err)
	}

	fmt.Printf("GPIO initialized.\n")

	// Parse font for clock text
	font, err := bdf.Parse(options.Font)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Fonts loaded.\n")

	// Initialize the led matrix library
	m := matrix.Init(options.Matrix, 32, 32)
	defer m.Close()

	// Trap SIGINT aka Ctrl-C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Clock text color from flags
	textColor := [3]byte{byte(options.TextRed), byte(options.TextGreen), byte(options.TextBlue)}
	// Default color for the OSC field (black)
	tallyColor := [3]byte{0x00, 0x00, 0x00}

	// Create initial bitmaps for clock text
	hourBitmap := font.TextBitmap("15")
	minuteBitmap := font.TextBitmap("04")
	secondBitmap := font.TextBitmap("05")
	tallyBitmap := font.TextBitmap("  ")

	updateTicker := time.NewTicker(time.Millisecond * 10)
	send := make([]byte, 1)

	engine, err := clock.MakeEngine(options.EngineOptions)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-sigChan:
			// SIGINT received, shutdown gracefully
			m.Close()
			os.Exit(1)
		case <-updateTicker.C:
			engine.Update()
			seconds := engine.Leds
			hourBitmap = font.TextBitmap(engine.Hours)
			minuteBitmap = font.TextBitmap(engine.Minutes)
			secondBitmap = font.TextBitmap(engine.Seconds)

			tallyColor = [3]byte{byte(engine.TallyRed), byte(engine.TallyGreen), byte(engine.TallyBlue)}
			tallyBitmap = font.TextBitmap(engine.Tally)

			// Clear the matrix buffer
			m.Fill(matrix.ColorBlack())

			if engine.Dots {
				// Draw the dots between hours and minutes
				m.SetPixel(14, 15, textColor)
				m.SetPixel(14, 16, textColor)
				m.SetPixel(15, 15, textColor)
				m.SetPixel(15, 16, textColor)

				m.SetPixel(18, 15, textColor)
				m.SetPixel(18, 16, textColor)
				m.SetPixel(19, 15, textColor)
				m.SetPixel(19, 16, textColor)
			}
			// Draw the text
			m.Scroll(hourBitmap, textColor, 10, 0, 0, 14)
			m.Scroll(minuteBitmap, textColor, 10, 17, 0, 14)
			m.Scroll(secondBitmap, textColor, 21, 8, 0, 14)
			m.Scroll(tallyBitmap, tallyColor, 0, 2, 0, 30)

			// Update the led ring via serial
			send[0] = byte(seconds)
			_, err := serial.Write(send)
			if err != nil {
				panic(err)
			}
			m.Send()
		}
	}
}
