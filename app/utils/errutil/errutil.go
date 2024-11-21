package errutil

import (
	"context"
)

func ContextError(ctx context.Context) error {
	if err := context.Cause(ctx); err != nil {
		return err
	}
	return ctx.Err()
}
