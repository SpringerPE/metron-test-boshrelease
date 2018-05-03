package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/cloudfoundry/sonde-go/events"
)

var (
	hostport     = flag.String("hostport", "0.0.0.0:8082", "host port to use to listen for gRPC v1 messages")
	certFile     = flag.String("cert", "", "cert to use to listen for gRPC v1 messages")
	keyFile      = flag.String("key", "", "key to use to listen for gRPC v1 messages")
	caFile       = flag.String("ca", "", "ca cert to use to listen for gRPC v1 messages")
	workers      = flag.Int("workers", 10, "The number of workers to start")
	maxQueueSize = flag.Int("max_queue_size", 524288, "The size of job queue")
	diodesNumber = flag.Int("diodes", 524288, "Diodes counter")
	buffversion  = flag.Int("buffversion", 2, "1: v1Buff, 2: v2Buff")
	waitTime     = flag.Duration("wait_time", 30*time.Second, "seconds to wait after receiving the last log valid(origin) Envelope")
	origin       = flag.String("origin", "metron-throughput/zz", "Origin/AZ Envelope fields")
)

// Job represents the job to be run
type Job struct {
	PayloadV1 *events.Envelope
	PayloadV2 *loggregator_v2.Envelope
	Version   int
}

func main() {
	flag.Parse()

	fmt.Printf("*** Init listening on %s with origin=%s with %d threads, %d queue size, buffversion %s and keep alive time of %s\n", *hostport, *origin, *workers, *maxQueueSize, *buffversion, *waitTime)
	// Create the job queue.
	jobQueue := make(chan Job, *maxQueueSize)
	dispatcher := NewDispatcher(jobQueue, *workers)

	fmt.Printf("* Starting Doppler Router with %d diodes\n", *diodesNumber)
	doppler := NewDoppler(*diodesNumber, *certFile, *keyFile, *caFile)
	doppler.Start(*hostport)
	doppler.Run(*buffversion, *origin, jobQueue)

	dispatcher.Run(*origin, *waitTime)

	fmt.Printf("*** Done! Showing reports ...\n")
	doppler.Print()
	dispatcher.Print(*diodesNumber, doppler.GetCounter())
	fmt.Printf("*** End\n")
	os.Exit(0)
}
