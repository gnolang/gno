package pluginreputation

import (
	"errors"
	"std"
)

// NOTE: Think about implementing a reputation based policy
const (
	PolicyLinear Policy = iota
	PolicyTokenBased
)

const Name = "reputation"

var ErrNotSupported = errors.New("reputation not supported")

type (
	Policy int

	Plugin struct {
		Store             Store
		Policy            Policy
		AllowedPostLevels []int
	}

	Reputation struct {
		Upvotes   uint
		Downvotes uint
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

func (p Plugin) Votes(pst *post.Post) (upvotes uint64, downvotes uint64) {
	if !p.HasReputationSupport(pst) {
		return ErrNotSupported
	}

	r := pst.PluginStore[Name].(*Reputation)
	return r.Upvotes, r.Downvotes
}

func (p Plugin) Voters(pst *post.Post) []std.Address {
	if !p.HasReputationSupport(pst) {
		return ErrNotSupported
	}

	// TODO: Implement support for tracking voters
}

func (p *Plugin) Upvote(pst *post.Post) error {
	if !p.HasReputationSupport(pst) {
		return ErrNotSupported
	}

	// TODO: Handle accounts and change downvotes for existing accounts that downvoted
	r := pst.PluginStore[Name].(*Reputation)
	r.Upvotes++
	p.store.inc(pst.ID)
}

func (p *Plugin) Downvote(pst *post.Post) error {
	if !p.HasReputationSupport(pst) {
		return ErrNotSupported
	}

	// TODO: Handle accounts and change upvotes for existing accounts that upvoted
	r := pst.PluginStore[Name].(*Reputation)
	r.Downvotes++
	p.store.dec(pst.ID)
}
