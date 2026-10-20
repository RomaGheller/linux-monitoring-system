package api

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/application"
)

func TestUnixServer_CloseStopsServe(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "monitoring.sock")

	app := &application.Application{}
	server := NewServer(app)

	unixServer, err := NewUnixServer(
		UnixServerConfig{
			Path: socketPath,
			Mode: 0660,
		},
		server,
	)

	if err != nil {
		t.Fatalf("NewUnixServer() error = %v", err)
	}

	serveDone := make(chan error, 1)

	go func() {
		serveDone <- unixServer.Serve()
	}()

	if err := unixServer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	select {
	case err := <-serveDone:
		if err != nil {
			t.Fatalf("Serve() error = %v, want nil", err)
		}

	case <-time.After(time.Second):
		t.Fatal("Serve() did not stop after Close()")
	}
}

func TestUnixServer_CloseRemovesSocket(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "monitoring.sock")

	app := &application.Application{}
	server := NewServer(app)

	unixServer, err := NewUnixServer(
		UnixServerConfig{
			Path: socketPath,
			Mode: 0660,
		},
		server,
	)

	if err != nil {
		t.Fatalf("NewUnixServer() error = %v", err)
	}

	if _, err := os.Stat(socketPath); err != nil {
		t.Fatalf("socket does not exist after NewUnixServer(): %v", err)
	}

	if err := unixServer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	_, err = os.Stat(socketPath)

	if err == nil {
		t.Fatal("socket still exists after Close()")
	}

	if !os.IsNotExist(err) {
		t.Fatalf("Stat() error = %v, want os.ErrNotExist", err)
	}
}

func TestUnixServer_CloseIsIdempotent(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "monitoring.sock")

	app := &application.Application{}
	server := NewServer(app)

	unixServer, err := NewUnixServer(
		UnixServerConfig{
			Path: socketPath,
			Mode: 0660,
		},
		server,
	)

	if err != nil {
		t.Fatalf("NewUnixServer() error = %v", err)
	}

	if err := unixServer.Close(); err != nil {
		t.Fatalf("first Close() error = %v", err)
	}

	if err := unixServer.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
}

func TestNewUnixServerSetsSocketPermissions(t *testing.T) {
	socketPath := filepath.Join(
		t.TempDir(),
		"monitoring.sock",
	)

	server, err := NewUnixServer(
		UnixServerConfig{
			Path: socketPath,
			Mode: 0660,
		},
		&Server{},
	)
	if err != nil {
		t.Fatalf("NewUnixServer() error = %v", err)
	}
	defer server.Close()

	info, err := os.Stat(socketPath)
	if err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}

	if got := info.Mode().Perm(); got != 0660 {
		t.Errorf(
			"socket permissions = %04o, want %04o",
			got,
			os.FileMode(0660),
		)
	}
}
