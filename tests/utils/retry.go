package utils

import (
	"time"

	retry "github.com/avast/retry-go"
)

func WithRetry(utils *TestUtils, f func(utils *TestUtils) error) error {
	// TODO: think about using k8s.io/client-go/util/retry instead
	return retry.Do(
		func() error {
			return f(utils)
		},
		retry.Delay(5*time.Second),
		retry.DelayType(retry.FixedDelay),
		retry.Attempts(100),
		retry.Context(utils.Ctx),
		retry.LastErrorOnly(true),
	)
}
