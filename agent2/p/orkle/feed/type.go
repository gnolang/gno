package feed

type Type int

const (
	TypeStatic Type = iota
	TypeContinuous
	TypePeriodic
)
