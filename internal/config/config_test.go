package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/config"
)

func TestLoad(t *testing.T) {

	dir := t.TempDir()

	path := filepath.Join(dir, "services.yaml")

	content := `
services:
  - name: mts-mock-service
    type: systemd
  - name: nginx
    type: systemd
`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(cfg.Services))
	}

	if cfg.Services[0].Name != "mts-mock-service" {
		t.Errorf(
			"expected first service to be mts-mock-service, got %q",
			cfg.Services[0].Name,
		)
	}

	if cfg.Services[0].Type != "systemd" {
		t.Errorf(
			"expected first service type to be systemd, got %q",
			cfg.Services[0].Type,
		)
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := config.Load("/does/not/exist/services.yaml")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "services.yaml")

	content := `
services:
  - name: [
`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	_, err := config.Load(path)

	if err == nil {
		t.Fatal("expected parsing error, got nil")
	}
}

func TestValidateEmptyName(t *testing.T) {
	cfg := config.Config{
		Services: []config.ServiceConfig{
			{
				Name: "",
				Type: "systemd",
			},
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidateEmptyType(t *testing.T) {
	cfg := config.Config{
		Services: []config.ServiceConfig{
			{
				Name: "nginx",
				Type: "",
			},
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidateUnsupportedType(t *testing.T) {
	cfg := config.Config{
		Services: []config.ServiceConfig{
			{
				Name: "nginx",
				Type: "docker",
			},
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestServiceConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		service config.ServiceConfig
		wantErr bool
	}{
		{
			name: "valid systemd service",
			service: config.ServiceConfig{
				Name: "nginx",
				Type: "systemd",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			service: config.ServiceConfig{
				Name: "",
				Type: "systemd",
			},
			wantErr: true,
		},
		{
			name: "empty type",
			service: config.ServiceConfig{
				Name: "nginx",
				Type: "",
			},
			wantErr: true,
		},
		{
			name: "unsupported type",
			service: config.ServiceConfig{
				Name: "nginx",
				Type: "docker",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.service.Validate()

			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"Validate() error = %v, wantErr = %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestValidateDuplicateServices(t *testing.T) {
	cfg := config.Config{
		Services: []config.ServiceConfig{
			{
				Name: "nginx",
				Type: "systemd",
			},
			{
				Name: "nginx",
				Type: "systemd",
			},
		},
	}

	err := cfg.Validate()

	if err == nil {
		t.Fatal("expected duplicate service error, got nil")
	}
}

func TestSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "services.yaml")

	cfg := config.Config{
		Services: []config.ServiceConfig{
			{
				Name: "nginx",
				Type: "systemd",
			},
			{
				Name: "sshd",
				Type: "systemd",
			},
		},
	}

	if err := config.Save(path, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, err := config.Load(path)
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}

	if len(loaded.Services) != 2 {
		t.Fatalf(
			"expected 2 services, got %d",
			len(loaded.Services),
		)
	}

	if loaded.Services[0].Name != "nginx" {
		t.Errorf(
			"expected first service to be nginx, got %q",
			loaded.Services[0].Name,
		)
	}

	if loaded.Services[1].Name != "sshd" {
		t.Errorf(
			"expected second service to be sshd, got %q",
			loaded.Services[1].Name,
		)
	}
}

func TestSaveInvalidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "services.yaml")

	cfg := config.Config{
		Services: []config.ServiceConfig{
			{
				Name: "",
				Type: "systemd",
			},
		},
	}

	if err := config.Save(path, cfg); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestSaveReplacesExistingConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "services.yaml")

	first := config.Config{
		Services: []config.ServiceConfig{
			{
				Name: "nginx",
				Type: "systemd",
			},
		},
	}

	if err := config.Save(path, first); err != nil {
		t.Fatalf("failed to save first config: %v", err)
	}

	second := config.Config{
		Services: []config.ServiceConfig{
			{
				Name: "sshd",
				Type: "systemd",
			},
		},
	}

	if err := config.Save(path, second); err != nil {
		t.Fatalf("failed to save second config: %v", err)
	}

	loaded, err := config.Load(path)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if len(loaded.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(loaded.Services))
	}

	if loaded.Services[0].Name != "sshd" {
		t.Errorf(
			"expected sshd, got %q",
			loaded.Services[0].Name,
		)
	}
}
