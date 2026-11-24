package main

// A method body runs the converting assignment, for value and pointer
// receivers alike.

type NamedMap map[string]int

func fetch() (NamedMap, bool) { return NamedMap{"a": 1}, true }

type T struct{}

func (T) Method() (int, bool) {
	var m map[string]int
	var ok bool
	m, ok = fetch()
	return len(m), ok
}

func (*T) PMethod() (int, bool) {
	var m map[string]int
	var ok bool
	m, ok = fetch()
	return len(m), ok
}

func main() {
	println(T{}.Method())
	println((&T{}).PMethod())
}

// Output:
// 1 true
// 1 true
