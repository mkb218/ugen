package ugen

import "testing"
import "time"
import "sync"

type BufMaker struct {
	UGenBase
	*testing.T
	quitchan chan q
}

func (*BufMaker) GetParams() []ParamDesc {
	return []ParamDesc{}
}

func (b *BufMaker) Start(op OutputParams) error {
	b.quitchan = make(chan q)
	b.outchans = []chan []float32{make(chan []float32)}
	go func() {
		i := 0
		// b.Log("starting")
		for {
			buf := GetNewBuf(op)
			for i := range buf {
				logger.Println(i)
				buf[i] = float32(i)
			}
			select {
			case <-b.quitchan:
				b.Log("stopping")
				return
			case b.outchans[0] <- buf:
				// b.Log("sent a buf", i)
				i++
			}
		}
	}()
	return nil
}

func (b *BufMaker) Stop() error {
	b.quitchan <- q{}
	return nil
}

func TestSpreader(t *testing.T) {
	var b UGen = &BufMaker{T: t}
	spreader := NewSpreader(2)
	spreader.SetInput(0, b)
	op := OutputParams{SampleRate: 1000, BufferSize: 32}

	e := spreader.Start(op)
	if e != nil {
		t.Error(e)
	}
	okchan := make(chan q)
	var l0, r0, l1, r1 []float32
	
	var wg sync.WaitGroup
	go func() {
		wg.Add(2)
		go func() {
			l0 = <-spreader.OutputChannels()[0]
			wg.Done()
		}()
		go func() {
			r0 = <-spreader.OutputChannels()[1]
			wg.Done()
		}()
		wg.Wait()
		okchan <- q{}
	}()
	go func() {
		wg.Add(2)
		go func() {
			l1 = <-spreader.OutputChannels()[0]
			wg.Done()
		}()
		go func() {
			r1 = <-spreader.OutputChannels()[1]
			wg.Done()
		}()
		wg.Wait()
		okchan <- q{}
	}()
	for i := range []int{0, 1} {
		select {
		case <-okchan:
		case <-time.After(50 * time.Millisecond):
			t.Log("timeout, probably deadlocked", i)
			t.FailNow()
		}
	}

	for i, r := range [][]float32{l0, r0, l1, r1} {
		for j, s := range r {
			if float32(j) != s {
				t.Errorf("bad value in array %d, expected %f got %f\n", i, float32(j), s)
			}
		}
	}
	e = spreader.Stop()
	if e != nil {
		t.Error(e)
	}

}
