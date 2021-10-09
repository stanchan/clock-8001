package main

import (
	// "fmt"
	"crypto/subtle"
	"fmt"
	"gitlab.com/Depili/clock-8001/v4/clock"
	htmlTemplate "html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"text/template"
	"time"
)

func runHTTP() {
	if options.configFile == "" {
		// No config file specified, can't save the config
		log.Printf("No config specified, http config interface disabled")
		return
	}

	http.HandleFunc("/save", basicAuth(saveHandler))
	http.HandleFunc("/", basicAuth(indexHandler))
	http.HandleFunc("/export", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Disposition", "attachment;filename=clock.ini")
		http.ServeFile(res, req, options.configFile)
	})
	http.HandleFunc("/import", basicAuth(importHandler))

	log.Printf("HTTP config: listening on %v", options.HTTPPort)
	log.Fatal(http.ListenAndServe(options.HTTPPort, nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := htmlTemplate.New("config.html").Parse(configHTML)
	if err != nil {
		panic(err)
	}

	options.Raspberry = fileExists("/boot/config.txt")

	if options.Raspberry {
		// Read out various config files for editing
		bytes, err := ioutil.ReadFile("/boot/config.txt")
		if err != nil {
			panic(err)
		}
		options.ConfigTxt = string(bytes)
	}

	options.Fonts = make([]string, 0, 200)
	err = filepath.Walk(options.FontPath, func(path string, info os.FileInfo, err error) error {
		if matched, err := filepath.Match("*.ttf", filepath.Base(path)); err != nil {
			return err
		} else if matched {
			options.Fonts = append(options.Fonts, path)
		}
		return nil
	})

	options.Timezones = tzList

	log.Printf("fonts: %v", options.Fonts)
	err = tmpl.Execute(w, options)
	if err != nil {
		panic(err)
	}
}

func importHandler(w http.ResponseWriter, r *http.Request) {

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("import")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	dst, err := os.OpenFile(options.configFile, os.O_RDWR|os.O_CREATE, 0755)
	defer dst.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Writing new config ini file")
	// Copy the uploaded file to the created file on the filesystem
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go delayedExit()

	tmpl, err := htmlTemplate.New("confirm.html").Parse(confirmHTML)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		panic(err)
	}
}

