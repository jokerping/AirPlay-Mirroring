package media

/*
#include "playfair/playfair.c"
#include "playfair/omg_hax.c"
#include "playfair/modified_md5.c"
#include "playfair/sap_hash.c"
#include "playfair/hand_garble.c"
*/
import "C"
import (
	"AirPlayServer/rtsp"
	"unsafe"
)

func desryAesKey() []byte {
	desryAesKey := make([]byte, 16)
	//1.解密AES密钥，用到fpsetup（CSeq=5）接收到的keymessage
	C.playfair_decrypt(
		(*C.uchar)(unsafe.Pointer(&rtsp.Session.KeyMessage[0])),
		(*C.uchar)(unsafe.Pointer(&rtsp.Session.Ekey[0])),
		(*C.uchar)(unsafe.Pointer(&desryAesKey[0])))
	rtsp.Session.DesryAesKey = make([]byte, 16)
	copy(rtsp.Session.DesryAesKey, desryAesKey)
	return desryAesKey
}
