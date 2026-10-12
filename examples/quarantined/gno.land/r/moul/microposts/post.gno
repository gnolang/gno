package microposts

import (
	"time"
)

type Post struct {
	text      string
	author    address
	createdAt time.Time
}

func (p Post) String() string {
	out := p.text + "\n"
	out += "_" + p.createdAt.Format("02 Jan 2006, 15:04") + ", by " + p.author.String() + "_"
	return out
}
