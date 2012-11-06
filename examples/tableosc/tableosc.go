package main

import "time"
import "github.com/mkb218/ugen"
import "github.com/mkb218/gosndfile/sndfile"
import "fmt"
import "flag"

func main() {
	maxfreq := flag.Float64("maxfreq", 440, "max freq for sweep")
	sinchg := flag.Float64("sinchgrate", 15, "chg rate for sin wav")
	bufsize := flag.Int("bufsize", 4096, "buffer size")
	t := flag.Int64("time", 15000, "milliseconds")
	filename := flag.String("filename", "tableout.wav", "output filename")
	flag.Parse()
	tables := flag.Args()
	
	fmt.Println("start!")
	var sp ugen.UGen = ugen.NewSin(float32(*sinchg), 0, 440, float32(*maxfreq - 440))
	var ps = ugen.NewAudioParam()
	ps.SetUGenSource(0, sp, 0)

	var o = ugen.NewSndfileOut(*filename, sndfile.Info{Channels:2, Format:sndfile.SF_FORMAT_WAV|sndfile.SF_FORMAT_PCM_24},1, true)
	var op = ugen.OutputParams{SampleRate:44100, BufferSize: *bufsize}
	
	custombuf := make([]float32, 8192)
	
	go func() { 
		t := time.Tick(time.Second)
		for {
			<- t
			ugen.LogRecycleStats()
			// ugen.LogStackTrace()
		}
	}()
	TABLES: for _, table := range tables {
		var s0 ugen.UGen
		var err error
		switch table {
		case "sin":
			s0, err = ugen.NewTableOsc(1000, 0, 0, 1, false, ugen.SinTable, 8192, nil)
		case "tan":
			s0, err = ugen.NewTableOsc(1000, 0, 0, 1, false, ugen.TanTable, 8192, nil)
		case "pulse":
			s0, err = ugen.NewTableOsc(1000, 0, 0, 1, false, ugen.PulseTable, 8192, nil, 0.5)
		case "tri":
			s0, err = ugen.NewTableOsc(1000, 0, 0, 1, false, ugen.TriTable, 8192, nil)
		case "saw":
			s0, err = ugen.NewTableOsc(1000, 0, 0, 1, false, ugen.SawTable, 8192, nil, 0)
		case "custom":
			s0, err = ugen.NewTableOsc(1000, 0, 0, 1, false, ugen.CustomTable, 8192, custombuf)
		default:
			continue TABLES
		}
		
		if err != nil {
			fmt.Println(err)
			continue TABLES
		}

		spread := ugen.NewSpreaderWithUGen(2, s0)
		o.SetInput(0, spread)
		ps.SetDest(0, s0, 0)
		ps.Start(op)
		o.Start(op)
		fmt.Println("sleeping!")
		time.Sleep(time.Duration(*t)*time.Millisecond)
		fmt.Println("stopping!")
		o.Stop()
		ps.Stop()
	}
}