package ugen

import "sync"
import "fmt"
import "code.google.com/p/portaudio-go/portaudio"

type Universe struct {
	SampleRate float64
	BufferSize int
	Output *Output
	Input *Input
}

func NewUniverse(in, out int) *Universe {
	var u Universe
	u.BufferSize = 4096
	u.SampleRate = 44100
	u.Output = NewOutput(out, &u)
	u.Input = NewInput(in, &u)
	return &u
}

func (u *Universe) Start() {
	u.Output.Start()
}

func (u *Universe) Stop() {
	u.Output.Stop()
}

type UGen interface {
	Inputs() []UGen
	SetInput(UGen, int) error
	OutputChannels() []chan float32
	Start() error
	Stop() error
}

type UGenBase struct {
	inputs []UGen
	outchans []chan float32
}

func (u *UGenBase) OutputChannels() []chan float32 {
	return u.outchans
}

func (u *UGenBase) Inputs() []UGen {
	return u.inputs
}

type BadInputSet struct {
	index, length int
}

func (b BadInputSet) Error() string {
	return fmt.Sprintf("Attempted to set bad input %d on inputs of length %d", b.index, b.length)
}

func (u *UGenBase) SetInput(g UGen, i int) error {
	if i < 0 || i >= len(u.inputs) {
		return BadInputSet{i, len(u.inputs)}
	}
	
	u.inputs[i] = g
	return nil
}

type Output struct {
	UGenBase
	universe *Universe
	start, stop *sync.Once
	stream *portaudio.Stream
}

func NewOutput(inchan int, u *Universe) *Output {
	var o Output
	o.inputs = make([]UGen, inchan)
	o.resetonces()
	o.universe = u
	return &o
}

func (o *Output) resetonces() {
	o.start = new(sync.Once)
	o.stop = new(sync.Once)
}

func (o *Output) Start() error {
	var err error
	o.start.Do(func () {
		if o.stream == nil {
			o.stream, err = portaudio.OpenDefaultStream(0, len(o.inputs), o.universe.SampleRate, o.universe.BufferSize, o)
			if err != nil {
				return
			}
		}
		for _, u := range o.inputs {
			u.Start()
		}
		o.stream.Start()
	})
	return err
}

func (o *Output) ProcessAudio(_, out [][]float32) {
	var wg sync.WaitGroup
	for i, b := range out {
		wg.Add(1)
		go func(i int, b []float32) {
			for j := range b {
				b[j] = <- o.inputs[i].OutputChannels()[i]
			}
			wg.Done()
		}(i, b)
	}
	wg.Wait()
}

func (o *Output) Stop() error {
	o.stop.Do(func () {
		o.stream.Stop()
		for _, u := range o.inputs {
			u.Stop()
		}
	})
	o.resetonces()
	return nil
}

func init() {
	var u = new (Universe)
	var _ UGen = NewOutput(1, u)
	var _ UGen = NewInput(1, u)
}

type Input struct {
	UGenBase
	universe *Universe
	start, stop *sync.Once
	stream *portaudio.Stream
}


func NewInput(outchan int, u *Universe) *Input {
	var i Input
	i.outchans = make([]chan float32, outchan)
	for j := range i.outchans {
		i.outchans[j] = make(chan float32)
	}
	i.resetonces()
	i.universe = u
	return &i
}

func (i *Input) resetonces() {
	i.start = new(sync.Once)
	i.stop = new(sync.Once)
}

func (i *Input) Start() error {
	var err error
	i.start.Do(func () {
		if i.stream == nil {
			i.stream, err = portaudio.OpenDefaultStream(len(i.outchans), 0, i.universe.SampleRate, i.universe.BufferSize, i)
			if err != nil {
				return
			}
		}
		i.stream.Start()
	})
	return err
}

func (inp *Input) ProcessAudio(in, _ [][]float32) {
	var wg sync.WaitGroup
	for i, b := range in {
		wg.Add(1)
		go func(i int, b []float32) {
			for j := range b {
				inp.outchans[i] <- b[j]
			}
			wg.Done()
		}(i, b)
	}
	wg.Wait()
}

func (i *Input) Stop() error {
	var err error
	i.stop.Do(func () {
		err = i.stream.Stop()
	})
	i.resetonces()
	return err
}

