package monitoring

import (
	"context"
	"testing"
	"time"
)

type fakeChecker struct {
	result CheckResult
}

type fakeResultSink struct {
	results []CheckResult
	called  chan struct{}
}

func (s *fakeResultSink) Publish(results []CheckResult) {
	s.results = results
	close(s.called)
}

func (f fakeChecker) Check() CheckResult {
	return f.result
}

func (c fakeChecker) Name() string {
	return c.result.Name
}

func TestEngineCheck(t *testing.T) {
	engine := NewEngine()

	engine.AddChecker(fakeChecker{
		result: CheckResult{
			Name:    "checker-1",
			Status:  StatusHealthy,
			State:   "active",
			Message: "service is active",
		},
	})

	engine.AddChecker(fakeChecker{
		result: CheckResult{
			Name:    "checker-2",
			Status:  StatusUnhealthy,
			State:   "failed",
			Message: "service is not active",
		},
	})

	results := engine.Check()

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	expectedResults := []CheckResult{
		{
			Name:    "checker-1",
			Status:  StatusHealthy,
			State:   "active",
			Message: "service is active",
		},
		{
			Name:    "checker-2",
			Status:  StatusUnhealthy,
			State:   "failed",
			Message: "service is not active",
		},
	}

	for i, expected := range expectedResults {
		if results[i] != expected {
			t.Errorf(
				"expected result %d to be %+v, got %+v",
				i,
				expected,
				results[i],
			)
		}
	}
}

