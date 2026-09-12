// PKGPATH: gno.land/r/test
package test

type T struct{ x int }

func (t T) Get() int { return t.x }

var (
	M1  any // map with local-type key
	M2  any // map with local-type value
	SL  any // slice of local type
	P   any // pointer to local type
	RN  any // recursive local type
	NB  any // nested local types
	MI  any // primitive-based local type
	FT  any // local func type
	G   func() int
	ARR any // byte-based local type array
)

func main(cur realm) {
	// 1 map key
	type K struct{ k int }
	m := map[K]string{{1}: "one"}
	M1 = m
	println(m[K{1}])
	// 2 map value
	type V struct{ v int }
	mv := map[string]V{"a": {2}}
	M2 = mv
	println(mv["a"].v)
	// 3 slice of local type
	type S struct{ x int }
	sl := []S{{1}, {2}}
	SL = sl
	println(sl[0].x + sl[1].x)
	// 4 pointer to local type
	p := &S{5}
	P = p
	println(p.x)
	// 5 recursive local type
	type N struct {
		next *N
		v    int
	}
	n := &N{&N{nil, 2}, 1}
	RN = n
	sum := 0
	for ; n != nil; n = n.next {
		sum += n.v
	}
	println(sum)
	// 6 nested local types
	type A struct{ v int }
	type B struct{ a A }
	b := B{A{6}}
	NB = b
	println(b.a.v)
	// 7 primitive base
	type MyInt int
	MI = MyInt(42)
	println(int(MI.(MyInt)))
	// 8 local func type
	type F func() int
	var f F = func() int { return 8 }
	FT = f
	println(f())
	// 9 closure over var of local interface type
	type LI interface{ Get() int }
	var li LI = T{9}
	G = func() int { return li.Get() }
	println(G())
	// 10 byte-based local type array
	type MB byte
	arr := [3]MB{1, 2, 3}
	ARR = arr
	println(int(arr[0]) + int(arr[1]) + int(arr[2]))
}

// Output:
// one
// 2
// 3
// 5
// 3
// 6
// 42
// 8
// 9
// 6
