package api

import (
	"encoding/json"
	"net"
	"testing"

	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/application"
	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/config"
	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/monitoring"
	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/testutil"
)

func sendRequest(
	t *testing.T,
	server *Server,
	request Request,
) Response {
	t.Helper()

	serverConn, clientConn := net.Pipe()

	go func() {
		if err := server.Handle(serverConn); err != nil {
			t.Errorf("server.Handle() failed: %v", err)
		}
	}()

	if err := json.NewEncoder(clientConn).Encode(request); err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	var response Response

	if err := json.NewDecoder(clientConn).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	clientConn.Close()

	return response
}

func TestServerListServices(t *testing.T) {
	registry := &testutil.FakeRegistry{
		Services: []config.ServiceConfig{
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

	app := application.New(
		registry,
		engine,
		testutil.FakeCheckerFactory{},
	)

	server := NewServer(app)

	response := sendRequest(
		t,
		server,
		Request{
			Command: CommandListServices,
		},
	)

	if !response.Success {
		t.Fatalf("expected successful response, got error: %s", response.Error)
	}

	if len(response.Services) != 2 {
		t.Fatalf(
			"expected 2 services, got %d",
			len(response.Services),
		)
	}
}

func TestServerUnknownCommand(t *testing.T) {
	app := application.New(
		&testutil.FakeRegistry{},
		monitoring.NewEngine(),
		testutil.FakeCheckerFactory{},
	)

	server := NewServer(app)

	response := sendRequest(
		t,
		server,
		Request{
			Command: "something.invalid",
		},
	)

	if response.Success {
		t.Fatal("expected unsuccessful response")
	}

	if response.Error == "" {
		t.Fatal("expected error message")
	}
}

func TestServerAddServiceWithoutService(t *testing.T) {
	app := application.New(
		&testutil.FakeRegistry{},
		monitoring.NewEngine(),
		testutil.FakeCheckerFactory{},
	)

	server := NewServer(app)

	response := sendRequest(
		t,
		server,
		Request{
			Command: CommandAddService,
		},
	)

	if response.Success {
		t.Fatal("expected unsuccessful response")
	}

	if response.Error != "service is required" {
		t.Fatalf(
			"expected %q, got %q",
			"service is required",
			response.Error,
		)
	}
}

func TestServerAddService(t *testing.T) {
	registry := &testutil.FakeRegistry{}

	engine := monitoring.NewEngine()

	app := application.New(
		registry,
		engine,
		testutil.FakeCheckerFactory{},
	)

	server := NewServer(app)

	response := sendRequest(
		t,
		server,
		Request{
			Command: CommandAddService,
			Service: &ServiceRequest{
				Name: "sshd",
				Type: "systemd",
			},
		},
	)

	if !response.Success {
		t.Fatalf(
			"expected successful response, got error: %s",
			response.Error,
		)
	}

	if !engine.HasChecker("sshd") {
		t.Fatal("expected sshd checker to be added")
	}

	if len(registry.Services) != 1 {
		t.Fatalf(
			"expected 1 registered service, got %d",
			len(registry.Services),
		)
	}
}

func TestServerEnsureServiceAddsService(t *testing.T) {
	registry := &testutil.FakeRegistry{}

	engine := monitoring.NewEngine()

	app := application.New(
		registry,
		engine,
		testutil.FakeCheckerFactory{},
	)

	server := NewServer(app)

	response := sendRequest(
		t,
		server,
		Request{
			Command: CommandEnsureService,
			Service: &ServiceRequest{
				Name: "nginx",
				Type: "systemd",
			},
		},
	)

	if !response.Success {
		t.Fatalf(
			"expected successful response, got error: %s",
			response.Error,
		)
	}

	if response.Added == nil {
		t.Fatal("expected added field")
	}

	if !*response.Added {
		t.Fatal("expected added to be true")
	}
}

func TestServerEnsureServiceDoesNotDuplicate(t *testing.T) {
	registry := &testutil.FakeRegistry{}

	engine := monitoring.NewEngine()

	app := application.New(
		registry,
		engine,
		testutil.FakeCheckerFactory{},
	)

	server := NewServer(app)

	request := Request{
		Command: CommandEnsureService,
		Service: &ServiceRequest{
			Name: "nginx",
			Type: "systemd",
		},
	}

	first := sendRequest(t, server, request)

	if !first.Success {
		t.Fatalf("first request failed: %s", first.Error)
	}

	second := sendRequest(t, server, request)

	if !second.Success {
		t.Fatalf("second request failed: %s", second.Error)
	}

	if second.Added == nil {
		t.Fatal("expected added field")
	}

	if *second.Added {
		t.Fatal("expected added to be false on second request")
	}

	if got := engine.CheckerCount(); got != 1 {
		t.Fatalf("expected 1 checker, got %d", got)
	}
}

func TestServerReload(t *testing.T) {
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

	server := NewServer(app)

	response := sendRequest(
		t,
		server,
		Request{
			Command: CommandReload,
		},
	)

	if !response.Success {
		t.Fatalf(
			"expected successful response, got error: %s",
			response.Error,
		)
	}

	if !engine.HasChecker("sshd") {
		t.Fatal("expected sshd checker after reload")
	}
}
