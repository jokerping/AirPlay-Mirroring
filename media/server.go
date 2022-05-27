package media

import (
	"AirPlayServer/config"
	"AirPlayServer/global"
	"net"
	"strconv"
)

type media struct {
}

var e = &media{}

func RunServer() (err error) {
	port := ":" + strconv.FormatInt(int64(config.Config.DataPort), 10)
	l, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	global.Debug.Println("启动媒体服务器")
	defer l.Close()

	for {
		// listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			global.Debug.Println("Error accepting: ", err.Error())
			return err
		}

		go handleEventConnection(conn)
	}
	return nil
}

func handleEventConnection(conn net.Conn) {
	defer conn.Close()
	// Handle connections in a new goroutine.
	for {
		var buffer [4096]byte
		count, err := conn.Read(buffer[:])
		if err != nil {
			global.Debug.Printf("Event error : %v", err)
			break
		}
		global.Debug.Printf("接收到数据:%d", count)
	}
}
