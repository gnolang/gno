package ingester

// Type indicates an ingester type.
type Type int

const (
	// TypeSingle indicates an ingester that can only ingest a single within a given period or no period.
	TypeSingle Type = iota
	// TypeMulti indicates an ingester that can ingest multiple within a given period or no period
	TypeMulti
)
