package systemd

type State string

const (
	StateActive       State = "active"
	StateInactive     State = "inactive"
	StateFailed       State = "failed"
	StateActivating   State = "activating"
	StateDeactivating State = "deactivating"
)
