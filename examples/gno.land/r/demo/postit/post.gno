package postit

import (
	"time"
)

type Post struct {
	Id       uint64
	Username string
	Body     string

	Comments []Comment

	CreatedAt time.Time
}
