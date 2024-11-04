package reputationplugin

import "errors"

const (
	PolicyLinear = iota
	PolicyTokenBase
)

const Name = "boards:reputation"

var ErrNotSupported = errors.New("reputation not supported")

type (
	Plugin struct {
		Policy            int
		AllowedPostLevels []int
	}

	Storage struct {
		Upvotes            uint
		Downbotes          uint
		ListOfWhoVotedWhat avl.Tree // string(std.Address) -> ?? (TODO: define)
	}
)

func NewReputationPlugin(o ...Option) Plugin {
	var p Plugin
	for _, apply := range o {
		apply(&p)
	}
	return p
}

func (p Plugin) Name() string {
	return Name
}

func (p Plugin) HasReputationSupport(p *Post) bool {
	if len(p.AllowedPostLevels) == 0 {
		return true
	}

	for _, lvl := range p.AllowedPostLevels {
		if p.Level == lvl {
			return true
		}
	}
	return false
}

func (p *Plugin) Votes(p *Post) uint32 {
	if !p.HasReputationSupport(p) {
		return ErrNotSupported
	}

	// TODO: Implement
}

func (p *Plugin) Upvote(p *Post) error {
	if !p.HasReputationSupport(p) {
		return ErrNotSupported
	}

	// TODO: Modify global state
	// TODO: Modify local state
	// TODO: Implement
	st := p.MustGetPluginStorage(p.Name()).(*Storage)
}

func (p *Plugin) Downvote(p *Post) error {
	if !p.HasReputationSupport(p) {
		return ErrNotSupported
	}

	// TODO: Implement
}
