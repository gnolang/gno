package main

// The converting assignment in a for-init runs once, before the condition:
// ok starts true, so the body never runs.

type NamedMap map[string]int

func fetch() (NamedMap, bool) { return NamedMap{"a": 1}, true }

func main() {
	var m map[string]int
	var ok bool
	trips, n := 0, 0
	for m, ok = fetch(); !ok; n++ {
		trips++
	}
	println("trips", trips, len(m), ok)
}

// Output:
// trips 0 1 true
