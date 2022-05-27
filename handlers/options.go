package handlers

import (
	"AirPlayServer/rtsp"
)

func (r Rstp) onOptions(req *rtsp.Request) (*rtsp.Response, error) {
	//非appletv mode 可能请求，但是总是处理不对，直接改Apple TV 不走这个请求
	return &rtsp.Response{StatusCode: rtsp.StatusOK,
		Header: rtsp.Header{
			"Public": rtsp.HeaderValue{"ANNOUNCE, SETUP, RECORD, PAUSE, FLUSH, TEARDOWN, OPTIONS, GET_PARAMETER, SET_PARAMETER, POST, GET"},
		}, Body: nil}, nil
}
