package ugen

import "math"

type Sin struct {
	UGenBase
	freq float32
	phase float32
	freqChan chan float32
	phaseChan chan float32
	universe *Universe // needs samplerate
	quitchan chan q
}

func init() {
	var _ UGen = &Sin{}
}

func NewSin(f, p float32, u *Universe) *Sin {
	var s Sin
	s.freq = f
	s.phase = p
	s.freqChan = make(chan float32, 1)
	s.phaseChan = make(chan float32, 1)
	s.quitchan = make(chan q)
	s.universe = u
	s.outchans = append(s.outchans, make(chan float32, 1))
	return &s
}

func (s *Sin) Start() error {
	go func() {
		var samplenum int64
		for {
			select {
			case f := <- s.freqChan:
				s.freq = f
			case p := <- s.phaseChan:
				s.phase = p
			default:
			}
		
			y := math.Sin(float64(s.freq) * float64(samplenum) / s.universe.SampleRate + float64(s.phase))
			println("y",y)
			select {
			case <- s.quitchan:
				return
			case s.outchans[0] <- float32(y):
			}

			samplenum++
		}
	}()
	return nil
}

func (s *Sin) Stop() error {
	s.quitchan <- q{}
	return nil
}