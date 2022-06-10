package media

import (
	"AirPlayServer/config"
	"AirPlayServer/global"
	"AirPlayServer/rtsp"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net"
	"os"
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

func RunVideoServer() (err error) {
	session = &videoSession{}
	stopVideo = false
	port := ":" + strconv.FormatInt(int64(config.Config.DataPort), 10)
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
	//TODO 停止处理视频流，如关闭视频等操作。
	stopVideo = true
}

func handlVideoConnection(conn net.Conn) {
	all := make([]byte, 0)
	defer func() {
		conn.Close()
		file, _ := os.Create("xx.h264")
		file.Write(all)
		file.Sync()
		file.Close()
	}()
	for { //承载协议是apple自定义的，承载内容为H264的裸流
		if stopVideo {
			break
		}
		header := make([]byte, 128)
		_, err := io.ReadFull(conn, header)
		if err != nil {
			break
		}
		mirrorHeader, err := newMirroringHeader(header)
		if session.pts == 0 {
			session.pts = mirrorHeader.PayloadPts
		}
		if session.WidthSource == 0 {
			session.WidthSource = mirrorHeader.WidthSource
		}
		if session.HeightSource == 0 {
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
				h264, err := processVideo(video, session.SpsPps, session.pts, session.WidthSource, session.HeightSource)
				if err == nil {
					all = append(all, h264.Data...)
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
		h.PayloadPts = h.ntpToPts()
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
	return (((ntp >> 32) & 0xffffffff) * 1000000) + ((ntp & 0xffffffff) * 1000 * 1000 / math.MaxInt32)
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
	if int(h264.LengthOfSps+h264.LengthOfPps) < 102400 {
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

func byteTofloat32(b []byte) float32 {
	t1 := binary.LittleEndian.Uint32(b)
	return math.Float32frombits(t1)
}

func decryptionVideo(buffer []byte) (data []byte, err error) {
	if rtsp.Session.DesryAesKey == nil {
		desryAesKey := desryAesKey()
		//2.将解密的desryAesKey和pairing阶段计算出的curve25519 共享密钥进行hash得到eaesHash,方法同pair阶段
		eaesHash := sha512.Sum512(append(desryAesKey, rtsp.Session.EcdhShared[:]...))
		//3.使用eaesHash前16个字节与（“AirPlayStreamKey”+setup阶段获得的streamConnectionId）构成的字符串hash方法同上。得到keyHash
		sID := make([]byte, 8)
		binary.LittleEndian.PutUint64(sID, uint64(rtsp.Session.StreamConnectionID))
		k1 := append([]byte(global.AirPlayStreamKey), sID...)
		keyHash := sha512.Sum512(append(k1, eaesHash[:16]...))
		//4.同样方法与“AirPlayStreamIV”+streamConnectionId hash得到ivHash
		i1 := append([]byte(global.AirPlayStreamIV), sID...)
		ivHash := sha512.Sum512(append(i1, eaesHash[:16]...))
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
