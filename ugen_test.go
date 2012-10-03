package ugen

import "math"
import "fmt"
import "testing"

func TestSin(t *testing.T) {
	var u Universe
	u.SampleRate = 1000
	s := NewSin(1, 0, &u)
	s.Start()
	for i := 0; i < 1000; i++ {
		var y float32
		
		y = <- s.OutputChannels()[0]
		if y != float32(math.Sin(float64(i)/1000)) {
			t.Log(fmt.Sprintf("bad output from sin, expected %f got %f", float32(math.Sin(1000*float64(i))), y))
			t.FailNow()
		}
	}
	s.Stop()
}

func TestMixer(t *testing.T) {
	var u Universe
	u.SampleRate = 1000
	s0 := NewSin(1, 0, &u)
	s1 := NewSin(1, math.Pi, &u)
	m := NewMixer(2)
	m.SetInput(s0, 0)
	m.SetInput(s1, 1)
	m.SetGain(0, 1)
	m.SetGain(1, 0.5)
	m.Start()
	for i := 0; i < 1000; i++ {
		var y float32
		
		y = <- m.OutputChannels()[0]
		if math.Abs(float64(y - float32(0.5*math.Sin(float64(i)/1000)))) > 0.00001 {
			t.Log(fmt.Sprintf("bad output from sin, expected %f got %f", float32(math.Sin(1000*float64(i))), y))
			t.FailNow()
		}
	}
	m.Stop()
}