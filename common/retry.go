package common

import "time"

type RetryableFunc func() error

type RetryConfig struct {
	Attempts uint
	Delay    time.Duration
}

func Retry(retryableFunc RetryableFunc, config RetryConfig) error {
	var n uint
	var lastErr error

	// default option
	if config.Attempts < 1 {
		config.Attempts = 3
	}
	if config.Delay < 100 * time.Millisecond {
		config.Delay = 100 * time.Millisecond
	}

	for n < config.Attempts {
		lastErr = retryableFunc()
		n++
		if lastErr != nil {
			time.Sleep(config.Delay)
		} else {
			return nil
		}
	}
	return lastErr
}
