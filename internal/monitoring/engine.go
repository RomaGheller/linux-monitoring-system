package monitoring

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Engine struct {
	mu        sync.RWMutex
	running   bool
	checkers  []Checker
	launcher  CheckLauncher
	scheduler Scheduler
	interval  time.Duration
	sink      ResultSink
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

func NewEngine() *Engine {
	return &Engine{
		launcher:  SequentialLauncher{},
		scheduler: Schedule,
		interval:  time.Second * 10,
		sink:      StdoutResultSink{},
	}
}

func (e *Engine) SetLauncher(launcher CheckLauncher) error {
	if e.running {
		return errors.New("cannot change launcher while engine is running")
	}

	if launcher == nil {
		return errors.New("launcher cannot be nil")
	}

	e.launcher = launcher

	return nil
}

func (e *Engine) SetScheduler(scheduler Scheduler) error {
	if e.running {
		return errors.New("cannot change scheduler while engine is running")
	}

	if scheduler == nil {
		return errors.New("scheduler cannot be nil")
	}

	e.scheduler = scheduler

	return nil
}

func (e *Engine) SetInterval(interval time.Duration) error {
	if e.running {
		return errors.New("cannot change interval while engine is running")
	}

	if interval <= 0 {
		return errors.New("interval must be greater than zero")
	}

	e.interval = interval

	return nil
}

func (e *Engine) SetSink(sink ResultSink) error {
	if e.running {
		return errors.New("cannot change sink while engine is running")
	}

	if sink == nil {
		return errors.New("sink cannot be nil")
	}

	e.sink = sink

	return nil
}

func (e *Engine) Start() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	e.cancel = cancel
	e.running = true

	e.wg.Add(1)

	go func() {
		defer e.wg.Done()

		e.scheduler(ctx, func() {
			results := e.Check()
			e.sink.Publish(results)
		}, e.interval)
	}()
}

func (e *Engine) Stop() {
	e.mu.Lock()

	if !e.running {
		e.mu.Unlock()
		return
	}

	cancel := e.cancel

	e.cancel = nil
	e.running = false

	e.mu.Unlock()

	cancel()
	e.wg.Wait()
}

func (e *Engine) Check() []CheckResult {
	e.mu.RLock()

	checkers := make([]Checker, len(e.checkers))
	copy(checkers, e.checkers)

	e.mu.RUnlock()

	return e.launcher.Launch(checkers)
}

func (e *Engine) AddChecker(checker Checker) error {
	if checker == nil {
		return errors.New("checker cannot be nil")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.checkers = append(e.checkers, checker)

	return nil
}

func (e *Engine) HasChecker(name string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, checker := range e.checkers {
		if checker.Name() == name {
			return true
		}
	}

	return false
}

func (e *Engine) CheckerCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return len(e.checkers)
}
