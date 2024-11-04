package pluginreputation

import (
	"gno.land/p/demo/seqid"
)

type (
	VotesIterFn = func(votes uint64, path string) bool

	Store struct {
		votes avl.Tree // string(count) -> string(path)
	}
)

func (s Store) Iterate(fn VotesIterFn) bool {
	// TODO: Support pagination of votes?
	return s.votes.Iterate("", "", func(key string, v interface{}) bool {
		count, _ := seqid.FromBinary(key)
		return fn(uint64(count), v.(string))
	})
}

func (s *Store) inc(path string) uint64 {
	var (
		current  seqid.ID
		v, found = s.votes.Get(path)
	)
	if found {
		current = v.(seqid.ID)
		// TODO: Implement the right solution because this is not right, just showcase
		s.votes.Remove(current.Binary())
	}

	current.Next()
	s.votes.Set(current.Binary(), path)
	return uint64(current)
}

func (s *Store) dec(path string) uint64 {
	var (
		current  seqid.ID
		v, found = s.votes.Get(path)
	)
	if found {
		current = v.(seqid.ID)
	}

	if current == 0 {
		return current
	}

	s.votes.Remove(current.Binary())
	current--
	if current != 0 {
		s.votes.Set(current.Binary(), current)
	}
	return uint64(current)
}
