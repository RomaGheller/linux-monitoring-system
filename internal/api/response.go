package api

type Response struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`

	Services []ServiceResponse `json:"services,omitempty"`
	Added    *bool             `json:"added,omitempty"`
}

type ServiceResponse struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
