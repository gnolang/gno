package evaluation

// Points could be converted to rewards
type Points struct {
	total   int64
	factors map[string]int32
}

func NewPoints(t int64, f map[string]int32) Points {
	return Points{
		total:   t,
		factors: f,
	}
}
