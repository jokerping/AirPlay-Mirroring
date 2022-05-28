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
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"net"
	"strconv"
	"unsafe"
)

var stop = false

func RunServer() (err error) {
	stop = false
	port := ":" + strconv.FormatInt(int64(config.Config.DataPort), 10)
	l, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	global.Debug.Println("启动媒体服务器")
	defer l.Close()

	for {
		if stop {
			break
		}
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
	//TODO 停止处理视频流，如关闭视频等操作。
}

func handleEventConnection(conn net.Conn) {
	defer conn.Close()
	// Handle connections in a new goroutine.
	for {
		if stop {
			break
		}
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

var blockMode cipher.Stream

func decryption(buffer []byte) (data []byte, err error) {
	if rtsp.Session.VideoAESKey == nil && rtsp.Session.VideoAESIv == nil {
		desryAesKey := make([]byte, 16)
		//1.解密AES密钥，用到fpsetup（CSeq=5）接收到的keymessage
		C.playfair_decrypt(
			(*C.uchar)(unsafe.Pointer(&rtsp.Session.KeyMessage[0])),
			(*C.uchar)(unsafe.Pointer(&rtsp.Session.Ekey[0])),
			(*C.uchar)(unsafe.Pointer(&desryAesKey[0])))
		//2.讲解密的aeskey和pairing阶段计算出的curve25519 共享密钥进行hash得到eaesHash,方法同pair阶段
		eaesHash := sha512.Sum512(append([]byte(global.PairVerifyAESKey), rtsp.Session.EcdhShared[:]...))
		//3.使用eaesHash前16个字节与（“AirPlayStreamKey”+setup阶段获得的streamConnectionId）构成的字符串hash方法同上。得到keyHash
		keyHash := sha512.Sum512(append([]byte(global.AirPlayStreamKey), eaesHash[:]...))
		//4.同样方法与“AirPlayStreamIv”+streamConnectionId hash得到ivHash
		ivHash := sha512.Sum512(append([]byte(global.AirPlayStreamIv), eaesHash[:]...))
		//5.取keyHash和IVhash的前16字节作为key和iv执行aes-ctr-128 解密视频流，视频流是avcc格式的H264裸流
		rtsp.Session.VideoAESKey = make([]byte, 16)
		rtsp.Session.VideoAESIv = make([]byte, 16)
		copy(rtsp.Session.VideoAESKey, keyHash[:16])
		copy(rtsp.Session.VideoAESIv, ivHash[:16])
		//创建解码器 ctr 计数器特性，创建1次
		block, err := aes.NewCipher(rtsp.Session.VideoAESKey)
		if err != nil {
			return nil, err
		}
		//aes ctr模式加密
		blockMode = cipher.NewCTR(block, rtsp.Session.VideoAESIv)
	}
	//解密视频
	message := make([]byte, len(buffer))
	blockMode.XORKeyStream(message, buffer)
	return message, nil
}
