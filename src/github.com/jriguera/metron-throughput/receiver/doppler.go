package main


import (
	"fmt"
	//"time"
	"log"
	"net"
	"sync"

	"code.cloudfoundry.org/loggregator/plumbing"
	"code.cloudfoundry.org/loggregator/diodes"
	//"code.cloudfoundry.org/loggregator/plumbing/conversion"
	"github.com/jriguera/metron-throughput/receiver/internal/server/v1"
	gendiodes "code.cloudfoundry.org/go-diodes"
	"code.cloudfoundry.org/loggregator/metricemitter"
	//"github.com/cloudfoundry/sonde-go/events"
	//"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)


type SpyHealthRegistrar struct {
	mu sync.Mutex
	values map[string]float64
}


type Doppler struct {
		v1Buf *diodes.ManyToOneEnvelope
		v2Buf *diodes.ManyToOneEnvelopeV2
		manager *v1.IngestorServer
		server *grpc.Server
		listener net.Listener
		healthRegistrar *SpyHealthRegistrar
		ingressMetric *metricemitter.Counter
		diodesCounter int
		terminate chan bool
}


func NewDoppler(diodesCounter int, certFile string, keyFile string, caFile string) *Doppler {
	tlsConfig, err := plumbing.NewServerMutualTLSConfig(certFile, keyFile, caFile)
	if err != nil {
		log.Fatal(err)
	}
	v1Buf := diodes.NewManyToOneEnvelope(diodesCounter,
		gendiodes.AlertFunc(func(missed int) {
			fmt.Printf("* Diodes Dropped %d envelopes (v1 buffer)\n", missed)
		}))
	v2Buf := diodes.NewManyToOneEnvelopeV2(diodesCounter,
		gendiodes.AlertFunc(func(missed int) {
			fmt.Printf("* Diodes Dropped %d envelopes (v2 buffer)\n", missed)
		}))
	// metric-documentation-v2: (loggregator.doppler.ingress) Number of received
	// envelopes from Metron on Doppler's v2 gRPC server
	ingressMetric := metricemitter.NewCounter("ingress", "doppler")
	healthRegistrar := newSpyHealthRegistrar()

	manager := v1.NewIngestorServer(v1Buf, v2Buf, ingressMetric, healthRegistrar)
	transportCreds := credentials.NewTLS(tlsConfig)
	server := grpc.NewServer(grpc.Creds(transportCreds))
	plumbing.RegisterDopplerIngestorServer(server, manager)

	return &Doppler{
		v1Buf: v1Buf,
		v2Buf: v2Buf,
		manager: manager,
		server: server,
		healthRegistrar: healthRegistrar,
		ingressMetric: ingressMetric,
		diodesCounter: diodesCounter,
		terminate: make(chan bool, 1),
	}
}


func (d *Doppler) Start(hostport string) {
	listener, err := net.Listen("tcp", hostport)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Starting gRPC server on %s\n", listener.Addr().String())
	go d.server.Serve(listener)
	d.listener = listener
}


func (d *Doppler) Run(v int, origin string, jobQueue chan Job) {
	// loop to read logs
	go func() {
		for {
			select {
				case <-d.terminate:
					close(d.terminate)
					return
				default:
					switch v {
						case 1:
							job := Job{Version: v, PayloadV1: d.v1Buf.Next()}
							jobQueue <- job
						case 2:
							job := Job{Version: v, PayloadV2: d.v2Buf.Next()}
							jobQueue <- job
					}
			}
		}
	}()
}


func (d *Doppler) Stop() {
	d.terminate <- true
	log.Printf("Stopping gRPC server on %s\n", d.listener.Addr().String())
	d.server.Stop()
	err := d.listener.Close()
	if err != nil {
		log.Fatal(err)
	}
}


func (d *Doppler) Print() {
	fmt.Println("* Printing Doppler library info ...")
	d.healthRegistrar.Print()
}


func newSpyHealthRegistrar() *SpyHealthRegistrar {
	return &SpyHealthRegistrar{
		values: make(map[string]float64),
	}
}

func (s *SpyHealthRegistrar) Inc(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[name]++
}

func (s *SpyHealthRegistrar) Dec(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[name]--
}

func (s *SpyHealthRegistrar) Get(name string) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.values[name]
}

func (s *SpyHealthRegistrar) Print() {
	fmt.Println("* Doppler SpyHealthRegistrar: ")
	for key, value := range s.values {
		fmt.Println("  ", key, ": ", value)
	}
}

