package boards

import "plugin"

type Context struct {
	opts  []Option
	st    post.Storage
	plugs map[post.PluginName]plugin.Plugin
}

func newContext() Context {

}

func (c Context) Plugin(n post.PluginName) post.Plugin {

}

func (c Context) Set(p *Post) (updated bool) {
	key := newKey(p.Level, p.Slug())
	return b.posts.Set(key, p)
}

func (c Context) Remove(level int, path string) (_ *Post, removed bool) {
	key := newKey(level, path)
	if v, removed := b.posts.Remove(key); removed {
		return v.(*Post), true
	}
	return nil, false
}

func (c Context) Get(level int, path string, iterator func()) (_ *Post, found bool) {
	key := newKey(level, path)
	if v, found := b.posts.Get(key); found {
		return v.(*Post), true
	}
	return "", false
}
