package task

type Definition interface {
	MarshalJSON() ([]byte, error)
}
