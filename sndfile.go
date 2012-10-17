package ugen

import "github.com/mkb218/gosndfile/sndfile"
import "sync"
import "fmt"

type SndfileOut struct {
	UGenBase
	name string
	fi sndfile.Info
	file sndfile.File
	start, stop *sync.Once
	appendMode bool
	quitchan chan q
}

func NewSndfileOut(name string, fi sndfile.Info, appendMode bool) (s *SndfileOut, err error) {
	s = new(SndfileOut)
	s.inputs = make([]UGen, inputs)
	s.fi = fi
	s.name = name
	s.quitchan = make(chan q)
	s.start = new(sync.Once)
	s.stop = new(sync.Once)
	s.stop.Do(func(){})
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

	start.Do(func() {
		s.fi.Samplerate = int32(op.SampleRate)
		m := sndfile.Read
		if s.appendMode {
			m = sndfile.ReadWrite
		}
		bufsize := 0
		ochans := 0
		for _, u := range s.inputs {
			ochans += len(u.OutputChannels)
			bufsize += (op.BufferSize * len(u.OutputChannels))
		}
		logger.Print(s.name,ochans,"output channels")
		s.fi.Channels = ochans
		s.file, err = sndfile.Open(s.name, m, &s.fi)
		if err != nil {
			s.start = new(sync.Once)
			return
		}
		
		if s.appendMode {
			err = s.file.Seek(0, sndfile.End)
			if err != nil {
				s.start = new(sync.Once)
				return
			}
		}
		
		obuf := make([]float32, bufsize)
		for _, u := range s.inputs {
			u.Start()
		}
		
		s.stop = new(sync.Once)
		
		go func() {
			defer func() {
				for _, u := range s.inputs {
					u.Stop()
				}
			}()

			for {
				channum := 0
				for _, u := range s.inputs {
					for _, c := range u.OutputChannels() {
						select {
						case ibuf := <- c:
							for snum, sval := range ibuf {
								obuf[snum*ochans+channum] = sval
							}
						case <- s.quitchan:
							return
						}
					}
				}
			
				var n int64
				n, err = s.WriteItems(obuf)
				if n != len(obuf) || err != nil {
					logger.Print(s.name, "couldn't write from buf length", len(obuf), "wrote", n, "err", err)
					return
				}
			}			
		}()
	})
	
	return
}

