package media

import (
	"AirPlayServer/config"
	"AirPlayServer/global"
	"AirPlayServer/rtsp"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"fmt"
	"net"
)

var stopVoice = false

// RunVoiceServer 音频使用udp传输
func RunVoiceServer() (err error) {
	stopVoice = false
	l, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: int(config.Config.VoicePort),
	})
	if err != nil {
		return err
	}
	global.Debug.Println("启动事件服务器")
	defer l.Close()
	for {
		if stopVoice {
			break
		}
		handlVoiceConnection(l)
	}
	return err
}

func CloseVoiceServer() {
	stopVoice = true
	//TODO 停止处理音频流
}

func handlVoiceConnection(conn *net.UDPConn) {

	buf := make([]byte, 4096)
	_, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Println("conn.ReadFromUDP err:", err)
		return
	}
	decryption(buf)
	//TODO 处理音频流
	//这里可能存在的问题，每个序号的音频包可能会发送3次，需要做过滤处理
}

func decryption(buffer []byte) (data []byte, err error) {
	// 资料显示{0x0, 0x68, 0x34, 0x0}为没有音频数据
	if rtsp.Session.DesryAesKey == nil {
		desryAesKey()
	}
	//创建解码器
	keyHash := sha512.Sum512(append(rtsp.Session.DesryAesKey, rtsp.Session.EcdhShared...))
	block, err := aes.NewCipher(keyHash[:16])
	if err != nil {
		return nil, err
	}
	//aes cbc模式加密,此处IV直接取第1次连接传送过来的IV
	blockMode := cipher.NewCBCDecrypter(block, rtsp.Session.Eiv)
	message := make([]byte, len(buffer))
	blockMode.CryptBlocks(message, buffer)
	return message, nil
}
