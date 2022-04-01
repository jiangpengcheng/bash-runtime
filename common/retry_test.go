package common

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	counter := 0
	countFunc := func() error {
		counter++
		return nil
	}
	errorFunc := func() error {
		counter++
		return errors.New(fmt.Sprintf("error after %d calls", counter))
	}
	type args struct {
		retryableFunc RetryableFunc
		config        RetryConfig
	}
	tests := []struct {
		name    string
		args    args
		expectedErr error
		callCount int
	}{
		{
			name: "it should call func only once if there is no problem",
			args: args{
				retryableFunc: countFunc,
				config: RetryConfig{
					Attempts: 3,
					Delay: 100 * time.Millisecond,
				},
			},
			expectedErr: nil,
			callCount: 1,
		},
		{
			name: "it should retry to call function and return last call's error if failed to exec given function",
			args: args{
				retryableFunc: errorFunc,
				config: RetryConfig{
					Attempts: 3,
					Delay: 100 * time.Millisecond,
				},
			},
			expectedErr: errors.New("error after 3 calls"),
			callCount: 3,
		},
		{
			name: "it should use the default retry config when provided retry is not valid",
			args: args{
				retryableFunc: errorFunc,
				config: RetryConfig{
					Attempts: 0,
					Delay: 100 * time.Millisecond,
				},
			},
			expectedErr: errors.New("error after 3 calls"),
			callCount: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Retry(tt.args.retryableFunc, tt.args.config)
			assert.Equal(t, err, tt.expectedErr)
			assert.Equal(t, counter, tt.callCount)
			counter = 0 // reset counter
		})
	}
}
