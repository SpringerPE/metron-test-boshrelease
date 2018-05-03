package main

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/cloudfoundry/sonde-go/events"
)

// Worker represents the worker that executes the job
type Worker struct {
	id              int
	workerCounter   *uint64
	workerErrors    *uint64
	jobQueue        chan Job
	Quit            chan bool
	StartT          time.Time
	EndT            time.Time
	sleepMs         time.Duration
	waitSeconds     time.Duration
	MetricsCounters map[string]uint64
	wg              *sync.WaitGroup
}

func NewWorker(waitSeconds time.Duration, id int, jobQueue chan Job, wg *sync.WaitGroup) *Worker {
	return &Worker{
		id:              id,
		jobQueue:        jobQueue,
		waitSeconds:     waitSeconds,
		sleepMs:         1,
		MetricsCounters: make(map[string]uint64),
		wg:              wg,
	}
}

//https://github.com/cloudfoundry/loggregator-api/blob/master/README.md#v2---mapping-v1
func (w *Worker) readLog(data Job) bool {
	v := data.Version
	switch v {
	case 1:
		if data.PayloadV1.GetEventType() == events.Envelope_LogMessage {
			atomic.AddUint64(w.workerCounter, 1)
			//(*w.workerCounter)++
			// time for the last log
			w.EndT = time.Now()
			if *w.workerCounter == 1 {
				// time for first log (to calculate rate)
				w.StartT = w.EndT
			}
			//tripTime := time.Since(time.Unix(0, data.GetTimestamp()))
			//log.Printf("t=%s : %s\n", tripTime, data.GetLogMessage().GetMessage())
			return true
		}
		if data.PayloadV1.GetEventType() == events.Envelope_CounterEvent {

			event := (*events.CounterEvent)(data.PayloadV1.CounterEvent)

			_, ok := w.MetricsCounters[*event.Name]

			if !ok {
				w.MetricsCounters[*event.Name] = uint64(*event.Total)
			} else if w.MetricsCounters[*event.Name] < uint64(*event.Total) {
				w.MetricsCounters[*event.Name] = uint64(*event.Total)
			}

		}
		// anything else goes to a different type
		atomic.AddUint64(w.workerErrors, 1)
		//(*w.workerErrors)++
		return false
	case 2:
		switch (data.PayloadV2.Message).(type) {
		case *loggregator_v2.Envelope_Log:
			atomic.AddUint64(w.workerCounter, 1)
			//(*w.workerCounter)++
			// time for the last log
			w.EndT = time.Now()
			if *w.workerCounter == 1 {
				// time for first log (to calculate rate)
				w.StartT = w.EndT
			}
			//tripTime := time.Since(time.Unix(0, data.GetTimestamp()))
			//log.Printf("t=%s : %s\n", tripTime, data.GetLogMessage().GetMessage())
			return true
		case *loggregator_v2.Envelope_Counter:

			event := data.PayloadV2.GetCounter()

			_, ok := w.MetricsCounters[event.Name]

			if !ok {
				w.MetricsCounters[event.Name] = uint64(event.Total)
			} else if w.MetricsCounters[event.Name] < uint64(event.Total) {
				w.MetricsCounters[event.Name] = uint64(event.Total)
			}

			// anything else goes to a different type
			atomic.AddUint64(w.workerErrors, 1)
			//(*w.workerErrors)++
			return false
		default:
			// anything else is different type for us
			atomic.AddUint64(w.workerErrors, 1)
			//(*w.workerErrors)++
			return false
		}
	default:
		// just count
		// time for the last log
		atomic.AddUint64(w.workerCounter, 1)
		//(*w.workerCounter)++

		w.EndT = time.Now()
		if *w.workerCounter == 1 {
			// time for first log (to calculate rate)
			w.StartT = w.EndT
		}
	}
	return false
}

// Start method starts the run loop for the worker, listening for a quit channel in
// case we need to stop it
func (w *Worker) Start() {
	var counter, errors uint64 = 0, 0
	w.workerCounter = &counter
	w.workerErrors = &errors
	w.Quit = make(chan bool)
	//log.Printf("Starting worker=%d\n", w.id)
	//
	timer1 := time.NewTimer(w.waitSeconds)
	go func() {
		<-timer1.C
		log.Printf("Worker %d about to stop due to inactivity for %s", w.id, w.waitSeconds)
		w.Stop()
	}()

	go func() {
		run := true
		for run {

			select {
			case job := <-w.jobQueue:
				// we have received a work request.
				if w.readLog(job) {
					timer1.Reset(w.waitSeconds)
				}

			case terminate := <-w.Quit:
				run = !terminate
			}
			// Add worker's jobQueue to main worker pool (keeps waiting if no select)

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
