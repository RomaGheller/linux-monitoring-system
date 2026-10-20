package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/api"
	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/application"
	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/monitoring"
	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/registry"
	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/systemd"
)

type engineLifecycle interface {
	Start()
	Stop()
}

type apiServerLifecycle interface {
	Serve() error
	Close() error
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {

	serviceRegistry := registry.NewFileServiceRegistry(
		"/etc/monitoring-system/services.yaml",
	)

	engine := monitoring.NewEngine()

	factory := application.NewSystemdCheckerFactory(
		systemd.RunCommand,
	)

	app := application.New(
		serviceRegistry,
		engine,
		factory,
	)

	if err := app.Reload(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	apiServer := api.NewServer(app)

	unixServer, err := api.NewUnixServer(
		api.UnixServerConfig{
			Path: api.DefaultSocketPath,
			Mode: 0660,
		},
		apiServer,
	)

	if err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	return runDaemon(
		ctx,
		engine,
		unixServer,
	)
}

func runDaemon(
	ctx context.Context,
	engine engineLifecycle,
	server apiServerLifecycle,
) error {
	engine.Start()

	serveErr := make(chan error, 1)

	go func() {
		serveErr <- server.Serve()
	}()

	select {
	case <-ctx.Done():
		if err := server.Close(); err != nil {
			engine.Stop()
			return fmt.Errorf("failed to close API server: %w", err)
		}

		if err := <-serveErr; err != nil {
			engine.Stop()
			return fmt.Errorf(
				"API server stopped during shutdown: %w",
				err,
			)
		}

	case err := <-serveErr:
		engine.Stop()

		if err != nil {
			return fmt.Errorf(
				"API server stopped unexpectedly: %w",
				err,
			)
		}

		return errors.New("API server stopped unexpectedly")
	}

	engine.Stop()

	return nil
}
