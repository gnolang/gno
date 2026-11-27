package gen

import "fmt"

// Feature snippets are type-correct and stay inside the Gno subset.
// Names carry the builder sequence ($ID) so two snippets in one main do not
// clash, and values/branches vary with g.r so distinct seeds explore distinct
// runtime paths. Arithmetic that should wrap uses variables, not typed
// constants: a typed constant that overflows is constOverflowProgram, which
// gc must reject.

func (g *builder) closure() {
	a := g.r.IntN(11) - 5
	b := g.r.IntN(11) - 5
	n := 2 + g.r.IntN(3)
	g.emit(`
func add$ID(n int) func(int) int {
	return func(x int) int { return x + n }
}
`, fmt.Sprintf(`
	println(add$ID(%d)(%d))
	fs$ID := make([]func() int, %d)
	for i := 0; i < %d; i++ {
		fs$ID[i] = func() int { return i + %d }
	}
	for i := 0; i < len(fs$ID); i++ {
		println(fs$ID[i]())
	}
`, a, b, n, n, a))
}

func (g *builder) deferRecover() {
	doPanic := g.r.IntN(2) == 0
	tag := g.r.IntN(10)
	after := 10 + g.r.IntN(10)
	body := fmt.Sprintf(`
	func() {
		defer func() {
			r := recover()
			println(r != nil, %d)
		}()
`, tag)
	if doPanic {
		body += fmt.Sprintf("\t\tpanic(%q)\n", fmt.Sprintf("p%d", tag))
	} else {
		body += fmt.Sprintf("\t\tprintln(%d)\n", tag)
	}
	body += fmt.Sprintf("\t}()\n\tdefer func() { println(%d) }()\n", after)
	g.emit("", body)
}

func (g *builder) ifaceSwitch() {
	v := g.r.IntN(21) - 10
	useNilPtr := g.r.IntN(3) == 0
	body := fmt.Sprintf(`
	var i$ID face$ID = box$ID{V: %d}
	switch v := i$ID.(type) {
	case box$ID:
		println(v.V)
	case *box$ID:
		println(v == nil)
	default:
		println(0)
	}
	println(i$ID == nil, i$ID.S())
`, v)
	if useNilPtr {
		body += `
	var p$ID *box$ID
	var i2$ID face$ID = p$ID
	println(p$ID == nil, i2$ID == nil)
`
	} else {
		body += fmt.Sprintf(`
	var i2$ID face$ID
	println(i2$ID == nil, box$ID{V: %d}.S())
`, v+1)
	}
	g.emit(`
type face$ID interface{ S() int }
type box$ID struct{ V int }

func (b box$ID) S() int { return b.V }
`, body)
}

func (g *builder) embedMethod() {
	x := g.r.IntN(20) - 10
	n := g.r.IntN(20) - 10
	d1 := g.r.IntN(9) - 4
	d2 := g.r.IntN(9) - 4
	g.emit(`
type in$ID struct{ X int }

func (i in$ID) M() int { return i.X }

type out$ID struct{ in$ID }

type cell$ID struct{ N int }

func (c *cell$ID) Add(x int) int { return c.N + x }
`, fmt.Sprintf(`
	o$ID := out$ID{in$ID{X: %d}}
	println(o$ID.M(), o$ID.X)
	c$ID := cell$ID{N: %d}
	f$ID := c$ID.Add
	println(f$ID(%d))
	g$ID := (*cell$ID).Add
	println(g$ID(&c$ID, %d))
`, x, n, d1, d2))
}

func (g *builder) labeledBreak() {
	rows := 2 + g.r.IntN(3)
	cols := 2 + g.r.IntN(3)
	bi := g.r.IntN(rows)
	bj := g.r.IntN(cols)
	g.emit("", fmt.Sprintf(`
	n$ID := 0
outer$ID:
	for i := 0; i < %d; i++ {
		for j := 0; j < %d; j++ {
			n$ID++
			if i == %d && j == %d {
				break outer$ID
			}
		}
	}
	println(n$ID)
`, rows, cols, bi, bj))
}

func (g *builder) labeledContinue() {
	rows := 2 + g.r.IntN(3)
	cols := 2 + g.r.IntN(3)
	skip := g.r.IntN(cols)
	g.emit("", fmt.Sprintf(`
	n$ID := 0
outer$ID:
	for i := 0; i < %d; i++ {
		for j := 0; j < %d; j++ {
			if j == %d {
				continue outer$ID
			}
			n$ID++
		}
	}
	println(n$ID)
`, rows, cols, skip))
}

func (g *builder) gotoStmt() {
	n := g.r.IntN(2)
	val := 1 + g.r.IntN(9)
	g.emit("", fmt.Sprintf(`
	n$ID := %d
	if n$ID == 0 {
		goto done$ID
	}
	n$ID = %d
done$ID:
	println(n$ID)
`, n, val))
}

func (g *builder) fallthroughStmt() {
	target := 1 + g.r.IntN(3)
	g.emit("", fmt.Sprintf(`
	switch %d {
	case 1:
		println(1)
		fallthrough
	case 2:
		println(2)
	default:
		println(9)
	}
`, target))
}

func (g *builder) arraySlice() {
	a, b, c := g.r.IntN(10), g.r.IntN(10), g.r.IntN(10)
	m1, m2 := 10+g.r.IntN(10), 20+g.r.IntN(10)
	g.emit("", fmt.Sprintf(`
	a$ID := [3]int{%d, %d, %d}
	b$ID := a$ID
	b$ID[0] = %d
	println(a$ID[0], b$ID[0], a$ID == [3]int{%d, %d, %d})
	s$ID := a$ID[:]
	s$ID[1] = %d
	println(a$ID[1], len(s$ID), cap(s$ID))
`, a, b, c, m1, a, b, c, m2))
}

