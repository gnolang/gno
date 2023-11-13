package postit

import (
	"time"
)

type Comment struct {
	parent    uint64
	username  string
	body      string
	timestamp time.Time
}
