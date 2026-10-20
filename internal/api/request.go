package api

const (
	CommandListServices  = "services.list"
	CommandAddService    = "services.add"
	CommandEnsureService = "services.ensure"
	CommandReload        = "reload"
)

type Request struct {
	Command string          `json:"command"`
	Service *ServiceRequest `json:"service,omitempty"`
}

type ServiceRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
