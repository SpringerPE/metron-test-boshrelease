package main

import (
	"time"
	"fmt"
	"sync"
	"runtime"
)


type Dispatcher struct {
	interval int
	maxWorkers int
	workerInstances []*Emiter
	wg *sync.WaitGroup
	running bool
}


func NewDispatcher(maxWorkers int, ms int) *Dispatcher {
	var wg sync.WaitGroup
	d := &Dispatcher{
		interval: ms,
		maxWorkers: maxWorkers,
		workerInstances: make([]*Emiter, maxWorkers),
		wg: &wg,
		running: false,
	}
	return d
}


func (d *Dispatcher) Run(appid string, msg string) bool {
	if !d.running {
		for i := 0; i < d.maxWorkers; i++ {
			worker := NewEmiter(d.interval, i, d.wg)
			d.workerInstances[i] = worker
			worker.Start(appid, msg)
		}
		d.running = true
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


// This does not use atomic operations
// Run after Stop()!!!!!!!
func (d *Dispatcher) Print() bool {
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
	fmt.Printf("  Logs sent = %d\n", totalCounter)
	fmt.Printf("  Errors = %d\n", totalErrors)
	fmt.Printf("  Start time = %s\n", startTime.Format(time.RFC1123))
	fmt.Printf("  End time = %s\n", endTime.Format(time.RFC1123))
	elapsed := endTime.Sub(*startTime)
	fmt.Printf("  Elapsed seconds = %f s\n", elapsed.Seconds())
	rate := float64(totalCounter)/elapsed.Seconds()
	fmt.Printf("  Rate = %f\n", rate)
	theoricalRate := (elapsed.Seconds()/(float64(d.interval)/1000000.0))*float64(d.maxWorkers)
	fmt.Printf("* STATS NumCPUs Workers Interval(us) TheoreticalRate(logs/s) LogsSent Errors Duration Rate(logs/s)\n")
	fmt.Printf("--STATS %d %d %d %f %d %d %f %f\n", runtime.NumCPU(), d.maxWorkers, d.interval, theoricalRate, totalCounter, totalErrors, elapsed.Seconds(), rate)
	return true
}

