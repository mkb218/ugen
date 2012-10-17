package ugen

import "time"

func TickerFor(op OutputParams) (t *time.Ticker) {
	d := float64(op.BufferSize) / op.SampleRate
	t = time.NewTicker(time.Duration(d * float64(time.Second)))
	return
}