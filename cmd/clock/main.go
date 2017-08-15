package main

import (
	"fmt"
	"github.com/SpComb/osc-tally/clock"
	"github.com/depili/go-rgb-led-matrix/bdf"
	"github.com/depili/go-rgb-led-matrix/matrix"
	"github.com/hypebeast/go-osc/osc"
	"github.com/jessevdk/go-flags"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"github.com/tarm/serial"
	"log"
	"os"
	"os/signal"
	"time"
)

var Options struct {
	Font        string `short:"F" long:"font" description:"Font for event name" default:"fonts/7x13.bdf"`
	Matrix      string `short:"m" long:"matrix" description:"Matrix to connect to" required:"true"`
	SerialName  string `long:"serial-name" description:"Serial device for arduino" default:"/dev/ttyUSB0"`
	SerialBaud  int    `long:"serial-baud" value-name:"BAUD" default:"57600"`
	TextRed     int    `short:"r" long:"red" description:"Red component of text color" default:"128"`
	TextGreen   int    `short:"g" long:"green" description:"Green component of text color" default:"128"`
	TextBlue    int    `short:"b" long:"blue" description:"Blue component of text color" default:"0"`
	LocalTime   string `short:"t" long:"local-time" description:"Local timezone" default:"Europe/Helsinki"`
	ForeignTime string `short:"T" long:"foreign-time" description:"Foreign timezone" default:"Europe/Moscow"`
	TimePin     int    `short:"p" long:"time-pin" description:"Pin to select foreign timezone, active low" default:"15"`
	ListenAddr  string `long:"osc-listen" description:"Address to listen for incoming osc messages" required:"true"`
}

var parser = flags.NewParser(&Options, flags.Default)

func runOSC(oscServer *osc.Server) {
	err := oscServer.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func main() {
	if _, err := parser.Parse(); err != nil {
		panic(err)
	}

	// Initialize osc listener
	var oscServer = osc.Server{
		Addr: Options.ListenAddr,
	}

	var clockServer = clock.MakeServer(&oscServer)
	log.Printf("osc server: listen %v", oscServer.Addr)

	go runOSC(&oscServer)

	// Serial connection for the led ring
	serialConfig := serial.Config{
		Name: Options.SerialName,
		Baud: Options.SerialBaud,
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

	timePin, err := embd.NewDigitalPin(Options.TimePin)
	if err != nil {
		panic(err)
	} else if err := timePin.SetDirection(embd.In); err != nil {
		panic(err)
	}

	fmt.Printf("GPIO initialized.\n")

	// Load timezones
	local, err := time.LoadLocation(Options.LocalTime)
	if err != nil {
		panic(err)
	}

	foreign, err := time.LoadLocation(Options.ForeignTime)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Timezones loaded.\n")

	// Parse font for clock text
	font, err := bdf.Parse(Options.Font)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Fonts loaded.\n")

	// Initialize the led matrix library
	m := matrix.Init(Options.Matrix, 32, 32)
	defer m.Close()

	// Trap SIGINT aka Ctrl-C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Clock text color from flags
	textColor := [3]byte{byte(Options.TextRed), byte(Options.TextGreen), byte(Options.TextBlue)}
	// Default color for the OSC field (black)
	tallyColor := [3]byte{0x00, 0x00, 0x00}

	// Create initial bitmaps for clock text
	hourBitmap := font.TextBitmap("15")
	minuteBitmap := font.TextBitmap("04")
	secondBitmap := font.TextBitmap("05")
	tallyBitmap := font.TextBitmap("  ")

	updateTicker := time.NewTicker(time.Millisecond * 10)
	send := make([]byte, 1)
	oscChan := clockServer.Listen()

	for {
		select {
		case msg := <-oscChan:
			// New OSC message received
			fmt.Printf("Got new osc data.\n")
			tallyColor = [3]byte{byte(msg.ColorRed), byte(msg.ColorGreen), byte(msg.ColorBlue)}
			tallyBitmap = font.TextBitmap(fmt.Sprintf("%1s%02d%1s", msg.Symbol, msg.Count, msg.Unit))
		case <-sigChan:
			// SIGINT received, shutdown gracefully
			m.Close()
			os.Exit(1)
		case <-updateTicker.C:
			// Update the time shown on the matrix
			var t time.Time
			if i, _ := timePin.Read(); err != nil {
				panic(err)
			} else if i == 1 {
				t = time.Now().In(local)
			} else {
				t = time.Now().In(foreign)
			}

			// Check that the rpi has valid time
			if t.Year() > 2000 {
				hourBitmap = font.TextBitmap(t.Format("15"))
				minuteBitmap = font.TextBitmap(t.Format("04"))
				secondBitmap = font.TextBitmap(t.Format("05"))
			} else {
				// No valid time, indicate it with "XX" as the time
				hourBitmap = font.TextBitmap("XX")
				minuteBitmap = font.TextBitmap("XX")
				secondBitmap = font.TextBitmap("")
			}
			seconds := t.Second()
			// Clear the matrix buffer
			m.Fill(matrix.ColorBlack())

			// Draw the dots between hours and minutes
			m.SetPixel(14, 15, textColor)
			m.SetPixel(14, 16, textColor)
			m.SetPixel(15, 15, textColor)
			m.SetPixel(15, 16, textColor)

			m.SetPixel(18, 15, textColor)
			m.SetPixel(18, 16, textColor)
			m.SetPixel(19, 15, textColor)
			m.SetPixel(19, 16, textColor)

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
