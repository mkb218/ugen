package ugen

import "math"
import "fmt"
import "errors"
import "sync"

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
	var _ UGen = &TableOsc{}
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

type TableOsc struct {
	UGenBase
	freq      float32
	phase     float32
	add       float32
	amp       float32
	buf       []float32
	interp    bool
	start, stop *sync.Once
	quitchan  chan q
}

const (
	SinTable = iota
	TanTable = iota
	PulseTable = iota
	TriTable = iota
	SawTable = iota
	CustomTable = iota
)

func NewTableOsc(f, p, a, amp float32, interp bool, table int, bufsize int, buf []float32, args ...float32, ) (t *TableOsc, err error) {
	defer func() {
		if err != nil {
			t = nil
		}
	}()
	t = new(TableOsc)
	t.freq = f
	t.phase = p
	t.add = a
	t.amp = amp
	t.interp = interp
	t.outchans = make([]chan []float32,1)
	t.outchans[0] = make(chan []float32)
	switch table {
	case SinTable:
		t.buf = fillSin(bufsize)
	case TanTable:
		t.buf = fillTan(bufsize)
	case PulseTable:
		t.buf = fillPulse(bufsize, args[0])
	case TriTable:
		t.buf = fillTri(bufsize)
	case SawTable:
		t.buf = fillSaw(bufsize, args[0])
	case CustomTable:
		if buf == nil {
			err = errors.New("NewTableOsc: no buffer provided for CustomTable TableOsc")
			return
		}
		t.buf = buf
	default:
		err = fmt.Errorf("NewTableOsc: unknown table type %d", table)
	}
	
	t.start = new(sync.Once)
	t.stop = new(sync.Once)
	t.stop.Do(func(){})
	
	t.quitchan = make(chan q)
	t.paramchannel = make(chan ParamValue)

		
	return
}

func fill6464Base(bufsize int, f func(float64) float64) []float32 {
	b := make([]float32, bufsize)
	for i := range b {
		pos := float64(i) / float64(bufsize)
		b[i] = float32(f(pos / 2 / math.Pi))
	}
	return b
}

func fillSin(bufsize int) []float32 {
	return fill6464Base(bufsize, math.Sin)
}

func fillTan(bufsize int) []float32 {
	return fill6464Base(bufsize, math.Tan)
}

func fillPulse(bufsize int, duty float32) []float32 {
	thresh := int(duty / float32(bufsize))
	// slewrate is number of samples for transition
	b := make([]float32, bufsize)
	for i := range b {
		if i > thresh {
			b[i] = -1
		} else {
			b[i] = 1
		}
	}
	return b
}

func fillTri(bufsize int) []float32 {
	inc := 1/float32(bufsize/4)
	// slewrate is number of samples for transition
	b := make([]float32, bufsize)
	v := float32(0)
	for i := range b {
		b[i] = v
		if i < bufsize/4 || i >= 3*bufsize/4 {
			v += inc
		} else {
			v -= inc
		}
	}
	return b
}

func fillSaw(bufsize int, slope float32) []float32 {
	inc := 2/float32(bufsize)
	// slewrate is number of samples for transition, someday
	b := make([]float32, bufsize)
	v := float32(-1)
	for i := range b {
		b[i] = v
		v += inc
	}
	return b
}
	
func calcInc(b int, f, s float32) int {
	return int(float32(b) * f / s)	
}

func (t *TableOsc) GetParams() []ParamDesc {
	return []ParamDesc{ParamDesc{Name:"freq",Size:1},ParamDesc{Name:"add",Size:1},ParamDesc{Name:"amp",Size:1}}
}

func (t *TableOsc) Start(op OutputParams) (err error) {
	t.start.Do(func() {
		t.stop = new(sync.Once)
		go func() {
			// defer func() {
			// 	logger.Println("leaving")
			// }()
			inc := calcInc(len(t.buf), t.freq, float32(op.SampleRate))
			pos := int(t.phase / 2 / math.Pi * float32(len(t.buf)))
			for {
				
				// logger.Println("startloop")
				ob := GetNewBuf(op)
				// logger.Println("newbuf got")
				for i := range ob {
					// logger.Println("inner loop")
					select {
					case <- t.quitchan:
						return
					case pv := <- t.paramchannel:
						switch pv.Index {
						case 0:
							t.freq = pv.Value
							inc = calcInc(len(t.buf), t.freq, float32(op.SampleRate))
						case 1:
							t.add = pv.Value
						case 2:
							t.amp = pv.Value
						}
					default:
					}
					ob[i] = t.buf[pos]
					pos = (pos + inc) % len(t.buf)
				}
				select {
				case <- t.quitchan:
					return
				case t.outchans[0] <- ob:
					// logger.Println("hi")
				}
				// logger.Println("bye")
			}
		}()
	})
	return
}
			
func (t *TableOsc) Stop() (err error) {
	t.stop.Do(func() {
		t.quitchan <- q{}
	})
	return
}
			
			
			
			
			
			
			
			
			
			
			
			