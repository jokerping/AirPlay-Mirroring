package handlers

import (
	"AirPlayServer/rtsp"
)

func (r *Rstp) OnRecordWeb(req *rtsp.Request) (*rtsp.Response, error) {
	audio := req.Header["2205"]
	if audio == nil {
		audio = rtsp.HeaderValue{"0"}
	}
	return &rtsp.Response{StatusCode: rtsp.StatusOK, Header: rtsp.Header{
		"Audio-Latency": audio,
	}}, nil
}
