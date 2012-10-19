package main

import "time"
import "github.com/mkb218/ugen"
import "fmt"
import "flag"
func main() {
	maxfreq := flag.Float64("maxfreq", 440, "max freq for sweep")
	// sinchg := flag.Float64("sinchgrate", 15, "chg rate for sin wav")
	bufsize := flag.Int("bufsize", 4096, "buffer size")
	t := flag.Int64("time", 15000, "milliseconds")
	flag.Parse()
	fmt.Println("start!")
	var s0 ugen.UGen = ugen.NewSin(float32(*maxfreq), 0, 0, 1)
	// var s1 ugen.UGen = ugen.NewSin(1004, 0, 0, 1)
	// var s2 ugen.UGen = ugen.NewSin(1008, 0, 0, 1)
	// var s3 ugen.UGen = ugen.NewSin(1016, 0, 0, 1)
	// var sp ugen.UGen = ugen.NewSin(float32(*sinchg), 0, 440, float32(*maxfreq - 440))
	
	// ml := ugen.NewMixer(4,2)
	// 
	// ml.SetInput(0, ugen.NewSpreaderWithUGen(2, s0))
	// ml.SetInput(1, ugen.NewSpreaderWithUGen(2, s1))
	// ml.SetInput(2, ugen.NewSpreaderWithUGen(2, s2))
	// ml.SetInput(3, ugen.NewSpreaderWithUGen(2, s3))
	
	// ml.ParamChannel() <- ugen.ParamValue{0, 0.05}
	// ml.ParamChannel() <- ugen.ParamValue{1, 0.05}
	// ml.ParamChannel() <- ugen.ParamValue{2, 0.05}
	// ml.ParamChannel() <- ugen.ParamValue{3, 0.05}
	
	// var ps = ugen.NewAudioParam()
	// ps.SetUGenSource(0, sp, 0)
	// ps.SetDest(0, s0, 0)

	var o = ugen.NewPortAudioOutput(2)
	var op = ugen.OutputParams{SampleRate:44100, BufferSize: *bufsize}
	
	// o.SetInput(0, ml)
	o.SetInput(0, ugen.NewSpreaderWithUGen(2, s0))
//	ps.Start(op)
	o.Start(op)
	go func() { 
		t := time.Tick(time.Second)
		for {
			<- t
			ugen.LogRecycleStats()
			// ugen.LogStackTrace()
		}
	}()
	fmt.Println("sleeping!")
	time.Sleep(time.Duration(*t)*time.Millisecond)
	fmt.Println("stopping!")
	o.Stop()
	// ps.Stop()
}