func TestEngineCheckWithoutCheckers(t *testing.T) {
	engine := NewEngine()

	results := engine.Check()

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestEngineSetInterval(t *testing.T) {
	engine := NewEngine()

	if err := engine.SetInterval(5 * time.Second); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if engine.interval != 5*time.Second {
		t.Errorf(
			"expected interval %v, got %v",
			5*time.Second,
			engine.interval,
		)
	}
}

func TestEngineSetIntervalRejectsInvalidValue(t *testing.T) {
	engine := NewEngine()

	if err := engine.SetInterval(0); err == nil {
		t.Fatal("expected error for zero interval")
	}

	if err := engine.SetInterval(-time.Second); err == nil {
		t.Fatal("expected error for negative interval")
	}
}

func TestEngineSetLauncher(t *testing.T) {
	engine := NewEngine()

	launcher := SequentialLauncher{}

	if err := engine.SetLauncher(launcher); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if engine.launcher != launcher {
		t.Fatal("launcher was not updated")
	}
}

func TestEngineSetScheduler(t *testing.T) {
	engine := NewEngine()

	scheduler := func(
		ctx context.Context,
		task func(),
		interval time.Duration,
	) {
	}

	if err := engine.SetScheduler(scheduler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if engine.scheduler == nil {
		t.Fatal("scheduler was not updated")
	}
}

func TestEngineSetSink(t *testing.T) {
	engine := NewEngine()

	sink := &fakeResultSink{}

	if err := engine.SetSink(sink); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if engine.sink != sink {
		t.Fatal("sink was not updated")
	}
}

func TestEngineCannotBeConfiguredWhileRunning(t *testing.T) {
	engine := NewEngine()

	schedulerStarted := make(chan struct{})

	scheduler := func(
		ctx context.Context,
		task func(),
		interval time.Duration,
	) {
		close(schedulerStarted)
		<-ctx.Done()
	}

	if err := engine.SetScheduler(scheduler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	engine.Start()

	select {
	case <-schedulerStarted:
	case <-time.After(time.Second):
		t.Fatal("scheduler did not start")
	}

	defer engine.Stop()

	if err := engine.SetInterval(time.Second); err == nil {
		t.Fatal("expected error when changing interval while engine is running")
	}
}

func TestEngineStartChecksAndPublishesResults(t *testing.T) {
	engine := NewEngine()

	sink := &fakeResultSink{
		called: make(chan struct{}),
	}

	if err := engine.SetSink(sink); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	engine.AddChecker(fakeChecker{
		result: CheckResult{
			Name:    "test-service",
			Status:  StatusHealthy,
			State:   "active",
			Message: "service is active",
		},
	})

	scheduler := func(
		ctx context.Context,
		task func(),
		interval time.Duration,
	) {
		task()
		<-ctx.Done()
	}

	if err := engine.SetScheduler(scheduler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	engine.Start()

	defer engine.Stop()

	select {
	case <-sink.called:
	case <-time.After(time.Second):
		t.Fatal("result sink was not called")
	}

	if len(sink.results) != 1 {
		t.Fatalf(
			"expected 1 result, got %d",
			len(sink.results),
		)
	}

	result := sink.results[0]

	if result.Name != "test-service" {
		t.Errorf(
			"expected name %q, got %q",
			"test-service",
			result.Name,
		)
	}

	if result.Status != StatusHealthy {
		t.Errorf(
			"expected status %q, got %q",
			StatusHealthy,
			result.Status,
		)
	}

	if result.State != "active" {
		t.Errorf(
			"expected state %q, got %q",
			"active",
			result.State,
		)
	}
}

func TestEngineStop(t *testing.T) {
	engine := NewEngine()

	stopped := make(chan struct{})

	scheduler := func(
		ctx context.Context,
		task func(),
		interval time.Duration,
	) {
		<-ctx.Done()
		close(stopped)
	}

	if err := engine.SetScheduler(scheduler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	engine.Start()
	engine.Stop()

	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("scheduler did not stop")
	}

	if engine.running {
		t.Fatal("expected engine to be stopped")
	}
}

func TestEngineAddChecker(t *testing.T) {
	engine := NewEngine()

	checker := fakeChecker{
		result: CheckResult{
			Name:   "test-service",
			Status: StatusHealthy,
			State:  "active",
		},
	}

	if err := engine.AddChecker(checker); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(engine.checkers) != 1 {
		t.Fatalf("expected 1 checker, got %d", len(engine.checkers))
	}

	if engine.checkers[0] != checker {
		t.Fatal("checker was not added")
	}
}

func TestEngineAddCheckerRejectsNil(t *testing.T) {
	engine := NewEngine()

	if err := engine.AddChecker(nil); err == nil {
		t.Fatal("expected error when adding nil checker")
	}

	if len(engine.checkers) != 0 {
		t.Fatalf("expected 0 checkers, got %d", len(engine.checkers))
	}
}

func TestEngineAddCheckerWhileRunning(t *testing.T) {
	engine := NewEngine()

	if err := engine.SetInterval(time.Millisecond); err != nil {
		t.Fatalf("SetInterval() failed: %v", err)
	}

	checker := fakeChecker{
		result: CheckResult{
			Name:   "test-service1",
			Status: StatusHealthy,
		},
	}

	if err := engine.AddChecker(checker); err != nil {
		t.Fatalf("AddChecker() failed: %v", err)
	}

	engine.Start()
	defer engine.Stop()

	secondChecker := fakeChecker{
		result: CheckResult{
			Name:   "test-service2",
			Status: StatusHealthy,
		},
	}

	if err := engine.AddChecker(secondChecker); err != nil {
		t.Fatalf("AddChecker() while running failed: %v", err)
	}
}

func TestEngineHasChecker(t *testing.T) {
	engine := NewEngine()

	checker := fakeChecker{
		result: CheckResult{
			Name:   "test-service",
			Status: StatusHealthy,
		},
	}

	if err := engine.AddChecker(checker); err != nil {
		t.Fatalf("AddChecker() failed: %v", err)
	}

	if !engine.HasChecker("test-service") {
		t.Fatal("HasChecker() returned false for existing checker")
	}

	if engine.HasChecker("missing-service") {
		t.Fatal("HasChecker() returned true for missing checker")
	}
}

func TestEngineHasCheckerAfterAddingMultipleCheckers(t *testing.T) {
	engine := NewEngine()

	checkers := []fakeChecker{
		{
			result: CheckResult{Name: "service-1"},
		},
		{
			result: CheckResult{Name: "service-2"},
		},
	}

	for _, checker := range checkers {
		if err := engine.AddChecker(checker); err != nil {
			t.Fatalf("AddChecker() failed: %v", err)
		}
	}

	for _, name := range []string{"service-1", "service-2"} {
		if !engine.HasChecker(name) {
			t.Fatalf("HasChecker(%q) returned false", name)
		}
	}
}

func TestEngineStopWaitsForScheduler(t *testing.T) {
	engine := NewEngine()

	schedulerStarted := make(chan struct{})
	allowSchedulerExit := make(chan struct{})

	scheduler := func(
		ctx context.Context,
		task func(),
		interval time.Duration,
	) {
		close(schedulerStarted)

		<-ctx.Done()
		<-allowSchedulerExit
	}

	if err := engine.SetScheduler(scheduler); err != nil {
		t.Fatalf("SetScheduler() error = %v", err)
	}

	engine.Start()

	<-schedulerStarted

	stopDone := make(chan struct{})

	go func() {
		engine.Stop()
		close(stopDone)
	}()

	select {
	case <-stopDone:
		t.Fatal("Stop() returned before scheduler exited")

	case <-time.After(50 * time.Millisecond):
		// Expected: Stop() is still waiting.
	}

	close(allowSchedulerExit)

	select {
	case <-stopDone:
		// Expected.

	case <-time.After(time.Second):
		t.Fatal("Stop() did not return after scheduler exited")
	}
}
