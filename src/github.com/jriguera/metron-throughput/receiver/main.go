package main

import (
	"os"
	"flag"
	"fmt"

	"github.com/cloudfoundry/sonde-go/events"
)


var (
	hostport = flag.String("hostport", "0.0.0.0:8082", "host port to use to listen for gRPC v1 messages")
	certFile = flag.String("cert", "", "cert to use to listen for gRPC v1 messages")
	keyFile = flag.String("key", "", "key to use to listen for gRPC v1 messages")
	caFile = flag.String("ca", "", "ca cert to use to listen for gRPC v1 messages")
	workers = flag.Int("workers", 10, "The number of workers to start")
	maxQueueSize = flag.Int("max_queue_size", 1000, "The size of job queue")
	diodesNumber = flag.Int("diodes", 1000, "Diodes counter")
	waitTime = flag.Int("wait_time", 30, "seconds to wait after receiving the last log valid(origin) Envelope")
	origin = flag.String("origin", "metron-throughput/zz", "Origin/AZ Envelope fields")
)



// Job represents the job to be run
type Job struct {
	Payload *events.Envelope
	Name string
}


func main() {
	flag.Parse()

	fmt.Printf("*** Init listening on %s with origin=%s with %d threads, %d queue size and keep alive time of %ds\n", *hostport, *origin, *workers, *maxQueueSize, *waitTime)
	// Create the job queue.
	jobQueue := make(chan Job, *maxQueueSize)
	dispatcher := NewDispatcher(jobQueue, *workers)
	dispatcher.Run(*origin, *waitTime)

	fmt.Printf("* Starting Doppler Router with %d diodes\n", *diodesNumber)
	doppler := NewDoppler(*diodesNumber, *certFile, *keyFile, *caFile)
	doppler.Start(*hostport)
	doppler.Run(*origin, jobQueue)

	dispatcher.WaitStop()
	fmt.Printf("*** Done! Showing reports ...\n")
	doppler.Print()
	dispatcher.Print(*diodesNumber)
	fmt.Printf("*** End\n")
	os.Exit(0)
}


