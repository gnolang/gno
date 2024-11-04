package pluginfork

import (
	"gno.land/p/demo/boards/post"
)

const Name = "fork"

// TODO: Implement fork plugin to support thread forking
type Plugin struct {
	AllowedPostLevels []int
}

func New(o ...Option) Plugin {
	var p Plugin
	for _, apply := range o {
		apply(&p)
	}
	return p
}

func (p Plugin) Name() string {
	return Name
}

func (p Plugin) Render() string {
	// TODO: Implement render support for text
	return ""
}

func (p Plugin) HasForkSupport(pst *post.Post) bool {
	if len(p.AllowedPostLevels) == 0 {
		return true
	}

	for _, lvl := range p.AllowedPostLevels {
		if pst.Level == lvl {
			return true
		}
	}
	return false
}

func (p Plugin) Fork(pst *post.Post, newPath string) error {
	// TODO: Implement fork support
	return nil
}
