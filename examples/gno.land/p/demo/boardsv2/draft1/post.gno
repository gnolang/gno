package boardsv2

import (
	"strconv"
)

const (
	LevelBoard = iota
	LevelPost
	LevelComment
)

type (
	Content interface {
		Type() string
		Render() string
	}

	// TODO: Still have to consider how to add custom features like up/down voting, reputation, fork, lock, ...
	Post struct {
		ID        string  // NOTE: It would be nice to use a type alias for the ID field type
		Content   Content // NOTE: Maybe should be called Type (board as content is odd)
		Parent    *Post   // NOTE: If all is a post we need to have a parent
		Level     int
		Base      *Post
		Children  []*Post
		Forks     []*Post
		UpdatedAt time.Time
		CreatedAt time.Time
		Creator   std.Address
	}
)

func (p Post) Slug() string { // NOTE: Not optimal, calculate slug dynamically
	if p.Parent == nil {
		return p.ID
	}
	return p.Parent.Slug() + "/" + p.ID
}

func (p Post) NextIncrementalKey(baseKey string) string {
	return baseKey + "/" + strconv.Itoa(len(p.Children))
}
