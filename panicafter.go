package ugen

import "time"

func panicAfter(t time.Duration) chan q {
	cancel := make(chan q)
	go func() {
		select {
		case <- time.After(t*time.Second):
			logger.Panic("timeout!")
		case <- cancel:
		}
	}()
	return cancel
}