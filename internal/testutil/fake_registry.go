package testutil

import (
	"fmt"

	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/config"
)

type FakeRegistry struct {
	Services []config.ServiceConfig
	Err      error
}

func (r *FakeRegistry) List() ([]config.ServiceConfig, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	return r.Services, nil
}

func (r *FakeRegistry) Add(service config.ServiceConfig) error {
	if r.Err != nil {
		return r.Err
	}

	for _, existing := range r.Services {
		if existing.Name == service.Name {
			return fmt.Errorf("service %q already exists", service.Name)
		}
	}

	r.Services = append(r.Services, service)

	return nil
}

func (r *FakeRegistry) Ensure(service config.ServiceConfig) (bool, error) {
	if r.Err != nil {
		return false, r.Err
	}

	for _, existing := range r.Services {
		if existing.Name == service.Name {
			return false, nil
		}
	}

	r.Services = append(r.Services, service)

	return true, nil
}
