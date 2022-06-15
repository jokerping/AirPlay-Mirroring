package ntp

import (
	"AirPlayServer/config"
	ntp "AirPlayServer/lib"
	"AirPlayServer/rtsp"
	"sort"
	"time"
)

type NtServer struct {
	SyncOffset int64
	data       [ntpDataCount]ntData
	dataIdx    int
}

type ntData struct {
	offset int64
	rtt    int64
}

var Server *NtServer
var timeTicker *time.Ticker
var stopNtp = false

const ntpDataCount = 8

func RunServer() {
	timeTicker = time.NewTicker(time.Second * 3)
	var dataSorted [ntpDataCount]ntData
	Server = &NtServer{}
	for !stopNtp {
		options := ntp.QueryOptions{
			LocalAddress: rtsp.Session.LocalIp,
			Port:         int(rtsp.Session.TimePort),
			LocalPort:    int(config.Config.TimingPort)}
		response, err := ntp.QueryWithOptions(rtsp.Session.RomteIP, options)
		if err == nil {
			Server.dataIdx = (Server.dataIdx + 1) % ntpDataCount
			Server.data[Server.dataIdx].offset = int64(response.ClockOffset)
			Server.data[Server.dataIdx].rtt = int64(response.RTT)
			copy(dataSorted[:], Server.data[:])
			sort.Slice(dataSorted[:], func(i, j int) bool {
				return dataSorted[i].rtt < dataSorted[j].rtt
			})
			Server.SyncOffset = dataSorted[0].offset
			<-timeTicker.C
		}
	}
}

func CloseNTPServer() {
	stopNtp = true
	timeTicker.Stop()
}
