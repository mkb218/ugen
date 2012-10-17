package ugen

import "sync"
import "fmt"

type ParamSource interface {
	SetUGenSource(s int, u UGen, cindex int)
	Sources() int
	SetDest(d int, u UGen, param int)
	Dests() int
	Start(OutputParams) error
	Stop() error
}

type AudioParam struct {
	src UGen
	cindex int
	dest chan<- ParamValue
	param int
	start, stop *sync.Once
	quitchan chan q
}

func (a *AudioParam) SetUGenSource(s int, u UGen, cindex int) {
	if s != 0 {
		return
	}
	a.src = u
	a.cindex = cindex
}

func (a *AudioParam) Sources() int {
	return 1
}

func (a *AudioParam) SetDest(d int, u UGen, index int) {
	if d != 0 {
		return
	}
	a.dest = u.ParamChannel()
	a.param = index
}

func (a *AudioParam) Dests() int {
	return 1
}

func NewAudioParam() *AudioParam {
	var a AudioParam
	a.start = new(sync.Once)
	a.stop = new(sync.Once)
	a.stop.Do(func(){})
	a.quitchan = make(chan q)
	return &a
}

type RequiredChannelWasNil struct {
	Location string
}

func (r RequiredChannelWasNil) Error() string {
	return fmt.Sprintf("%s: src or dest channel was nil", r.Location)
}

func (a *AudioParam) Start(op OutputParams) (err error) {
	a.start.Do(func() {
		if a.src == nil || a.dest == nil {
			err = &RequiredChannelWasNil{"AudioParam.Start()"}
			a.start = new(sync.Once)
			return
		}
		a.stop = new(sync.Once)
		a.src.Start(op)
		go func() {
			t := TickerFor(op)
			defer t.Stop()
			for {
				var inbuf []float32
				select {
				case inbuf = <- a.src.OutputChannels()[a.cindex]:
				case <-a.quitchan:
					return
				}
				select {
				case <- t.C:
				case <- a.quitchan:
					return
				}
				select {
				case a.dest <- ParamValue{Index:a.param, Value:inbuf[0]}:
					go RecycleBuf(inbuf, op)
				case <- a.quitchan:
					return
				}
			}
		}()
	})
	return err
}

func (a *AudioParam) Stop() (err error) {
	a.stop.Do(func() {
		a.quitchan <- q{}
		a.start = new(sync.Once)
	})
	return nil
}

func init() {
	var _ ParamSource = NewAudioParam()
}