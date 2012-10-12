package main

import "time"
import "github.com/mkb218/ugen"
import "fmt"

func main() {
	fmt.Println("start!")
	var s0 ugen.UGen = ugen.NewSin(1000, 0)
	var s1 ugen.UGen = ugen.NewSin(1004, 0)
	var s2 ugen.UGen = ugen.NewSin(1008, 0)
	var s3 ugen.UGen = ugen.NewSin(1016, 0)
	
	ml := ugen.NewMixer(4,2)
	
	ml.SetInput(0, s0)
	ml.SetInput(1, s1)
	ml.SetInput(2, ugen.NewSpreaderWithUGen(2, s2))
	ml.SetInput(3, ugen.NewSpreaderWithUGen(2, s3))
	
	ml.ParamChannel() <- ugen.ParamValue{0, 0.2}
	ml.ParamChannel() <- ugen.ParamValue{1, 0.2}
	ml.ParamChannel() <- ugen.ParamValue{2, 0.2}
	ml.ParamChannel() <- ugen.ParamValue{3, 0.2}

	var o = ugen.NewPortAudioOutput(2)
	var op = ugen.OutputParams{SampleRate:44100, BufferSize: 4096}
	
	o.SetInput(0, ml)
	o.Start(op)
	go func() { 
		t := time.Tick(time.Second)
		for {
			<- t
			ugen.LogRecycleStats()
//			ugen.LogStackTrace()
		}
	}()
	fmt.Println("sleeping!")
	time.Sleep(time.Second/2)
	fmt.Println("stopping!")
	o.Stop()
}