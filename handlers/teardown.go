package handlers

import (
	"AirPlayServer/media"
	"AirPlayServer/ntp"
	"AirPlayServer/rtsp"
	"howett.net/plist"
)

// OnTeardownWeb 客户端通知关闭连接
func (r *Rstp) OnTeardownWeb(req *rtsp.Request) (*rtsp.Response, error) {
	content := map[string]interface{}{}
	if _, err := plist.Unmarshal(req.Body, &content); err != nil {
		return &rtsp.Response{StatusCode: rtsp.StatusBadRequest}, nil
	}
	if len(content) > 0 {
		stream := content["streams"].([]interface{})[0].(map[string]interface{})
		//关闭音频时还有别的字段，不做处理
		switch stream["type"].(uint64) {
		case videoType:
			media.CloseVideoServer()
			ntp.CloseVideoServer()
		case voiceType:
			media.CloseVoiceServer()
		}
	}
	return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil
}
