// TODO: Document how plugins work and best practices
package plugin

import (
	"gno.land/p/demo/avl"
)

type (
	// NOTE: Consider adding lifecycle methods like `Post` creation, deletion, ...
	Plugin interface {
		Name() string
		Render() string
	}

	Registry struct {
		plugins avl.Tree // string(name) -> Plugin
	}
)

func NewRegistry(plugins ...Plugin) *Registry {
	r := &Registry{}
	for _, p := range plugins {
		r.plugins.Set(p.Name(), p)
	}
	return r
}

func (r Registry) Has(name string) bool {
	return r.posts.Has(name)
}

func (r Registry) Get(name string) (_ Plugin, found bool) {
	if v, found := r.plugins.Get(name); found {
		return v.(Plugin), true
	}
	return nil, false
}

func (r *Registry) Add(p Plugin) {
	r.plugins.Set(p.Name(), p)
}

func (r *Registry) Remove(name string) (_ Plugin, removed bool) {
	if v, removed := r.plugins.Remove(name, p); removed {
		return v.(Plugin), false
	}
	return nil, false
}