// TODO: validation
func saveHandler(w http.ResponseWriter, r *http.Request) {
	var newOptions clockOptions
	var errors string
	var err error

	newOptions.EngineOptions = &clock.EngineOptions{}
	newOptions.EngineOptions.Source1 = &clock.SourceOptions{}
	newOptions.EngineOptions.Source2 = &clock.SourceOptions{}
	newOptions.EngineOptions.Source3 = &clock.SourceOptions{}
	newOptions.EngineOptions.Source4 = &clock.SourceOptions{}

	// Booleans, no validation on them
	newOptions.Debug = r.FormValue("Debug") != ""
	newOptions.DisableHTTP = r.FormValue("DisableHTTP") != ""
	newOptions.NoARCorrection = r.FormValue("NoARCorrection") != ""
	newOptions.EngineOptions.DisableOSC = r.FormValue("DisableOSC") != ""
	newOptions.EngineOptions.DisableFeedback = r.FormValue("DisableFeedback") != ""
	newOptions.EngineOptions.DisableLTC = r.FormValue("DisableLTC") != ""
	newOptions.EngineOptions.LTCSeconds = r.FormValue("LTCSeconds") != ""
	newOptions.EngineOptions.LTCFollow = r.FormValue("LTCFollow") != ""
	newOptions.EngineOptions.Format12h = r.FormValue("Format12h") != ""

	newOptions.EngineOptions.Source1.LTC = r.FormValue("source1-ltc") != ""
	newOptions.EngineOptions.Source1.Timer = r.FormValue("source1-timer") != ""
	newOptions.EngineOptions.Source1.Tod = r.FormValue("source1-tod") != ""
	newOptions.EngineOptions.Source1.Hidden = r.FormValue("source1-hidden") != ""

	newOptions.EngineOptions.Source2.LTC = r.FormValue("source2-ltc") != ""
	newOptions.EngineOptions.Source2.Timer = r.FormValue("source2-timer") != ""
	newOptions.EngineOptions.Source2.Tod = r.FormValue("source2-tod") != ""
	newOptions.EngineOptions.Source2.Hidden = r.FormValue("source2-hidden") != ""

	newOptions.EngineOptions.Source3.LTC = r.FormValue("source3-ltc") != ""
	newOptions.EngineOptions.Source3.Timer = r.FormValue("source3-timer") != ""
	newOptions.EngineOptions.Source3.Tod = r.FormValue("source3-tod") != ""
	newOptions.EngineOptions.Source3.Hidden = r.FormValue("source3-hidden") != ""

	newOptions.EngineOptions.Source4.LTC = r.FormValue("source4-ltc") != ""
	newOptions.EngineOptions.Source4.Timer = r.FormValue("source4-timer") != ""
	newOptions.EngineOptions.Source4.Tod = r.FormValue("source4-tod") != ""
	newOptions.EngineOptions.Source4.Hidden = r.FormValue("source4-hidden") != ""

	newOptions.DrawBoxes = r.FormValue("DrawBoxes") != ""

	newOptions.EngineOptions.AutoSignals = r.FormValue("auto-signals") != ""
	newOptions.EngineOptions.SignalStart = r.FormValue("signal-start") != ""
	newOptions.SignalFollow = r.FormValue("signal-hw-follow") != ""

	newOptions.AudioEnabled = r.FormValue("AudioEnabled") != ""
	newOptions.TODBeep = r.FormValue("TODBeep") != ""

	// Strings, will not be validated
	newOptions.HTTPUser = r.FormValue("HTTPUser")
	newOptions.HTTPPassword = r.FormValue("HTTPPassword")
	newOptions.EngineOptions.Source1.Text = r.FormValue("source1-text")
	newOptions.EngineOptions.Source2.Text = r.FormValue("source2-text")
	newOptions.EngineOptions.Source3.Text = r.FormValue("source3-text")
	newOptions.EngineOptions.Source4.Text = r.FormValue("source4-text")

	// Clock face type
	newOptions.Face = r.FormValue("Face")
	if f := newOptions.Face; (f != "round") && (f != "dual-round") && (f != "text") && (f != "small") && (f != "single") && (f != "144") && (f != "192") && (f != "288x144") {
		errors += fmt.Sprintf("<li>Clock face selection is invalid (%s)</li>", newOptions.Face)
	}

	// Signal hardware type
	newOptions.SignalType = r.FormValue("signal-hw-type")
	if t := newOptions.SignalType; (t != "unicorn-hd") && (t != "none") {
		errors += fmt.Sprintf("<li>Signal hardware type selection is invalid (%s)</li>", newOptions.SignalType)
	}

	// UDPTime
	newOptions.EngineOptions.UDPTime = r.FormValue("udp-time")
	if f := newOptions.EngineOptions.UDPTime; (f != "off") && (f != "send") && (f != "receive") {
		errors += fmt.Sprintf("<li>UDP time selection is invalid (%s)</li>", newOptions.EngineOptions.UDPTime)
	}

	// Overtime count mode
	newOptions.EngineOptions.OvertimeCountMode = r.FormValue("overtime-count-mode")
	if f := newOptions.EngineOptions.OvertimeCountMode; (f != "zero") && (f != "blank") && (f != "continue") {
		errors += fmt.Sprintf("<li>Overtime count mode selection is invalid (%s)</li>", newOptions.EngineOptions.OvertimeCountMode)
	}

	// Overtime visibility
	newOptions.EngineOptions.OvertimeVisibility = r.FormValue("overtime-visibility")
	if f := newOptions.EngineOptions.OvertimeVisibility; (f != "blink") && (f != "none") && (f != "background") && (f != "both") {
		errors += fmt.Sprintf("<li>Overtime visibility selection is invalid (%s)</li>", newOptions.EngineOptions.OvertimeVisibility)
	}

	// Filenames
	newOptions.NumberFont = r.FormValue("NumberFont")
	errors += validateFile(newOptions.NumberFont, "Number font")
	newOptions.LabelFont = r.FormValue("LabelFont")
	errors += validateFile(newOptions.LabelFont, "Label font")
	newOptions.IconFont = r.FormValue("IconFont")
	errors += validateFile(newOptions.IconFont, "Icon font")
	newOptions.Background = r.FormValue("Background")
	// Missing BG is totally OK
	newOptions.BackgroundPath = r.FormValue("BackgroundPath")
	// Missing BG path is totally OK
	newOptions.Font = r.FormValue("Font")
	errors += validateFile(newOptions.Font, "Font for round clocks")

	// Addresses
	newOptions.EngineOptions.ListenAddr = r.FormValue("ListenAddr")
	errors += validateAddr(newOptions.EngineOptions.ListenAddr, "OSC listen address")
	newOptions.EngineOptions.Connect = r.FormValue("Connect")
	errors += validateAddr(newOptions.EngineOptions.Connect, "OSC feedback address")
	newOptions.HTTPPort = r.FormValue("HTTPPort")
	errors += validateAddr(newOptions.HTTPPort, "HTTP config interface address")

	// Timezones
	newOptions.EngineOptions.Source1.TimeZone = r.FormValue("source1-timezone")
	errors += validateTZ(newOptions.EngineOptions.Source1.TimeZone, "Source 1 timezone")
	newOptions.EngineOptions.Source2.TimeZone = r.FormValue("source2-timezone")
	errors += validateTZ(newOptions.EngineOptions.Source2.TimeZone, "Source 2 timezone")
	newOptions.EngineOptions.Source3.TimeZone = r.FormValue("source3-timezone")
	errors += validateTZ(newOptions.EngineOptions.Source3.TimeZone, "Source 3 timezone")
	newOptions.EngineOptions.Source4.TimeZone = r.FormValue("source4-timezone")
	errors += validateTZ(newOptions.EngineOptions.Source4.TimeZone, "Source 4 timezone")

	// Regexp
	newOptions.EngineOptions.Ignore = r.FormValue("millumin-ignore")
	_, err = regexp.Compile("(?i)" + newOptions.EngineOptions.Ignore)
	if err != nil {
		errors += fmt.Sprintf("<li>Millumin layer ignore regexp: %v</li>", err)
	}

	// Integers
	newOptions.EngineOptions.Flash, err = strconv.Atoi(r.FormValue("Flash"))
	validateNumber(err, "Flash time")
	newOptions.EngineOptions.Timeout, err = strconv.Atoi(r.FormValue("Timeout"))
	validateNumber(err, "Tally message timeout")
	newOptions.EngineOptions.Source1.Counter, err = strconv.Atoi(r.FormValue("source1-counter"))
	validateNumber(err, "Source 1 timer")
	validateTimer(newOptions.EngineOptions.Source1.Counter, "Source 1 timer")
	newOptions.EngineOptions.Source2.Counter, err = strconv.Atoi(r.FormValue("source2-counter"))
	validateNumber(err, "Source 2 timer")
	validateTimer(newOptions.EngineOptions.Source2.Counter, "Source 2 timer")
	newOptions.EngineOptions.Source3.Counter, err = strconv.Atoi(r.FormValue("source3-counter"))
	validateNumber(err, "Source 3 timer")
	validateTimer(newOptions.EngineOptions.Source3.Counter, "Source 3 timer")
	newOptions.EngineOptions.Source4.Counter, err = strconv.Atoi(r.FormValue("source4-counter"))
	validateNumber(err, "Source 4 timer")
	validateTimer(newOptions.EngineOptions.Source4.Counter, "Source 4 timer")
	newOptions.EngineOptions.Mitti, err = strconv.Atoi(r.FormValue("mitti"))
	validateNumber(err, "Mitti destination timer")
	validateTimer(newOptions.EngineOptions.Mitti, "Mitti destination timer")
	newOptions.EngineOptions.Millumin, err = strconv.Atoi(r.FormValue("millumin"))
	validateNumber(err, "Millumin destination timer")
	validateTimer(newOptions.EngineOptions.Millumin, "Millumin destination timer")
	newOptions.EngineOptions.ShowInfo, err = strconv.Atoi(r.FormValue("ShowInfo"))
	validateNumber(err, "Time to show clock info on startup")

	newOptions.NumberFontSize, err = strconv.Atoi(r.FormValue("NumberFontSize"))
	validateNumber(err, "Number font size")

	newOptions.EngineOptions.UDPTimer1, err = strconv.Atoi(r.FormValue("udp-timer-1"))
	validateNumber(err, "UDP Timer 1")

	newOptions.EngineOptions.UDPTimer2, err = strconv.Atoi(r.FormValue("udp-timer-2"))
	validateNumber(err, "UDP Timer 2")

	alpha, err := strconv.Atoi(r.FormValue("row1-alpha"))
	validateNumber(err, "Row1 alpha")
	newOptions.Row1Alpha = uint8(alpha)

	alpha, err = strconv.Atoi(r.FormValue("row2-alpha"))
	validateNumber(err, "Row2 alpha")
	newOptions.Row2Alpha = uint8(alpha)

	alpha, err = strconv.Atoi(r.FormValue("row3-alpha"))
	validateNumber(err, "Row3 alpha")
	newOptions.Row3Alpha = uint8(alpha)

	alpha, err = strconv.Atoi(r.FormValue("label-alpha"))
	validateNumber(err, "Label alpha")
	newOptions.LabelAlpha = uint8(alpha)

	alpha, err = strconv.Atoi(r.FormValue("label-bg-alpha"))
	validateNumber(err, "Label background alpha")
	newOptions.LabelBGAlpha = uint8(alpha)

	alpha, err = strconv.Atoi(r.FormValue("timer-bg-alpha"))
	validateNumber(err, "Timer bg alpha")
	newOptions.TimerBGAlpha = uint8(alpha)

	newOptions.EngineOptions.SignalThresholdWarning, err = strconv.Atoi(r.FormValue("signal-threshold-warning"))
	validateNumber(err, "Warning signal threshold")

	newOptions.EngineOptions.SignalThresholdEnd, err = strconv.Atoi(r.FormValue("signal-threshold-end"))
	validateNumber(err, "End signal threshold")

	newOptions.EngineOptions.SignalHardware, err = strconv.Atoi(r.FormValue("signal-hw-group"))
	validateNumber(err, "Signal hardware group")

	newOptions.SignalBrightness, err = strconv.Atoi(r.FormValue("signal-hw-brightness"))
	validateNumber(err, "Signal hardware brightness")

	// Colors
	newOptions.TextColor = r.FormValue("TextColor")
	errors += validateColor(newOptions.TextColor, "Round clock text color")
	newOptions.SecondColor = r.FormValue("SecColor")
	errors += validateColor(newOptions.SecondColor, "Round clock second ring color")
	newOptions.StaticColor = r.FormValue("StaticColor")
	errors += validateColor(newOptions.StaticColor, "Round clock static ring color")
	newOptions.CountdownColor = r.FormValue("CountdownColor")
	errors += validateColor(newOptions.CountdownColor, "Round clock countdown color")

	newOptions.Row1Color = r.FormValue("Row1Color")
	errors += validateColor(newOptions.Row1Color, "Text clock row 1 color")
	newOptions.Row2Color = r.FormValue("Row2Color")
	errors += validateColor(newOptions.Row2Color, "Text clock row 2 color")
	newOptions.Row3Color = r.FormValue("Row3Color")
	errors += validateColor(newOptions.Row3Color, "Text clock row 3 color")

	newOptions.LabelColor = r.FormValue("LabelColor")
	errors += validateColor(newOptions.LabelColor, "Text clock label color")
	newOptions.LabelBG = r.FormValue("LabelBG")
	errors += validateColor(newOptions.LabelBG, "Text clock label background color")
	newOptions.TimerBG = r.FormValue("TimerBG")
	errors += validateColor(newOptions.TimerBG, "Text clock timer background color")

	newOptions.BackgroundColor = r.FormValue("BackgroundColor")
	errors += validateColor(newOptions.BackgroundColor, "Background color")

	newOptions.EngineOptions.SignalColorStart = r.FormValue("signal-color-start")
	errors += validateColor(newOptions.EngineOptions.SignalColorStart, "Signal color: start")
	newOptions.EngineOptions.SignalColorWarning = r.FormValue("signal-color-warning")
	errors += validateColor(newOptions.EngineOptions.SignalColorStart, "Signal color: warning")
	newOptions.EngineOptions.SignalColorEnd = r.FormValue("signal-color-end")
	errors += validateColor(newOptions.EngineOptions.SignalColorStart, "Signal color: end")

	newOptions.EngineOptions.Source1.OvertimeColor = r.FormValue("source1-overtime-color")
	errors += validateColor(newOptions.EngineOptions.Source1.OvertimeColor, "Source1 overtime color")
	newOptions.EngineOptions.Source2.OvertimeColor = r.FormValue("source2-overtime-color")
	errors += validateColor(newOptions.EngineOptions.Source2.OvertimeColor, "Source2 overtime color")
	newOptions.EngineOptions.Source3.OvertimeColor = r.FormValue("source3-overtime-color")
	errors += validateColor(newOptions.EngineOptions.Source3.OvertimeColor, "Source3 overtime color")
	newOptions.EngineOptions.Source4.OvertimeColor = r.FormValue("source4-overtime-color")
	errors += validateColor(newOptions.EngineOptions.Source4.OvertimeColor, "Source4 overtime color")

	if errors != "" {
		tmpl, err := htmlTemplate.New("config.html").Parse(configHTML)
		if err != nil {
			panic(err)
		}
		newOptions.Errors = htmlTemplate.HTML(fmt.Sprintf("<ul>%s</ul>", errors))
		err = tmpl.Execute(w, newOptions)
		if err != nil {
			panic(err)
		}
	} else {
		log.Printf("Writing new config ini file")
		newOptions.writeConfig(options.configFile)

		// TODO render success page

		if r.FormValue("configtxt") != "" {
			bytes, err := ioutil.ReadFile("/boot/config.txt")
			check(err)
			currentConfig := string(bytes)

			if r.FormValue("configtxt") != currentConfig {
				log.Printf("Writing /boot/config.txt")
				f, err := os.Create("/boot/config.txt")
				check(err)
				_, err = f.WriteString(r.FormValue("configtxt"))
				check(err)
				f.Close()

				// reboot the rpi
				go delayedReboot()
			}

		}
		go delayedExit()

		// Render success page

		tmpl, err := htmlTemplate.New("confirm.html").Parse(confirmHTML)
		if err != nil {
			panic(err)
		}
		err = tmpl.Execute(w, nil)
		if err != nil {
			panic(err)
		}
	}
}

