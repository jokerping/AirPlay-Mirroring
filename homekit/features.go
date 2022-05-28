package homekit

import (
	"fmt"
	"math/big"
)

type Features struct {
	Value big.Int
}

const (
	SupportsAirPlayVideoV1                = 0
	SupportsAirPlayPhoto                  = 1
	supportAirplayVideoFairPlay           = 2
	SupportsAirPlaySlideshow              = 5
	SupportsAirPlayUnKnow6                = 6 //不知道是啥，文档里是空的
	SupportsAirPlayScreen                 = 7
	SupportsAirPlayAudio                  = 9
	SupportsAirPlayUnKnow10               = 10 //不知道是啥，文档里是空的
	AudioRedundant                        = 11
	FPSAPv2pt5_AES_GCM                    = 12
	PhotoCaching                          = 13
	Authentication_4                      = 14
	MetadataFeatures_0                    = 15
	MetadataFeatures_1                    = 16
	MetadataFeatures_2                    = 17
	AudioFormats_0                        = 18
	AudioFormats_1                        = 19
	AudioFormats_2                        = 20
	AudioFormats_3                        = 21
	SupportsAirPlayUnKnow22               = 22
	Authentication_1                      = 23
	SupportsAirPlayUnKnow25               = 25
	Authentication_8                      = 26
	SupportsLegacyPairing                 = 27
	SupportsAirPlayUnKnow28               = 28
	HasUnifiedAdvertiserInfo              = 30
	IsCarPlay                             = 32
	SupportsVolume                        = 32
	SupportsAirPlayVideoPlayQueue         = 33
	SupportsAirPlayFromCloud              = 34
	SupportsTLS_PSK                       = 35
	SupportsUnifiedMediaControl           = 38
	SupportsBufferedAudio                 = 40
	SupportsPTP                           = 41
	SupportsScreenMultiCodec              = 42
	SupportsSystemPairing                 = 43
	IsAPValeriaScreenSender               = 44
	SupportsHKPairingAndAccessControl     = 46
	SupportsHKPeerManagement              = 47
	SupportsCoreUtilsPairingAndEncryption = 48
	SupportsAirPlayVideoV2                = 49
	MetadataFeatures_3                    = 50
	SupportsUnifiedPairSetupAndMFi        = 51
	SupportsSetPeersExtendedMessage       = 52
	SupportsAPSync                        = 54
	SupportsWoL1                          = 55
	SupportsWoL2                          = 56
	SupportsHangdogRemoteControl          = 58
	SupportsAudioStreamConnectionSetup    = 59
	SupportsAudioMediaDataControl         = 60
	SupportsRFC2198Redundancy             = 61
)

func (flag Features) Set(i int) Features {
	flag.Value.SetBit(&flag.Value, i, 1)
	return flag
}

func (flag Features) UnSet(i int) Features {
	flag.Value.SetBit(&flag.Value, i, 0)
	return flag
}

func (flag Features) ToRecord() string {
	return fmt.Sprintf("0x%x,0x%x", flag.Value.Int64()&0xffffffff, flag.Value.Int64()>>32&0xffffffff)
}

func (flag Features) ToUint64() uint64 {
	return flag.Value.Uint64()
}

func NewFeatures() Features {
	return Features{Value: *big.NewInt(0)}
}
