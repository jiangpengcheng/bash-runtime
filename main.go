package main

import (
	"bash-runtime/common"
	"bash-runtime/runner"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	pulsarUrl := common.GetEnv("PULSAR_URL", "pulsar://localhost:6650")
	outTopic := common.GetEnv("OUT_TOPIC", "bash-runtime-out")
	logTopic := common.GetEnv("LOG_TOPIC", "bash-runtime-log")
	inTopics := common.GetEnv("IN_TOPICS", "bash-runtime-in")
	subscription := common.GetEnv("SUBSCRIPTION", "bash-runtime-sub")
	script := common.GetEnv("SCRIPT", "./scripts/exec.sh")

	scriptRunner, err := runner.NewRunner(pulsarUrl, logTopic, inTopics, subscription, outTopic)
	if err != nil {
		logrus.Errorf("Failed to initialize script runner: %s", err)
		os.Exit(1)
	}
	defer scriptRunner.Close()
	_ = scriptRunner.Run(script)
}