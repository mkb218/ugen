package main

import "time"
import "github.com/mkb218/ugen"

func main() {
	u := ugen.NewUniverse(2, 2)
	s0 := ugen.NewSin(1000, 0, u)
	s1 := ugen.NewSin(1004, 0, u)
	s2 := ugen.NewSin(1008, 0, u)
	s3 := ugen.NewSin(1016, 0, u)
	ml := ugen.NewMixer(2)
	mr := ugen.NewMixer(2)
	u.Output.SetInput(ml, 0)
	u.Output.SetInput(mr, 1)
	ml.SetInput(s0, 0)
	ml.SetGain(0, 0.5)
	ml.SetInput(s1, 1)
	ml.SetGain(1, 0.5)
	mr.SetInput(s2, 0)
	mr.SetGain(0, 0.5)
	mr.SetInput(s3, 1)
	mr.SetGain(1, 0.5)
	u.Start()
	time.Sleep(5 * time.Second)
	u.Stop()
}