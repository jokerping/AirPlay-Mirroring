package handlers

import "AirPlayServer/rtsp"

func (r *Rstp) OnPostWeb(conn *rtsp.Conn, req *rtsp.Request) (*rtsp.Response, error) {

	switch req.Path {
	case "pair-setup":
		return r.OnPairSetup(req)
	case "pair-verify":
		return r.OnPairVerify(req)
	case "fp-setup":
		return r.OnFpSetup(req)
	case "feedback":
		return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil //心跳包
	}
	return &rtsp.Response{StatusCode: rtsp.StatusNotImplemented}, nil
}
