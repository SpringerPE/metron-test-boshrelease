package main


import (
	"time"
	"log"
	"fmt"
	"sync/atomic"
	"sync"

	"github.com/cloudfoundry/sonde-go/events"
)


// Worker represents the worker that executes the job
type Worker struct {
	id int
	workerCounter *uint64
	workerErrors *uint64
	WorkerPool chan chan Job
	jobQueue chan Job
	Quit chan bool
	StartT time.Time
	EndT time.Time
	sleepMs time.Duration
	waitSeconds float64
	done *int32
	mu sync.Mutex
}


func NewWorker(waitSeconds int, id int, workerPool chan chan Job) *Worker {
	var counter, errors uint64 = 0, 0
	var done int32 = 0
	return &Worker{
		id: id,
		workerErrors: &errors,
		workerCounter: &counter,
		WorkerPool: workerPool,
		jobQueue: make(chan Job),
		Quit: make(chan bool),
		waitSeconds: (time.Duration(waitSeconds)*time.Second).Seconds(),
		done: &done,
		sleepMs: 1,
	}
}



func (w *Worker) readOne(running *int32) bool {
	select {
		case job := <-w.jobQueue:
			// we have received a work request.
			data := job.Payload
			if (data.GetEventType() == events.Envelope_LogMessage) {
				w.mu.Lock()
				// time for the last log
				(*w.workerCounter)++
				w.EndT = time.Now()
				if *running == 0 {
					// time for first log (to calculate rate)
					w.StartT = w.EndT
					(*running)++
				}
				w.mu.Unlock()
				//tripTime := time.Since(time.Unix(0, data.GetTimestamp()))
				//log.Printf("t=%s : %s\n", tripTime, data.GetLogMessage().GetMessage())
			} else if (data.GetEventType() == events.Envelope_Error) {
				atomic.AddUint64(w.workerErrors, 1)
				log.Printf("ERROR: %s\n", data.GetError())
			}
			return true
		case <-w.Quit:
			// we have received a signal to stop
			log.Printf("Stopping reader in worker=%d\n", w.id)
			// send again
			w.Quit <- true
			return false
	}
}


// Start method starts the run loop for the worker, listening for a quit channel in
// case we need to stop it
func (w *Worker) Start() {
	go func() {
		var running int32 = 0
		for {
			select {
			// Add worker's jobQueue to main worker pool and keeps waiting
			case w.WorkerPool <- w.jobQueue:
				go w.readOne(&running)
			case <-w.Quit:
				// we have received a signal to stop
				log.Printf("Stopping worker=%d\n", w.id)
				close(w.Quit)
				return
			default:
				time.Sleep(w.sleepMs * time.Millisecond)
				if atomic.LoadInt32(&running) > 0 {
					// Only disconnect after receive something
					// It will keep waiting forever until at least one log is processed!
					w.mu.Lock()
					elapsed := time.Now().Sub(w.EndT)
					w.mu.Unlock()
					if (elapsed.Seconds() > w.waitSeconds) {
						log.Printf("Worker %d about to stop due to inactivity for %f s", w.id, elapsed.Seconds())
						atomic.AddInt32(w.done, 1)
						w.Stop()
					}
				}
			}
		}
	}()
}


// Stop signals the worker to stop listening for work requests.
func (w *Worker) Stop() {
	w.Quit <- true
}


// Stop signals the worker to stop listening for work requests.
func (w *Worker) IsDone() bool {
	done := atomic.LoadInt32(w.done)
	if (done > 0) {
		return true
	}
	return false
}


func (w *Worker) Counter() uint64 {
	return atomic.LoadUint64(w.workerCounter)
}


func (w *Worker) Errors() uint64 {
	return atomic.LoadUint64(w.workerErrors)
}


func (w *Worker) Print() {
	counter := atomic.LoadUint64(w.workerCounter)
	errors := atomic.LoadUint64(w.workerErrors)
	elapsed := w.EndT.Sub(w.StartT)
	fmt.Printf("* Operations done by worker %d: %d (%d errors) in %f s, rate=%f ops/s\n", w.id, counter, errors, elapsed.Seconds(), float64(counter)/elapsed.Seconds())
}

