// Extends assign40 to the other LHS shapes that go through PopAsPointer:
// IndexExpr (slice), IndexExpr (map), StarExpr, SelectorExpr. Each duplicate
// LHS must observe left-to-right assignment (Go spec §Assignments).
package main

type T struct{ f int }

func indexSlice() {
	s := []int{0, 0}
	s[0], s[0], s[0] = 1, 2, 3
	if s[0] != 3 {
		panic(s[0])
	}
}

func indexMap() {
	m := map[string]int{}
	m["k"], m["k"], m["k"] = 1, 2, 3
	if m["k"] != 3 {
		panic(m["k"])
	}
}

func starExpr() {
	x := 0
	p := &x
	*p, *p, *p = 1, 2, 3
	if x != 3 {
		panic(x)
	}
}

func selectorExpr() {
	s := &T{}
	s.f, s.f, s.f = 1, 2, 3
	if s.f != 3 {
		panic(s.f)
	}
}

func main() {
	indexSlice()
	indexMap()
	starExpr()
	selectorExpr()
	println("ok")
}

// Output:
// ok
