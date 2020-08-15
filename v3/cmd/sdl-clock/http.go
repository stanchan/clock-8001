package main

import (
	// "fmt"
	"crypto/subtle"
	"encoding/hex"
	"gitlab.com/Depili/clock-8001/v3/clock"
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

	// Booleans
	newOptions.Small = r.FormValue("Small") != ""
	newOptions.DualClock = r.FormValue("DualClock") != ""
	newOptions.Debug = r.FormValue("Debug") != ""
	newOptions.DisableHTTP = r.FormValue("DisableHTTP") != ""
	newOptions.NoARCorrection = r.FormValue("NoARCorrection") != ""
	newOptions.EngineOptions.DisableOSC = r.FormValue("DisableOSC") != ""
	newOptions.EngineOptions.DisableFeedback = r.FormValue("DisableFeedback") != ""
	newOptions.EngineOptions.DisableLTC = r.FormValue("DisableLTC") != ""
	newOptions.EngineOptions.LTCSeconds = r.FormValue("LTCSeconds") != ""
	newOptions.EngineOptions.LTCFollow = r.FormValue("LTCFollow") != ""
	newOptions.EngineOptions.Format12h = r.FormValue("Format12h") != ""

	// Strings
	newOptions.Font = r.FormValue("Font")
	newOptions.EngineOptions.Timezone = r.FormValue("Timezone")
	newOptions.EngineOptions.ListenAddr = r.FormValue("ListenAddr")
	newOptions.EngineOptions.Connect = r.FormValue("Connect")
	newOptions.HTTPPort = r.FormValue("HTTPPort")
	newOptions.HTTPUser = r.FormValue("HTTPUser")
	newOptions.HTTPPassword = r.FormValue("HTTPPassword")
	newOptions.Background = r.FormValue("Background")

	// Integers
	newOptions.EngineOptions.Flash, _ = strconv.Atoi(r.FormValue("Flash"))
	newOptions.EngineOptions.Timeout, _ = strconv.Atoi(r.FormValue("Timeout"))

	// Colors
	newOptions.TextRed, newOptions.TextGreen, newOptions.TextBlue =
		parseColor(r.FormValue("TextColor"))
	newOptions.SecRed, newOptions.SecGreen, newOptions.SecBlue =
		parseColor(r.FormValue("SecColor"))
	newOptions.StaticRed, newOptions.StaticGreen, newOptions.StaticBlue =
		parseColor(r.FormValue("StaticColor"))
	newOptions.EngineOptions.CountdownRed, newOptions.EngineOptions.CountdownGreen, newOptions.EngineOptions.CountdownBlue =
		parseColor(r.FormValue("CountdownColor"))

	log.Printf("Writing new config ini file")
	newOptions.writeConfig(options.configFile)

	if r.FormValue("configtxt") != "" {
		bytes, err := ioutil.ReadFile("/boot/config.txt")
		check(err)
		current_config := string(bytes)

		if r.FormValue("configtxt") != current_config {
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

func parseColor(str string) (uint8, uint8, uint8) {
	bytes, _ := hex.DecodeString(str[1:])
	return uint8(bytes[0]), uint8(bytes[1]), uint8(bytes[2])
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

// fileExists checks if a file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
