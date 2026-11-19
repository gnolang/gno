package main

// `var a, b T = f()` converts each result to T when the two differ only by
// name, as `var a T = x` does; an interface T boxes the result unchanged.

type NamedMap map[string]int

func two() (NamedMap, NamedMap)              { return NamedMap{"a": 1}, NamedMap{"b": 2} }
func twoU() (map[string]int, map[string]int) { return map[string]int{"a": 1}, map[string]int{"b": 2} }

func kind(x any) string {
	switch x.(type) {
	case NamedMap:
		return "NamedMap"
	case map[string]int:
		return "map"
	}
	return "other"
}

// viaReturn is declared to return T, so the declaration below converts
// nothing, whatever type `return two()` leaves on the results; they still end
// up typed map[string]int, as in Go.
func viaReturn() (map[string]int, map[string]int) { return two() }

var pa, pb map[string]int = two()

func main() {
	var a, b map[string]int = two()
	println("named to unnamed", kind(a), kind(b))
	var c, d NamedMap = twoU()
	println("unnamed to named", kind(c), kind(d))
	var e, f any = two()
	println("interface keeps", kind(e), kind(f))
	println("package level", kind(pa), kind(pb))
	var g, h map[string]int = viaReturn()
	println("via return", kind(g), kind(h))
	m := map[string]bool{"k": true}
	var v, ok bool = m["k"]
	println("comma-ok", v, ok)
}

// Output:
// named to unnamed map map
// unnamed to named NamedMap NamedMap
// interface keeps NamedMap NamedMap
// package level map map
// via return map map
// comma-ok true true
