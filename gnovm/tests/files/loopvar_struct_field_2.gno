package main

// Struct composite literal: field names that exactly match for-loop init vars.
// Tests that TRANS_COMPOSITE_KEY NameExprs are not renamed to i.loopvar.
type S struct {
	i int
}

func main() {
	for i := 0; i < 3; i++ {
		s := S{i: i * 10}
		println(s.i)
	}
}

// Output:
// 0
// 10
// 20
