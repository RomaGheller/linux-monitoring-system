package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeEngine struct {
	started bool
	stopped bool
}

func (e *fakeEngine) Start() {
	e.started = true
}

func (e *fakeEngine) Stop() {
	e.stopped = true
}

type fakeAPIServer struct {
	serveStarted chan struct{}
	closeCalled  chan struct{}
	serveDone    chan struct{}

	serveErr error
	closeErr error
}

func newFakeAPIServer() *fakeAPIServer {
	return &fakeAPIServer{
		serveStarted: make(chan struct{}),
		closeCalled:  make(chan struct{}),
		serveDone:    make(chan struct{}),
	}
}

func (s *fakeAPIServer) Serve() error {
	close(s.serveStarted)

	<-s.serveDone

	return s.serveErr
}

func (s *fakeAPIServer) Close() error {
	close(s.closeCalled)
	close(s.serveDone)

	return s.closeErr
}

func (s *fakeAPIServer) stopServing() {
	close(s.serveDone)
}

func TestRunDaemonGracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	engine := &fakeEngine{}
	server := newFakeAPIServer()

	done := make(chan error, 1)

	go func() {
		done <- runDaemon(ctx, engine, server)
	}()

	<-server.serveStarted

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("runDaemon() error = %v", err)
		}

	case <-time.After(time.Second):
		t.Fatal("runDaemon() did not stop")
	}

	if !engine.started {
		t.Error("engine was not started")
	}

	if !engine.stopped {
		t.Error("engine was not stopped")
	}

	select {
	case <-server.closeCalled:
		// Expected.
	default:
		t.Error("API server was not closed")
	}
}

func TestRunDaemonReturnsErrorWhenServerFails(t *testing.T) {
	ctx := context.Background()

	engine := &fakeEngine{}
	server := newFakeAPIServer()

	serveErr := errors.New("listener failure")
	server.serveErr = serveErr

	done := make(chan error, 1)

	go func() {
		done <- runDaemon(ctx, engine, server)
	}()

	<-server.serveStarted

	server.stopServing()

	select {
	case err := <-done:
		if !errors.Is(err, serveErr) {
			t.Fatalf(
				"runDaemon() error = %v, want error wrapping %v",
				err,
				serveErr,
			)
		}

	case <-time.After(time.Second):
		t.Fatal("runDaemon() did not stop")
	}

	if !engine.started {
		t.Error("engine was not started")
	}

	if !engine.stopped {
		t.Error("engine was not stopped")
	}
}

func TestRunDaemonReturnsErrorWhenServerStopsUnexpectedly(t *testing.T) {
	ctx := context.Background()

	engine := &fakeEngine{}
	server := newFakeAPIServer()

	done := make(chan error, 1)

	go func() {
		done <- runDaemon(ctx, engine, server)
	}()

	<-server.serveStarted

	server.stopServing()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("runDaemon() error = nil, want error")
		}

	case <-time.After(time.Second):
		t.Fatal("runDaemon() did not stop")
	}

	if !engine.stopped {
		t.Error("engine was not stopped")
	}
}

func TestRunDaemonReturnsErrorWhenServerCloseFails(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	engine := &fakeEngine{}
	server := newFakeAPIServer()

	closeErr := errors.New("close failure")
	server.closeErr = closeErr

	done := make(chan error, 1)

	go func() {
		done <- runDaemon(ctx, engine, server)
	}()

	<-server.serveStarted

	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, closeErr) {
			t.Fatalf(
				"runDaemon() error = %v, want error wrapping %v",
				err,
				closeErr,
			)
		}

	case <-time.After(time.Second):
		t.Fatal("runDaemon() did not stop")
	}

	if !engine.started {
		t.Error("engine was not started")
	}

	if !engine.stopped {
		t.Error("engine was not stopped")
	}
}
