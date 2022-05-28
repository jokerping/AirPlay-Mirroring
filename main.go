package main

import (
	"AirPlayServer/airplay"
	"AirPlayServer/config"
	"AirPlayServer/global"
	"flag"
	"log"
	"os"
)

//本例只用于展示基本流程，实际工程中可使用成熟库代替自己实现的
//比如认证部分使用https://github.com/brutella/hap
func main() {
	debug := flag.Bool("debug", true, "Show debug output.")
	flag.StringVar(&config.Config.DeviceName, "n", "fuckAndroid", "Specify device name")
	flag.Parse()

	if *debug {
		global.Debug = log.New(os.Stderr, "DEBUG ", log.LstdFlags)
	}

	err := airplay.RunAirPlayServer()
	if err != nil {
		log.Fatal(err)
	}

}
