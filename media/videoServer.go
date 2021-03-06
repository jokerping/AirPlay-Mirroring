package media

import (
	"AirPlayServer/config"
	"AirPlayServer/global"
	"AirPlayServer/lib"
	"AirPlayServer/rtsp"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strconv"
)

var stopVideo = false
var blockMode cipher.Stream //每次只处理一个连接，放这也没什么问题

type mirroringHeader struct {
	PayloadSize   uint32
	PayloadType   uint16
	PayloadOption uint16
	PayloadNtp    uint64
	PayloadPts    uint64
	WidthSource   uint32
	HeightSource  uint32
	Width         uint32
	Height        uint32
}

type videoSession struct {
	pts          uint64
	WidthSource  uint32
	HeightSource uint32
	SpsPps       []byte
}

var session *videoSession
var nextDecryptCount = 0
var og [16]byte

func RunVideoServer() (err error) {
	session = &videoSession{}
	stopVideo = false
	port := ":" + strconv.FormatUint(config.Config.DataPort, 10)
	l, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	global.Debug.Println("启动媒体服务器")
	defer func() {
		l.Close()
		rtsp.Session.DesryAesKey = nil
		session = nil
	}()
	for stopVideo == false {
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
	//TODO 停止处理视频流，如关闭视频等操作。
	stopVideo = true
}

func handlVideoConnection(conn net.Conn) {
	//all := make([]byte, 0)
	initAes()
	defer func() {
		//FIXME 偷懒在这写，实际上可以一边接收一边写。需要改，这样占用更多内存
		conn.Close()
		//file, _ := os.Create("test.h264")
		//file.Write(all)
		//file.Sync()
		//file.Close()
	}()
	for { //承载协议是apple自定义的，承载内容为H264的裸流
		if stopVideo {
			break
		}
		header := make([]byte, 128)
		ret, err := io.ReadFull(conn, header[:4])
		if ret < 4 {
			continue
		}
		if (header[0] == 80 && header[1] == 79 && header[2] == 83 && header[3] == 84) || (header[0] == 71 && header[1] == 69 && header[2] == 84) {
			global.Debug.Println("别的请求")
			// Request is POST or GET (skip)
		} else {
			_, err = io.ReadFull(conn, header[4:])
			if err != nil {
				break
			}
			mirrorHeader, err := newMirroringHeader(header)

			if mirrorHeader.WidthSource > 0 {
				session.WidthSource = mirrorHeader.WidthSource
			}
			if mirrorHeader.HeightSource > 0 {
				session.HeightSource = mirrorHeader.HeightSource
			}
			if err == nil {
				//创建数据缓冲区
				payload := make([]byte, mirrorHeader.PayloadSize)
				_, err = io.ReadFull(conn, payload)
				if err != nil {
					break
				}
				if mirrorHeader.PayloadType == 0 {
					//TODO 处理视频
					video, _ := decryptionVideo(payload)
					h264, err := processVideo(video, session.SpsPps, mirrorHeader.PayloadPts, session.WidthSource, session.HeightSource)
					if err == nil {
						PlayVideo(h264)
						//all = append(all, h264.Data...)
					} else {
						global.Debug.Println(err)
					}
				} else if mirrorHeader.PayloadType == 1 {
					spsPps := processSpsPps(payload)
					session.SpsPps = spsPps
				}
			}
		}
	}
}

func newMirroringHeader(header []byte) (mirroringHeader, error) {
	r := bytes.NewReader(header)
	h := mirroringHeader{}
	err := binary.Read(r, binary.LittleEndian, &h.PayloadSize)
	if err != nil {
		return mirroringHeader{}, err
	}
	err = binary.Read(r, binary.LittleEndian, &h.PayloadType)
	if err != nil {
		return mirroringHeader{}, err
	}
	h.PayloadType &= 0xff
	err = binary.Read(r, binary.LittleEndian, &h.PayloadOption)
	if err != nil {
		return mirroringHeader{}, err
	}
	if h.PayloadType == 0 {
		err = binary.Read(r, binary.LittleEndian, &h.PayloadNtp)
		if err != nil {
			return mirroringHeader{}, err
		}
		ptsRemote := h.ntpToPts()
		//不设置NTP修正感觉更准一点，这里没有搞清楚
		//pts := int64(ptsRemote) - ntp.Server.SyncOffset
		//h.PayloadPts = uint64(pts)
		h.PayloadPts = ptsRemote
	}
	if h.PayloadType == 1 {
		r = bytes.NewReader(header[40:])
		err = binary.Read(r, binary.LittleEndian, &h.WidthSource)
		if err != nil {
			return mirroringHeader{}, err
		}
		err = binary.Read(r, binary.LittleEndian, &h.HeightSource)
		if err != nil {
			return mirroringHeader{}, err
		}
		r = bytes.NewReader(header[56:])
		err = binary.Read(r, binary.LittleEndian, &h.Width)
		if err != nil {
			return mirroringHeader{}, err
		}
		err = binary.Read(r, binary.LittleEndian, &h.Height)
		if err != nil {
			return mirroringHeader{}, err
		}
	}
	return h, nil
}

func (h *mirroringHeader) ntpToPts() uint64 {
	ntp := h.PayloadNtp
	sec := (ntp >> 32) & 0xffffffff
	frac := ntp & 0xffffffff
	return (sec * 100_0000) + ((frac * 100_0000) >> 32)
}

func processSpsPps(payload []byte) []byte {
	h264 := H264Codec{}
	h264.Version = payload[0]
	h264.ProfileHigh = payload[1]
	h264.Compatibility = payload[2]
	h264.Level = payload[3]
	h264.Reserved6AndNal = payload[4]
	h264.Reserved3AndSps = payload[5]
	h264.LengthOfSps = binary.BigEndian.Uint16(payload[6:8])
	sequence := make([]byte, h264.LengthOfSps)
	copy(sequence, payload[8:8+h264.LengthOfSps])
	h264.SequenceParameterSet = sequence
	h264.NumberOfPps = uint16(payload[h264.LengthOfSps+8])
	h264.LengthOfPps = binary.BigEndian.Uint16(payload[h264.LengthOfSps+9 : h264.LengthOfSps+11])
	picture := make([]byte, h264.LengthOfPps)
	copy(picture, payload[h264.LengthOfSps+11:h264.LengthOfSps+11+h264.LengthOfPps])
	h264.PictureParameterSet = picture
	if int(h264.LengthOfSps)+int(h264.LengthOfPps) < 102400 {
		spsPpsLen := int(h264.LengthOfSps + h264.LengthOfPps + 8)
		spsPps := make([]byte, spsPpsLen)
		spsPps[0] = 0
		spsPps[1] = 0
		spsPps[2] = 0
		spsPps[3] = 1
		copy(spsPps[4:], h264.SequenceParameterSet[:h264.LengthOfSps])
		spsPps[h264.LengthOfSps+4] = 0
		spsPps[h264.LengthOfSps+5] = 0
		spsPps[h264.LengthOfSps+6] = 0
		spsPps[h264.LengthOfSps+7] = 1
		copy(spsPps[h264.LengthOfSps+8:], h264.PictureParameterSet[:h264.LengthOfPps])
		return spsPps
	}
	return nil
}

func processVideo(payload []byte, spsPps []byte, pts uint64, widthSource uint32, heightSource uint32) (H264Data, error) {
	naluSize := 0
	for naluSize < len(payload) {
		ncLen := binary.BigEndian.Uint32(payload[naluSize : naluSize+4])
		if ncLen > 0 {
			payload[naluSize] = 0
			payload[naluSize+1] = 0
			payload[naluSize+2] = 0
			payload[naluSize+3] = 1
			naluSize += int(ncLen) + 4
		}
		if len(payload)-int(ncLen) > 4 {
			return H264Data{}, errors.New("获取h264错误")
		}
	}
	if len(spsPps) != 0 {
		h264Data := H264Data{}
		h264Data.FrameType = int(payload[4] & 0x1f)
		if h264Data.FrameType == 5 {
			payloadOut := make([]byte, len(payload)+len(spsPps))
			copy(payloadOut, spsPps)
			copy(payloadOut[len(spsPps):], payload)
			h264Data.Data = payloadOut
			h264Data.Length = len(payload) + len(spsPps)
		} else {
			h264Data.Data = payload
			h264Data.Length = len(payload)
		}
		h264Data.Pts = pts
		h264Data.Width = widthSource
		h264Data.Height = heightSource
		return h264Data, nil
	}
	return H264Data{}, errors.New("获取h264错误")
}

func decryptionVideo(videoData []byte) (data []byte, err error) {
	//解密视频
	if nextDecryptCount > 0 {
		for i := 0; i < nextDecryptCount; i++ {
			videoData[i] = videoData[i] ^ og[(16-nextDecryptCount)+i]
		}
	}
	encryptlen := ((len(videoData) - nextDecryptCount) / 16) * 16
	blockMode.XORKeyStream(videoData[nextDecryptCount:], videoData[nextDecryptCount:nextDecryptCount+encryptlen])
	restlen := (len(videoData) - nextDecryptCount) % 16
	reststart := len(videoData) - restlen
	nextDecryptCount = 0
	if restlen > 0 {
		og = [16]byte{}
		copy(og[:restlen], videoData[reststart:reststart+restlen])
		blockMode.XORKeyStream(og[:], og[:16])
		copy(videoData[reststart:reststart+restlen], og[:restlen])
		nextDecryptCount = 16 - restlen
	}
	output := make([]byte, len(videoData))
	copy(output, videoData)
	return output, nil
}

func initAes() {
	desryAesKey := lib.DesryAesKey()
	//2.将解密的desryAesKey和pairing阶段计算出的curve25519 共享密钥进行hash得到eaesHash,方法同pair阶段
	eaesHash := sha512.Sum512(append(desryAesKey, rtsp.Session.EcdhShared[:]...))
	//3.使用eaesHash前16个字节与（“AirPlayStreamKey”+setup阶段获得的streamConnectionId）构成的字符串hash方法同上。得到keyHash
	k1 := global.AirPlayStreamKey + strconv.FormatUint(rtsp.Session.StreamConnectionID, 10)
	keyHash := sha512.Sum512(append([]byte(k1), eaesHash[:16]...))
	//4.同样方法与“AirPlayStreamIV”+streamConnectionId hash得到ivHash
	i1 := global.AirPlayStreamIV + strconv.FormatUint(rtsp.Session.StreamConnectionID, 10)
	ivHash := sha512.Sum512(append([]byte(i1), eaesHash[:16]...))
	//5.取keyHash和IVhash的前16字节作为key和iv执行aes-ctr-128 解密视频流，视频流是avcc格式的H264裸流
	//创建解码器
	block, _ := aes.NewCipher(keyHash[:16])
	//aes ctr模式加密
	blockMode = cipher.NewCTR(block, ivHash[:16])
}
