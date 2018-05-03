package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	dropsonde "github.com/cloudfoundry/dropsonde"
)

var (
	destination = flag.String("destination", "127.0.0.1:3457", "Logging destination (metron_agent listen address)")
	origin      = flag.String("origin", "metron-throughput", "Dropsonde origin envelope field")
	az          = flag.String("az", "zz", "Dropsonde AZ envelope field")
	runTime     = flag.Int("runtime", 60, "Running time in seconds")
	interval    = flag.Int("interval", 1000, "Microseconds to sleep between two (consecutive) log events, with runtime and threads, it defines the rate. Default is 1ms")
	threads     = flag.Int("threads", 1, "Number of worker threads (go subs)")
)

func main() {
	flag.Parse()

	fmt.Printf("*** Init emiter sending to %s with origin=%s/%s with %d threads and interval %d microseconds for %d seconds\n", *destination, *origin, *az, *threads, *interval, *runTime)
	err := dropsonde.Initialize(*destination, *origin, *az)
	if err != nil {
		log.Printf("ERROR initializing dropsonde: %s", err.Error())
		os.Exit(1)
	}
	d := NewDispatcher(*threads, *interval)
	d.Run("metron-test-emiter", "loooonnnnnng looooooog 000 111 222 333 444 555 666 777 888 999")
	time.Sleep(time.Second * time.Duration(*runTime))
	d.Stop()
	d.Print()
	fmt.Printf("*** End\n")
	os.Exit(0)
}
