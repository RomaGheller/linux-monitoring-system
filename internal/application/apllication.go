package application

import (
	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/config"
	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/monitoring"
	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/registry"
)

type Application struct {
	registry registry.ServiceRegistry
	engine   *monitoring.Engine
	factory  CheckerFactory
}

type CheckerFactory interface {
	Create(service config.ServiceConfig) (monitoring.Checker, error)
}

func New(
	serviceRegistry registry.ServiceRegistry,
	engine *monitoring.Engine,
	factory CheckerFactory,
) *Application {
	return &Application{
		registry: serviceRegistry,
		engine:   engine,
		factory:  factory,
	}
}

func (a *Application) ListServices() ([]config.ServiceConfig, error) {
	return a.registry.List()
}

func (a *Application) Reload() error {
	services, err := a.registry.List()
	if err != nil {
		return err
	}

	for _, service := range services {
		if a.engine.HasChecker(service.Name) {
			continue
		}

		checker, err := a.factory.Create(service)
		if err != nil {
			return err
		}

		if err := a.engine.AddChecker(checker); err != nil {
			return err
		}
	}

	return nil
}

func (a *Application) AddService(
	service config.ServiceConfig,
) error {
	if err := a.registry.Add(service); err != nil {
		return err
	}

	return a.Reload()
}

func (a *Application) EnsureService(
	service config.ServiceConfig,
) (bool, error) {
	added, err := a.registry.Ensure(service)
	if err != nil {
		return false, err
	}

	if !added {
		return false, nil
	}

	if err := a.Reload(); err != nil {
		return false, err
	}

	return true, nil
}
