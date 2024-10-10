package post

func NewStore() Store {
	return Store{}
}

// TODO: Implement posts store
type Store struct {
	posts avl.Tree // string(level + creation timestamp + slug) -> *Post
	slugs avl.Tree // string(slug) -> *Post
}

func (s Store) Get(path string) (_ *Post, found bool) {
	if v, found := s.slugs.Get(path); found {
		return v.(*Post), true
	}
	return nil, false
}