// Reboot the pi after a short delay
// delay needs to be shorter than the
// delayedExit()...
func delayedReboot() {
	time.Sleep(time.Millisecond * 500)
	cmd := exec.Command("reboot")
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func delayedExit() {
	time.Sleep(time.Second)
	os.Exit(0)
}

func (options *clockOptions) writeConfig(path string) {
	tmpl, err := template.New("config.ini").Parse(configTemplate)
	if err != nil {
		panic(err)
	}
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(f, options)
}

func (options *clockOptions) createHTML() {
	tmpl, err := htmlTemplate.New("config.html").Parse(configHTML)
	if err != nil {
		panic(err)
	}

	options.Raspberry = fileExists("/boot/config.txt")

	if options.Raspberry {
		// Read out various config files for editing
		bytes, err := ioutil.ReadFile("/boot/config.txt")
		if err != nil {
			panic(err)
		}
		options.ConfigTxt = string(bytes)
	}

	err = tmpl.Execute(os.Stdout, options)
	if err != nil {
		panic(err)
	}
}

func basicAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if !ok || subtle.ConstantTimeCompare([]byte(options.HTTPUser), []byte(user)) != 1 || subtle.ConstantTimeCompare([]byte(options.HTTPPassword), []byte(pass)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Clock-8001 config"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}

		handler(w, r)
	}
}

