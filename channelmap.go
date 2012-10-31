package ugen

// A Spreader takes a single UGen and expands its channels.
// Simplest case is a single channel UGen expanded to multiple channels. Each output buffer is identical.
// Less simple is mindless expansion of a two channel UGen to five. 0 1 -> 0 1 0 1 0. Mindless.
type Spreader struct {
	UGenBase
	quitchan chan q
}

func init() {
	var _ UGen = NewSpreader(1)
}

func NewSpreader(out int) *Spreader {
	var s Spreader
	prepareoutchans(&(s.UGenBase), out)
	s.inputs = make([]UGen, 1)
	s.quitchan = make(chan q)
	return &s
}

func NewSpreaderWithUGen(out int, u UGen) *Spreader {
	s := NewSpreader(out)
	s.SetInput(0, u)
	return s
}


func (s *Spreader) Stop() error {
	s.quitchan <- q{}
	for _, u := range s.inputs {
		u.Stop()
	}
	return nil
}

func (s *Spreader) GetParams() []ParamDesc {
	return []ParamDesc{}
}

func (s *Spreader) Start(op OutputParams) error {
	for _, u := range s.inputs {
		u.Start(op)
	}
	go func() {
		for {
			// fetch all input buffers
			bufs := make([][]float32, len(s.inputs[0].OutputChannels()))
			for i := range bufs {
				bufs[i] = GetNewBuf(op)
				select {
				case <-s.quitchan:
					return
				case bufs[i] = <- s.inputs[0].OutputChannels()[i]:
				}
			}

			func() {
				first := true
				var inum int
				for _, oc := range s.outchans {
					// logger.Println("i", inum, oc)
					if first {
						// logger.Println("zing")
						defer func(i int, oc chan []float32) {
							// logger.Println("send first")
							oc <- bufs[i]
						}(inum, oc)
					} else {
						// logger.Println("getting newbuf")
						nb := GetNewBuf(op)
						// logger.Println("got newbuf")
						copy(nb, bufs[inum])
						select {
						case <-s.quitchan:
							return
						case oc <- nb:
							// logger.Println("send n",inum)
						}
					}

					inum++
					if inum == len(s.inputs[0].OutputChannels()) {
						first = false
						inum = 0
					}
				}
			}()
		}
	}()
	return nil
}
