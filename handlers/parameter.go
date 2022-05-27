package handlers

import (
	"AirPlayServer/config"
	"AirPlayServer/rtsp"
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

// OnGetParameterWeb 客户端在想知道服务器的音量级别
func (r *Rstp) OnGetParameterWeb(req *rtsp.Request) (*rtsp.Response, error) {
	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "text/parameters") {
		var body string
		scanner := bufio.NewScanner(bytes.NewReader(req.Body))
		for scanner.Scan() {
			switch scanner.Text() {
			case "volume":
				body += fmt.Sprintf("volume: %f\r\n", config.Config.Volume)
			}
		}
		return &rtsp.Response{StatusCode: rtsp.StatusOK, Header: rtsp.Header{
			"Content-Type": rtsp.HeaderValue{"text/parameters"},
		}, Body: []byte(body)}, nil

	}
	return &rtsp.Response{StatusCode: rtsp.StatusInternalServerError}, nil
}

// OnSetParameterWeb 客户端设置服务器音量
func (r *Rstp) OnSetParameterWeb(req *rtsp.Request) (*rtsp.Response, error) {

	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "text/parameters") {
		scanner := bufio.NewScanner(bytes.NewReader(req.Body))
		for scanner.Scan() {
			var vol float64
			line := scanner.Text()
			if strings.HasPrefix(line, "volume") {
				if c, err := fmt.Sscanf(line, "volume: %f", &vol); c != 1 || err != nil {
					fmt.Printf("erreur parsing volume parameters : %s\n", line)
				} else {
					config.Config.Volume = vol
					//TODO 此处设置音量
				}
			}
		}
		return &rtsp.Response{StatusCode: rtsp.StatusOK}, nil
	}

	return &rtsp.Response{StatusCode: rtsp.StatusInternalServerError}, nil
}
