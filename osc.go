package ugen

import "math"

// TODO: methods to set phase and freq
type Sin struct {
	UGenBase
	freq      float32
	phase     float32
	add       float32
	amp       float32
	quitchan  chan q
}

func init() {
	var _ UGen = &Sin{}
}

func NewSin(f, p, a, amp float32) *Sin {
	var s Sin
	s.freq = f
	s.phase = p
	s.add = a
	s.amp = amp
	s.paramchannel = make(chan ParamValue)
	s.quitchan = make(chan q)
	prepareoutchans(&(s.UGenBase), 1)
	return &s
}

func (s *Sin) GetParams() []ParamDesc {
	return []ParamDesc{ParamDesc{Name:"freq", Size:1}, ParamDesc{Name:"phase",Size:1}, ParamDesc{Name:"add",Size:1}, ParamDesc{Name:"amp",Size:1}}
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
						logger.Println("got new freq", s.freq)
						s.freq = p.Value
					case 1:
						s.phase = p.Value
					case 2:
						s.add = p.Value
					case 3:
						s.amp = p.Value
					default:
						logger.Printf("Sin: Bad Index %d in ParamValue", p.Index)
					}
				default:
					break PARAM
				}
			}

			b := GetNewBuf(op)
			for x := range b {
				b[x] = s.amp * float32(math.Sin(2*math.Pi*float64(s.freq)*float64(samplenum+int64(x))/op.SampleRate + float64(s.phase))) + s.add
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
