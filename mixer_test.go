package ugen

import "testing"
import "sync"
import "time"
import "math"
import "fmt"

func TestMixer(t *testing.T) {
	testMixer(t,2,1)
	testMixer(t,2,2)
}

func testMixer(t *testing.T, inputs, buswidth int) {
	ugens := make([]UGen, inputs)
	for i := range ugens {
		ugens[i] = &BufMaker{T:t}
	}
	
	quitchan := make(chan q)
	go func() {
		testMixerSub(t, ugens, buswidth, "single")
		t.Log("back from sub, sending to quitchan")
		quitchan <- q{}
	}()
	
	select {
	case <- time.After(time.Second):
		t.Log("single timed out, possibly deadlocked")
		t.Fail()
	case <- quitchan:
	}
	
	if buswidth > 1 {
		for i := range ugens {
			ugens[i] = NewSpreaderWithUGen(buswidth,ugens[i])
		}
		go func() {
			testMixerSub(t, ugens, buswidth, "full")
			quitchan <- q{}
		}()

		select {
		case <- time.After(500*time.Millisecond):
			t.Log("full timed out, probably deadlocked")
			t.FailNow()
		case <- quitchan:
		}
	} 
}

func testMixerSub(t *testing.T, inputs []UGen, buswidth int, id string) {
	t.Log(len(inputs), buswidth, id)
	mx := NewMixer(len(inputs), buswidth)
	op := OutputParams{BufferSize:10, SampleRate:1000}
	
	// static gains
	for i := range inputs {
		mx.SetInput(i, inputs[i])
		mx.SetGain(i, 0.1)
	}
	
	mx.Start(op)
	
	var wg sync.WaitGroup
	var ob = make([][]float32,buswidth)
	for i := 0; i < buswidth; i++ {
		wg.Add(1)
		go func(i int) {
			ob[i] = <- mx.OutputChannels()[i]
			wg.Done()
		}(i)
	}
	wg.Wait()
	t.Log("got one output buffer")
	
	for i, b := range ob {
		for j, v := range b {
			if math.Abs(float64(float32(j*len(inputs)) * 0.1 - v)) > 0.001 {
				t.Logf("got unexpected value %v instead of %v for channel %v sample %v", v, float32(j*len(inputs)) * 0.1, i, j)
				t.Fail()
			}
		}
	}
	
	
	// double all gains
	for i := range inputs {
		mx.SetInput(i, inputs[i])
		mx.SetGain(i, 0.2)
	}
	for i := 0; i < buswidth; i++ {
		wg.Add(1)
		go func(i int) {
			ob[i] = <- mx.OutputChannels()[i]
			wg.Done()
		}(i)
	}
	wg.Wait()
	t.Log("got one output buffer")

	for i, b := range ob {
		for j, v := range b {
			if math.Abs(float64(float32(j*len(inputs)) * 0.1 - v)) > 0.001 {
				t.Logf("got unexpected value %v instead of %v for channel %v sample %v", v, float32(j*len(inputs)) * 0.2, i, j)
				t.Fail()
			}
		}
	}

	for i := 0; i < buswidth; i++ {
		wg.Add(1)
		go func(i int) {
			ob[i] = <- mx.OutputChannels()[i]
			wg.Done()
		}(i)
	}
	wg.Wait()
	t.Log("got one output buffer")

	for i, b := range ob {
		for j, v := range b {
			if math.Abs(float64(float32(j*len(inputs)) * 0.2 - v)) > 0.001 {
				t.Logf("got unexpected value %v instead of %v for channel %v sample %v", v, float32(j*len(inputs)) * 0.2, i, j)
				t.Fail()
			}
		}
	}
	mx.Stop()
	
	fmt.Println("returning from",id)
}
	
	
		