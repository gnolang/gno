package pluginlock

import "errors"

const Name = "lock"

var ErrInvalidPostType = errors.New("post type is not a board or thread")

type (
	Plugin struct{}
	Lock   struct {
		IsLocked bool
	}
)

func New() Plugin {
	return Plugin{}
}

func (p Plugin) Name() string {
	return Name
}

func (p Plugin) Render() string {
	return ""
}

func (p *Plugin) Lock(pst *post.Post) error {
	if !isBoardOrThread(pst) {
		return ErrInvalidPostType
	}

	pst.PluginStore[Name].(*Lock).IsLocked = true
}

func (p *Plugin) Unlock(pst *post.Post) error {
	if !isBoardOrThread(pst) {
		return ErrInvalidPostType
	}

	pst.PluginStore[Name].(*Lock).IsLocked = false
}

func (p Plugin) IsLocked(pst *post.Post) bool {
	if !isBoardOrThread(pst) {
		return ErrInvalidPostType
	}

	// TODO: Check parents if current post is not locked
	return pst.PluginStore[Name].(*Lock).IsLocked
}

func isBoardOrThread(pst *post.Post) bool {
	return pst.Level == post.LevelBoard || pst.Level == post.LevelPost
}
