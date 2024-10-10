package pluginreputation

import "errors"

const (
	PolicyLinear Policy = iota
	PolicyTokenBased
)

const Name = "reputation"

var ErrNotSupported = errors.New("reputation not supported")

type (
	Policy int

	// TODO: Implement reputation plugin
	Plugin struct {
		Policy            Policy
		AllowedPostLevels []int
	}

	Reputation struct {
		Upvotes   uint
		Downbotes uint
	}
)

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
	return ""
}

func (p Plugin) HasReputationSupport(pst *post.Post) bool {
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

func (p *Plugin) Votes(pst *post.Post) uint32 {
	if !p.HasReputationSupport(pst) {
		return ErrNotSupported
	}

	// TODO: Implement
}

func (p *Plugin) Upvote(pst *post.Post) error {
	if !p.HasReputationSupport(pst) {
		return ErrNotSupported
	}

	// TODO: Modify global state
	// TODO: Modify local state
	r := pst.PluginStore[Name].(*Reputation)
}

func (p *Plugin) Downvote(pst *post.Post) error {
	if !p.HasReputationSupport(pst) {
		return ErrNotSupported
	}

	// TODO: Implement
}
