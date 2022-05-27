package airplay

import (
	"AirPlayServer/config"
	"AirPlayServer/global"
	"AirPlayServer/handlers"
	"AirPlayServer/homekit"
	"AirPlayServer/rtsp"
	"errors"
	"github.com/grandcat/zeroconf"
	"net"
	"strconv"
	"strings"
)

var handler = &handlers.Rstp{}

func RunAirPlayServer() error {
	address := config.Config.Port //服务运行的端口，随便
	_, portstr, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	port, err := strconv.Atoi(portstr)
	if err != nil {
		return err
	}

	iFace, err := net.InterfaceByName("en1")
	if err != nil {
		all, err := net.Interfaces()
		if err != nil {
			return err
		}
		for _, face := range all {
			if iFace == nil {
				addrs, _ := face.Addrs()
				for _, addr := range addrs {
					ipNet, isIpNet := addr.(*net.IPNet)
					if isIpNet && !ipNet.IP.IsLoopback() {
						if ipNet.IP.To4() != nil && face.HardwareAddr != nil {
							iFace = &face
							break
						}
					}
				}
			}
		}
	}

	if iFace.HardwareAddr.String() == "" {
		return errors.New("找不到可用的网卡")
	}
	macAddress := strings.ToUpper(iFace.HardwareAddr.String())
	homekit.Device = homekit.NewAccessory(macAddress, config.Config.DeviceUUID, homekit.AirplayDevice())
	global.Debug.Printf("Starting %s for device %v", config.Config.DeviceName, homekit.Device)
	//raopName := hex.EncodeToString(iFace.HardwareAddr) + "@" + config.Config.DeviceName //按文档说必须是这种格式
	server, err := zeroconf.Register(config.Config.DeviceName, "_airplay._tcp", "local.",
		port, homekit.Device.ToRecords(), nil)
	if err != nil {
		return err
	}
	defer server.Shutdown()

	global.Debug.Println("Service", config.Config.DeviceName, "registered on address", address)

	err = rtsp.RunRtspServer(handler)
	if err != nil {
		return err
	}
	return nil
}
