package main

import (
	// "fmt"
	"crypto/subtle"
	"gitlab.com/Depili/clock-8001/v4/clock"
	htmlTemplate "html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"text/template"
)

func runHTTP() {
	if options.configFile == "" {
		// No config file specified, can't save the config
		log.Printf("No config specified, http config interface disabled")
		return
	}

	http.HandleFunc("/save", basicAuth(saveHandler))
	http.HandleFunc("/", basicAuth(indexHandler))
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

	err = tmpl.Execute(w, options)
	if err != nil {
		panic(err)
	}
}

// TODO: validation
func saveHandler(w http.ResponseWriter, r *http.Request) {
	var newOptions clockOptions
	newOptions.EngineOptions = &clock.EngineOptions{}
	newOptions.EngineOptions.Source1 = &clock.SourceOptions{}
	newOptions.EngineOptions.Source2 = &clock.SourceOptions{}
	newOptions.EngineOptions.Source3 = &clock.SourceOptions{}
	newOptions.EngineOptions.Source4 = &clock.SourceOptions{}

	// Booleans
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
	newOptions.EngineOptions.Source1.UDP = r.FormValue("source1-udp") != ""
	newOptions.EngineOptions.Source1.Timer = r.FormValue("source1-timer") != ""
	newOptions.EngineOptions.Source1.Tod = r.FormValue("source1-tod") != ""

	newOptions.EngineOptions.Source2.LTC = r.FormValue("source2-ltc") != ""
	newOptions.EngineOptions.Source2.UDP = r.FormValue("source2-udp") != ""
	newOptions.EngineOptions.Source2.Timer = r.FormValue("source2-timer") != ""
	newOptions.EngineOptions.Source2.Tod = r.FormValue("source2-tod") != ""

	newOptions.EngineOptions.Source3.LTC = r.FormValue("source3-ltc") != ""
	newOptions.EngineOptions.Source3.UDP = r.FormValue("source3-udp") != ""
	newOptions.EngineOptions.Source3.Timer = r.FormValue("source3-timer") != ""
	newOptions.EngineOptions.Source3.Tod = r.FormValue("source3-tod") != ""

	newOptions.EngineOptions.Source4.LTC = r.FormValue("source4-ltc") != ""
	newOptions.EngineOptions.Source4.UDP = r.FormValue("source4-udp") != ""
	newOptions.EngineOptions.Source4.Timer = r.FormValue("source4-timer") != ""
	newOptions.EngineOptions.Source4.Tod = r.FormValue("source4-tod") != ""

	newOptions.DrawBoxes = r.FormValue("DrawBoxes") != ""

	// Strings
	newOptions.Face = r.FormValue("Face")
	newOptions.NumberFont = r.FormValue("NumberFont")
	newOptions.LabelFont = r.FormValue("LabelFont")
	newOptions.IconFont = r.FormValue("IconFont")

	newOptions.Font = r.FormValue("Font")
	newOptions.EngineOptions.ListenAddr = r.FormValue("ListenAddr")
	newOptions.EngineOptions.Connect = r.FormValue("Connect")
	newOptions.HTTPPort = r.FormValue("HTTPPort")
	newOptions.HTTPUser = r.FormValue("HTTPUser")
	newOptions.HTTPPassword = r.FormValue("HTTPPassword")
	newOptions.Background = r.FormValue("Background")

	newOptions.EngineOptions.Source1.Text = r.FormValue("source1-text")
	newOptions.EngineOptions.Source1.TimeZone = r.FormValue("source1-timezone")

	newOptions.EngineOptions.Source2.Text = r.FormValue("source2-text")
	newOptions.EngineOptions.Source2.TimeZone = r.FormValue("source2-timezone")

	newOptions.EngineOptions.Source3.Text = r.FormValue("source3-text")
	newOptions.EngineOptions.Source3.TimeZone = r.FormValue("source3-timezone")

	newOptions.EngineOptions.Source4.Text = r.FormValue("source4-text")
	newOptions.EngineOptions.Source4.TimeZone = r.FormValue("source4-timezone")

	newOptions.EngineOptions.Ignore = r.FormValue("millumin-ignore")

	// Integers
	newOptions.EngineOptions.Flash, _ = strconv.Atoi(r.FormValue("Flash"))
	newOptions.EngineOptions.Timeout, _ = strconv.Atoi(r.FormValue("Timeout"))

	newOptions.EngineOptions.Source1.Counter, _ = strconv.Atoi(r.FormValue("source1-counter"))
	newOptions.EngineOptions.Source2.Counter, _ = strconv.Atoi(r.FormValue("source2-counter"))
	newOptions.EngineOptions.Source3.Counter, _ = strconv.Atoi(r.FormValue("source3-counter"))
	newOptions.EngineOptions.Source4.Counter, _ = strconv.Atoi(r.FormValue("source4-counter"))

	newOptions.EngineOptions.Mitti, _ = strconv.Atoi(r.FormValue("mitti"))
	newOptions.EngineOptions.Millumin, _ = strconv.Atoi(r.FormValue("millumin"))

	// Colors
	newOptions.TextColor = r.FormValue("TextColor")
	newOptions.SecondColor = r.FormValue("SecColor")
	newOptions.StaticColor = r.FormValue("StaticColor")
	newOptions.CountdownColor = r.FormValue("CountdownColor")

	newOptions.Row1Color = r.FormValue("Row1Color")
	newOptions.Row2Color = r.FormValue("Row2Color")
	newOptions.Row3Color = r.FormValue("Row3Color")

	newOptions.LabelColor = r.FormValue("LabelColor")
	newOptions.LabelBG = r.FormValue("LabelBG")
	newOptions.TimerBG = r.FormValue("TimerBG")

	log.Printf("Writing new config ini file")
	newOptions.writeConfig(options.configFile)

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
			cmd := exec.Command("reboot")
			cmd.Env = os.Environ()
			if err := cmd.Run(); err != nil {
				panic(err)
			}
		}
	}

	os.Exit(0)
}

/*
func parseColor(str string) (uint8, uint8, uint8) {
	bytes, _ := hex.DecodeString(str[1:])
	return uint8(bytes[0]), uint8(bytes[1]), uint8(bytes[2])
}
*/

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

// fileExists checks if a file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
