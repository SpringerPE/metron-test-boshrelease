package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
	//"log"
	//"code.cloudfoundry.org/loggregator/plumbing/conversion"
)

type Dispatcher struct {
	// A pool of workers channels that are registered with the dispatcher
	maxWorkers   int
	aliveSeconds time.Duration
	// A buffered channel that we can send work requests on.
	jobQueue        chan Job
	filterOrigin    string
	workerInstances []*Worker
	wg              *sync.WaitGroup
	running         bool
	metronEvents    uint64
}

// NewDispatcher creates, and returns a new Dispatcher object.
func NewDispatcher(jobQueue chan Job, maxWorkers int) *Dispatcher {
	var wg sync.WaitGroup
	d := Dispatcher{
		jobQueue:        jobQueue,
		maxWorkers:      maxWorkers,
		workerInstances: make([]*Worker, maxWorkers),
		wg:              &wg,
		running:         false,
		metronEvents:    0,
	}
	return &d
}

func (d *Dispatcher) Run(filterOrigin string, aliveSeconds time.Duration) {
	if !d.running {
		d.aliveSeconds = aliveSeconds
		for i := 0; i < d.maxWorkers; i++ {
			worker := NewWorker(d.aliveSeconds, i, d.jobQueue, d.wg)
			d.workerInstances[i] = worker
			worker.Start()
			d.wg.Add(1)
		}
		d.running = true
	}
	d.wg.Wait()
	d.running = false
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

// This does not use atomic operations
// Run it after Stop()!!!!!!!
func (d *Dispatcher) Print(diodesNumber int, dopplerIngress uint64) bool {
	var startTime, endTime *time.Time
	var totalCounter, totalErrors uint64 = 0, 0

	metricCounters := make(map[string]uint64)

	if d.running {
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

		for metric, value := range worker.MetricsCounters {

			_, ok := metricCounters[metric]

			if !ok {
				metricCounters[metric] = value
			} else if worker.MetricsCounters[metric] > value {
				metricCounters[metric] = value
			}
		}

		totalCounter = totalCounter + worker.Counter()
		totalErrors = totalErrors + worker.Errors()
		worker.Print()
	}

	fmt.Printf("* Totals:\n")
	for metric, value := range metricCounters {
		fmt.Printf("  %s = %d\n", metric, value)
	}
	fmt.Printf("  Logs received = %d\n", totalCounter)
	fmt.Printf("  Errors = %d\n", totalErrors)
	fmt.Printf("  Start time = %s\n", startTime.Format(time.RFC1123))
	fmt.Printf("  End time = %s\n", endTime.Format(time.RFC1123))
	elapsed := endTime.Sub(*startTime)
	fmt.Printf("  Elapsed seconds = %f s\n", elapsed.Seconds())
	rate := float64(totalCounter) / elapsed.Seconds()
	fmt.Printf("  Rate = %f\n", rate)
	fmt.Printf("* STATS Date Time NumCPUs diodes Workers LogsReceived Errors Duration Rate(logs/s) logSenderTotalMessagesRead grpc.sendErrorCount dropsondeUnmarshaller.logMessageTotal dopplerIngress\n")
	fmt.Printf("--STATS %s %s %d %d %d %d %d %f %f %d %d %d %d\n", startTime.Format("01-02-2006"), startTime.Format("15:04:05"), runtime.NumCPU(), diodesNumber, d.maxWorkers, totalCounter, totalErrors, elapsed.Seconds(), rate, metricCounters["logSenderTotalMessagesRead"], metricCounters["grpc.sendErrorCount"], metricCounters["dropsondeUnmarshaller.logMessageTotal"], dopplerIngress)
	return true
}
