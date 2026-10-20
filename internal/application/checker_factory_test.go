package application

import (
	"testing"

	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/config"
	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/systemd"
)

func TestSystemdCheckerFactoryCreate(t *testing.T) {
	factory := NewSystemdCheckerFactory(systemd.RunCommand)

	service := config.ServiceConfig{
		Name: "sshd",
		Type: "systemd",
	}

	checker, err := factory.Create(service)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if checker == nil {
		t.Fatal("Create() returned nil checker")
	}

	if checker.Name() != "sshd" {
		t.Fatalf(
			"expected checker name %q, got %q",
			"sshd",
			checker.Name(),
		)
	}
}

func TestSystemdCheckerFactoryUnsupportedType(t *testing.T) {
	factory := NewSystemdCheckerFactory(systemd.RunCommand)

	service := config.ServiceConfig{
		Name: "nginx",
		Type: "docker",
	}

	checker, err := factory.Create(service)

	if err == nil {
		t.Fatal("expected error for unsupported service type")
	}

	if checker != nil {
		t.Fatal("expected nil checker on error")
	}
}
