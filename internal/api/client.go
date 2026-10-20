package api

import (
	"encoding/json"
	"errors"
	"net"
)

type Client struct {
	socketPath string
}

func NewClient(socketPath string) (*Client, error) {
	if socketPath == "" {
		return nil, errors.New("socket path cannot be empty")
	}

	return &Client{
		socketPath: socketPath,
	}, nil
}

func NewDefaultClient() (*Client, error) {
	return NewClient(DefaultSocketPath)
}

func (c *Client) do(request Request) (Response, error) {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return Response{}, err
	}
	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(request); err != nil {
		return Response{}, err
	}

	var response Response

	if err := json.NewDecoder(conn).Decode(&response); err != nil {
		return Response{}, err
	}

	return response, nil
}

func (c *Client) ListServices() ([]ServiceResponse, error) {
	response, err := c.do(Request{
		Command: CommandListServices,
	})
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, errors.New(response.Error)
	}

	return response.Services, nil
}

func (c *Client) AddService(service ServiceRequest) error {
	response, err := c.do(Request{
		Command: CommandAddService,
		Service: &service,
	})
	if err != nil {
		return err
	}

	if !response.Success {
		return errors.New(response.Error)
	}

	return nil
}

func (c *Client) EnsureService(service ServiceRequest) (bool, error) {
	response, err := c.do(Request{
		Command: CommandEnsureService,
		Service: &service,
	})
	if err != nil {
		return false, err
	}

	if !response.Success {
		return false, errors.New(response.Error)
	}

	if response.Added == nil {
		return false, errors.New("response missing added field")
	}

	return *response.Added, nil
}

func (c *Client) Reload() error {
	response, err := c.do(Request{
		Command: CommandReload,
	})
	if err != nil {
		return err
	}

	if !response.Success {
		return errors.New(response.Error)
	}

	return nil
}
