package feed

import "time"

// Value represents a value published by a feed. The `Time` is when the value was published.
type Value struct {
	String string
	Time   time.Time
}
