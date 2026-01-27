package netbox

import (
	"errors"
	"time"
)

// retryOnRetryable retries the given function on transient NetBox or network errors (HTTP 429/5xx) with simple backoff.
func retryOnRetryable(fn func() error) error {
	const maxAttempts = 6
	var err error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		if !isRetryableError(err) || attempt == maxAttempts {
			return err
		}

		// simple linear backoff
		time.Sleep(time.Duration(attempt) * time.Second)
	}

	return err
}

func isRetryableError(err error) bool {
	var coder interface {
		Code() int
	}
	if errors.As(err, &coder) {
		code := coder.Code()
		if code == 429 || (code >= 500 && code < 600) {
			return true
		}
	}

	// Fallback: check for temporary network errors
	var temporary interface {
		Temporary() bool
	}
	if errors.As(err, &temporary) && temporary.Temporary() {
		return true
	}

	return false
}
