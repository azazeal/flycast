// Package loop implements a persevering func runner.
package loop

import (
	"context"
	"time"

	"github.com/azazeal/pause"
)

// Func loops executions of fn, with ctx as its param, for as long as the ctx is
// not done pausing for p after each execution.
func Func(ctx context.Context, p time.Duration, fn func(context.Context)) {
	for ctx.Err() == nil {
		fn(ctx)

		pause.For(ctx, p)
	}
}
