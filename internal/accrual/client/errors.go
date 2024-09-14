package client

import (
	"errors"
	"fmt"
	"time"
)

type ErrUnexpectedStatus struct {
	Status int
}

func (err ErrUnexpectedStatus) Error() string {
	return fmt.Sprintf("unexpected status code: %d", err.Status)
}

func newErrUnexpectedStatus(status int) ErrUnexpectedStatus {
	return ErrUnexpectedStatus{
		Status: status,
	}
}

type ErrTooManyRequests struct {
	RetryAfterTime time.Time
}

func (err ErrTooManyRequests) Error() string {
	return "too many requests"
}

func newErrTooManyRequests(retryAfter time.Duration) ErrTooManyRequests {
	return ErrTooManyRequests{
		RetryAfterTime: time.Now().Add(retryAfter),
	}
}

var ErrOrderNotFound = errors.New("order not found")
var ErrInternalServerError = errors.New("internal server error")
