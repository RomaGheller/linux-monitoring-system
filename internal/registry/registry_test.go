package registry

import (
	"path/filepath"
	"testing"

	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/config"
)

func TestFileServiceRegistryList(t *testing.T) {
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
		t.Fatalf("failed to save test config: %v", err)
	}

	registry := NewFileServiceRegistry(path)

	services, err := registry.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(services))
	}

	if services[0].Name != "nginx" {
		t.Errorf("expected nginx, got %q", services[0].Name)
	}

	if services[1].Name != "sshd" {
		t.Errorf("expected sshd, got %q", services[1].Name)
	}
}

func TestFileServiceRegistryAdd(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "services.yaml")

	cfg := config.Config{
		Services: []config.ServiceConfig{
			{
				Name: "nginx",
				Type: "systemd",
			},
		},
	}

	if err := config.Save(path, cfg); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	registry := NewFileServiceRegistry(path)

	err := registry.Add(config.ServiceConfig{
		Name: "sshd",
		Type: "systemd",
	})
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	services, err := registry.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(services))
	}

	if services[1].Name != "sshd" {
		t.Errorf("expected sshd, got %q", services[1].Name)
	}
}

func TestFileServiceRegistryAddDuplicate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "services.yaml")

	cfg := config.Config{
		Services: []config.ServiceConfig{
			{
				Name: "nginx",
				Type: "systemd",
			},
		},
	}

	if err := config.Save(path, cfg); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	registry := NewFileServiceRegistry(path)

	err := registry.Add(config.ServiceConfig{
		Name: "nginx",
		Type: "systemd",
	})

	if err == nil {
		t.Fatal("expected duplicate service error, got nil")
	}

	services, err := registry.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(services) != 1 {
		t.Fatalf(
			"expected config to still contain 1 service, got %d",
			len(services),
		)
	}
}

func TestFileServiceRegistryEnsureAddsService(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "services.yaml")

	cfg := config.Config{
		Services: []config.ServiceConfig{
			{
				Name: "nginx",
				Type: "systemd",
			},
		},
	}

	if err := config.Save(path, cfg); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	registry := NewFileServiceRegistry(path)

	added, err := registry.Ensure(config.ServiceConfig{
		Name: "sshd",
		Type: "systemd",
	})

	if err != nil {
		t.Fatalf("Ensure() failed: %v", err)
	}

	if !added {
		t.Fatal("expected service to be added")
	}

	services, err := registry.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(services))
	}
}

func TestFileServiceRegistryEnsureExistingService(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "services.yaml")

	cfg := config.Config{
		Services: []config.ServiceConfig{
			{
				Name: "nginx",
				Type: "systemd",
			},
		},
	}

	if err := config.Save(path, cfg); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	registry := NewFileServiceRegistry(path)

	added, err := registry.Ensure(config.ServiceConfig{
		Name: "nginx",
		Type: "systemd",
	})

	if err != nil {
		t.Fatalf("Ensure() failed: %v", err)
	}

	if added {
		t.Fatal("expected existing service not to be added")
	}

	services, err := registry.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}
}

func TestFileServiceRegistryEnsureDifferentType(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "services.yaml")

	cfg := config.Config{
		Services: []config.ServiceConfig{
			{
				Name: "nginx",
				Type: "systemd",
			},
		},
	}

	if err := config.Save(path, cfg); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	registry := NewFileServiceRegistry(path)

	_, err := registry.Ensure(config.ServiceConfig{
		Name: "nginx",
		Type: "docker",
	})

	if err == nil {
		t.Fatal("expected type conflict error, got nil")
	}
}
