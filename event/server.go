package event

import (
	"AirPlayServer/config"
	"AirPlayServer/global"
	"fmt"
	"net"
)

type event struct {
}

var e = event{}

func RunServer() (err error) {

	l, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: int(config.Config.EventPort),
	})
	if err == nil {
		global.Debug.Println("启动事件服务器")
		defer l.Close()
		for {
			e.handleConnection(l)
		}
	}
	return err
}

func (e *event) handleConnection(conn *net.UDPConn) {

	buf := make([]byte, 1024)
	n, raddr, err := conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Println("conn.ReadFromUDP err:", err)
		return
	}
	global.Debug.Printf("接收到客户端[%s]：%s", raddr, string(buf[:n]))

	//conn.WriteToUDP([]byte("I-AM-SERVER"), raddr) // 简单回写数据给客户端
}
