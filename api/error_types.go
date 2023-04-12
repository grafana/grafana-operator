package api

import (
	"fmt"
	"time"
)

type BackoffError struct {
	Attempts  int
	RetryTime time.Time
}

func (e *BackoffError) Error() string {
	return fmt.Sprintf("error backof attempt %d, retrying after %s", e.Attempts, e.RetryTime)
}

func (e *BackoffError) RetryDelay() time.Duration {
	return time.Until(e.RetryTime)
}
