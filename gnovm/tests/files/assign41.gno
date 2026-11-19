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

// Blanks interleaved with duplicate real targets: blanks must not
// disturb the left-to-right write order of the real LHS.
func blankInterleaved() {
	var a int
	a, _, a, _ = 1, 2, 3, 4
	if a != 3 {
		panic(a)
	}
}

// Multi-value RHS (single call returning multiple values) into a
// duplicate LHS. Goes through a different RHS push pattern (one
// OpEval for the call, multiple values produced).
func twoVal() (int, int) { return 1, 2 }

func multiReturnRHS() {
	var a int
	a, a = twoVal()
	if a != 2 {
		panic(a)
	}
}

// Multi-blank with mixed-type RHS. All blanks alias the same shared slot
// (Block.Blank), so the reorder changes which RHS value/type lands there
// last. This smoke-tests that the statement completes without crashing;
// the realm-finalization-level effect (assertTypeIsPublic against the
// surviving type) only manifests in realm code and is out of scope here.
func multiBlankMixedType() {
	_, _ = "hello", 42
	_, _ = 1, "two"
}

func main() {
	indexSlice()
	indexMap()
	starExpr()
	selectorExpr()
	blankInterleaved()
	multiReturnRHS()
	multiBlankMixedType()
	println("ok")
}

// Output:
// ok
