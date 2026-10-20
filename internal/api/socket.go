package api

import (
	"errors"
	"fmt"
	"net"
	"os"
)

const DefaultSocketPath = "/run/monitoring/monitoring.sock"

type UnixServer struct {
	listener   net.Listener
	server     *Server
	socketPath string
}

type UnixServerConfig struct {
	Path string
	Mode os.FileMode
}

func NewUnixServer(
	config UnixServerConfig,
	server *Server,
) (*UnixServer, error) {
	if server == nil {
		return nil, errors.New("server cannot be nil")
	}

	if config.Path == "" {
		return nil, errors.New("socket path cannot be empty")
	}

	if err := os.Remove(config.Path); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf(
			"failed to remove existing socket: %w",
			err,
		)
	}

	listener, err := net.Listen("unix", config.Path)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to listen on Unix socket: %w",
			err,
		)
	}

	if err := os.Chmod(config.Path, config.Mode); err != nil {
		_ = listener.Close()
		_ = os.Remove(config.Path)

		return nil, fmt.Errorf(
			"failed to set socket permissions: %w",
			err,
		)
	}

	return &UnixServer{
		listener:   listener,
		server:     server,
		socketPath: config.Path,
	}, nil
}

func (s *UnixServer) Serve() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}

			return err
		}

		go func() {
			_ = s.server.Handle(conn)
		}()
	}
}

func (s *UnixServer) Close() error {
	if err := s.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return err
	}

	if err := os.Remove(s.socketPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
