package main

import (
	"bash-runtime/common"
	"bash-runtime/runner"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
	"time"
)

func main() {
	pulsarUrl := common.GetEnv("PULSAR_URL", "pulsar://localhost:6650")
	outTopic := common.GetEnv("OUT_TOPIC", "bash-runtime-out")
	logTopic := common.GetEnv("LOG_TOPIC", "bash-runtime-log")
	inTopics := common.GetEnv("IN_TOPICS", "bash-runtime-in")
	subscription := common.GetEnv("SUBSCRIPTION", "bash-runtime-sub")

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: pulsarUrl,
	})
	if err != nil {
		logrus.Errorf("Faild to connect pulsar, %s", err)
		os.Exit(1)
	}

	if logTopic != "" {
		pulsarWriter, err := common.NewPulsarWriter(logTopic, client)
		if err != nil {
			logrus.Errorf("Faild to create log producer, %s", err)
			os.Exit(1)
		}
		logrus.SetOutput(io.MultiWriter(os.Stdout, pulsarWriter))
	}

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: outTopic,
	})
	if err != nil {
		logrus.Errorf("Faild to create producer, %s", err)
		os.Exit(1)
	}

	topics := strings.Split(inTopics, ",")
	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topics:            topics,
		SubscriptionName: subscription,
		Type:             pulsar.Shared,
	})
	if err != nil {
		logrus.Errorf("Faild to create consumer, %s", err)
		os.Exit(1)
	}

	runner.Run(consumer, producer, "./scripts/exec.sh", 5, 100 * time.Millisecond)

	defer client.Close()
	defer producer.Close()
	defer consumer.Close()
}