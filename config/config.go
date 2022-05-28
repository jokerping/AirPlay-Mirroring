package config

import (
	"encoding/json"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Configuration struct {
	Volume       float64 `json:"sound-volume"`
	DeviceUUID   string  `json:"device-uuid"`
	PulseSink    string  `json:"-"`
	DeviceName   string  `json:"-"`
	Port         string  `json:"port"`
	exitsSignals chan os.Signal
}

var Config = &Configuration{
	PulseSink:  "",
	Volume:     -999,
	DeviceUUID: uuid.NewString(),
	Port:       ":7100",
}

func (c *Configuration) Load() {
	data, err := ioutil.ReadFile(c.DeviceName + "/config.json")
	if err != nil || json.Unmarshal(data, &c) != nil {
		log.Printf("%s is not valid - at new file will be created at program exit\n", c.DeviceName+"/config.json")
	}
	c.exitsSignals = make(chan os.Signal, 1)
	signal.Notify(c.exitsSignals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-c.exitsSignals
		c.Store()
		os.Exit(0)
	}()
}

func (c *Configuration) Store() {
	data, err := json.Marshal(&c)
	if err != nil {
		log.Printf("Warning: impossible to marshal configuration in json")
	}
	err = ioutil.WriteFile(c.DeviceName+"/config.json", data, 0660)
	if err != nil {
		log.Printf("Warning : impossible to store config file %s \n", c.DeviceName+"/config.json")
	}
}
