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
		t.Log("samplenum test",i)
		
		y = <- s.OutputChannels()[0]//:
		if y != float32(math.Sin(float64(i)/1000)) {
			t.Log(fmt.Sprintf("bad output from sin, expected %f got %f", float32(math.Sin(1000*float64(i))), y))
			t.FailNow()
		}
	}
	s.Stop()
}