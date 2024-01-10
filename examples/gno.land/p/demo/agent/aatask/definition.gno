package aatask

type Definition interface {
	MarshalJSON() ([]byte, error)
	Type() string
}
