package main

import (
	"fmt"
	"time"
	"runtime"
	"sync"
	//"log"
	//"code.cloudfoundry.org/loggregator/plumbing/conversion"
)


type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	workerPool chan chan Job
	maxWorkers int
	aliveSeconds int
	// A buffered channel that we can send work requests on.
	jobQueue chan Job
	filterOrigin string
	workerInstances []*Worker
	wg *sync.WaitGroup
	running bool
}


// NewDispatcher creates, and returns a new Dispatcher object.
func NewDispatcher(jobQueue chan Job, maxWorkers int) *Dispatcher {
	var wg sync.WaitGroup
	workerPool := make(chan chan Job, maxWorkers)
	d := Dispatcher{
		jobQueue: jobQueue,
		maxWorkers: maxWorkers,
		workerPool: workerPool,
		workerInstances: make([]*Worker, maxWorkers),
		wg: &wg,
		running: false,
	}
	return &d
}


func (d *Dispatcher) Run(filterOrigin string, aliveSeconds int) bool {
	if !d.running {
		d.aliveSeconds = aliveSeconds
		for i := 0; i < d.maxWorkers; i++ {
			worker := NewWorker(d.aliveSeconds, i, d.workerPool, d.wg)
			d.workerInstances[i] = worker
			worker.Start()
		}
		d.running = true
		go d.dispatch(filterOrigin)
	}
	return d.running
}


func (d *Dispatcher) Stop() bool {
	if d.running {
		for i := 0; i < d.maxWorkers; i++ {
			worker := d.workerInstances[i]
			worker.Stop()
		}
		d.wg.Wait()
		d.running = false
	}
	return !d.running
}


func (d *Dispatcher) dispatch(filterOrigin string) {
	for {
		select {
		case job := <-d.jobQueue:
			// a job request has been received
			//go func() {
				v := job.Version
				switch v {
						case 1:
							//fmt.Printf("> %s\n", job.PayloadV1)
							if (*job.PayloadV1.Origin == filterOrigin) {
								// try to obtain a worker job channel that is available.
								// this will block until a worker is idle
								//fmt.Printf("> %s\n", job.Payload.GetLogMessage())
								workerJobQueue := <-d.workerPool
								// dispatch the job to the worker job channel
								workerJobQueue <- job
							}
						case 2:
							//fmt.Printf("> %s\n", job.PayloadV2)
							if (job.PayloadV2.GetTags()["origin"] == filterOrigin) {
								// try to obtain a worker job channel that is available.
								// this will block until a worker is idle
								//fmt.Printf("> %s\n", job.Payload.GetLogMessage())
								workerJobQueue := <-d.workerPool
								// dispatch the job to the worker job channel
							workerJobQueue <- job
						}
					}
			//}
		}
	}
}


// Keeps waiting for all workers until they decide to stop
func (d *Dispatcher) WaitStop() {
	var doneCounter int = 0
	if d.running {
		for {
			for i := 0; i < d.maxWorkers; i++ {
				worker := d.workerInstances[i]
				if worker.IsDone() {
					doneCounter++
				}
			}
			if doneCounter >= d.maxWorkers {
				break
			}
			time.Sleep(1 * time.Second)
		}
		d.running = false
	}
}


// This does not use atomic operations
// Run it after Stop()!!!!!!!
func (d *Dispatcher) Print(diodesNumber int) bool {
	var startTime, endTime *time.Time
	var totalCounter, totalErrors uint64 = 0, 0
	if (d.running) {
		return false
	}
	fmt.Println("*** INFO:")
	for i := 0; i < d.maxWorkers; i++ {
		worker := d.workerInstances[i]
		if i == 0 {
			startTime = &worker.StartT
			endTime = &worker.EndT
		} else {
			if worker.EndT.After(*endTime) {
				endTime = &worker.EndT
			}
			if worker.StartT.Before(*startTime) {
				startTime = &worker.StartT
			}
		}
		totalCounter = totalCounter + worker.Counter()
		totalErrors = totalErrors + worker.Errors()
		worker.Print()
	}
	fmt.Printf("* Totals:\n")
	fmt.Printf("  Logs received = %d\n", totalCounter)
	fmt.Printf("  Errors = %d\n", totalErrors)
	fmt.Printf("  Start time = %s\n", startTime.Format(time.RFC1123))
	fmt.Printf("  End time = %s\n", endTime.Format(time.RFC1123))
	elapsed := endTime.Sub(*startTime)
	fmt.Printf("  Elapsed seconds = %f s\n", elapsed.Seconds())
	rate := float64(totalCounter)/elapsed.Seconds()
	fmt.Printf("  Rate = %f\n", rate)
	fmt.Printf("* STATS Date Time NumCPUs diodes Workers LogsReceived Errors Duration Rate(logs/s)\n")
	fmt.Printf("--STATS %s %s %d %d %d %d %d %f %f\n", startTime.Format("01-02-2006"), startTime.Format("15:04:05"), runtime.NumCPU(), diodesNumber, d.maxWorkers, totalCounter, totalErrors, elapsed.Seconds(), rate)
	return true
}

