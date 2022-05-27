package handlers

import (
	"AirPlayServer/config"
	"AirPlayServer/rtsp"
	"howett.net/plist"
	"strings"
)

//第一次setup请求头
type setupRequest struct {
	eiv                      []byte `plist:"eiv"`                      //AES iv  需要保存会在后面使用
	eKey                     []byte `plist:"ekey"`                     //AES 密钥 需要保存会在后面使用
	timingProtocol           string `plist:"timingProtocol"`           //用于发送计时数据的协议
	timingPort               uint64 `plist:"timingPort"`               //心跳的端口,可以在第一次setup返回中更改
	isScreenMirroringSession bool   `plist:"isScreenMirroringSession"` //用于指示流的类型（仅限视频或音频）
}

//第2次setup请求头
type setupRequest2 struct {
	streams []setupStream `plist:"streams"`
}

type setupStream struct {
	streamType         uint64 `plist:"type"`               //流媒体类型96：实时音频 103：缓冲音频 110：屏幕镜像 120：播放 130：遥控器
	streamConnectionID int64  `plist:"streamConnectionID"` //当前连接的id,需要保存后面会用
}

var handler = &Rstp{}

func (r *Rstp) OnSetupWeb(req *rtsp.Request) (*rtsp.Response, error) {
	//镜像setup,会请求两次，第一次和第二次并不顺序发生
	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "application/x-apple-binary-plist") {
		temp := map[string]interface{}{} //总是解析不成功，用个字典存一下，只取关键数据
		plist.Unmarshal(req.Body, &temp)
		if temp["eiv"] != nil { //判断是第一次
			//content := setupRequest{
			//	eiv:                      temp["eiv"].([]byte),
			//	eKey:                     temp["ekey"].([]byte),
			//	timingProtocol:           temp["timingProtocol"].(string),
			//	timingPort:               temp["timingPort"].(uint64),
			//	isScreenMirroringSession: temp["isScreenMirroringSession"].(bool),
			//}
			return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil
		} else {
			var resutStreams []setupStream
			arr := temp["streams"].([]interface{})
			for _, s := range arr {
				value := s.(map[string]interface{})
				var streamConnectionID64 int64
				switch value["streamConnectionID"].(type) {
				case int64:
					streamConnectionID64 = value["streamConnectionID"].(int64)
				case uint64:
					streamConnectionID64 = int64(value["streamConnectionID"].(uint64))
				}
				stream := setupStream{
					streamType:         value["type"].(uint64),
					streamConnectionID: streamConnectionID64,
				}
				resutStreams = append(resutStreams, stream)
			}
			//setupRequest := setupRequest2{streams: resutStreams}
			//stream := responseStream{
			//	dataPort:   7120,
			//	streamType: 110,
			//}
			//streams := []responseStream{
			//	stream,
			//}
			//responseBody := &setupResponse2{
			//	streams:    streams,
			//	timingPort: 7320,
			//	eventPort:  7220,
			//}

			stream := map[string]uint64{
				"dataPort": config.Config.DataPort,
				"type":     110,
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

			return &rtsp.Response{StatusCode: rtsp.StatusOK, Header: rtsp.Header{
				"Content-Type": rtsp.HeaderValue{"application/x-apple-binary-plist"},
			}, Body: []byte(body)}, nil
		}
	}
	return &rtsp.Response{StatusCode: rtsp.StatusInternalServerError}, nil
}
