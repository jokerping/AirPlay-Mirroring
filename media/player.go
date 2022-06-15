package media

import (
	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/go-gst/gst/app"
	"time"
)

const h264Caps = "video/x-h264,colorimetry=bt709,stream-format=(string)byte-stream,alignment=(string)au"

var src *app.Source

func PlayVideo(data H264Data) {
	buffer := gst.NewBufferWithSize(int64(data.Length))
	buffer.SetPresentationTimestamp(time.Duration(data.Pts))
	buffer.FillBytes(0, data.Data)
	src.PushBuffer(buffer)
}

func CreatePipeline() (*gst.Pipeline, error) {
	gst.Init(nil)
	launch := "appsrc name=video_source ! "
	launch += "queue ! "
	launch += "h264parse"
	launch += " ! "
	launch += "decodebin"
	launch += " ! "
	launch += "videoconvert"
	launch += " ! "
	launch += "autovideosink"
	launch += " name=video_sink sync=false"
	pipeline, err := gst.NewPipelineFromString(launch)
	element, _ := pipeline.GetElementByName("video_source")
	caps := gst.NewCapsFromString(h264Caps)
	element.SetProperty("caps", caps)
	element.SetProperty("stream-type", 0)
	element.SetProperty("is-live", true)
	element.SetProperty("format", gst.FormatTime)
	src = app.SrcFromElement(element)
	pipeline.SetState(gst.StateReady)
	return pipeline, err
}
