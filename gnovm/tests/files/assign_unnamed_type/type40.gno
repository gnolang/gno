package main

type Set map[string]struct{}

func NewSet(items ...string) Set {
	return map[string]struct{}{}
}

func (s Set) Has(key string) bool {
	_, ok := s[key]
	return ok
}

func main() {
	s := NewSet()
	println(s.Has("a"))
}

// Output:
// false
