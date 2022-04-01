// +build system

package common

import (
	"context"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPulsarWriter_Write(t *testing.T) {
	pulsarUrl := GetEnv("PULSAR_URL", "pulsar://localhost:6650")
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: pulsarUrl,
	})
	defer client.Close()
	if err != nil {
		t.Errorf("failed to connect to pulsar: %s", err)
	}

	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		logTopic string
		args    args
	}{
		{
			name: "it should write bytes to pulsar topic",
			logTopic: "test-log-topic",
			args: args{
				p: []byte("hello world"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, err := NewPulsarWriter(tt.logTopic, client)
			assert.Equal(t, err, nil)

			consumer, err := client.Subscribe(pulsar.ConsumerOptions{
				Topic: tt.logTopic,
				SubscriptionName: "test-sub",
				Type:             pulsar.Exclusive,
			})
			assert.Equal(t, err, nil)

			got, err := writer.Write(tt.args.p)
			assert.Equal(t, err, nil)
			assert.Equal(t, got, len(tt.args.p))

			msg, err := consumer.Receive(context.TODO())
			assert.Equal(t, err, nil)
			assert.Equal(t, tt.args.p, msg.Payload())

			consumer.Ack(msg)
			consumer.Close()
			writer.Close()
		})
	}
}

func TestPulsarWriter_Close(t *testing.T) {
	pulsarUrl := GetEnv("PULSAR_URL", "pulsar://localhost:6650")
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: pulsarUrl,
	})
	defer client.Close()
	if err != nil {
		t.Errorf("failed to connect to pulsar: %s", err)
	}

	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		logTopic string
		args    args
	}{
		{
			name: "it should close the pulsar producer when close",
			logTopic: "test-log-topic",
			args: args{
				p: []byte("hello world"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, err := NewPulsarWriter(tt.logTopic, client)
			assert.Equal(t, err, nil)

			consumer, err := client.Subscribe(pulsar.ConsumerOptions{
				Topic: tt.logTopic,
				SubscriptionName: "test-sub",
				Type:             pulsar.Exclusive,
			})
			assert.Equal(t, err, nil)

			got, err := writer.Write(tt.args.p)
			assert.Equal(t, err, nil)
			assert.Equal(t, got, len(tt.args.p))

			msg, err := consumer.Receive(context.TODO())
			assert.Equal(t, err, nil)
			assert.Equal(t, tt.args.p, msg.Payload())
			consumer.Ack(msg)
			consumer.Close()

			writer.Close()
			got, err = writer.Write(tt.args.p)
			assert.Equal(t, got, 0)
			assert.Equal(t, err != nil, true)
		})
	}
}
