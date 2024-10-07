package boardsv2

import (
	"strconv"
	"time"
)

type (
	Body interface {
		Type() string // NOTE: The idea is to avoid casting
	}

	// TODO: ID is the full path to the post

	Post struct {
		ID            string   // NOTE: It would be nice to use a type alias for the ID field type
		Body       Body  // NOTE: Maybe should be called Type (board as content is odd)
		Parent        *Post    // NOTE: If all is a post we need to have a parent
		Level         int
		Base          *Post
		Children      []*Post
		Forks         []*Post
		UpdatedAt     time.Time
		CreatedAt     time.Time
		Creator       std.Address
	}
)

func (p Post) NextIncrementalKey(baseKey string) string {
	return baseKey + "/" + strconv.Itoa(len(p.Children))
}
