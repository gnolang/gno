package feed

type Task interface {
	MarshalToJSON() ([]byte, error)
}
