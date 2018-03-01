package main

import (
	"fmt"
	"time"
	//"log"
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
}


// NewDispatcher creates, and returns a new Dispatcher object.
func NewDispatcher(filterOrigin string, jobQueue chan Job, maxWorkers int, aliveSeconds int) *Dispatcher {
	workerPool := make(chan chan Job, maxWorkers)
	return &Dispatcher{
		jobQueue: jobQueue,
		aliveSeconds: aliveSeconds,
		maxWorkers: maxWorkers,
		workerPool: workerPool,
		filterOrigin: filterOrigin,
		workerInstances: make([]*Worker, maxWorkers),
	}
}



func (d *Dispatcher) Run() {
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(d.aliveSeconds, i, d.workerPool)
		worker.Start()
		d.workerInstances[i] = worker
	}
	go d.dispatch()
}



func (d *Dispatcher) Stop() {
	for i := 0; i < d.maxWorkers; i++ {
		worker := d.workerInstances[i]
		worker.Stop()
	}
}


func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-d.jobQueue:
			// a job request has been received
			go func() {
				//fmt.Printf("> %s\n", job.Payload)
				if (*job.Payload.Origin == d.filterOrigin) {
					//log.Printf("Fetching workerJobQueue for: %s\n", job.Name)
					// try to obtain a worker job channel that is available.
					// this will block until a worker is idle
					//fmt.Printf("> %s\n", job.Payload.GetLogMessage())
					workerJobQueue := <-d.workerPool
					// dispatch the job to the worker job channel
					workerJobQueue <- job
				}
			}()
		}
	}
}


// Keeps waiting for all workers until they decide to stop
func (d *Dispatcher) WaitStop() {
	var doneCounter int = 0
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
}


// This does not use atomic operations
// Run it after Stop()!!!!!!!
func (d *Dispatcher) Print() {
	var startTime, endTime *time.Time
	var totalCounter, totalErrors uint64 = 0, 0
	fmt.Println("* Printing info ...")
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
	fmt.Println("* Totals:")
	fmt.Printf("  Logs processed = %d\n", totalCounter)
	fmt.Printf("  Errors = %d\n", totalErrors)
	fmt.Printf("  Start time = %s\n", startTime.Format(time.RFC1123))
	fmt.Printf("  End time = %s\n", endTime.Format(time.RFC1123))
	elapsed := endTime.Sub(*startTime)
	fmt.Printf("  Elapsed time = %f s\n", elapsed.Seconds())
	fmt.Printf("  Rate = %f\n", float64(totalCounter)/elapsed.Seconds())
}

