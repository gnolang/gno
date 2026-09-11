package feed

// Type indicates the type of a feed.
type Type int

const (
	// TypeStatic indicates a feed cannot be changed once the first value is committed.
	TypeStatic Type = iota
	// TypeContinuous indicates a feed can continuously ingest values and will publish
	// a new value on request using the values it has ingested.
	TypeContinuous
	// TypePeriodic indicates a feed can accept one or more values within a certain period
	// and will proceed to commit these values at the end up each period to produce an
	// aggregate value before starting a new period.
	TypePeriodic
)
