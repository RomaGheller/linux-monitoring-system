package registry

import (
	"fmt"

	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/config"
)

type ServiceRegistry interface {
	List() ([]config.ServiceConfig, error)
	Add(service config.ServiceConfig) error
	Ensure(service config.ServiceConfig) (bool, error)
}

type FileServiceRegistry struct {
	path string
}

var _ ServiceRegistry = (*FileServiceRegistry)(nil)

func NewFileServiceRegistry(path string) *FileServiceRegistry {
	return &FileServiceRegistry{
		path: path,
	}
}

func (r *FileServiceRegistry) List() ([]config.ServiceConfig, error) {
	cfg, err := config.Load(r.path)
	if err != nil {
		return nil, err
	}

	return cfg.Services, nil
}

func (r *FileServiceRegistry) Add(service config.ServiceConfig) error {
	if err := service.Validate(); err != nil {
		return err
	}

	cfg, err := config.Load(r.path)
	if err != nil {
		return err
	}

	for _, existing := range cfg.Services {
		if existing.Name == service.Name {
			return fmt.Errorf(
				"service %q already exists",
				service.Name,
			)
		}
	}

	cfg.Services = append(cfg.Services, service)

	return config.Save(r.path, cfg)
}

func (r *FileServiceRegistry) Ensure(
	service config.ServiceConfig,
) (bool, error) {
	if err := service.Validate(); err != nil {
		return false, err
	}

	cfg, err := config.Load(r.path)
	if err != nil {
		return false, err
	}

	for _, existing := range cfg.Services {
		if existing.Name == service.Name {
			if existing.Type != service.Type {
				return false, fmt.Errorf(
					"service %q already exists with type %q",
					service.Name,
					existing.Type,
				)
			}

			return false, nil
		}
	}

	cfg.Services = append(cfg.Services, service)

	if err := config.Save(r.path, cfg); err != nil {
		return false, err
	}

	return true, nil
}
