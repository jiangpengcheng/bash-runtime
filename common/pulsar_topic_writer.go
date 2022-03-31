package common

import (
	"context"
	"github.com/apache/pulsar-client-go/pulsar"
)

type PulsarWriter struct {
	producer pulsar.Producer
}

func NewPulsarWriter(topic string, client pulsar.Client) (*PulsarWriter, error) {
	logProducer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: topic,
	})
	if err != nil {
		return nil, err
	}
	return &PulsarWriter{
		producer: logProducer,
	}, nil
}

func (writer *PulsarWriter) Write(p []byte) (int, error)  {
	_, err := writer.producer.Send(context.TODO(), &pulsar.ProducerMessage{
		Payload: p,
	})
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
