package handlers

import (
	"AirPlayServer/rtsp"
)

func (r Rstp) onOptions(req *rtsp.Request) (*rtsp.Response, error) {
	return &rtsp.Response{StatusCode: rtsp.StatusOK,
		Header: rtsp.Header{
			"Public": rtsp.HeaderValue{"ANNOUNCE, SETUP, RECORD, PAUSE, FLUSH, TEARDOWN, OPTIONS, GET_PARAMETER, SET_PARAMETER, POST, GET"},
		}, Body: nil}, nil
}
