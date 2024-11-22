package errutil

import (
	"context"
	"errors"
)

func ContextError(ctx context.Context, errs ...error) error {
	for _, err := range errs {
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
	}

	if err := context.Cause(ctx); err != nil {
		return err
	}
	return ctx.Err()
}

type ClientError struct {
	Err error
}

var ErrClient = &ClientError{}

func NewClientError(err error) error {
	return &ClientError{Err: err}
}

func (ce *ClientError) Is(err error) (ok bool) {
	_, ok = err.(*ClientError)
	return
}

func (ce *ClientError) Error() string {
	return ce.Err.Error()
}

func (ce *ClientError) Unwrap() error {
	return ce.Err
}
