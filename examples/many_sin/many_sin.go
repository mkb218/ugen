package main

import "time"
import "github.com/mkb218/ugen"
import "fmt"
import "flag"

func main() {
	count := flag.Int("count", 0, "Number of sines")
	interval := flag.Float64("interval", 2, "Multiplier for each increment. If <= 1 increase arithmetically, not geometrically")
	outch := flag.Int("channels", 2, "channels in output device")
	bufsize := flag.Int("bufsize", 4096, "buffer size")
	samplerate := flag.Float64("samplerate", 44100, "sample rate")
	minfreq := flag.Float64("minfreq", 20, "starting frequency")
	maxfreq := flag.Float64("maxfreq", -1, "ending frequency")
	t := flag.Int64("time", 15000, "milliseconds")
	flag.Parse()


	freqs := make([]float32, 0)
	if *maxfreq == -1 {
		if *count <= 0 {
			fmt.Println("No maxfreq, no count. Can't read your mind, human.")
			return
		}
		if *interval > 1 {
			for f, i := float32(*minfreq), 0; i < *count; f, i = f*float32(*interval), i+1 {
				freqs = append(freqs, f)
			}
		} else {
			for f, i := float32(*minfreq), 0; i < *count; f, i = f+float32(*minfreq), i+1 {
				freqs = append(freqs, f)
			}
		}
	} else {
		if *interval > 1 {
			for f, i := float32(*minfreq), 0; i < *count && f < float32(*maxfreq) ; f, i = f*float32(*interval), i+1 {
				freqs = append(freqs, f)
			}
		} else {
			for f, i := float32(*minfreq), 0; i < *count && f < float32(*maxfreq); f, i = f+float32(*minfreq), i+1 {
				freqs = append(freqs, f)
			}
		}
	}
	
	fmt.Println("frequencies",freqs)
	
	mx := ugen.NewMixer(len(freqs), *outch)
	
	for i := 0; i < len(freqs); i++ {
		mx.SetInput(i, ugen.NewSpreaderWithUGen(*outch, ugen.NewSin(freqs[i], 0, 0, 1/float32(len(freqs)))))
	}

	fmt.Println("start!")

	var o = ugen.NewPortAudioOutput(*outch)
	var op = ugen.OutputParams{SampleRate:*samplerate, BufferSize: *bufsize}
	
	o.SetInput(0, mx)
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
}