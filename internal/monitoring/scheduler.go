package monitoring

import (
	"context"
	"time"
)

type Scheduler func(
	ctx context.Context,
	task func(),
	interval time.Duration,
)

func Schedule(ctx context.Context, task func(), interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	runSchedule(ctx, task, ticker.C)
}

func runSchedule(ctx context.Context, task func(), ticks <-chan time.Time) {
	for {
		select {
		case <-ticks:
			task()

		case <-ctx.Done():
			return
		}
	}
}
