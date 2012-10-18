package ugen

import "github.com/mkb218/gosndfile/sndfile"
import "sync"

type SndfileOut struct {
	UGenBase
	name string
	fi sndfile.Info
	file *sndfile.File
	start, stop *sync.Once
	appendMode bool
	quitchan chan q
}

func NewSndfileOut(name string, fi sndfile.Info, inputugens int, appendMode bool) (s *SndfileOut) {
	s = new(SndfileOut)
	s.inputs = make([]UGen, inputugens)
	s.fi = fi
	s.name = name
	s.quitchan = make(chan q)
	s.start = new(sync.Once)
	s.stop = new(sync.Once)
	s.stop.Do(func(){})
	return
}

func init() {
	var _ UGen = new(SndfileOut)
}

func (s *SndfileOut) Start(op OutputParams) (err error) {
	defer func() {
		if err != nil {
			if s.file != nil {
				s.file.Close()
			}
		}
	}()

	s.start.Do(func() {
		s.fi.Samplerate = int32(op.SampleRate)
		m := sndfile.Write
		if s.appendMode {
			m = sndfile.ReadWrite
		}
		bufsize := 0
		ochans := 0
		for _, u := range s.inputs {
			ochans += len(u.OutputChannels())
			bufsize += (op.BufferSize * len(u.OutputChannels()))
		}
		logger.Println(s.name,ochans,"output channels")
		s.fi.Channels = int32(ochans)
		s.file, err = sndfile.Open(s.name, m, &s.fi)
		if err != nil {
			logger.Print(err)
			s.start = new(sync.Once)
			return
		}
		
		if s.appendMode {
			_, err = s.file.Seek(0, sndfile.End)
			if err != nil {
				s.start = new(sync.Once)
				return
			}
		}
		
		for _, u := range s.inputs {
			u.Start(op)
		}
		
		s.stop = new(sync.Once)

		obuf := make([]float32, bufsize)
		
		go func() {
			// create ticker
			t := TickerFor(op)
			defer func() {
				for _, u := range s.inputs {
					u.Stop()
				}
				// stop ticker
				t.Stop()
				s.stop.Do(func() {
					s.start = new(sync.Once)
				})
			}()

			for {
				// wait for ticker
				select {
				case <- t.C:
				case <- s.quitchan:
					return
				}
				
				channum := 0
				for _, u := range s.inputs {
					for _, c := range u.OutputChannels() {
						select {
						case ibuf := <- c:
							for snum, sval := range ibuf {
								// if sval != 0 {
								// 	logger.Println(obuf[snum*ochans+channum], sval)
								// }
								obuf[snum*ochans+channum] = sval
								// if sval != 0 {
								// 	logger.Println(obuf[snum*ochans+channum], sval)
								// }
							}
							go RecycleBuf(ibuf, op)
						case <- s.quitchan:
							return
						}
						channum++
					}
				}
			
				var n int64
				n, err = s.file.WriteItems(obuf)
				
				if n != int64(len(obuf)) || err != nil {
					logger.Print(s.name, "couldn't write from buf length", len(obuf), "wrote", n, "err", err)
					return
				}
			}			
		}()
	})
	
	return
}

func (s *SndfileOut) Stop() (err error) {
	s.stop.Do(func() {
		s.quitchan <- q{}
		s.start = new(sync.Once)
		err = s.file.Close()
	})
	return
}