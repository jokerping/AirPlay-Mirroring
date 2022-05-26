package handlers

import (
	"AirPlayServer/config"
	"AirPlayServer/global"
	"AirPlayServer/homekit"
	"AirPlayServer/rtsp"
	"howett.net/plist"
	"strings"
)

type getInfoContent struct {
	Qualifier []string `plist:"qualifier"`
}

type audioLatenciesResponse struct {
	AudioType           string `plist:"audioType"`
	InputLatencyMicros  uint64 `plist:"inputLatencyMicros"`
	OutputLatencyMicros uint64 `plist:"outputLatencyMicros"`
	Type                uint64 `plist:"type"`
}

type displaysResponse struct {
	Height         int    `plist:"height"`
	Width          int    `plist:"width"`
	Rotation       bool   `plist:"rotation"`
	WidthPhysical  bool   `plist:"widthPhysical"`
	HeightPhysical bool   `plist:"heightPhysical"`
	WidthPixels    int    `plist:"widthPixels"`
	HeightPixels   int    `plist:"heightPixels"`
	RefreshRate    int    `plist:"refreshRate"`
	Features       int64  `plist:"features"`
	MaxFPS         int    `plist:"maxFPS"`
	Overscanned    bool   `plist:"overscanned"`
	UUID           string `plist:"uuid"`
}

type getInfoResponse struct {
	AudioLatencies  []audioLatenciesResponse `plist:"audioLatencies"`
	Displays        []displaysResponse       `plist:"displays"` //看起来像显示数据
	DeviceId        string                   `plist:"deviceID"`
	Features        string                   `plist:"features"`
	Pi              string                   `plist:"pi"`
	Psi             string                   `plist:"psi"`
	ProtocolVersion string                   `plist:"protocolVersion"`
	Sdk             string                   `plist:"sdk"`           //解码出来的，不知道什么影响
	SourceVersion   string                   `plist:"sourceVersion"` //服务器版本
	StatusFlags     string                   `plist:"statusFlags"`
	Name            string                   `plist:"name"`       //设备名称
	TxtAirPlay      []byte                   `plist:"txtAirPlay"` //广播发送字典的base64字符串
	PTPInfo         string                   `plist:"PTPInfo"`    //不知道是什么
	Pk              []byte                   `plist:"pk"`         //公钥，应该跟广播时用的一致
}

func (r *Rstp) OnGetWeb(req *rtsp.Request) (*rtsp.Response, error) {

	switch req.Path {
	case "info":
		return r.OnGetInfo(req)
	case "stream.xml":
		global.Debug.Println("请求流")
	}
	return &rtsp.Response{StatusCode: rtsp.StatusNotImplemented}, nil
}

func (r *Rstp) OnGetInfo(req *rtsp.Request) (*rtsp.Response, error) {

	if contentType, found := req.Header["Content-Type"]; found && strings.EqualFold(contentType[0], "application/x-apple-binary-plist") {
		var content getInfoContent
		if _, err := plist.Unmarshal(req.Body, &content); err != nil {
			return &rtsp.Response{StatusCode: rtsp.StatusBadRequest}, err
		}
	}
	latencies := [6]audioLatenciesResponse{{
		InputLatencyMicros:  0,
		OutputLatencyMicros: 400000,
		Type:                100,
	}, {
		InputLatencyMicros:  0,
		OutputLatencyMicros: 400000,
		Type:                100,
		AudioType:           "default",
	}, {
		InputLatencyMicros:  0,
		OutputLatencyMicros: 400000,
		Type:                100,
		AudioType:           "media",
	}, {
		InputLatencyMicros:  0,
		OutputLatencyMicros: 400000,
		Type:                100,
		AudioType:           "telephony",
	}, {
		InputLatencyMicros:  0,
		OutputLatencyMicros: 400000,
		Type:                100,
		AudioType:           "speechRecognition",
	}, {
		InputLatencyMicros:  0,
		OutputLatencyMicros: 400000,
		Type:                100,
		AudioType:           "alerts",
	},
	}
	displays := [1]displaysResponse{{
		Height:         1080,
		Width:          1920,
		Rotation:       false,
		WidthPhysical:  false,
		HeightPhysical: false,
		WidthPixels:    1920,
		HeightPixels:   1080,
		RefreshRate:    60,
		Features:       homekit.Device.Features.Value.Int64(),
		MaxFPS:         60,
		Overscanned:    false,
		UUID:           config.Config.DeviceUUID,
	}}
	var all string
	for _, re := range homekit.Device.ToRecords() {
		all += re
	}
	//base64.StdEncoding.EncodeToString(homekit.Device.ToRecords())
	//构建返回数据
	responseBody := &getInfoResponse{
		AudioLatencies:  latencies[:],
		Displays:        displays[:],
		DeviceId:        homekit.Device.Deviceid,
		Features:        homekit.Device.Features.ToRecord(),
		Pi:              homekit.Device.Pi.String(),
		Psi:             homekit.Device.Psi.String(),
		ProtocolVersion: "1.1",
		Sdk:             "AirPlay;2.1.1-f.1", //这都是抓包出来的不知道怎么设置的
		SourceVersion:   homekit.Device.Srcvers,
		StatusFlags:     homekit.Device.Flags,
		Name:            config.Config.DeviceName,
		TxtAirPlay:      []byte(all),
		PTPInfo:         "OpenAVNU ArtAndLogic-aPTP-changes Commit: 17f0335 on Sep 22, 2018",
		Pk:              []byte(homekit.Device.Pk),
	}

	if body, err := plist.Marshal(*responseBody, plist.AutomaticFormat); err == nil {
		return &rtsp.Response{StatusCode: rtsp.StatusOK, Header: rtsp.Header{
			"Content-Type": rtsp.HeaderValue{"application/x-apple-binary-plist"},
		}, Body: body}, nil
	}

	return &rtsp.Response{StatusCode: rtsp.StatusNotImplemented}, nil
}