func validateFile(filename, title string) (msg string) {
	if !fileExists(filename) {
		msg = fmt.Sprintf("<li>%s: file does not exist (%s)</li>", title, filename)
	}
	return
}

func validateNumber(err error, title string) (msg string) {
	if err != nil {
		msg = fmt.Sprintf("<li>%s: error parsing number</li>", title)
	}
	return
}

func validateTimer(timer int, title string) (msg string) {
	if timer < 0 || timer > 9 {
		msg = fmt.Sprintf("<li>%s: timer number not in range 0-9 (%d)</li>", title, timer)
	}
	return
}

func validateColor(color string, title string) (msg string) {
	match, err := regexp.MatchString(`^#([0-9a-fA-F]{3}){1,2}$`, color)

	if !match {
		msg = fmt.Sprintf("<li>%s: incorrect format for a color (%s)</li>", title, color)
	} else if err != nil {
		msg = fmt.Sprintf("<li>%s: %v</li>", title, err)
	}
	return
}

func validateAddr(addr, title string) (msg string) {
	match, err := regexp.MatchString(`^.*:\d*$`, addr)

	if !match {
		msg = fmt.Sprintf("<li>%s: address not formatted correctly (%s)</li>", title, addr)
	} else if err != nil {
		msg = fmt.Sprintf("<li>%s: %v</li>", title, err)
	}
	return
}

func validateTZ(zone, title string) (msg string) {
	_, err := time.LoadLocation(zone)
	if err != nil {
		msg = fmt.Sprintf("<li>%s: %v</li>", title, err)
	}
	return
}

// fileExists checks if a file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
