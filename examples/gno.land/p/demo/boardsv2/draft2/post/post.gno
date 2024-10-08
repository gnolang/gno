package boardsv2

import (
	"plugin"
	"strconv"
	"time"
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
