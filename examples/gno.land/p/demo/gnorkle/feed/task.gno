package feed

// Task is a unit of work that can be part of a `Feed` definition. Tasks
// are executed by agents.
type Task interface {
	MarshalJSON() ([]byte, error)
}
