package monitoring

import "testing"

func TestSequentialLauncherLaunch(t *testing.T) {
	launcher := SequentialLauncher{}

	checkers := []Checker{
		fakeChecker{
			result: CheckResult{
				Name:    "checker-1",
				Status:  StatusHealthy,
				State:   "active",
				Message: "service is active",
			},
		},
		fakeChecker{
			result: CheckResult{
				Name:    "checker-2",
				Status:  StatusUnhealthy,
				State:   "failed",
				Message: "service is not active",
			},
		},
	}

	results := launcher.Launch(checkers)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].Name != "checker-1" {
		t.Errorf("expected first result from checker-1, got %s", results[0].Name)
	}

	if results[1].Name != "checker-2" {
		t.Errorf("expected second result from checker-2, got %s", results[1].Name)
	}
}
