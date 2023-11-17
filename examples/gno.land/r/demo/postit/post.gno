package postit

import (
	"time"
)

type Post struct {
	Id        uint64
	Username  string
	Body      string
	Timestamp time.Time
	Comments  []Comment
}
