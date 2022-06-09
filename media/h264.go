package media

type H264Data struct {
	FrameType int
	Data      []byte
	Length    int
	Pts       uint64
	Width     uint32
	Height    uint32
}

type H264Codec struct {
	Compatibility        byte
	LengthOfPps          uint16
	LengthOfSps          uint16
	Level                byte
	NumberOfPps          uint16
	PictureParameterSet  []byte
	ProfileHigh          byte
	Reserved3AndSps      byte
	Reserved6AndNal      byte
	SequenceParameterSet []byte
	Version              byte
}
