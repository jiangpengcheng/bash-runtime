package runner

import (
	"bash-runtime/common"
	"bytes"
	"context"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/sirupsen/logrus"
	"os/exec"
	"time"
)

func execScript(file string, param string) (*bytes.Buffer, *bytes.Buffer, error)  {
	if _, err := exec.LookPath(file); err != nil {
		return nil, nil, common.ErrScriptNotExist
	}
	var outb, errb bytes.Buffer
	cmd := exec.Command(file, param)
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		return nil, &errb, common.ErrScriptExecError
	}
	return &outb, &errb, nil
}

func Run(consumer pulsar.Consumer, producer pulsar.Producer, scriptFile string, retry uint, delay time.Duration) {
	for {
		msg, err := consumer.Receive(context.Background())
		if err != nil {
			logrus.Errorf("consumer is closed or context is done")
			break
		}
		consumer.AckID(msg.ID())

		stdout, stderr, err := execScript(scriptFile, string(msg.Payload()))
		if err != nil {
			logrus.Errorf("failed to process message: %s", err)
			continue
		}

		if stderr.Len() > 0 {
			logrus.Errorf("error: %s", stderr.String())
		}
		logrus.Infof("process message '%s' succeefully", msg.Payload())

		// retry sending message
		err = common.Retry(func() error {
			_, err = producer.Send(context.Background(), &pulsar.ProducerMessage{
				Payload: stdout.Bytes(),
			})
			if err != nil {
				return err
			}
			return nil
		}, common.RetryConfig{Attempts: retry, Delay: delay})

		if err != nil {
			logrus.Errorf("failed to send message to topic: %s, skip", err)
		}
	}
}
