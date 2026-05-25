package main

// Concrete struct compare (outer not via interface) where a field's
// declared type is interface{} — viaInterface must propagate into
// the field so the uncomparable dynamic type panics.
type S struct{ a interface{} }

func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()
	var m map[int]int
	s1 := S{a: m}
	s2 := S{a: m}
	_ = s1 == s2
}

// Output:
// recover: comparing uncomparable type map[int]int
