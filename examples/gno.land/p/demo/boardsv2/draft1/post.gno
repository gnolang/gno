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
		ID        string
		Content   Content // NOTE: Maybe should be Type (board as content is odd)
		Level     int
		Base      *Post
		Children  []*Post
		Forks     []*Post
		UpdatedAt time.Time
		CreatedAt time.Time
		Creator   std.Address
	}
)

func (p Post) NextIncrementalKey(baseKey string) string {
	return baseKey + "/" + strconv.Itoa(len(p.Children))
}
