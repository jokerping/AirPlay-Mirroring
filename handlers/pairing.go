package handlers

import (
	"AirPlayServer/rtsp"
	"crypto/ed25519"
)

func (r *Rstp) OnPairSetup(conn *rtsp.Conn, req *rtsp.Request) (*rtsp.Response, error) {
	//根据Length 生成指定长度的公钥
	var publicKey ed25519.PublicKey
	var err error
	publicKey, _, err = ed25519.GenerateKey(nil)
	if err != nil {
		return &rtsp.Response{StatusCode: rtsp.StatusServiceUnavailable}, err
	}
	contentType, found := req.Header["Content-Type"]
	if !found {
		contentType = rtsp.HeaderValue{"application/octet-stream"}
	}
	return &rtsp.Response{StatusCode: rtsp.StatusOK, Header: rtsp.Header{
		"Content-Type": contentType,
	}, Body: publicKey}, err

}
