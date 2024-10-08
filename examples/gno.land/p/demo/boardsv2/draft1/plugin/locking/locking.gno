package lockingplugin

import "errors"

var ErrInvalidPostType = errors.New("post type is not a board or thread")

type (
	Plugin  struct{}
	Storage struct {
		IsLocked bool
	}
)

func New() Plugin {
	return Plugin{}
}

func (p Plugin) Name() string {
	return "boards:locking"
}

func (p *Plugin) Lock(p *Post) error {
	if !isBoardOrThread(p) {
		return ErrInvalidPostType
	}

	p.MustGetPluginStorage(p.Name()).(*Storage).IsLocked = true
}

func (p *Plugin) Unlock(p *Post) error {
	if !isBoardOrThread(p) {
		return ErrInvalidPostType
	}

	p.MustGetPluginStorage(p.Name()).(*Storage).IsLocked = false
}

func (p Plugin) IsLocked(p *Post) bool {
	if !isBoardOrThread(p) {
		return ErrInvalidPostType
	}

	return p.MustGetPluginStorage(p.Name()).(*Storage).IsLocked
}

func isBoardOrThread(p *Post) bool {
	return p.Level == LevelBoard || p.Level == LevelPost
}
