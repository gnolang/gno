package boardsv2

import (
	"strconv"
	"time"
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

	Post struct {
		ID            string
		Content       Content
		PluginStorage avl.Tree // string(plugin name) -> interface{}(plugin storage)
		Parent        *Post
		Level         int
		Base          *Post
		Children      []*Post
		Forks         []*Post
		UpdatedAt     time.Time
		CreatedAt     time.Time
		Creator       std.Address
	}
)

func (p Post) MustGetPluginStorage(name string) interface{} {
	if v, found := p.pluginStorage.Get(name); found {
		return v
	}

	panic("plugin storage not found: " + name)
}

func (p Post) NextIncrementalKey(baseKey string) string {
	return baseKey + "/" + strconv.Itoa(len(p.Children))
}
