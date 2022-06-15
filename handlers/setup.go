package handlers

import (
	"AirPlayServer/config"
	"AirPlayServer/media"
	"AirPlayServer/ntp"
	"AirPlayServer/rtsp"
	"howett.net/plist"
	"strings"
)

//第2次setup请求头
type setupRequest2 struct {
	streams []setupStream `plist:"streams"` //虽然是数组但是每次只有1个
}

type setupStream struct {
	streamType         uint64 `plist:"type"`               //流媒体类型96：实时音频 103：缓冲音频 110：屏幕镜像 120：播放 130：遥控器
	streamConnectionID uint64 `plist:"streamConnectionID"` //当前连接的id,需要保存后面会用
}

const (
	voiceType         = 96  //音频
	videoType         = 110 //屏幕镜像
	bufferedAudioType = 103 //缓冲音频
	playType          = 120 //播放，应该是用在镜像播放视频？没试过
	RemoteType        = 130 //遥控
)

func (r *Rstp) OnSetupWeb(req *rtsp.Request) (*rtsp.Response, error) {
	//镜像setup,会请求两次，第一次和第二次并不顺序发生
	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "application/x-apple-binary-plist") {
		temp := map[string]interface{}{} //总是解析不成功，用个字典存一下，只取关键数据
		plist.Unmarshal(req.Body, &temp)
		if temp["eiv"] != nil { //判断是第一次
			//只取了有用的数据，详细见setup1.plist。每次不太一样，有的数据不知道干吗的
			rtsp.Session.Eiv = make([]byte, len(temp["eiv"].([]byte)))
			copy(rtsp.Session.Eiv, temp["eiv"].([]byte)) //解密视频、音频用的iv
			rtsp.Session.Ekey = make([]byte, len(temp["ekey"].([]byte)))
			copy(rtsp.Session.Ekey, temp["ekey"].([]byte)) //解密视频、音频用的key
			rtsp.Session.TimePort = temp["timingPort"].(uint64)
			//启动媒体服务
			go media.RunVideoServer()
			//TODO 启动事件服务

			return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil
		} else {
			var resutStreams []setupStream
			arr := temp["streams"].([]interface{})
			value := arr[0].(map[string]interface{})
			if value["type"].(uint64) == videoType {
				//第二次视频setup，如果仅声音的AirPlay 并不会有第二次，而是两次/command 请求，本项目没有做处理（设置不同flag、特征值、不同的功能请求都是不同的，
				//AirPlay协议是个非常负责的协议）
				//只取了有用的，详见setup2.plist
				for _, s := range arr {
					value := s.(map[string]interface{})
					var streamConnectionID64 uint64
					switch value["streamConnectionID"].(type) { //这里要注意，streamConnectionID要是uint64
					case int64:
						streamConnectionID64 = uint64(value["streamConnectionID"].(int64))
					case uint64:
						streamConnectionID64 = value["streamConnectionID"].(uint64)
					}
					stream := setupStream{
						streamType:         value["type"].(uint64),
						streamConnectionID: streamConnectionID64,
					}
					resutStreams = append(resutStreams, stream)
				}
				rtsp.Session.StreamConnectionID = resutStreams[0].streamConnectionID
				stream := map[string]uint64{
					"dataPort": config.Config.DataPort,
					"type":     videoType,
				}

				streams := [1]map[string]uint64{
					stream,
				}
				responseBody := map[string]interface{}{
					"streams":    streams[:],
					"eventPort":  config.Config.EventPort,
					"timingPort": config.Config.TimingPort,
				}

				body, err := plist.MarshalIndent(&responseBody, plist.AutomaticFormat, "\t")
				if err != nil {
					return &rtsp.Response{StatusCode: rtsp.StatusInternalServerError}, err
				}
				go ntp.RunServer()
				return &rtsp.Response{StatusCode: rtsp.StatusOK, Header: rtsp.Header{
					"Content-Type": rtsp.HeaderValue{"application/x-apple-binary-plist"},
				}, Body: body}, nil
			} else if value["type"].(uint64) == voiceType {
				//镜像时播放声音会有第三次
				//接收数据不重要，格式如下见文件setup-voice.plist,此处不需要，就不处理了
				stream := map[string]uint64{
					"dataPort":    config.Config.VoicePort,        //服务器接收音频数据的接口
					"controlPort": config.Config.VoiceControlPort, //音频重传包会通过这个接口给
					"type":        voiceType,
				}
				streams := [1]map[string]uint64{
					stream,
				}
				responseBody := map[string]interface{}{
					"streams":    streams[:],
					"timingPort": config.Config.TimingPort, //同样用于NTP配对，本项目没有做NTP对时，使用的是标准NTP协议，相见文件ntp
				}
				body, err := plist.MarshalIndent(&responseBody, plist.AutomaticFormat, "\t")
				if err != nil {
					return &rtsp.Response{StatusCode: rtsp.StatusInternalServerError}, err
				}

				return &rtsp.Response{StatusCode: rtsp.StatusOK, Header: rtsp.Header{
					"Content-Type": rtsp.HeaderValue{"application/x-apple-binary-plist"},
				}, Body: body}, nil
			}
			return &rtsp.Response{StatusCode: rtsp.StatusInternalServerError}, nil

		}
		return &rtsp.Response{StatusCode: rtsp.StatusInternalServerError}, nil
	}
	return &rtsp.Response{StatusCode: rtsp.StatusInternalServerError}, nil
}
