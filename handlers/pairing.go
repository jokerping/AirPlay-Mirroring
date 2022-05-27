package handlers

import (
	"AirPlayServer/rtsp"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"golang.org/x/crypto/curve25519"
	"io"
)

var edTheirs [32]byte           //用于第二次PairVerify 验签
var edPubKey ed25519.PublicKey  //ed25519签名使用的公钥
var edPriKey ed25519.PrivateKey //ed25519签名使用的私钥
var privateKey [32]byte         //服务器curve25519算法生成的私钥
var publicKey [32]byte          // 服务器curve25519算法生成的公钥
var ecdhTheir [32]byte          //第一次传输过来curve25519公钥
var blockMode cipher.Stream     //aes 加解密器，由于ctr特性，两次需要使用同一个aes
var pairVerifyAESKey = "Pair-Verify-AES-Key"
var pairVerifyAESIV = "Pair-Verify-AES-IV"

func (r *Rstp) OnPairSetup(req *rtsp.Request) (*rtsp.Response, error) {
	//根据Length 生成指定长度的公钥
	var err error
	edPubKey, edPriKey, err = ed25519.GenerateKey(nil) //此处签名需要保存，后面ed签名都需要使用这个公钥和私钥
	if err != nil {
		return &rtsp.Response{StatusCode: rtsp.StatusServiceUnavailable}, err
	}
	contentType, found := req.Header["Content-Type"]
	if !found {
		contentType = rtsp.HeaderValue{"application/octet-stream"}
	}
	return &rtsp.Response{StatusCode: rtsp.StatusOK, Header: rtsp.Header{
		"Content-Type": contentType,
	}, Body: edPubKey}, err
}

func (r *Rstp) OnPairVerify(req *rtsp.Request) (*rtsp.Response, error) {

	if edPriKey == nil {
		return &rtsp.Response{StatusCode: rtsp.StatusBadRequest}, nil
	}
	if len(req.Body) < 68 { //这一步必须这么多字节
		return &rtsp.Response{StatusCode: rtsp.StatusBadRequest}, errors.New("PairVerify 接收数据为空")
	}
	//前四个字节用作标志位判断,取第一个字节用于判断
	temp := int(req.Body[0])
	if temp == 1 {
		//将剩下64个字节按每32个字节分两组,前32位手机Curve25519算法的公钥，后32位Ed25519加签公钥
		copy(ecdhTheir[:], req.Body[4:36])
		copy(edTheirs[:], req.Body[36:])
		//使用curve25519算法将服务器curve25519私钥+手机传过来的公钥（前32个字节）进行计算，得到服务器key
		if _, err := io.ReadFull(rand.Reader, privateKey[:]); err != nil {
			return &rtsp.Response{StatusCode: rtsp.StatusNotImplemented}, err
		}
		curve25519.ScalarBaseMult(&publicKey, &privateKey)
		//使用生成的服务器curve25519私钥作为key得到aes-ctr-128的key
		ecdhShared, err := curve25519.X25519(privateKey[:], ecdhTheir[:])
		if err != nil {
			return &rtsp.Response{StatusCode: rtsp.StatusNotImplemented}, err
		}
		//构建返回数据，前32字节是服务器使用curve25519算法产生的公钥
		respBody := make([]byte, 32)
		copy(respBody, publicKey[:])
		//ed25519 签名，签名内容为服务器curve25519公钥+手机传送过来的公钥共64字节
		//后64字节使用aes-ctr-128算法加密ed25519签名
		data := append(publicKey[:], ecdhTheir[:]...) //签名内容为服务器curve25519公钥+手机公钥
		edSign := ed25519.Sign(edPriKey, data)
		//使用aes-ctr-128 加密签名
		//message, err := aesCtr(edSign)
		//将32位的ecdhShared 散列成16位ecdhShared作为aes的key使用从而得到128位的aes。
		//必须这么散列，否则接受不到。使用sha512讲 拼接后的字符串散列，然后截取前16个字符。
		hash := sha512.Sum512(append([]byte(pairVerifyAESKey), ecdhShared[:]...))
		block, err := aes.NewCipher(hash[:16])
		if err != nil {
			return nil, err
		}
		//aes ctr模式加密
		iv := sha512.Sum512(append([]byte(pairVerifyAESIV), ecdhShared[:]...)) //给到一个iv值，长度等于block的块尺寸，即16byte
		blockMode = cipher.NewCTR(block, iv[:16])
		message := make([]byte, len(edSign))
		blockMode.XORKeyStream(message, edSign)
		if err != nil {
			return &rtsp.Response{StatusCode: rtsp.StatusNotImplemented}, nil
		}
		respBody = append(respBody, message...)
		contentType, found := req.Header["Content-Type"]
		if !found {
			contentType = rtsp.HeaderValue{"application/octet-stream"}
		}
		return &rtsp.Response{StatusCode: rtsp.StatusOK, Header: rtsp.Header{
			"Content-Type": contentType,
		}, Body: respBody}, err

	} else if temp == 0 {
		//第二次请求验证客户端签名
		sig := make([]byte, len(req.Body[4:]))
		blockMode.XORKeyStream(sig, req.Body[4:])
		//验签
		message := append(ecdhTheir[:], publicKey[:]...)
		verify := ed25519.Verify(edTheirs[:], message, sig)
		if verify {
			contentType, found := req.Header["Content-Type"]
			if !found {
				contentType = rtsp.HeaderValue{"application/octet-stream"}
			}
			return &rtsp.Response{StatusCode: rtsp.StatusOK, Header: rtsp.Header{
				"Content-Type": contentType,
			}, Body: nil}, nil
		} else {
			return &rtsp.Response{StatusCode: rtsp.StatusBadRequest}, errors.New("PairVerify 服务器验签失败")
		}
	}

	return &rtsp.Response{StatusCode: rtsp.StatusNotImplemented}, nil
}
