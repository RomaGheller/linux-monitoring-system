package monitoring

type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusError     Status = "error"
)

type CheckResult struct {
	Name    string
	Status  Status
	State   string
	Message string
}

type Checker interface {
	Name() string
	Check() CheckResult
}
