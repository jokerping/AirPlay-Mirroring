package homekit

func AirplayDevice() Features {
	var features = NewFeatures().Set(SupportsAirPlayPhoto).Set(supportAirplayVideoFairPlay)
	features = features.Set(SupportsAirPlayUnKnow6).Set(SupportsAirPlayScreen)
	features = features.Set(SupportsAirPlayAudio).Set(SupportsAirPlayUnKnow10)
	features = features.Set(AudioRedundant).Set(FPSAPv2pt5_AES_GCM)
	features = features.Set(PhotoCaching).Set(Authentication_4)
	features = features.Set(MetadataFeatures_0).Set(MetadataFeatures_1)
	features = features.Set(MetadataFeatures_2).Set(AudioFormats_0)
	features = features.Set(AudioFormats_1).Set(AudioFormats_2)
	features = features.Set(AudioFormats_3).Set(SupportsAirPlayUnKnow22)
	features = features.Set(SupportsAirPlayUnKnow25).Set(SupportsLegacyPairing)
	features = features.Set(SupportsAirPlayUnKnow28).Set(HasUnifiedAdvertiserInfo)
	features = features.Set(SupportsAirPlaySlideshow)
	return features
}
