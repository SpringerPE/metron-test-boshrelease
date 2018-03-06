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
	wg *sync.WaitGroup
}


func NewWorker(waitSeconds int, id int, workerPool chan chan Job, wg *sync.WaitGroup) *Worker {
	return &Worker{
		id: id,
		WorkerPool: workerPool,
		jobQueue: make(chan Job),
		waitSeconds: (time.Duration(waitSeconds)*time.Second).Seconds(),
		sleepMs: 1,
		wg: wg,
	}
}


func (w *Worker) readLog(data *events.Envelope, first *int32) bool {
	if (data.GetEventType() == events.Envelope_LogMessage) {
		//atomic.AddUint64(w.workerCounter, 1)
		(*w.workerCounter)++
		// time for the last log
		w.EndT = time.Now()
		if *first == 0 {
			// time for first log (to calculate rate)
			w.StartT = w.EndT
			//atomic.AddInt32(first, 1)
			(*first)++
		}
		//tripTime := time.Since(time.Unix(0, data.GetTimestamp()))
		//log.Printf("t=%s : %s\n", tripTime, data.GetLogMessage().GetMessage())
		return true
	} else if (data.GetEventType() == events.Envelope_Error) {
		//atomic.AddUint64(w.workerErrors, 1)
		(*w.workerErrors)++
		log.Printf("ERROR: %s\n", data.GetError())
	}
	return false
}


// Start method starts the run loop for the worker, listening for a quit channel in
// case we need to stop it
func (w *Worker) Start() {
	var counter, errors uint64 = 0, 0
	var done int32 = 0
	w.workerCounter = &counter
	w.workerErrors = &errors
	w.done = &done
	w.Quit = make(chan bool)
	w.wg.Add(1)
	//log.Printf("Starting worker=%d\n", w.id)
	go func() {
		var first int32 = 0
		run := true
		for run {
			processLog := true
			for processLog {
				select {
					case job := <-w.jobQueue:
						// we have received a work request.
						w.readLog(job.Payload, &first)
					case terminate := <-w.Quit:
						run = ! terminate
					default:
						processLog = false
				}
			}
			select {
				// Add worker's jobQueue to main worker pool (keeps waiting if no select)
				case w.WorkerPool <- w.jobQueue:
					continue
				default:
					time.Sleep(w.sleepMs * time.Millisecond)
					if first > 0 {
						// Only disconnect after receive something
						// It will keep waiting forever until at least one log is processed!
						elapsed := time.Now().Sub(w.EndT)
						if (elapsed.Seconds() > w.waitSeconds) {
							log.Printf("Worker %d about to stop due to inactivity for %f s", w.id, elapsed.Seconds())
							atomic.AddInt32(w.done, 1)
							run = false
						}
					}
			}
		}
		w.wg.Done()
	}()
}


// Stop signals the worker to stop listening for work requests.
func (w *Worker) Stop() {
	//log.Printf("Stopping worker=%d\n", w.id)
	w.Quit <- true
	close(w.Quit)
}


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

