package media

/*
#include "playfair/playfair.c"
#include "playfair/omg_hax.c"
#include "playfair/modified_md5.c"
#include "playfair/sap_hash.c"
#include "playfair/hand_garble.c"
*/
import "C"
import (
	"AirPlayServer/config"
	"AirPlayServer/global"
	"AirPlayServer/rtsp"
	"net"
	"strconv"
	"unsafe"
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
		_, err := conn.Read(buffer[:])
		if err != nil {
			global.Debug.Printf("Event error : %v", err)
			break
		}
		decryption(buffer[:])

	}
}

var desryAesKey []byte //解密后的aes key

func decryption(buffer []byte) {
	if desryAesKey == nil {
		desryAesKey = make([]byte, 16)
		//1.解密AES密钥，用到fpsetup（CSeq=5）接收到的keymessage
		C.playfair_decrypt(
			(*C.uchar)(unsafe.Pointer(&rtsp.Session.KeyMessage[0])),
			(*C.uchar)(unsafe.Pointer(&rtsp.Session.Ekey[0])),
			(*C.uchar)(unsafe.Pointer(&desryAesKey[0])))

		global.Debug.Printf("解码前:%s,解码后%s", rtsp.Session.Ekey, desryAesKey)
	}
}
