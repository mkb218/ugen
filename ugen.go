package ugen

import "sync"
import "fmt"
import "log"
import "os"
import "runtime"
import "sync/atomic"

var logger *log.Logger

var ReserveBufferMax = 1024*1024*512

type ParamDesc struct {
	Name string
	Size int
}

type ParamValue struct {
	Index int
	Value float32
}

type OutputParams struct {
	SampleRate float64
	BufferSize int
}

type UGen interface {
	Inputs() []UGen
	ParamChannel() chan<- ParamValue
	GetParams() []ParamDesc
	SetInput(int, UGen) error
	OutputChannels() []chan []float32
	Start(OutputParams) error
	Stop() error
}

var RecycleStats rs

type rs struct {
	Alloced uint64
	Lost uint64
	Issued uint64
	Recycled uint64
}

func (r rs) String() string {
	return fmt.Sprintf("a: %d l: %d i: %d r %d", r.Alloced, r.Lost, r.Issued, r.Recycled)
}

type UGenBase struct {
	inputs       []UGen
	paramchannel chan ParamValue
	outchans     []chan []float32
}
// 
// func (u *UGenBase) GetParams() []ParamDesc {
// 	return []ParamDesc{}
// }

func (u *UGenBase) GetChannelIndexForParam(_ string) (int, int) {
	return 0, 0
}

func (u *UGenBase) ParamChannel() chan<- ParamValue {
	return u.paramchannel
}

func (u *UGenBase) OutputChannels() []chan []float32 {
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

func (u *UGenBase) SetInput(i int, g UGen) error {
	if i < 0 || i >= len(u.inputs) {
		return BadInputSet{i, len(u.inputs)}
	}

	u.inputs[i] = g
	return nil
}

var recyclers struct {
	sync.RWMutex
	m map[int]chan []float32
}

func init() {
	recyclers.m = make(map[int]chan []float32)
	dn, err := os.Create(os.DevNull)
	if err != nil {
		panic(err)
	}
	logger = log.New(dn, "ugen: ", log.LstdFlags|log.Lshortfile)
	logger = log.New(os.Stderr, "ugen: ", log.LstdFlags|log.Lshortfile)
//	runtime.GOMAXPROCS(runtime.NumCPU())
}

func prepareoutchans(u *UGenBase, c int) {
	u.outchans = make([]chan []float32, c)
	for i := range u.outchans {
		u.outchans[i] = make(chan []float32)
	}
}

func MakeRecycleChannel(op OutputParams) {
	recyclers.Lock() // possible that some other writer got lock between runlock and here
	defer recyclers.Unlock()
	if _, ok := recyclers.m[op.BufferSize]; !ok {
		recyclers.m[op.BufferSize] = make(chan []float32, ReserveBufferMax / op.BufferSize)
	}
}

// GetNewBuf either returns a recycled buffer or newly allocated buffer with the desired buffer size.
// This should never block.
func GetNewBuf(op OutputParams) (b []float32) {
	recyclers.RLock()
	defer recyclers.RUnlock()
	// pc, filename, line, ok := runtime.Caller(1)
	// var logstr string
	// if ok {
		// f := runtime.FuncForPC(pc)
		// logstr = fmt.Sprintf("%s:%d:%s", filename, line, f.Name())
	// } else {
		// logstr = "???"
	// }
	if c, ok := recyclers.m[op.BufferSize]; ok {
		select {
		case b = <-c:
			// logger.Println("recycled a buf for ",logstr)
		default:
			// logger.Println("alloced new for", logstr)
			atomic.AddUint64(&RecycleStats.Alloced, 1)
			b = make([]float32, op.BufferSize)
		}
	} else {
		// logger.Println("alloced new for", logstr)
		atomic.AddUint64(&RecycleStats.Alloced, 1)
		b = make([]float32, op.BufferSize)
	}
	atomic.AddUint64(&RecycleStats.Issued, 1)
	return
}

// RecycleBuf will put a used buffer into the recycling queue for the given BufferSize. It can block, so you should always call it from its own goroutine.
func RecycleBuf(b []float32, op OutputParams) {
	// pc, filename, line, ok := runtime.Caller(1)
	// var logstr string
	// if ok {
	// 	f := runtime.FuncForPC(pc)
	// 	logstr = fmt.Sprintf("%s:%d:%s", filename, line, f.Name())
	// } else {
	// 	logstr = "???"
	// }
	if len(b) == op.BufferSize && cap(b) == op.BufferSize {
		recyclers.RLock()
		var ok bool
		if _, ok = recyclers.m[op.BufferSize]; !ok {
			recyclers.RUnlock()
			MakeRecycleChannel(op)
			recyclers.RLock()
		} 
		select {
		case recyclers.m[op.BufferSize] <- b:
			// logger.Println("sending buffer to recycler for",logstr)
			atomic.AddUint64(&RecycleStats.Recycled, 1)
		default:
			// logger.Println("dropping buffer from",logstr)
			atomic.AddUint64(&RecycleStats.Lost, 1)
			// logger.Println("recycling channel full", len(recyclers.m[op.BufferSize]))
		}
		recyclers.RUnlock()
	} else {
		// logger.Println("no recycler for buffer from",logstr)
		atomic.AddUint64(&RecycleStats.Lost, 1)
	}
}

func LogRecycleStats() {
	recyclers.RLock()
	defer recyclers.RUnlock()
	for k, v := range recyclers.m {
		logger.Println("RecycleStats", k, len(v))
	}
	logger.Println("RecycleStats", RecycleStats)
//	var m runtime.MemStats
//	runtime.ReadMemStats(&m)
//	logger.Println("RecycleStats GC", m.GC)
}


func LogStackTrace() {
	var b [4096]byte
	c := runtime.Stack(b[:], true)
	log.Print(string(b[:c]))
}