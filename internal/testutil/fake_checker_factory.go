package testutil

import (
	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/config"
	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/monitoring"
)

type FakeCheckerFactory struct {
	Err error
}

func (f FakeCheckerFactory) Create(
	service config.ServiceConfig,
) (monitoring.Checker, error) {
	if f.Err != nil {
		return nil, f.Err
	}

	return FakeChecker{
		result: monitoring.CheckResult{
			Name: service.Name,
		},
	}, nil
}
