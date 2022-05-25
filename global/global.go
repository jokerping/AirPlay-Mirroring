package global

import (
	"io/ioutil"
	"log"
)

var Debug = log.New(ioutil.Discard, "DEBUG ", log.LstdFlags)
