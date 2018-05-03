package main_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/jriguera/metron-throughput/receiver"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
)

var _ = Describe("Dispatcher", func() {

	It("Can process multiple v1 log messages", func() {

		jobQueue := make(chan Job, 10)
		go func() {
			for i := 0; i < 10; i++ {
				envelope, _ := buildV1LogMessage()

				job := Job{
					PayloadV1: envelope,
					PayloadV2: nil,
					Version:   1,
				}

				jobQueue <- job
				time.Sleep(10 * time.Millisecond)
			}
		}()
		dispatcher := NewDispatcher(jobQueue, 10)
		dispatcher.Run("metron-throughput/zz", 10*time.Second)

		dispatcher.Print(1, 100)
	})

	It("Can process multiple v2 log messages", func() {

		jobQueue := make(chan Job, 10)
		go func() {
			for i := 0; i < 10; i++ {
				envelope, _ := buildV2LogMessage()

				job := Job{
					PayloadV1: nil,
					PayloadV2: envelope,
					Version:   2,
				}

				jobQueue <- job
				time.Sleep(10 * time.Millisecond)
			}
		}()
		dispatcher := NewDispatcher(jobQueue, 10)
		dispatcher.Run("metron-throughput/zz", 10*time.Second)

		dispatcher.Print(1, 100)
	})
})

func buildV1LogMessage() (*events.Envelope, []byte) {
	envelope := &events.Envelope{
		Origin:    proto.String("metron-throughput/zz"),
		EventType: events.Envelope_LogMessage.Enum(),
		Timestamp: proto.Int64(time.Now().UnixNano()),
		LogMessage: &events.LogMessage{
			Message:     []byte("some-log-message"),
			MessageType: events.LogMessage_OUT.Enum(),
			Timestamp:   proto.Int64(time.Now().UnixNano()),
		},
	}
	data, err := proto.Marshal(envelope)
	Expect(err).ToNot(HaveOccurred())
	return envelope, data
}

func buildV2LogMessage() (*loggregator_v2.Envelope, []byte) {
	envelope := &loggregator_v2.Envelope{
		Timestamp: *(proto.Int64(time.Now().UnixNano())),
		SourceId:  *proto.String("metron-throughput/zz"),
		Message: &loggregator_v2.Envelope_Log{
			Log: &loggregator_v2.Log{
				Payload: []byte("some-log-message"),
				Type:    loggregator_v2.Log_OUT,
			},
		},
	}
	data, err := proto.Marshal(envelope)
	Expect(err).ToNot(HaveOccurred())
	return envelope, data
}
