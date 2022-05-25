package airplay

import (
	"AirPlayServer/config"
	"AirPlayServer/global"
	"AirPlayServer/handlers"
	"AirPlayServer/homekit"
	"AirPlayServer/rtsp"
	"encoding/hex"
	"errors"
	"github.com/grandcat/zeroconf"
	"net"
	"strconv"
	"strings"
)

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

	var iface, err2 = net.InterfaceByName("en1")
	if err2 != nil {
		//找可用网卡，兼容windows
		allIface, err := net.Interfaces()
		if err != nil {
			return err
		}
		for _, inter := range allIface {
			if inter.Flags%2 != 0 &&
				inter.HardwareAddr.String() != "" {
				iface = &inter
				break
			}
		}
	}

	if iface.HardwareAddr.String() == "" {
		return errors.New("找不到可用的网卡")
	}
	macAddress := strings.ToUpper(iface.HardwareAddr.String())
	homekit.Device = homekit.NewAccessory(macAddress, config.Config.DeviceUUID, homekit.AirplayDevice())
	global.Debug.Printf("Starting %s for device %v", config.Config.DeviceName, homekit.Device)
	raopName := hex.EncodeToString(iface.HardwareAddr) + "@" + config.Config.DeviceName //按文档说必须是这种格式
	server, err := zeroconf.Register(raopName, "_airplay._tcp", "local.",
		port, homekit.Device.ToRecords(), nil)
	if err != nil {
		return err
	}
	defer server.Shutdown()

	global.Debug.Println("Service", raopName, "registered on address", address)
	var handler = &handlers.Rstp{}
	err = rtsp.RunRtspServer(handler)
	if err != nil {
		return err
	}
	return nil
}
