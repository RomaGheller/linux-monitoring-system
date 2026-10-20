package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Services []ServiceConfig `yaml:"services"`
}

type ServiceConfig struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	var config Config

	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("parse config file: %w", err)
	}

	if err := config.Validate(); err != nil {
		return Config{}, fmt.Errorf("validate config: %w", err)
	}

	return config, nil
}

func (s ServiceConfig) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("service name is required")
	}

	if s.Type == "" {
		return fmt.Errorf("service %q: type is required", s.Name)
	}

	if s.Type != "systemd" {
		return fmt.Errorf(
			"service %q: unsupported type %q",
			s.Name,
			s.Type,
		)
	}

	return nil
}

func (c Config) Validate() error {

	seen := make(map[string]struct{})

	for _, service := range c.Services {
		if err := service.Validate(); err != nil {
			return err
		}

		if _, exists := seen[service.Name]; exists {
			return fmt.Errorf(
				"duplicate service name %q",
				service.Name,
			)
		}

		seen[service.Name] = struct{}{}
	}

	return nil
}

func Save(path string, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	dir := filepath.Dir(path)

	tmpFile, err := os.CreateTemp(dir, ".services-*.yaml")
	if err != nil {
		return fmt.Errorf("create temporary config file: %w", err)
	}

	tmpPath := tmpFile.Name()

	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("write temporary config file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temporary config file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace config file: %w", err)
	}

	return nil
}
