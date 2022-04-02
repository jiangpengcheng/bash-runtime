package runner

import (
	"bash-runtime/common"
	"bytes"
	"context"
	"errors"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Runner struct {
	pulsarWriter *common.PulsarWriter
	client pulsar.Client
	consumer pulsar.Consumer
	producer pulsar.Producer
	logger *logrus.Logger
	running bool
}

func NewRunner(pulsarUrl string, logTopic string, inputTopics string, subscription string, outputTopic string) (*Runner, error) {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: pulsarUrl,
	})
	if err != nil {
		logrus.Errorf("Faild to connect pulsar, %s", err)
		return nil, err
	}

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: outputTopic,
	})
	if err != nil {
		logrus.Errorf("Faild to create producer, %s", err)
		return nil, err
	}

	topics := strings.Split(inputTopics, ",")
	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topics:            topics,
		SubscriptionName: subscription,
		Type:             pulsar.Shared, // make parallel processing available
	})
	if err != nil {
		logrus.Errorf("Faild to create consumer, %s", err)
		return nil, err
	}

	var pulsarWriter *common.PulsarWriter
	logger := logrus.StandardLogger()
	if logTopic != "" {
		pulsarWriter, err = common.NewPulsarWriter(logTopic, client)
		if err != nil {
			logrus.Errorf("Faild to create log producer, %s", err)
			return nil, err
		}
		logger = logrus.New()
		logger.SetOutput(io.MultiWriter(os.Stdout, pulsarWriter))
	}

	return &Runner{
		pulsarWriter: pulsarWriter,
		producer: producer,
		consumer: consumer,
		client: client,
		logger: logger,
	}, nil
}

func (runner *Runner) Run(scriptFile string) error  {
	// do not allow running in parallel using a same instance
	if runner.running {
		return errors.New("runner is already running")
	}
	runner.running = true
	for {
		msg, err := runner.consumer.Receive(context.Background())
		if err != nil {
			runner.logger.Errorf("consumer is closed or context is done")
			break
		}
		runner.consumer.AckID(msg.ID())

		stdout, stderr, err := execScript(scriptFile, string(msg.Payload()))
		if err != nil {
			runner.logger.Errorf("failed to process message: %s", err)
			continue
		}

		if len(stderr) > 0 {
			runner.logger.Errorf("error: %s", stderr)
		}
		runner.logger.Infof("process message '%s' successfully", msg.Payload())

		// retry sending message
		err = common.Retry(func() error {
			_, err = runner.producer.Send(context.Background(), &pulsar.ProducerMessage{
				Payload: stdout,
			})
			if err != nil {
				return err
			}
			return nil
		}, common.RetryConfig{Attempts: 3, Delay: 100 * time.Millisecond})

		if err != nil {
			runner.logger.Errorf("failed to send message to topic: %s, skip", err)
		}
	}
	runner.running = false
	return nil
}

func (runner *Runner) Close() {
	if runner == nil {
		return
	}
	runner.pulsarWriter.Close()
	runner.consumer.Close()
	runner.producer.Close()
	runner.client.Close()
}

func execScript(file string, param string) ([]byte, []byte, error)  {
	if _, err := exec.LookPath(file); err != nil {
		return nil, nil, common.ErrScriptNotExist
	}
	var outb, errb bytes.Buffer
	cmd := exec.Command(file, param)
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	outBytes := bytes.TrimRight(outb.Bytes(), "\n")
	errBytes := bytes.TrimRight(errb.Bytes(), "\n")
	if err != nil {
		return nil, errBytes, common.ErrScriptExecError
	}
	return outBytes, errBytes, nil
}
