package systemd

import (
	"errors"
	"os/exec"
)

type CommandResult struct {
	Output   string
	ExitCode int
	Err      error
}

type CommandRunner func(name string, args ...string) CommandResult

func RunCommand(name string, args ...string) CommandResult {
	cmd := exec.Command(name, args...)

	output, err := cmd.CombinedOutput()

	if err == nil {
		return CommandResult{
			Output:   string(output),
			ExitCode: 0,
		}
	}

	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return CommandResult{
			Output:   string(output),
			ExitCode: exitError.ExitCode(),
		}
	}

	return CommandResult{
		Output: string(output),
		Err:    err,
	}
}
