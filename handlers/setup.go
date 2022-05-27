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

//第一次setup返回body
type setupResponse struct {
	eventPort  uint64 `plist:"eventPort"`  //客户端用来向服务器发送事件的端口
	timingPort uint64 `plist:"timingPort"` //客户端用于向服务器发送心跳的端口（可以直接用客户端给的，也可以修改成服务指定的）
}

//第2次setup请求头
type setupRequest2 struct {
	streams []setupStream `plist:"streams"`
}

type setupStream struct {
	streamType         uint64 `plist:"type"`               //流媒体类型96：实时音频 103：缓冲音频 110：屏幕镜像 120：播放 130：遥控器
	streamConnectionID int64  `plist:"streamConnectionID"` //当前连接的id,需要保存后面会用
}

type timestampInfo struct {
	name string `plist:"name"`
}

//第二次setup返回body
type setupResponse2 struct {
	streams []responseStream `plist:"streams"`
}

type responseStream struct {
	dataPort   uint64 `plist:"dataPort"` //客户端用于向接收器发送视频流数据的端口，此处必须初始化镜像服务 以处理 H264 数据。
	streamType int    `plist:"type"`     //流媒体类型，同上。此处返回110，屏幕镜像
}

func (r *Rstp) OnSetupWeb(req *rtsp.Request) (*rtsp.Response, error) {
	//镜像setup,会请求两次，第一次和第二次并不顺序发生
	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "application/x-apple-binary-plist") {
		temp := map[string]interface{}{} //总是解析不成功，用个字典存一下，只取关键数据
		plist.Unmarshal(req.Body, &temp)
		if temp["eiv"] != nil { //判断是第一次
			content := setupRequest{
				eiv:                      temp["eiv"].([]byte),
				eKey:                     temp["ekey"].([]byte),
				timingProtocol:           temp["timingProtocol"].(string),
				timingPort:               temp["timingPort"].(uint64),
				isScreenMirroringSession: temp["isScreenMirroringSession"].(bool),
			}
			//构造返回数据
			responseBody := &setupResponse{
				eventPort:  config.Config.EventPort,
				timingPort: content.timingPort,
			}
			body, err := plist.Marshal(*responseBody, plist.AutomaticFormat)
			if err != nil {
				return &rtsp.Response{StatusCode: rtsp.StatusInternalServerError}, err
			}

			return &rtsp.Response{StatusCode: rtsp.StatusOK, Header: rtsp.Header{
				"Content-Type": rtsp.HeaderValue{"application/x-apple-binary-plist"},
			}, Body: body}, nil
		} else {
			var resutStreams []setupStream
			arr := temp["streams"].([]interface{})
			for _, s := range arr {
				value := s.(map[string]interface{})
				stream := setupStream{
					streamType:         value["type"].(uint64),
					streamConnectionID: value["streamConnectionID"].(int64),
				}
				resutStreams = append(resutStreams, stream)
			}
			//setupRequest := setupRequest2{streams: resutStreams}
			stream := responseStream{
				dataPort:   config.Config.DataPort,
				streamType: 110,
			}
			streams := []responseStream{
				stream,
			}
			responseBody := &setupResponse2{streams: streams}
			body, err := plist.Marshal(*responseBody, plist.AutomaticFormat)
			if err != nil {
				return &rtsp.Response{StatusCode: rtsp.StatusInternalServerError}, err
			}
			//TODO 此处要启动监听流媒体服务
			return &rtsp.Response{StatusCode: rtsp.StatusOK, Header: rtsp.Header{
				"Content-Type": rtsp.HeaderValue{"application/x-apple-binary-plist"},
			}, Body: body}, nil
		}
	}
	return &rtsp.Response{StatusCode: rtsp.StatusInternalServerError}, nil
}
