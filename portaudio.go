package ugen

import "code.google.com/p/portaudio-go/portaudio"
import "sync"

// A PortAudioOutput takes its input from a single UGen (likely a Mixer, Spreader, or Mapper)
// and sends it out on the the default portaudio device
type PortAudioOutput struct {
	UGenBase
	op OutputParams
	channels int
	start, stop *sync.Once
	stream *portaudio.Stream
}

func NewPortAudioOutput(channels int) *PortAudioOutput {
	var o PortAudioOutput
	o.inputs = make([]UGen, 1)
	o.channels = channels
	o.start = new(sync.Once)
	o.stop = new(sync.Once)
	o.stop.Do(func(){})
	return &o
}

func (o *PortAudioOutput) GetParams() []ParamDesc {
	return []ParamDesc{}
}

func (i *PortAudioInput) GetParams() []ParamDesc {
	return []ParamDesc{}
}

func (o *PortAudioOutput) Start(op OutputParams) error {
	var err error
	o.start.Do(func () {
		o.op = op
		MakeRecycleChannel(op)
		if o.stream == nil {
			o.stream, err = portaudio.OpenDefaultStream(0, o.channels, op.SampleRate, op.BufferSize, o)
			if err != nil {
				logger.Println("portaudio output creation failed", err)
				return
			}
		}
		err = o.stream.Start()
		if err != nil {
			logger.Println("portaudio output start failed", err)
			return
		}
		err = o.inputs[0].Start(op)
		if err != nil {
			logger.Println("input 0 start failed",err)
			o.stream.Stop()
			return
		}
		logger.Printf("portaudio Start %T %p", o.inputs[0], o.inputs[0])
		o.stop = new(sync.Once)
	})
	return err
}

func (o *PortAudioOutput) ProcessAudio(_, out [][]float32) {
	var wg sync.WaitGroup
	for i, b := range out {
		wg.Add(1)
		go func(i int, b []float32) {
			defer func() {
				if r := recover(); r != nil {
					logger.Println(i, len(o.inputs[i].OutputChannels()))
					panic(r)
				}
			}()
			ib := <- o.inputs[0].OutputChannels()[i]
//			logger.Println("pa test", ib[123])
			copy(b, ib)
			go func() { RecycleBuf(ib, o.op) }()
			wg.Done()
		}(i, b)
	}
	wg.Wait()
//	logger.Println("ProcessAudio exiting")
}

func (o *PortAudioOutput) Stop() error {
	o.stop.Do(func () {
		logger.Println("PAO STOP")
		o.stream.Stop()
		for i, u := range o.inputs {
			logger.Println("PAO CHILDREN STOP", i)
			u.Stop()
		}
		o.start = new(sync.Once)
	})
	return nil
}

func init() {
	var _ UGen = NewPortAudioOutput(1)
	var _ UGen = NewPortAudioInput(1)
}

type PortAudioInput struct {
	UGenBase
	channels int
	op OutputParams
	start, stop *sync.Once
	stream *portaudio.Stream
}

func NewPortAudioInput(outchan int) *PortAudioInput {
	var i PortAudioInput
	i.outchans = make([]chan []float32, outchan)
	for j := range i.outchans {
		i.outchans[j] = make(chan []float32)
	}
	i.start = new(sync.Once)
	return &i
}

func (i *PortAudioInput) Start(op OutputParams) error {
	var err error
	i.start.Do(func () {
		if i.stream == nil {
			i.stream, err = portaudio.OpenDefaultStream(len(i.outchans), 0, op.SampleRate, op.BufferSize, i)
			if err != nil {
				return
			}
		}
		i.stop = new(sync.Once)
		i.stream.Start()
	})
	return err
}

func (inp *PortAudioInput) ProcessAudio(in, _ [][]float32) {
	var wg sync.WaitGroup
	for i, b := range in {
		wg.Add(1)
		go func(i int, b []float32) {
			for _ = range b {
				ob := GetNewBuf(inp.op)
				copy(ob, b)
				inp.outchans[i] <- ob
			}
			wg.Done()
		}(i, b)
	}
	wg.Wait()
}

func (i *PortAudioInput) Stop() error {
	var err error
	i.stop.Do(func () {
		err = i.stream.Stop()
		i.start = new(sync.Once)
	})
	return err
}
