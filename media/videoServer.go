package media

import (
	"AirPlayServer/config"
	"AirPlayServer/global"
	"AirPlayServer/rtsp"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"encoding/binary"
	"net"
	"strconv"
)

var stopVideo = false

func RunVideoServer() (err error) {
	stopVideo = false
	port := ":" + strconv.FormatInt(int64(config.Config.DataPort), 10)
	l, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	global.Debug.Println("启动媒体服务器")
	defer l.Close()

	for {
		if stopVideo {
			break
		}
		// listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			global.Debug.Println("Error accepting: ", err.Error())
			return err
		}
		go handlVideoConnection(conn)
	}
	return nil
}

func CloseVideoServer() {
	stopVideo = true
	//TODO 停止处理视频流，如关闭视频等操作。
}

func handlVideoConnection(conn net.Conn) {
	defer conn.Close()
	// Handle connections in a new goroutine.
	for {
		if stopVideo {
			break
		}
		buf := make([]byte, 4096)
		_, err := conn.Read(buf)
		if err != nil {
			global.Debug.Printf("Event error : %v", err)
			break
		}
		//TODO 此处处理视频流 H264 avcc
		decryptionVideo(buf)
	}
}

var blockMode cipher.Stream

func decryptionVideo(buffer []byte) (data []byte, err error) {
	if rtsp.Session.DesryAesKey == nil {
		desryAesKey := desryAesKey()
		//2.将解密的desryAesKey和pairing阶段计算出的curve25519 共享密钥进行hash得到eaesHash,方法同pair阶段
		eaesHash := sha512.Sum512(append(desryAesKey, rtsp.Session.EcdhShared[:]...))
		//3.使用eaesHash前16个字节与（“AirPlayStreamKey”+setup阶段获得的streamConnectionId）构成的字符串hash方法同上。得到keyHash
		sID := make([]byte, 8)
		binary.LittleEndian.PutUint64(sID, uint64(rtsp.Session.StreamConnectionID))
		k1 := append([]byte(global.AirPlayStreamKey), sID...)
		keyHash := sha512.Sum512(append(k1, eaesHash[:]...))
		//4.同样方法与“AirPlayStreamIv”+streamConnectionId hash得到ivHash
		i1 := append([]byte(global.AirPlayStreamIv), sID...)
		ivHash := sha512.Sum512(append(i1, eaesHash[:]...))
		//5.取keyHash和IVhash的前16字节作为key和iv执行aes-ctr-128 解密视频流，视频流是avcc格式的H264裸流
		//创建解码器
		block, err := aes.NewCipher(keyHash[:16])
		if err != nil {
			return nil, err
		}
		//aes ctr模式加密
		blockMode = cipher.NewCTR(block, ivHash[:16])
	}
	//解密视频
	message := make([]byte, len(buffer))
	blockMode.XORKeyStream(message, buffer)
	return message, nil
}
