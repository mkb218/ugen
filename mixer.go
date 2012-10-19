package ugen

import "sync"

type q struct{}

// A Mixer mixes all channels of all input UGens into a single set of output channels, with per-ugen gains. A UGen which does not output a given channel will put nothing on that channel.
type Mixer struct {
	UGenBase
	gains    []float32
	quitchan chan q
}

func (m *Mixer) GetParams() []ParamDesc {
	return []ParamDesc{ParamDesc{Name:"gain",Size:8}}
}

func init() {
	var _ UGen = NewMixer(1,1)
}

func NewMixer(inchan, buswidth int) *Mixer {
	var m Mixer
	m.outchans = make([]chan []float32, buswidth)
	for i := range m.outchans {
		m.outchans[i] = make(chan []float32)
	}

	m.gains = make([]float32, inchan)
	m.paramchannel = make(chan ParamValue, inchan*buswidth)
	m.inputs = make([]UGen, inchan)
	m.quitchan = make(chan q)
	return &m
}

func (m *Mixer) Stop() error {
	logger.Println("MIXER STOP")
	m.quitchan <- q{}
	logger.Println("MIXER SENT QUIT")
	for _, u := range m.inputs {
		u.Stop()
	}
	return nil
}

type bufchanpair struct {
	a []float32
	c int
}

func (m *Mixer) Start(op OutputParams) error {
	for _, u := range m.inputs {
		u.Start(op)
	}
	go func() {
		obs := make([]bufchanpair,0)
		ocs := make([][]float32,len(m.outchans))
		for {
			select {
			case <-m.quitchan:
				return
			default:
			}

			// grab all new gain settings
		GAIN:
			for {
				select {
				case gp := <- m.paramchannel:
					// check is done in SetGain
					// if gp.channel > len(m.inputs)-1 {
					// go logger.Printf("Gain supplied for nonexistent input\n")
					// }
					m.gains[gp.Index] = gp.Value
				default:
					break GAIN
				}
			}

			var wg sync.WaitGroup
			for cnum := range m.outchans {
				for i, iu := range m.inputs {
					wg.Add(1)
					go func(iu UGen, i, cnum int) {
						if len(iu.OutputChannels())-1 < cnum {
							return
						}
						ob := <-iu.OutputChannels()[cnum]
						
						for j := range ob {
							ob[j] *= m.gains[i]
						}
						obs = append(obs, bufchanpair{a:ob,c:cnum})
						wg.Done()
					}(iu, i, cnum)
				}
			}
			wg.Wait()
			for _, bpc := range obs {
				if ocs[bpc.c] == nil {
					ocs[bpc.c] = bpc.a
				} else {
					for i := range ocs[bpc.c] {
						ocs[bpc.c][i] += bpc.a[i]
					}
					go func() { RecycleBuf(bpc.a,op) }()
				}
			}
			
			for i, c := range m.outchans {
				select {
				case <- m.quitchan:
					return
				case c <- ocs[i]: 
					ocs[i] = nil
				}
			}
		}
	}()
	return nil
}

func (m *Mixer) SetGain(i int, g float32) error {
	if i < 0 || i >= len(m.gains) {
		return BadInputSet{i, len(m.gains)}
	}
	go func() { m.paramchannel <- ParamValue{Value: g, Index: i} }()
	return nil
}
