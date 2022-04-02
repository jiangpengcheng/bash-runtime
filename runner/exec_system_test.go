// +build system

package runner

import (
	"bash-runtime/common"
	"context"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func TestNewRunner(t *testing.T) {
	pulsarUrl := common.GetEnv("PULSAR_URL", "pulsar://localhost:6650")
	type args struct {
		pulsarUrl    string
		logTopic     string
		inputTopics  string
		subscription string
		outputTopic  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "it should create the runner successfully",
			args: args{
				pulsarUrl:    pulsarUrl,
				logTopic:     "system-test-new-log1",
				inputTopics:  "system-test-new-input1",
				subscription: "system-test-new-sub1",
				outputTopic:  "system-test-new-output1",
			},
			wantErr: false,
		},
		{
			name: "it should create the runner successfully when log topic is not specified",
			args: args{
				pulsarUrl:    pulsarUrl,
				logTopic:     "",
				inputTopics:  "system-test-new-input2",
				subscription: "system-test-new-sub2",
				outputTopic:  "system-test-new-output2",
			},
			wantErr: false,
		},
		{
			name: "it should create the runner successfully when specified multiple input topics",
			args: args{
				pulsarUrl:    pulsarUrl,
				logTopic:     "system-test-new-log3",
				inputTopics:  "system-test-new-input3-1,system-test-new-input3-2,system-test-new-input3-3",
				subscription: "system-test-new-sub3",
				outputTopic:  "system-test-new-output3",
			},
			wantErr: false,
		},
		{
			name: "it should failed to create runner when output topic is not specified",
			args: args{
				pulsarUrl:    pulsarUrl,
				logTopic:     "system-test-new-log4",
				inputTopics:  "system-test-new-input4",
				subscription: "system-test-new-sub4",
				outputTopic:  "",
			},
			wantErr: true,
		},
		{
			name: "it should failed to create runner when input topic is not specified",
			args: args{
				pulsarUrl:    pulsarUrl,
				logTopic:     "system-test-new-log5",
				inputTopics:  "",
				subscription: "system-test-new-sub5",
				outputTopic:  "system-test-new-output5",
			},
			wantErr: true,
		},
		{
			name: "it should failed to create runner when subscription is not specified",
			args: args{
				pulsarUrl:    pulsarUrl,
				logTopic:     "system-test-new-log6",
				inputTopics:  "system-test-new-input6",
				subscription: "",
				outputTopic:  "system-test-new-output6",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRunner(tt.args.pulsarUrl, tt.args.logTopic, tt.args.inputTopics, tt.args.subscription, tt.args.outputTopic)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRunner() error = %v, wantErr %v", err, tt.wantErr)
			}
			got.Close()
		})
	}
}

func TestRunner_Run(t *testing.T) {
	ctx := context.TODO()
	pulsarUrl := common.GetEnv("PULSAR_URL", "pulsar://localhost:6650")
	type args struct {
		pulsarUrl    string
		logTopic     string
		inputTopics  string
		subscription string
		outputTopic  string
	}
	tests := []struct {
		name     string
		args     args
		script   string
		messages []string
	}{
		{
			name: "it should process the script file successfully",
			args: args{
				pulsarUrl:    pulsarUrl,
				logTopic:     "",
				inputTopics:  "system-test-runner-input1",
				subscription: "system-test-runner-sub1",
				outputTopic:  "system-test-runner-output1",
			},
			script:   "../scripts/exec.sh",
			messages: []string{"hello", "world", "hello world"},
		},
		{
			name: "it should process the script file successfully with multiple input topics",
			args: args{
				pulsarUrl:    pulsarUrl,
				logTopic:     "",
				inputTopics:  "system-test-runner-input2-1,system-test-runner-input2-2,system-test-runner-input2-3",
				subscription: "system-test-runner-sub2",
				outputTopic:  "system-test-runner-output2",
			},
			script:   "../scripts/exec.sh",
			messages: []string{"hello", "world", "hello world"},
		},
		{
			name: "it should print log to the log topic if specified",
			args: args{
				pulsarUrl:    pulsarUrl,
				logTopic:     "system-test-runner-log3",
				inputTopics:  "system-test-runner-input3",
				subscription: "system-test-runner-sub3",
				outputTopic:  "system-test-runner-output3",
			},
			script:   "../scripts/stderr.sh",
			messages: []string{"hello", "world", "hello world"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scriptRunner, err := NewRunner(tt.args.pulsarUrl, tt.args.logTopic, tt.args.inputTopics, tt.args.subscription, tt.args.outputTopic)
			assert.Equal(t, err, nil)
			go func() {
				scriptRunner.Run(tt.script)
			}()

			time.Sleep(1 * time.Second) // wait for scriptRunner.consumer to start retrieve message

			// producers for input topics
			producers := []pulsar.Producer{}
			topics := strings.Split(tt.args.inputTopics, ",")
			for _, topic := range topics {
				producer, err := scriptRunner.client.CreateProducer(pulsar.ProducerOptions{
					Topic: topic,
				})
				assert.Equal(t, err, nil)
				producers = append(producers, producer)
			}

			// consumer for output topic
			consumer, err := scriptRunner.client.Subscribe(pulsar.ConsumerOptions{
				Topic:            tt.args.outputTopic,
				SubscriptionName: "system-test-runner-output-consumer",
			})
			assert.Equal(t, err, nil)

			// create log topic consumer when log topic is specified
			var logConsumer pulsar.Consumer
			if tt.args.logTopic != "" {
				logConsumer, err = scriptRunner.client.Subscribe(pulsar.ConsumerOptions{
					Topic:            tt.args.logTopic,
					SubscriptionName: "system-test-runner-log-consumer",
				})
				assert.Equal(t, err, nil)
			}

			// send all messages to every input topic
			for _, producer := range producers {
				for _, message := range tt.messages {
					_, err := producer.Send(ctx, &pulsar.ProducerMessage{
						Payload: []byte(message),
					})
					assert.Equal(t, err, nil)

					// output topic should receive processed message
					msg, err := consumer.Receive(ctx)
					assert.Equal(t, []byte(message+"!"), msg.Payload())
					assert.Equal(t, err, nil)
					consumer.Ack(msg)

					if logConsumer != nil {
						log, err := logConsumer.Receive(ctx)
						assert.Equal(t, true, strings.Contains(string(log.Payload()), "data: command not found"))
						assert.Equal(t, err, nil)
						logConsumer.Ack(log)

						log2, err := logConsumer.Receive(ctx)
						assert.Equal(t, true, strings.Contains(string(log2.Payload()), "process message '"+message+"' successfully"))
						assert.Equal(t, err, nil)
						logConsumer.Ack(log2)
					}
				}
			}

			consumer.Close()
			if logConsumer != nil {
				logConsumer.Close()
			}
			for _, producer := range producers {
				producer.Close()
			}
			scriptRunner.Close()
		})
	}
}

