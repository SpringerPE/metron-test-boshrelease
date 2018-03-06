package main

import (
	"time"
	"log"
	"fmt"
	"sync/atomic"
	"sync"

	logs "github.com/cloudfoundry/dropsonde/logs"
)


type Emiter struct {
	CounterLogs *uint64
	StartT time.Time
	EndT time.Time
	interval time.Duration
	Terminate chan bool
	ErrorLogs *uint64
	id int
	wg *sync.WaitGroup
}


func NewEmiter(ms int, id int, wg *sync.WaitGroup) *Emiter {
	e := &Emiter{
		interval: time.Duration(ms)*time.Microsecond,
		id: id,
		wg: wg,
	}
	return e
}


func (e *Emiter) Say(appid string, msg string) {
	e.wg.Add(1)
	go func() {
		ticker := time.NewTicker(e.interval)
		e.StartT = time.Now()
		run := true
		for run {
			counter := atomic.LoadUint64(e.CounterLogs)
			select {
				case <-ticker.C:
					// SendAppLog sends a log message with the given appid, log message, source type
					// and source instance, with a message type of std out.
					// Returns an error if one occurs while sending the event.
					// SendAppLog(appID, message, sourceType, sourceInstance string) error 
					err := logs.SendAppLog(appid, fmt.Sprintf("%d", counter), msg, fmt.Sprintf("%d", e.id))
					if err != nil {
						log.Printf("ERROR emiter=%d sending log %d: %s\n", e.id, counter, err.Error())
						atomic.AddUint64(e.ErrorLogs, 1)
					}
					atomic.AddUint64(e.CounterLogs, 1)
				case terminate := <-e.Terminate:
					if terminate {
						e.EndT = time.Now()
					}
					run = ! terminate
			}
		}
		ticker.Stop()
		e.wg.Done()
	}()
}


func (e *Emiter) Start(appid string, msg string) {
	var counter, errors uint64 = 0, 0
	e.CounterLogs = &counter
	e.ErrorLogs = &errors
	e.Terminate = make(chan bool)
	//log.Printf("Starting worker=%d\n", e.id)
	e.Say(appid, msg)
}


func (e *Emiter) Stop() {
	//log.Printf("Stopping worker=%d\n", e.id)
	e.Terminate <- true
	close(e.Terminate)
}


func (e *Emiter) Counter() uint64 {
	return atomic.LoadUint64(e.CounterLogs)
}


func (e *Emiter) Errors() uint64 {
	return atomic.LoadUint64(e.ErrorLogs)
}


func (e *Emiter) Print() {
	elapsed := e.EndT.Sub(e.StartT)
	counter := e.Counter()
	errors := e.Errors()
	fmt.Printf("* Logs sent by worker %d: %d (%d errors) in %fs, rate=%f logs/s\n", e.id, counter, errors, elapsed.Seconds(), float64(counter)/elapsed.Seconds())
}

