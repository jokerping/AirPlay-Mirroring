package voice

import (
	"AirPlayServer/config"
	"AirPlayServer/global"
	"net"
	"strconv"
)

var stop = false

func RunServer() (err error) {
	stop = false
	port := ":" + strconv.FormatInt(int64(config.Config.VoicePort), 10)
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

func CloseServer() {
	stop = true
	//TODO 停止处理音频流
}

func handleEventConnection(conn net.Conn) {
	defer conn.Close()
	// Handle connections in a new goroutine.
	for {
		var buffer [4096]byte
		_, err := conn.Read(buffer[:])
		if err != nil {
			global.Debug.Printf("Event error : %v", err)
			break
		}
		//TODO 此处处理视频流 H264 avcc
		//data, err := decryption(buffer[:])
		//if err != nil {
		//	break
		//}
		//global.Debug.Printf("接收数据量%d", len(data))
	}
}

func decryption(buffer []byte) (data []byte, err error) {
	return nil, nil
}
