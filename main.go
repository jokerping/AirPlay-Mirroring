package main

import (
	"AirPlayServer/airplay"
	"AirPlayServer/config"
	"AirPlayServer/global"
	"AirPlayServer/media"
	"flag"
	"github.com/tinyzimmer/go-glib/glib"
	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/go-gst/gst/app"
	"log"
	"os"
)

//本例只用于展示基本流程，实际工程中可使用成熟库代替自己实现的
//比如认证部分使用https://github.com/brutella/hap
func main() {
	debug := flag.Bool("debug", true, "Show debug output.")
	flag.StringVar(&config.Config.DeviceName, "n", "fuckAndroid", "Specify device name")
	flag.Parse()

	if *debug {
		global.Debug = log.New(os.Stderr, "DEBUG ", log.LstdFlags)
	}

	go func() {
		err := airplay.RunAirPlayServer()
		if err != nil {
			log.Fatal(err)
		}
	}()
	media.RunLoop(func(loop *glib.MainLoop) error {
		var pipeline *gst.Pipeline
		var err error
		if pipeline, err = media.CreatePipeline(); err != nil {
			return err
		}
		return mainLoop(loop, pipeline)
	})

}

func mainLoop(loop *glib.MainLoop, pipeline *gst.Pipeline) error {
	// Start the pipeline

	// Due to recent changes in the bindings - the finalizers might fire on the pipeline
	// prematurely when it's passed between scopes. So when you do this, it is safer to
	// take a reference that you dispose of when you are done. There is an alternative
	// to this method in other examples.
	pipeline.Ref()
	defer pipeline.Unref()

	pipeline.SetState(gst.StatePlaying)

	// Retrieve the bus from the pipeline and add a watch function
	pipeline.GetPipelineBus().AddWatch(func(msg *gst.Message) bool {
		if err := handleMessage(msg); err != nil {
			global.Debug.Println(err)
			loop.Quit()
			return false
		}
		return true
	})

	loop.Run()

	return nil
}

func handleMessage(msg *gst.Message) error {
	switch msg.Type() {
	case gst.MessageEOS:
		return app.ErrEOS
	case gst.MessageError:
		gerr := msg.ParseError()
		if debug := gerr.DebugString(); debug != "" {
			global.Debug.Println(debug)
		}
		return gerr
	}
	return nil
}
