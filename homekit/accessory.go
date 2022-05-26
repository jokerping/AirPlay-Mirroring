package homekit

import (
	"AirPlayServer/global"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

type UUID struct {
	uuid.UUID
}

type Accessory struct {
	Srcvers   string
	Deviceid  string
	Features  Features
	Flags     string //16进制
	model     string //目前看来不同的model显示的图标不同
	protovers string
	acl       string
	rsf       string
	Pi        UUID
	Gid       UUID
	Psi       UUID
	gcgl      string
	igl       string
	Pk        string
}

func NewAccessory(deviceId string, currentUuid string, features Features) *Accessory {
	var deviceUuid = uuid.MustParse(currentUuid)
	return &Accessory{
		Srcvers:   "366.0",
		Deviceid:  deviceId,
		Features:  features,
		Flags:     "0x20044",    //最好不用动，不同的flag会过来不同链接，坑不少。比如设置标志位9后手机会请求一次性匹配/pair-pin-start链接
		model:     "AppleTV5,3", //只能是appleTV ，写别的会先请求options，但是按网上找的格式返回后不对
		protovers: "1.1",
		acl:       "0",
		rsf:       "0x0",
		Pi:        UUID{deviceUuid},
		Gid:       UUID{deviceUuid},
		Psi:       UUID{deviceUuid},
		gcgl:      "1",
		igl:       "1",
		Pk:        "b07727d6f6cd6e08b58ede525ec3cdeaa252ad9f683feb212ef8a205246554e7",
	}
}

func (t *Accessory) String() string {
	return fmt.Sprintf("Pi: %s, guid: %s, Psi: %s", t.Pi, t.Gid, t.Psi)
}

func (uid UUID) ToRecord() string {
	return uid.String()
}

func (t *Accessory) ToRecords() []string {

	fields := reflect.TypeOf(*t)
	values := reflect.ValueOf(*t)

	numField := values.NumField()
	results := make([]string, numField)

	for i := 0; i < numField; i++ {
		results[i] = strings.ToLower(fields.Field(i).Name) + "="
		value := values.Field(i)
		switch fields.Field(i).Type.Name() {
		case "string":
			results[i] += value.String()
		case "UUID":
			results[i] += value.Interface().(UUID).ToRecord()
		case "Features":
			results[i] += value.Interface().(Features).ToRecord()
		default:
			panic(fields.Field(i).Type.Name())
		}
	}
	global.Debug.Println("服务发现特征值:\n", results)
	return results

}

var Device *Accessory
