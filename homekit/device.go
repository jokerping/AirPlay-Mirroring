package homekit

func AirplayDevice() Features {
	var features = NewFeatures().Set(SupportsAirPlayVideoV1).Set(SupportsAirPlayPhoto)
	features = features.Set(SupportsAirPlayScreen).Set(SupportsAirPlayAudio)
	features = features.Set(HasUnifiedAdvertiserInfo)
	return features
}
