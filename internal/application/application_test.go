package application

import (
	"errors"
	"fmt"
	"testing"

	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/config"
	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/monitoring"
)

type fakeRegistry struct {
	services []config.ServiceConfig
}

func (r *fakeRegistry) Add(service config.ServiceConfig) error {
	for _, existing := range r.services {
		if existing.Name == service.Name {
			return fmt.Errorf("service %q already exists", service.Name)
		}
	}

	r.services = append(r.services, service)

	return nil
}

type fakeCheckerFactory struct {
	err error
}

type fakeChecker struct {
	result monitoring.CheckResult
}

func (c fakeChecker) Name() string {
	return c.result.Name
}

func (c fakeChecker) Check() monitoring.CheckResult {
	return c.result
}

func (f fakeCheckerFactory) Create(
	service config.ServiceConfig,
) (monitoring.Checker, error) {
	if f.err != nil {
		return nil, f.err
	}

	return fakeChecker{
		result: monitoring.CheckResult{
			Name:   service.Name,
			Status: monitoring.StatusHealthy,
		},
	}, nil
}

func (r *fakeRegistry) List() ([]config.ServiceConfig, error) {
	return r.services, nil
}

func (r *fakeRegistry) Ensure(
	service config.ServiceConfig,
) (bool, error) {
	for _, existing := range r.services {
		if existing.Name == service.Name {
			return false, nil
		}
	}

	r.services = append(r.services, service)
	return true, nil
}

func TestApplicationListServices(t *testing.T) {
	registry := &fakeRegistry{
		services: []config.ServiceConfig{
			{
				Name: "sshd",
				Type: "systemd",
			},
			{
				Name: "NetworkManager",
				Type: "systemd",
			},
		},
	}

	engine := monitoring.NewEngine()

	app := New(registry, engine, fakeCheckerFactory{})

	services, err := app.ListServices()
	if err != nil {
		t.Fatalf("ListServices() failed: %v", err)
	}

	if len(services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(services))
	}

	if services[0].Name != "sshd" {
		t.Fatalf("expected sshd, got %q", services[0].Name)
	}
}

func TestApplicationReloadDoesNotDuplicateChecker(t *testing.T) {
	registry := &fakeRegistry{
		services: []config.ServiceConfig{
			{
				Name: "sshd",
				Type: "systemd",
			},
		},
	}

	engine := monitoring.NewEngine()

	factory := fakeCheckerFactory{}

	app := New(registry, engine, factory)

	if err := app.Reload(); err != nil {
		t.Fatalf("Reload() failed: %v", err)
	}

	if !engine.HasChecker("sshd") {
		t.Fatal("expected sshd checker to be added")
	}

	if err := app.Reload(); err != nil {
		t.Fatalf("second Reload() failed: %v", err)
	}

	if got := engine.CheckerCount(); got != 1 {
		t.Fatalf("expected 1 checker, got %d", got)
	}
}

func TestApplicationReloadAddsMissingChecker(t *testing.T) {
	registry := &fakeRegistry{
		services: []config.ServiceConfig{
			{
				Name: "sshd",
				Type: "systemd",
			},
			{
				Name: "NetworkManager",
				Type: "systemd",
			},
		},
	}

	engine := monitoring.NewEngine()

	app := New(
		registry,
		engine,
		fakeCheckerFactory{},
	)

	if err := app.Reload(); err != nil {
		t.Fatalf("Reload() failed: %v", err)
	}

	if !engine.HasChecker("sshd") {
		t.Fatal("expected sshd checker")
	}

	if !engine.HasChecker("NetworkManager") {
		t.Fatal("expected NetworkManager checker")
	}

	if got := engine.CheckerCount(); got != 2 {
		t.Fatalf("expected 2 checkers, got %d", got)
	}
}

func TestApplicationReloadFactoryError(t *testing.T) {
	expectedErr := errors.New("factory error")

	registry := &fakeRegistry{
		services: []config.ServiceConfig{
			{
				Name: "sshd",
				Type: "systemd",
			},
		},
	}

	engine := monitoring.NewEngine()

	app := New(
		registry,
		engine,
		fakeCheckerFactory{
			err: expectedErr,
		},
	)

	err := app.Reload()
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}

	if engine.CheckerCount() != 0 {
		t.Fatal("checker should not have been added")
	}
}

func TestApplicationAddService(t *testing.T) {
	registry := &fakeRegistry{}

	engine := monitoring.NewEngine()

	app := New(
		registry,
		engine,
		fakeCheckerFactory{},
	)

	service := config.ServiceConfig{
		Name: "sshd",
		Type: "systemd",
	}

	if err := app.AddService(service); err != nil {
		t.Fatalf("AddService() failed: %v", err)
	}

	if !engine.HasChecker("sshd") {
		t.Fatal("expected sshd checker to be added")
	}

	if got := engine.CheckerCount(); got != 1 {
		t.Fatalf("expected 1 checker, got %d", got)
	}
}

func TestApplicationEnsureServiceAddsNewService(t *testing.T) {
	registry := &fakeRegistry{}

	engine := monitoring.NewEngine()

	app := New(
		registry,
		engine,
		fakeCheckerFactory{},
	)

	service := config.ServiceConfig{
		Name: "sshd",
		Type: "systemd",
	}

	added, err := app.EnsureService(service)
	if err != nil {
		t.Fatalf("EnsureService() failed: %v", err)
	}

	if !added {
		t.Fatal("expected service to be added")
	}

	if !engine.HasChecker("sshd") {
		t.Fatal("expected sshd checker to be added")
	}

	if got := engine.CheckerCount(); got != 1 {
		t.Fatalf("expected 1 checker, got %d", got)
	}
}

func TestApplicationEnsureServiceDoesNotDuplicate(t *testing.T) {
	registry := &fakeRegistry{}

	engine := monitoring.NewEngine()

	app := New(
		registry,
		engine,
		fakeCheckerFactory{},
	)

	service := config.ServiceConfig{
		Name: "sshd",
		Type: "systemd",
	}

	if _, err := app.EnsureService(service); err != nil {
		t.Fatalf("first EnsureService() failed: %v", err)
	}

	added, err := app.EnsureService(service)
	if err != nil {
		t.Fatalf("second EnsureService() failed: %v", err)
	}

	if added {
		t.Fatal("expected second EnsureService() to report service already exists")
	}

	if got := engine.CheckerCount(); got != 1 {
		t.Fatalf("expected 1 checker, got %d", got)
	}
}

func TestApplicationAddServiceDuplicate(t *testing.T) {
	registry := &fakeRegistry{
		services: []config.ServiceConfig{
			{
				Name: "sshd",
				Type: "systemd",
			},
		},
	}

	engine := monitoring.NewEngine()

	app := New(
		registry,
		engine,
		fakeCheckerFactory{},
	)

	service := config.ServiceConfig{
		Name: "sshd",
		Type: "systemd",
	}

	if err := app.AddService(service); err == nil {
		t.Fatal("expected error when adding duplicate service")
	}
}
