package handlers

import (
	"AirPlayServer/global"
	"AirPlayServer/rtsp"
)

type Rstp struct {
}

func (r *Rstp) OnConnOpen(conn *rtsp.Conn) {
}

func (r *Rstp) OnRequest(conn *rtsp.Conn, request *rtsp.Request) {
	global.Debug.Printf("request received : %s %s body %d", request.Method, request.URL, len(request.Body))
}

func (r *Rstp) Handle(conn *rtsp.Conn, req *rtsp.Request) (*rtsp.Response, error) {
	switch req.Method {
	case rtsp.Options:
		return r.onOptions(req)
	case rtsp.Get:
		return r.OnGetWeb(req)
	case rtsp.Post:
		return r.OnPostWeb(conn, req)
	case rtsp.Setup:
		return r.OnSetupWeb(req)
	case rtsp.GetParameter:
		return r.OnGetParameterWeb(req)
	case rtsp.Record:
		return r.OnRecordWeb(req)
	case rtsp.SetParameter:
		return r.OnSetParameterWeb(req)
	case rtsp.Teardown:
		return r.OnTeardownWeb(req)

	}
	return &rtsp.Response{StatusCode: rtsp.StatusNotImplemented}, nil
}

func (r *Rstp) OnResponse(conn *rtsp.Conn, resp *rtsp.Response) {
	global.Debug.Printf("response sent :head %d body %d", len(resp.Header), len(resp.Body))
}
