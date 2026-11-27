package gen

func languageItems() []item {
	return []item{
		{"float", "trunc", mainOnly(`
	var f float64 = 1.9
	println(int(f))
	var g float64 = -1.9
	println(int(g))
	var h float32 = 2.9
	println(int(h))
	var n float32 = -2.9
	println(int(n))
	pos := 1.25
	println(int(pos * 4))
`)},
		{"float", "nan_inf", mainOnly(`
	var z float64 = 0
	nan := z / z
	println(nan == nan)
	println(nan != nan)
	pos := 1 / z
	neg := -1 / z
	println(pos > 1e300)
	println(pos == pos)
	println(neg < -1e300)
	println(neg == neg)
	println(pos > neg)
	nz := -z
	println(z == nz)
`)},
		{"float", "cmp", mainOnly(`
	var a float64 = 1
	var b float64 = 1.5
	println(a < b)
	println(a == a)
	println(b > 2)
	var z float64
	println(z == 0)
`)},
		{"const", "untyped", `package main

const (
	shift = 1 << 40
	div   = 10 / 3
	mix   = 2*3 + 4
	bigf  = 1e2
	huge  = 1 << 62
)

func main() {
	println(int64(shift))
	println(div)
	println(mix)
	println(int(bigf))
	println(int64(huge))
}
`},
		{"const", "overflow_int8", `package main

func main() {
	const x int8 = 128
	println(x)
}
`},
		{"const", "overflow_shift", `package main

const huge = 1 << 100

func main() {
	var x int64 = huge
	println(x)
}
`},
		{"const", "typed_vs_var", mainOnly(`
	const c int8 = 3
	var v int8 = 3
	println(c+1, v+1)
	const div = -10 / 3
	var a int = -10
	var b int = 3
	println(div, a/b, a%b)
`)},
		{"string", "utf8", mainOnly(`
	s := "héllo"
	println(len(s))
	println(s[0] == 'h')
	n := 0
	var last rune
	for _, r := range s {
		n++
		last = r
	}
	println(n, last == 'o')
	b := []byte(s)
	b[0] = 'H'
	println(s == "héllo")
	println(string(rune(65)))
	println(string(rune(0x1F600)) == "😀")
	println(string(rune(0xD800)) == "\uFFFD")
	var big uint64 = 0x110000
	println(string(big) == "\uFFFD")
	println(string(-1) == "\uFFFD")
	println(string(rune(0x10FFFF)) != "\uFFFD")
`)},
		{"string", "invalid_utf8", mainOnly(`
	raw := string([]byte{0xff, 0xfe, 'A'})
	repl := 0
	var sizes int
	for _, r := range raw {
		sizes++
		if r == '\uFFFD' {
			repl++
		}
	}
	println(len(raw), repl, sizes)
	println(raw[2] == 'A')
	println(string([]byte{0x00}) == "\x00")
`)},
		{"string", "slice_index", mainOnly(`
	s := "héllo"
	println(s[:1] == "h")
	println(len(s[1:3]))
	println(s[len(s)-1] == 'o')
	empty := ""
	println(len(empty), empty == "")
`)},
		{"string", "index_panic", mainOnly(`
	s := "ab"
	println("before")
	println(s[5])
`)},
		{"slice", "append_alias", mainOnly(`
	var s []int
	println(len(s), cap(s), s == nil)
	s = append(s, 1, 2, 3)
	println(len(s), s[0], s[2])
	t := s[:2]
	t[0] = 9
	println(s[0], len(t))
	u := append([]int{}, s...)
	u[0] = 1
	println(s[0], u[0], u[2])
`)},
		{"slice", "full_expr", mainOnly(`
	a := [5]int{1, 2, 3, 4, 5}
	s := a[1:3:4]
	println(len(s), cap(s), s[0], s[1])
	s[0] = 9
	println(a[1])
	b := a[:]
	println(len(b), cap(b), b[4])
`)},
		{"slice", "copy_overlap", mainOnly(`
	s := []int{1, 2, 3, 4}
	n := copy(s[1:], s[0:3])
	println(n, s[0], s[1], s[2], s[3])
	d := []int{0, 0}
	println(copy(d, []int{7, 8, 9}), d[0], d[1])
	println(copy([]int{}, []int{1}))
`)},
		{"slice", "nil_index", mainOnly(`
	var s []int
	println("before")
	println(s[0])
`)},
		{"slice", "make_bounds", mainOnly(`
	n, c := 2, 1
	println("before")
	_ = make([]int, n, c)
`)},
		{"slice", "reslice_low_high", mainOnly(`
	s := []int{1, 2, 3}
	lo, hi := 2, 1
	println("before")
	_ = s[lo:hi]
`)},
		{"map", "nil_ops", mainOnly(`
	var m map[string]int
	println(m == nil, len(m))
	_, ok := m["a"]
	println(ok, m["a"])
	delete(m, "a")
	println(len(m))
	sum := 0
	for _, v := range m {
		sum += v
	}
	println(sum)
`)},
		{"map", "lookup_delete", mainOnly(`
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	sum := 0
	keys := 0
	for k, v := range m {
		sum += v
		if k != "" {
			keys++
		}
	}
	println(sum, keys, len(m))
	println(m["a"], m["missing"])
	delete(m, "b")
	delete(m, "missing")
	_, ok := m["b"]
	println(len(m), ok, m["c"])
`)},
		{"map", "delete_during_range", mainOnly(`
	m := map[int]int{1: 10, 2: 20, 3: 30}
	for k := range m {
		if k == 2 {
			delete(m, k)
		}
	}
	println(len(m), m[1], m[2], m[3])
`)},
		{"struct", "value_copy", `package main

type pair struct {
	X int
	Y int
}

type box struct {
	Ns []int
}

func main() {
	a := pair{1, 2}
	b := a
	b.X = 9
	println(a.X, a.Y, b.X, b.Y)
	c := &a
	c.Y = 8
	println(a.Y, c.X)
	arr := [2]int{1, 2}
	arr2 := arr
	arr2[0] = 9
	println(arr[0], arr2[0])
	println(arr == [2]int{1, 2})
	s := box{Ns: []int{1, 2}}
	t := s
	t.Ns[0] = 5
	println(s.Ns[0], t.Ns[1])
	println(pair{1, 2} == pair{1, 2})
	println(pair{1, 2} == pair{2, 2})
}
`},
		{"struct", "array_compare", mainOnly(`
	var a [3]int
	b := [3]int{0, 1, 0}
	println(a == b)
	a[1] = 1
	println(a == b)
	var z [0]int
	var y [0]int
	println(z == y)
`)},
		{"iface", "nil_and_assert", `package main

type tally struct{ N int }

func (t tally) V() int { return t.N }

func main() {
	var i interface{} = tally{3}
	v, ok := i.(tally)
	println(ok, v.V())
	_, ok = i.(*tally)
	println(ok)
	var p *tally
	var j interface{} = p
	println(p == nil, j == nil)
	var k interface{}
	println(k == nil)
	_, ok = k.(tally)
	println(ok)
	switch i.(type) {
	case tally:
		println(1)
	case *tally:
		println(2)
	default:
		println(3)
	}
	switch k.(type) {
	case nil:
		println(0)
	default:
		println(9)
	}
}
`},
		{"iface", "embedded", `package main

type reader interface{ Read() int }

type both interface {
	reader
	Write() int
}

type mem struct{ n int }

func (m mem) Read() int  { return m.n }
func (m mem) Write() int { return m.n + 1 }

func main() {
	var b both = mem{4}
	println(b.Read(), b.Write())
	r := b.Read
	println(r())
	_, ok := interface{}(b).(reader)
	println(ok)
}
`},
		{"method", "pointer_receiver", `package main

type cell struct{ N int }

func (c cell) Val() int { return c.N }
func (c *cell) Add(n int) int {
	c.N += n
	return c.N
}

func main() {
	c := cell{2}
	println(c.Val())
	println(c.Add(3))
	println(c.N)
	f := c.Val
	println(f())
	g := (*cell).Add
	println(g(&c, 1))
	println(c.N)
}
`},
		{"method", "nil_pointer", `package main

type cell struct{ N int }

func (c *cell) Twice() int { return c.N * 2 }

func main() {
	var c *cell
	println("before")
	println(c.Twice())
}
`},
		{"closure", "loopvar_for", mainOnly(`
	fs := make([]func() int, 3)
	for i := 0; i < 3; i++ {
		fs[i] = func() int { return i }
	}
	println(fs[0](), fs[1](), fs[2]())
	x := 1
	inc := func() { x++ }
	inc()
	inc()
	println(x)
`)},
		{"closure", "loopvar_range", mainOnly(`
	xs := []int{10, 20, 30}
	fs := make([]func() int, 0)
	for _, v := range xs {
		fs = append(fs, func() int { return v })
	}
	println(fs[0](), fs[1](), fs[2]())
	gs := make([]func() int, 0)
	for i := range xs {
		gs = append(gs, func() int { return i })
	}
	println(gs[0](), gs[1](), gs[2]())
`)},
		{"closure", "capture_param", mainOnly(`
	add := func(n int) func(int) int {
		return func(m int) int { return n + m }
	}
	println(add(10)(3))
	println(add(1)(1))
`)},
		{"defer", "order_named", `package main

func main() {
	println(copied())
	println(named())
	defer println(1)
	defer println(2)
	println(3)
}

func copied() int {
	n := 0
	defer func() { n += 1 }()
	defer func() { n += 10 }()
	n = 100
	return n
}

func named() (n int) {
	defer func() { n += 1 }()
	defer func() { n += 10 }()
	n = 100
	return
}
`},
		{"defer", "recover", `package main

func main() {
	println(caught())
	println(swallowed())
	println(fromCallee())
}

func caught() (n int) {
	defer func() {
		if recover() != nil {
			n = 7
		}
	}()
	panic("boom")
}

func swallowed() int {
	defer func() { recover() }()
	panic("boom")
}

func fromCallee() (n int) {
	defer func() {
		if recover() != nil {
			n = 5
		}
	}()
	return inner()
}

func inner() int { panic("inner") }
`},
		{"defer", "repanic", `package main

func main() {
	println(outer())
}

func outer() (n int) {
	defer func() {
		if recover() != nil {
			n = 3
		}
	}()
	func() {
		defer func() {
			println(1)
			panic("re")
		}()
		panic("first")
	}()
	return
}
`},
		{"defer", "userpanic", mainOnly(`
	println("before")
	panic("boom")
`)},
		{"control", "goto", `package main

func flag() bool { return true }

func main() {
	n := 0
	if flag() {
		goto skip
	}
	n = 1
skip:
	n += 2
	println(n)
	i := 0
loop:
	if i == 3 {
		goto done
	}
	i++
	goto loop
done:
	println(i)
}
`},
		{"control", "break_continue", mainOnly(`
	n := 0
	for i := 0; i < 4; i++ {
		if i == 1 {
			continue
		}
		if i == 3 {
			break
		}
		n += i
	}
	println(n)
`)},
		{"control", "labeled_break", mainOnly(`
	n := 0
outer:
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			if i == 2 {
				break outer
			}
			n++
		}
	}
	println(n)
`)},
		{"control", "labeled_continue", mainOnly(`
	n := 0
outer:
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if j == 1 {
				continue outer
			}
			n++
		}
	}
	println(n)
`)},
		{"control", "switch_fallthrough", `package main

func main() {
	println(kind(0))
	println(kind("a"))
	println(kind(1.5))
	println(fall(1))
	println(fall(2))
	println(fall(3))
	println(fall(9))
}

func kind(v interface{}) int {
	switch v.(type) {
	case nil:
		return 0
	case int:
		return 1
	case string:
		return 2
	default:
		return 3
	}
}

func fall(n int) int {
	x := 0
	switch n {
	case 1:
		x += 1
		fallthrough
	case 2:
		x += 10
	case 3:
		x += 100
	default:
		x += 1000
	}
	return x
}
`},
		{"shadow", "block", mainOnly(`
	n := 1
	{
		n := n + 1
		println(n)
	}
	println(n)
	func() {
		n := 3
		println(n)
	}()
	println(n)
`)},
		{"shadow", "param", `package main

func f(n int) int {
	n = n + 1
	return n
}

func main() {
	n := 1
	println(f(n))
	println(n)
}
`},
		{"recur", "sum20", `package main

func sum(n int) int {
	if n == 0 {
		return 0
	}
	return n + sum(n-1)
}

func main() {
	println(sum(20))
	println(sum(40))
}
`},
		{"recur", "sum80", `package main

func sum(n int) int {
	if n == 0 {
		return 0
	}
	return n + sum(n-1)
}

func main() {
	println(sum(80))
}
`},
		{"zerosize", "pointers", mainOnly(`
	a := new(struct{})
	b := new(struct{})
	println(a == b)
	println(a == a)
`)},
		{"stdlib", "strings_basic", `package main

import "strings"

func main() {
	s := "héllo, world"
	println(strings.Contains(s, "lo"))
	println(strings.HasPrefix(s, "hé"))
	println(strings.HasSuffix(s, "world"))
	println(strings.Count("aaa", "a"))
	println(strings.Index(s, "lo"))
	println(strings.ToUpper("AbCé"))
	println(strings.ToLower("AbCé"))
	println(strings.TrimSpace("  a  "))
	println(strings.Repeat("ab", 3))
	println(strings.Replace("aaa", "a", "b", 2))
	parts := strings.Split("a,b,c", ",")
	println(len(parts), parts[0], parts[1], parts[2])
	println(strings.Join(parts, "-"))
	println(strings.EqualFold("AbC", "abc"))
	fields := strings.Fields("a  b\tc")
	println(len(fields), fields[0], fields[1], fields[2])
}
`},
		{"stdlib", "strconv_basic", `package main

import "strconv"

func main() {
	println(strconv.Itoa(-42))
	n, err := strconv.Atoi("42")
	println(n, err == nil)
	n, err = strconv.Atoi("nope")
	println(n, err == nil)
	println(strconv.FormatInt(-255, 16))
	println(strconv.FormatUint(255, 16))
	println(strconv.FormatBool(true))
	ok, err := strconv.ParseBool("1")
	println(ok, err == nil)
	v, err := strconv.ParseInt("-20", 10, 8)
	println(v, err == nil)
	v, err = strconv.ParseInt("200", 10, 8)
	println(v, err != nil)
	println(strconv.Quote("a\nb"))
	u, err := strconv.Unquote(strconv.Quote("héllo"))
	println(u, err == nil)
	println(strconv.FormatFloat(1.5, 'f', 1, 64))
	f, err := strconv.ParseFloat("1.5", 64)
	println(f == 1.5, err == nil)
	println(strconv.FormatFloat(-0.0, 'f', 1, 64))
}
`},
		{"stdlib", "math_basic", `package main

import "math"

func main() {
	println(math.Abs(-1.5) == 1.5)
	println(math.Floor(1.9) == 1)
	println(math.Ceil(-1.2) == -1)
	println(math.Trunc(-1.9) == -1)
	println(math.Round(1.5) == 2)
	println(math.IsNaN(math.NaN()))
	println(math.NaN() == math.NaN())
	println(math.IsInf(math.Inf(1), 1))
	println(math.IsInf(math.Inf(-1), -1))
	println(math.IsInf(math.Inf(1), -1))
	println(math.Signbit(math.Copysign(0, -1)))
	println(!math.Signbit(0))
	println(math.Float64bits(math.Copysign(0, -1)) == 0x8000000000000000)
	println(math.Sqrt(4) == 2)
	println(math.Pow(2, 10) == 1024)
	println(math.Max(1, 2) == 2)
	println(math.Min(-1, 2) == -1)
	println(math.Mod(5.5, 2) == 1.5)
	println(math.Dim(3, 1) == 2)
}
`},
		{"stdlib", "utf8_basic", `package main

import "utf8"

func main() {
	println(utf8.RuneCountInString("héllo"))
	println(utf8.ValidString("héllo"))
	println(utf8.Valid([]byte{0xff}))
	r, size := utf8.DecodeRuneInString("éx")
	println(r == 'é', size)
	println(utf8.RuneLen('é'))
	println(utf8.FullRune([]byte{0xc3}))
	r, size = utf8.DecodeRuneInString(string([]byte{0xff}))
	println(r == utf8.RuneError, size == 1)
}
`},
	}
}
