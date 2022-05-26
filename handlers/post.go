package handlers

import "AirPlayServer/rtsp"

func (r *Rstp) OnPostWeb(conn *rtsp.Conn, req *rtsp.Request) (*rtsp.Response, error) {

	switch req.Path {
	case "pair-setup":
		return r.OnPairSetup(conn, req)
	}
	return &rtsp.Response{StatusCode: rtsp.StatusNotImplemented}, nil
}
