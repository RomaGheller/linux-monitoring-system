package api

import (
	"encoding/json"
	"fmt"
	"net"

	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/application"
	"gitlab.com/linuxgamerpro/linux-monitoring-system/internal/config"
)

type Server struct {
	app *application.Application
}

func NewServer(app *application.Application) *Server {
	return &Server{
		app: app,
	}
}

func (s *Server) Handle(conn net.Conn) error {
	defer conn.Close()

	var request Request

	if err := json.NewDecoder(conn).Decode(&request); err != nil {
		return err
	}

	response := s.handleRequest(request)

	return json.NewEncoder(conn).Encode(response)
}

func (s *Server) handleRequest(request Request) Response {
	switch request.Command {
	case CommandListServices:
		return s.handleListServices()

	case CommandAddService:
		return s.handleAddService(request)

	case CommandEnsureService:
		return s.handleEnsureService(request)

	case CommandReload:
		return s.handleReload()

	default:
		return Response{
			Success: false,
			Error:   fmt.Sprintf("unknown command: %q", request.Command),
		}
	}
}

func (s *Server) handleListServices() Response {
	services, err := s.app.ListServices()
	if err != nil {
		return Response{
			Success: false,
			Error:   err.Error(),
		}
	}

	result := make([]ServiceResponse, 0, len(services))

	for _, service := range services {
		result = append(result, ServiceResponse{
			Name: service.Name,
			Type: service.Type,
		})
	}

	return Response{
		Success:  true,
		Services: result,
	}
}

func (s *Server) handleAddService(request Request) Response {
	if request.Service == nil {
		return Response{
			Success: false,
			Error:   "service is required",
		}
	}

	service := config.ServiceConfig{
		Name: request.Service.Name,
		Type: request.Service.Type,
	}

	if err := s.app.AddService(service); err != nil {
		return Response{
			Success: false,
			Error:   err.Error(),
		}
	}

	return Response{
		Success: true,
	}
}

func (s *Server) handleEnsureService(request Request) Response {
	if request.Service == nil {
		return Response{
			Success: false,
			Error:   "service is required",
		}
	}

	service := config.ServiceConfig{
		Name: request.Service.Name,
		Type: request.Service.Type,
	}

	added, err := s.app.EnsureService(service)
	if err != nil {
		return Response{
			Success: false,
			Error:   err.Error(),
		}
	}

	return Response{
		Success: true,
		Added:   &added,
	}
}

func (s *Server) handleReload() Response {
	if err := s.app.Reload(); err != nil {
		return Response{
			Success: false,
			Error:   err.Error(),
		}
	}

	return Response{
		Success: true,
	}
}
