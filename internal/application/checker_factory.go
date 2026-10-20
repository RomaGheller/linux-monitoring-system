package application

import (
	"fmt"

	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/config"
	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/monitoring"
	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/systemd"
)

type SystemdCheckerFactory struct {
	run systemd.CommandRunner
}

func NewSystemdCheckerFactory(
	run systemd.CommandRunner,
) *SystemdCheckerFactory {
	return &SystemdCheckerFactory{
		run: run,
	}
}

func (f *SystemdCheckerFactory) Create(
	service config.ServiceConfig,
) (monitoring.Checker, error) {
	switch service.Type {
	case "systemd":
		return systemd.NewChecker(
			service.Name,
			f.run,
		), nil

	default:
		return nil, fmt.Errorf(
			"unsupported service type: %q",
			service.Type,
		)
	}
}
