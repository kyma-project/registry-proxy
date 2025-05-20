package utils

import (
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"time"
)

func WithRetry(utils *TestUtils, f func(utils *TestUtils) error) error {
	backoff := wait.Backoff{
		Duration: 5 * time.Second,
		Steps:    100,
		Cap:      5 * time.Second,
	}

	err := retry.OnError(
		backoff,
		func(err error) bool {
			return true
		},
		func() error {
			return f(utils)
		},
	)

	if err != nil {
		return err
	}
	return nil
}
