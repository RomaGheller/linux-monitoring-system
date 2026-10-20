package systemd

import (
	"fmt"
	"strings"

	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/monitoring"
)

type Checker struct {
	ServiceName string
	run         CommandRunner
}

func NewChecker(serviceName string, run CommandRunner) *Checker {
	return &Checker{
		ServiceName: serviceName,
		run:         run,
	}
}

func (c *Checker) Check() monitoring.CheckResult {
	result := c.run("systemctl", "is-active", c.ServiceName)
	state := strings.TrimSpace(result.Output)

	if result.Err != nil {
		return monitoring.CheckResult{
			Name:    c.ServiceName,
			Status:  monitoring.StatusError,
			State:   state,
			Message: fmt.Sprintf("failed to execute systemctl: %v", result.Err),
		}
	}

	switch state {
	case "active":
		return monitoring.CheckResult{
			Name:    c.ServiceName,
			Status:  monitoring.StatusHealthy,
			State:   state,
			Message: "systemd service is active",
		}

	case "inactive", "failed", "activating", "deactivating":
		return monitoring.CheckResult{
			Name:    c.ServiceName,
			Status:  monitoring.StatusUnhealthy,
			State:   state,
			Message: "systemd service is not active",
		}

	default:
		return monitoring.CheckResult{
			Name:    c.ServiceName,
			Status:  monitoring.StatusError,
			State:   state,
			Message: "unknown systemd status",
		}
	}
}

func (c *Checker) Name() string {
	return c.ServiceName
}
