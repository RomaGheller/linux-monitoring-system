package systemd

import (
	"errors"
	"testing"

	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/monitoring"
)

func TestCheckerCheck(t *testing.T) {
	tests := []struct {
		name            string
		commandResult   CommandResult
		expectedStatus  monitoring.Status
		expectedState   string
		expectedMessage string
	}{
		{
			name: "active service",
			commandResult: CommandResult{
				Output:   "active\n",
				ExitCode: 0,
			},
			expectedStatus:  monitoring.StatusHealthy,
			expectedState:   "active",
			expectedMessage: "systemd service is active",
		},
		{
			name: "inactive service",
			commandResult: CommandResult{
				Output:   "inactive\n",
				ExitCode: 3,
			},
			expectedStatus:  monitoring.StatusUnhealthy,
			expectedState:   "inactive",
			expectedMessage: "systemd service is not active",
		},
		{
			name: "failed service",
			commandResult: CommandResult{
				Output:   "failed\n",
				ExitCode: 3,
			},
			expectedStatus:  monitoring.StatusUnhealthy,
			expectedState:   "failed",
			expectedMessage: "systemd service is not active",
		},
		{
			name: "activating service",
			commandResult: CommandResult{
				Output:   "activating\n",
				ExitCode: 0,
			},
			expectedStatus:  monitoring.StatusUnhealthy,
			expectedState:   "activating",
			expectedMessage: "systemd service is not active",
		},
		{
			name: "deactivating service",
			commandResult: CommandResult{
				Output:   "deactivating\n",
				ExitCode: 0,
			},
			expectedStatus:  monitoring.StatusUnhealthy,
			expectedState:   "deactivating",
			expectedMessage: "systemd service is not active",
		},
		{
			name: "unknown service state",
			commandResult: CommandResult{
				Output:   "unknown\n",
				ExitCode: 3,
			},
			expectedStatus:  monitoring.StatusError,
			expectedState:   "unknown",
			expectedMessage: "unknown systemd status",
		},
		{
			name: "execution error",
			commandResult: CommandResult{
				Err: errors.New("systemctl not found"),
			},
			expectedStatus:  monitoring.StatusError,
			expectedState:   "",
			expectedMessage: "failed to execute systemctl: systemctl not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := func(name string, args ...string) CommandResult {
				return tt.commandResult
			}

			checker := NewChecker("test-service", runner)

			result := checker.Check()

			if result.Status != tt.expectedStatus {
				t.Errorf(
					"expected status %q, got %q",
					tt.expectedStatus,
					result.Status,
				)
			}

			if result.State != tt.expectedState {
				t.Errorf(
					"expected state %q, got %q",
					tt.expectedState,
					result.State,
				)
			}

			if result.Message != tt.expectedMessage {
				t.Errorf(
					"expected message %q, got %q",
					tt.expectedMessage,
					result.Message,
				)
			}
		})
	}
}
