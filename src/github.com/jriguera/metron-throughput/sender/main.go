package main

import (
	"fmt"
	"time"
	"os"
	"sync"
	"flag"
	dropsonde "github.com/cloudfoundry/dropsonde"
	logs "github.com/cloudfoundry/dropsonde/logs"
)


var (
	destination = flag.String("destination", "127.0.0.1:3457", "Logging destination (metron listen address)")
	origin = flag.String("origin", "metron-throughput", "Origin Envelope")
	az = flag.String("az", "zz", "Origin Envelope")
	runtime = flag.Int("runtime", 60, "Running time in seconds")
	interval = flag.Int("ns", 1000000, "Interval time in nanoseconds. Default is 1ms")
)


type RateCounter struct {
	counter   int64
	start     time.Time
	end       time.Time
	interval  time.Duration
	terminate chan bool
	errors    int64
	wg        sync.WaitGroup
	appid     string
}


func Say(cnt *RateCounter) {
	ticker := time.NewTicker(cnt.interval)
	cnt.start = time.Now()
	for {
		select {
			case <-ticker.C:
				err := logs.SendAppLog(cnt.appid, fmt.Sprintf("%d", cnt.counter), "goooooo", "0")
				if err != nil {
					fmt.Printf("* Error in gosub sending log:%d \n", cnt.counter)
					cnt.errors++
				}
				cnt.counter++
			case <-cnt.terminate:
				ticker.Stop()
				cnt.end = time.Now()
				cnt.wg.Done()
				close(cnt.terminate)
				return
			}
		}
}


func Stop(cnt *RateCounter) {
	fmt.Println("* Sending termination to go sub ...")
	cnt.terminate <- true
	cnt.wg.Wait()
}


func Print(cnt *RateCounter) {
	fmt.Println("* Printing info ...")
	fmt.Printf("  Logs sent = %d (+1)\n", cnt.counter)
	fmt.Printf("  Errors = %d\n", cnt.errors)
	fmt.Printf("  Start time = %s\n", cnt.start.Format(time.RFC1123))
	fmt.Printf("  End time = %s\n", cnt.end.Format(time.RFC1123))
	elapsed := cnt.end.Sub(cnt.start)
	fmt.Printf("  Elapsed time = %f s\n", elapsed.Seconds())
	fmt.Printf("  Rate = %f\n", float64(cnt.counter)/elapsed.Seconds())
}


func Start(ns int, appid string) *RateCounter {
	cnt := &RateCounter{
		counter: 0,
		errors: 0,
		interval: time.Duration(ns)*time.Nanosecond,
		terminate: make(chan bool, 1),
		appid: appid,
	}
	cnt.wg.Add(1)
	fmt.Println("* Starting go sub ...")
	go Say(cnt)
	return cnt
}


func main() {
	flag.Parse()

	err := dropsonde.Initialize(*destination, *origin, *az)
	if err != nil {
		fmt.Println("* Error initializing: " + err.Error())
		os.Exit(1)
	}
	fmt.Println("*** Init, sending to " + *destination + " with origin " + *origin + "/" + *az)
	pid := os.Getpid()
	// SendAppLog sends a log message with the given appid, log message, source type
	// and source instance, with a message type of std out.
	// Returns an error if one occurs while sending the event.
	// SendAppLog(appID, message, sourceType, sourceInstance string) error {
	err = logs.SendAppLog("ratelogger", "About to start go thread", "main", fmt.Sprintf("%d", pid))
	if err != nil {
		fmt.Println("* Error sending first log: " + err.Error())
		os.Exit(1)
	}
	c := Start(*interval, "ratelogger")
	time.Sleep(time.Second * time.Duration(*runtime))
	Stop(c)
	Print(c)
	fmt.Printf("*** End\n")
	os.Exit(0)
}

