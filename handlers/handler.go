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
	}
	//switch req.Method {
	//case "GET":
	//	return r.OnGetWeb(req)
	//case "POST":
	//	return r.OnPostWeb(conn, req)
	//case "SETUP":
	//	return r.OnSetupWeb(req)
	//case "GET_PARAMETER":
	//	return r.OnGetParameterWeb(req)
	//case "SET_PARAMETER":
	//	return r.OnSetParameterWeb(req)
	//case "RECORD":
	//	return r.OnRecordWeb(req)
	//case "SETPEERS":
	//	return r.OnSetPeerWeb(req)
	//case "SETRATEANCHORTIME":
	//	return r.OnSetRateAnchorTime(req)
	//case "FLUSHBUFFERED":
	//	return r.OnFlushBuffered(req)
	//case "TEARDOWN":
	//	return r.OnTeardownWeb(req)
	//}
	return &rtsp.Response{StatusCode: rtsp.StatusNotImplemented}, nil
}

func (r *Rstp) OnResponse(conn *rtsp.Conn, resp *rtsp.Response) {
	global.Debug.Printf("response sent :head %d body %d", len(resp.Header), len(resp.Body))
}