func (g *builder) copyAppend() {
	v0, v1, v2, v3 := g.r.IntN(10), g.r.IntN(10), g.r.IntN(10), g.r.IntN(10)
	copyLine := "n$ID := copy(s$ID[1:], s$ID[0:3])"
	if g.r.IntN(2) == 0 {
		copyLine = "n$ID := copy(s$ID[0:3], s$ID[1:])"
	}
	g.emit("", fmt.Sprintf(`
	s$ID := []int{%d, %d, %d, %d}
	%s
	println(n$ID, s$ID[0], s$ID[1], s$ID[2], s$ID[3])
	var base$ID []int
	base$ID = append(base$ID, %d, %d, %d)
	t$ID := base$ID[:2]
	t$ID[0] = %d
	println(base$ID[0], len(t$ID))
	u$ID := append([]int{}, base$ID...)
	u$ID[0] = %d
	println(base$ID[0], u$ID[0], u$ID[2])
`, v0, v1, v2, v3, copyLine, v0, v1, v2, v3+10, v0+20))
}

func (g *builder) stringBytes() {
	samples := []string{`"héllo"`, `"gno"`, `"😀a"`, `""`, `"a\x00b"`}
	s := samples[g.r.IntN(len(samples))]
	b0 := 0x80 + g.r.IntN(0x80)
	g.emit("", fmt.Sprintf(`
	s$ID := %s
	b$ID := []byte(s$ID)
	println(len(s$ID), len(b$ID), len([]rune(s$ID)))
	println(string(b$ID) == s$ID)
	raw$ID := []byte{0x%02x, 0x00}
	println(len(string(raw$ID)), raw$ID[0] == 0x%02x)
	back$ID := []byte(string(raw$ID))
	println(len(back$ID), back$ID[0] == 0x%02x)
`, s, b0, b0, b0))
}

func (g *builder) structMap() {
	k1 := g.r.IntN(5)
	k2 := k1 + 1 + g.r.IntN(5)
	v1 := g.r.IntN(20)
	v2 := g.r.IntN(20)
	g.emit(`
type key$ID struct {
	A int
	B string
}
`, fmt.Sprintf(`
	m$ID := map[key$ID]int{{A: %d, B: "a"}: %d}
	m$ID[key$ID{A: %d, B: "b"}] = %d
	println(m$ID[key$ID{A: %d, B: "a"}], len(m$ID))
	delete(m$ID, key$ID{A: %d, B: "b"})
	println(len(m$ID), m$ID[key$ID{A: %d, B: "b"}])
`, k1, v1, k2, v2, k1, k2, k2))
}

func (g *builder) shadow() {
	outer := g.r.IntN(10) - 5
	delta := 1 + g.r.IntN(5)
	inner := 10 + g.r.IntN(10)
	g.emit(fmt.Sprintf(`
func shad$ID(x int) int {
	x = x + %d
	{
		x := %d
		println(x)
	}
	return x
}
`, delta, inner), fmt.Sprintf(`
	x$ID := %d
	{
		x$ID := x$ID + %d
		println(x$ID)
	}
	println(x$ID, shad$ID(%d))
`, outer, delta, outer))
}

func (g *builder) floats() {
	fvals := []string{"1.9", "-1.9", "0.0", "2.5", "-2.9", "100.75"}
	f1 := fvals[g.r.IntN(len(fvals))]
	f2 := fvals[g.r.IntN(len(fvals))]
	g.emit("", fmt.Sprintf(`
	var f$ID float64 = %s
	println(int(f$ID))
	var g$ID float32 = %s
	println(int(g$ID))
	var z$ID float64
	nan$ID := z$ID / z$ID
	println(nan$ID == nan$ID, nan$ID != nan$ID, z$ID == 0)
`, f1, f2))
}

func (g *builder) widthWrap() {
	id := g.seq
	g.seq++
	step := 1 + g.r.IntN(5)
	switch g.r.IntN(8) {
	case 0:
		fmt.Fprintf(&g.b, "\tvar a%d int8 = 127\n\tvar d%d int8 = %d\n\tprintln(a%d + d%d)\n", id, id, step, id, id)
	case 1:
		fmt.Fprintf(&g.b, "\tvar a%d int16 = -32768\n\tvar d%d int16 = %d\n\tprintln(a%d - d%d)\n", id, id, step, id, id)
	case 2:
		fmt.Fprintf(&g.b, "\tvar a%d uint8 = 255\n\tvar d%d uint8 = %d\n\tprintln(a%d + d%d)\n", id, id, step, id, id)
	case 3:
		fmt.Fprintf(&g.b, "\tvar a%d uint16 = 0\n\tvar d%d uint16 = %d\n\tprintln(a%d - d%d)\n", id, id, step, id, id)
	case 4:
		fmt.Fprintf(&g.b, "\tvar a%d int32 = 2147483647\n\tvar d%d int32 = %d\n\tprintln(a%d + d%d)\n", id, id, step, id, id)
	case 5:
		fmt.Fprintf(&g.b, "\tvar a%d uint32 = 0\n\tvar d%d uint32 = %d\n\tprintln(a%d - d%d)\n", id, id, step, id, id)
	case 6:
		fmt.Fprintf(&g.b, "\tvar a%d int64 = -9223372036854775808\n\tvar d%d int64 = -1\n\tprintln(a%d * d%d)\n", id, id, id, id)
	default:
		wide := 256 + g.r.IntN(500)
		neg := -(1 + g.r.IntN(300))
		fmt.Fprintf(&g.b, "\tvar wide%d int16 = %d\n\tprintln(int8(wide%d))\n\tvar neg%d int16 = %d\n\tprintln(uint8(neg%d))\n", id, wide, id, id, neg, id)
	}
}