func TestRunner_RunInParallel(t *testing.T) {
	ctx := context.TODO()
	pulsarUrl := common.GetEnv("PULSAR_URL", "pulsar://localhost:6650")
	messages := make([]int, 100)
	instances := 3

	type args struct {
		pulsarUrl    string
		logTopic     string
		inputTopic   string
		subscription string
		outputTopic  string
	}
	tests := []struct {
		name   string
		args   args
		script string
	}{
		{
			name: "it should process the script file successfully in parallel",
			args: args{
				pulsarUrl:    pulsarUrl,
				logTopic:     "",
				inputTopic:   "system-test-parallel-input1",
				subscription: "system-test-parallel-sub1",
				outputTopic:  "system-test-parallel-output1",
			},
			script: "../scripts/exec.sh",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runners := []*Runner{}
			for i := 0; i < instances; i++ {
				// send processed messages to different output topic, this is just for test purpose
				// normally we will keep all params same when user want to process messages parallel
				scriptRunner, err := NewRunner(tt.args.pulsarUrl, tt.args.logTopic, tt.args.inputTopic, tt.args.subscription, fmt.Sprintf("%s-%d", tt.args.outputTopic, i))
				assert.Equal(t, err, nil)
				go func() {
					_ = scriptRunner.Run(tt.script)
				}()
				runners = append(runners, scriptRunner)
			}

			time.Sleep(1 * time.Second) // wait for scriptRunner.consumer to start retrieve message

			// producer for input topic
			producer, err := runners[0].client.CreateProducer(pulsar.ProducerOptions{
				Topic: tt.args.inputTopic,
			})
			assert.Equal(t, err, nil)

			// consumer for output topics
			consumers := []pulsar.Consumer{}
			for i := 0; i < instances; i++ {
				consumer, err := runners[0].client.Subscribe(pulsar.ConsumerOptions{
					Topic:            fmt.Sprintf("%s-%d", tt.args.outputTopic, i),
					SubscriptionName: fmt.Sprintf("system-test-runner-output-consumer-%d", i),
				})
				assert.Equal(t, err, nil)
				consumers = append(consumers, consumer)
			}

			// send all messages to input topic
			for i := range messages {
				_, err := producer.Send(ctx, &pulsar.ProducerMessage{
					Payload: []byte(fmt.Sprintf("hello%d", i)),
				})
				assert.Equal(t, err, nil)

			}

			// each output topic should receive 33 messages
			for i := 0; i < 3; i++ {
				for j := 0; j < 33; j++ {
					msg, err := consumers[i].Receive(ctx)
					assert.Equal(t, err, nil)
					consumers[i].Ack(msg)
				}
			}

			for _, consumer := range consumers {
				consumer.Close()
			}
			producer.Close()
			for _, scriptRunner := range runners {
				scriptRunner.Close()
			}
		})
	}
}

func TestRunner_Close(t *testing.T) {
	ctx := context.TODO()
	pulsarUrl := common.GetEnv("PULSAR_URL", "pulsar://localhost:6650")
	type args struct {
		pulsarUrl    string
		logTopic     string
		inputTopic   string
		subscription string
		outputTopic  string
	}
	tests := []struct {
		name   string
		args   args
	}{
		{
			name: "it should close client/producers/consumers when close",
			args: args{
				pulsarUrl:    pulsarUrl,
				logTopic:     "system-test-close-log1",
				inputTopic:   "system-test-close-input1",
				subscription: "system-test-close-sub1",
				outputTopic:  "system-test-close-output1",
			},

		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, err := NewRunner(tt.args.pulsarUrl, tt.args.logTopic, tt.args.inputTopic, tt.args.subscription, tt.args.outputTopic)
			assert.Equal(t, err, nil)

			runner.Close()
			// failed to produce message
			_, err = runner.producer.Send(ctx, &pulsar.ProducerMessage{
				Payload: []byte("hello"),
			})
			assert.Equal(t, true, err != nil)

			// failed to retrieve message
			_, err = runner.consumer.Receive(ctx)
			assert.Equal(t, true, err != nil)

			// failed to produce message to log topic
			_, err = runner.pulsarWriter.Write([]byte("hello"))
			assert.Equal(t, true, err != nil)
		})
	}
}
