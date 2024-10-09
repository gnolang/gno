package post

import (
	"strconv"
	"time"

	"gno.land/p/demo/boards/post/plugin"
)

type (
	Post struct {
		ID          string
		PluginStore plugin.Plugin
		Parent      *Post
		Level       int
		Base        *Post
		Children    []*Post
		Forks       []*Post
		UpdatedAt   time.Time
		CreatedAt   time.Time
		Creator     std.Address
	}
)

func (p Post) NextIncrementalKey(baseKey string) string {
	return baseKey + "/" + strconv.Itoa(len(p.Children))
}
