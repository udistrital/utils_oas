package optimize

import (
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
func Digest(done <-chan interface{}, f func(interface{}, ...interface{}) interface{}, in <-chan interface{}, params []interface{}) (outchan <-chan interface{}) {
	out := make(chan interface{})
	var wg sync.WaitGroup
	const numDigesters = 1
	wg.Add(numDigesters)
	for i := 0; i < numDigesters; i++ {
		go func() {
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
