package ugen

type q struct{}

type Mixer struct {
	UGenBase
	gains []float32
	running bool
	quitchan chan q
}

func init() {
	var _ UGen = NewMixer(1)
}

func NewMixer(inchan int) *Mixer {
	var m Mixer
	m.outchans = append(m.outchans, make(chan float32))
	m.gains = make([]float32, inchan)
	m.inputs = make([]UGen, inchan)
	m.quitchan = make(chan q)
	return &m
}

func (m *Mixer) Stop() error {
	m.quitchan <- q{}
	for _, u := range m.inputs {
		u.Stop()
	}
	return nil
}

func (m *Mixer) Start() error {
	for _, u := range m.inputs {
		u.Start()
	}
	go func() {
		for {
			select {
			case <- m.quitchan:
				return
			default:
			}
			
			var o float32
			for i, u := range m.inputs {
				for _, c := range u.OutputChannels() {
					select {
					case <- m.quitchan:
						return
					default:
					}
					o += m.gains[i]* (<- c)
				}
			}
			select {
			case <- m.quitchan:
				return
			default:
			}
			m.outchans[0] <- o
		}
	}()
	return nil
}

func (m *Mixer) SetGain(i int, g float32) error {
	if i < 0 || i >= len(m.gains) {
		return BadInputSet{i, len(m.gains)}
	}
	m.gains[i] = g
	return nil
}