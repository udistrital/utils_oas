package optimize

import (
	"runtime"
	"sync"
)

//funcion para generar canales de interface{}
func GenChanInterface(mp ...interface{}) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		for _, ch := range mp {
			out <- ch
		}
		close(out)
	}()
	return out
}

func digester(done <-chan interface{}, f func(interface{}, ...interface{}) interface{}, params []interface{}, in <-chan interface{}, out chan<- interface{}) {
	for intfc := range in {
		res := f(intfc, params...)
		select {
		case out <- res:
		case <-done:
			return
		}
	}
}

//funcion para administrar las go rutines armadas para la consulta de solicitudes de rp.
func Digest(done <-chan interface{}, f func(interface{}, ...interface{}) interface{}, in <-chan interface{}, params []interface{}, maxConcurrency ...int) (outchan <-chan interface{}) {
	out := make(chan interface{})
	var wg sync.WaitGroup
	numDigesters := runtime.NumCPU()
	if len(maxConcurrency) > 0 {
		numDigesters = maxConcurrency[0]

	}
	wg.Add(numDigesters)
	for i := 0; i < numDigesters; i++ {
		go func() {
			defer func() {
				// recover from panic if one occured. Set err to nil otherwise.
				if recover() != nil {
					wg.Done()

				}
			}()
			digester(done, f, params, in, out)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func ProccDigest(data []interface{}, f func(interface{}, ...interface{}) interface{}, params []interface{}, maxConcurrency ...int) (res []interface{}) {
	done := make(chan interface{})
	defer close(done)
	resch := GenChanInterface(data...)
	chres := Digest(done, f, resch, params, maxConcurrency...)
	for resdata := range chres {
		if resdata != nil {
			res = append(res, resdata)
		}

	}
	return
}
