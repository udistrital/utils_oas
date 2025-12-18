package optimize

import (
	"fmt"
)

var WorkQueue = make(chan WorkRequest)
var WorkerQueue chan chan WorkRequest

type WorkRequest struct {
	JobParameter []interface{}
	Job          func(...interface{}) interface{}
}

func NewWorker(id int, workerQueue chan chan WorkRequest) Worker {
	// Create, and return the worker.
	worker := Worker{
		ID:          id,
		Work:        make(chan WorkRequest),
		WorkerQueue: workerQueue,
		QuitChan:    make(chan bool)}

	return worker
}

type Worker struct {
	ID          int
	Work        chan WorkRequest
	WorkerQueue chan chan WorkRequest
	QuitChan    chan bool
}

// This function "starts" the worker by starting a goroutine, that is
// an infinite "for-select" loop.
func (w *Worker) Start() {
	go func() {
		defer func() {
			// recover from panic if one occured. Set err to nil otherwise.
			if recover() != nil {
				fmt.Println("Stoped Work... ")
				return

			}
		}()
		for {
			// Add ourselves into the worker queue.
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
				work.Job(work.JobParameter...)
			case <-w.QuitChan:
				// We have been asked to stop.
				return
			}
		}
	}()
}

// Stop tells the worker to stop listening for work requests.
//
// Note that the worker will only stop *after* it has finished its work.
func (w *Worker) Stop() {
	go func() {
		w.QuitChan <- true
	}()
}

func StartDispatcher(nworkers int, limitW int) {
	// First, initialize the channel we are going to but the workers' work channels into.
	// WorkerQueue = make(chan chan WorkRequest, nworkers)
	WorkerQueue = make(chan chan WorkRequest, nworkers)
	if limitW > 0 {
		WorkQueue = make(chan WorkRequest, limitW)

	}

	// Now, create all of our workers.
	for i := 0; i < nworkers; i++ {
		worker := NewWorker(i+1, WorkerQueue)
		worker.Start()
	}

	go func() {
		// fmt.Println(WorkerQueue)
		for {
			select {
			case work := <-WorkQueue:

				go func() {

					worker := <-WorkerQueue

					worker <- work
				}()

			}
		}
	}()
}
