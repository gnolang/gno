package main

// Check that array and struct equality compares fields in declaration
// order and short-circuits on the first inequality. When the differing
// field is reached first, no panic occurs even if a later field has an
// uncomparable dynamic type via interface. When the uncomparable field
// is reached first (equal on both sides, or before any inequality),
// the comparison must panic.
//
// Mirrors Go's test/fixedbugs/issue8606.go.

type A [2]interface{}
type A2 [6]interface{}
type S struct{ x, y interface{} }
type S2 struct{ x, y, z, a, b, c interface{} }
type T1 struct {
	i interface{}
	a int64
	j interface{}
}
type T2 struct {
	i       interface{}
	a, b, c int64
	j       interface{}
}
type T3 struct {
	i interface{}
	s string
	j interface{}
}
type S3 struct {
	f interface{}
	i int
}
type S4 struct {
	a [1000]byte
	b interface{}
}

func main() {
	b := []byte{1}
	s1 := S3{func() {}, 0}
	s2 := S3{func() {}, 1}

	cases := []struct {
		name  string
		panic bool
		a, b  interface{}
	}{
		{"A_1b_2b", false, A{1, b}, A{2, b}},
		{"A_b1_b2", true, A{b, 1}, A{b, 2}},
		{"A_1b_s2b", false, A{1, b}, A{"2", b}},
		{"A_b1_bs2", true, A{b, 1}, A{b, "2"}},

		{"A2_1b_2b", false, A2{1, b}, A2{2, b}},
		{"A2_b1_b2", true, A2{b, 1}, A2{b, 2}},
		{"A2_1b_s2b", false, A2{1, b}, A2{"2", b}},
		{"A2_b1_bs2", true, A2{b, 1}, A2{b, "2"}},

		{"S_1b_2b", false, S{1, b}, S{2, b}},
		{"S_b1_b2", true, S{b, 1}, S{b, 2}},
		{"S_1b_s2b", false, S{1, b}, S{"2", b}},
		{"S_b1_bs2", true, S{b, 1}, S{b, "2"}},

		{"S2_1b_2b", false, S2{x: 1, y: b}, S2{x: 2, y: b}},
		{"S2_b1_b2", true, S2{x: b, y: 1}, S2{x: b, y: 2}},
		{"S2_1b_s2b", false, S2{x: 1, y: b}, S2{x: "2", y: b}},
		{"S2_b1_bs2", true, S2{x: b, y: 1}, S2{x: b, y: "2"}},

		{"T1_b1_b2", true, T1{i: b, a: 1}, T1{i: b, a: 2}},
		{"T1_1b_2b", false, T1{a: 1, j: b}, T1{a: 2, j: b}},
		{"T2_b1_b2", true, T2{i: b, a: 1}, T2{i: b, a: 2}},
		{"T2_1b_2b", false, T2{a: 1, j: b}, T2{a: 2, j: b}},
		{"T3_bfoo_bbar", true, T3{i: b, s: "foo"}, T3{i: b, s: "bar"}},
		{"T3_foob_barb", false, T3{s: "foo", j: b}, T3{s: "bar", j: b}},
		{"T3_bfooz_bbar", true, T3{i: b, s: "fooz"}, T3{i: b, s: "bar"}},
		{"T3_foozb_barb", false, T3{s: "fooz", j: b}, T3{s: "bar", j: b}},
		{"A_s1s2_s2s1", true, A{s1, s2}, A{s2, s1}},
		{"s1_s2", true, s1, s2},
		{"S4_0f_1f", false, S4{[1000]byte{0}, func() {}}, S4{[1000]byte{1}, func() {}}},
	}

	for _, tc := range cases {
		f := func() {
			defer func() {
				if recover() != nil {
					panic("comparing " + tc.name + " panicked unexpectedly")
				}
			}()
			if tc.a == tc.b {
				panic("values " + tc.name + " should not be equal")
			}
		}
		if tc.panic {
			shouldPanic("comparing "+tc.name+" did not panic", f)
		} else {
			f()
		}
	}
	println("OK")
}

func shouldPanic(name string, f func()) {
	defer func() {
		if recover() == nil {
			panic(name)
		}
	}()
	f()
}

// Output:
// OK
