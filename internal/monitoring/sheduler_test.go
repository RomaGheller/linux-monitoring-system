package monitoring

import (
	"context"
	"testing"
	"time"
)

func TestRunSchedule(t *testing.T) {
	t.Run("calls task on tick", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ticks := make(chan time.Time)
		called := make(chan struct{})

		task := func() {
			close(called)
		}

		go runSchedule(ctx, task, ticks)

		ticks <- time.Now()

		select {
		case <-called:
		case <-time.After(time.Second):
			t.Fatal("task was not called")
		}
	})

	t.Run("calls task for every tick", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ticks := make(chan time.Time)
		calls := make(chan struct{}, 3)

		task := func() {
			calls <- struct{}{}
		}

		go runSchedule(ctx, task, ticks)

		for range 3 {
			ticks <- time.Now()
		}

		for range 3 {
			select {
			case <-calls:
			case <-time.After(time.Second):
				t.Fatal("task was not called expected number of times")
			}
		}
	})

	t.Run("stops when context is cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		ticks := make(chan time.Time)
		done := make(chan struct{})

		task := func() {
			t.Fatal("task should not be called after cancellation")
		}

		go func() {
			runSchedule(ctx, task, ticks)
			close(done)
		}()

		cancel()

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("scheduler did not stop after context cancellation")
		}
	})
}
