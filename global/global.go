package global

import (
	"io/ioutil"
	"log"
)

var Debug = log.New(ioutil.Discard, "DEBUG ", log.LstdFlags)
var PairVerifyAESKey = "Pair-Verify-AES-Key"
var PairVerifyAESIV = "Pair-Verify-AES-IV"
var AirPlayStreamKey = "AirPlayStreamKey"
var AirPlayStreamIv = "AirPlayStreamIv"
