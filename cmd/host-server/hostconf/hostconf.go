package hostconf

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type HostConfiguration struct {
	GameserverBinaryPath string
}

var config HostConfiguration
var lastConfigMod time.Time

func init() {

	file, err := os.Open("host.config")
	if err != nil {
		log.Fatal(err)
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

	info, err := os.Stat("host.config")
	if err != nil {
		log.Fatal(err)
	}

	lastConfigMod = info.ModTime()
}

func GameserverBinaryPath() string {

	checkConfigFile()
	return config.GameserverBinaryPath
}

func checkConfigFile() {

	info, err := os.Stat("host.config")
	if err != nil {

		log.Fatal(err)
	}

	modTime := info.ModTime()

	if modTime.After(lastConfigMod) {

		file, err := os.Open("host.config")
		if err != nil {

			log.Fatal(err)
		}

		decoder := json.NewDecoder(file)
		err = decoder.Decode(&config)
		if err != nil {

			log.Fatal(err)
		}

		lastConfigMod = modTime

		log.Print("config reloaded")
	}
}
