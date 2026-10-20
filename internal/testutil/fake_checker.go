package testutil

import "gitlab.com/linuxgamerpro/linux-monitoring-system/internal/monitoring"

type FakeChecker struct {
	result monitoring.CheckResult
}

func (f FakeChecker) Check() monitoring.CheckResult {
	return f.result
}

func (c FakeChecker) Name() string {
	return c.result.Name
}
