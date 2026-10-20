package api

import (
	"path/filepath"
	"testing"

	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/application"
	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/config"
	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/monitoring"
	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/testutil"
)

func newTestClient(t *testing.T) (*Client, *UnixServer) {
	t.Helper()

	registry := &testutil.FakeRegistry{
		Services: []config.ServiceConfig{
			{
				Name: "sshd",
				Type: "systemd",
			},
		},
	}

	engine := monitoring.NewEngine()

	app := application.New(
		registry,
		engine,
		testutil.FakeCheckerFactory{},
	)

	apiServer := NewServer(app)

	socketPath := filepath.Join(
		t.TempDir(),
		"monitoring.sock",
	)

	unixServer, err := NewUnixServer(
		UnixServerConfig{
			Path: socketPath,
			Mode: 0660,
		},
		apiServer,
	)
	if err != nil {
		t.Fatalf("NewUnixServer() failed: %v", err)
	}

	go func() {
		_ = unixServer.Serve()
	}()

	client, err := NewClient(socketPath)
	if err != nil {
		unixServer.Close()
		t.Fatalf("NewClient() failed: %v", err)
	}

	t.Cleanup(func() {
		unixServer.Close()
	})

	return client, unixServer
}

func TestClientListServices(t *testing.T) {
	client, _ := newTestClient(t)

	services, err := client.ListServices()
	if err != nil {
		t.Fatalf("ListServices() failed: %v", err)
	}

	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}

	if services[0].Name != "sshd" {
		t.Fatalf(
			"expected sshd, got %q",
			services[0].Name,
		)
	}
}
