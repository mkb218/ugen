package ugen

import "math"

// TODO: methods to set phase and freq
type Sin struct {
	UGenBase
	freq      float32
	phase     float32
	quitchan  chan q
}

func init() {
	var _ UGen = &Sin{}
}

func NewSin(f, p float32) *Sin {
	var s Sin
	s.freq = f
	s.phase = p
	s.paramchannel = make(chan ParamValue)
	s.quitchan = make(chan q)
	prepareoutchans(&(s.UGenBase), 1)
	return &s
}

func (s *Sin) GetParams() []ParamDesc {
	return []ParamDesc{ParamDesc{Name:"freq", Size:1}, ParamDesc{Name:"phase",Size:1}}
}

func (s *Sin) Start(op OutputParams) error {
	go func() {
		var samplenum int64
		for {
			PARAM: for {
				select {
				case p := <-s.paramchannel:
					switch p.Index {
					case 0:
						s.freq = p.Value
					case 1:
						s.phase = p.Value
					default:
						logger.Printf("Sin: Bad Index %d in ParamValue", p.Index)
					}
				default:
					break PARAM
				}
			}

			b := GetNewBuf(op)
			for x := range b {
				b[x] = float32(math.Sin(2*math.Pi*float64(s.freq)*float64(samplenum+int64(x))/op.SampleRate + float64(s.phase)))
			}
			samplenum += int64(op.BufferSize)
			select {
			case <-s.quitchan:
				return
			case s.outchans[0] <- b:
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
