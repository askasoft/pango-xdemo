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
