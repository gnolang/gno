package gen

import "fmt"

// Feature snippets are type-correct and stay inside the Gno subset.
// Names carry the builder sequence so two snippets in one main do not clash.
// Arithmetic that should wrap uses variables, not typed constants: a typed
// constant that overflows is constOverflowProgram, which gc must reject.

func (g *builder) closure() {
	g.emit(`
func addID(n int) func(int) int {
	return func(x int) int { return x + n }
}
`, `
	println(addID(2)(3))
	var fsID [2]func() int
	for i := 0; i < 2; i++ {
		fsID[i] = func() int { return i }
	}
	println(fsID[0](), fsID[1]())
`)
}

func (g *builder) deferRecover() {
	g.emit("", `
	func() {
		defer func() {
			if recover() != nil {
				println(1)
			}
		}()
		panic("x")
	}()
	defer func() { println(2) }()
`)
}

func (g *builder) ifaceSwitch() {
	g.emit(`
type faceID interface{ S() int }
type boxID struct{ V int }

func (b boxID) S() int { return b.V }
`, `
	var iID faceID = boxID{V: 3}
	switch v := iID.(type) {
	case boxID:
		println(v.V)
	default:
		println(0)
	}
	println(iID == nil)
	var pID *boxID
	var i2ID faceID = pID
	println(i2ID == nil)
	println(iID.S())
`)
}

func (g *builder) embedMethod() {
	g.emit(`
type inID struct{ X int }

func (i inID) M() int { return i.X }

type outID struct{ inID }

type cellID struct{ N int }

func (c *cellID) Add(x int) int { return c.N + x }
`, `
	oID := outID{inID{X: 7}}
	println(oID.M(), oID.X)
	cID := cellID{N: 3}
	fID := cID.Add
	println(fID(4))
	gID := (*cellID).Add
	println(gID(&cID, 1))
`)
}

func (g *builder) labeledBreak() {
	g.emit("", `
	nID := 0
outerID:
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			nID++
			if i == 1 && j == 1 {
				break outerID
			}
		}
	}
	println(nID)
`)
}

func (g *builder) labeledContinue() {
	g.emit("", `
	nID := 0
outerID:
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if j == 1 {
				continue outerID
			}
			nID++
		}
	}
	println(nID)
`)
}

func (g *builder) gotoStmt() {
	g.emit("", `
	nID := 0
	if nID == 0 {
		goto doneID
	}
	nID = 1
doneID:
	println(nID)
`)
}

func (g *builder) fallthroughStmt() {
	g.emit("", `
	switch 1 {
	case 1:
		println(1)
		fallthrough
	case 2:
		println(2)
	default:
		println(9)
	}
`)
}

func (g *builder) arraySlice() {
	g.emit("", `
	aID := [3]int{1, 2, 3}
	bID := aID
	bID[0] = 9
	println(aID[0], bID[0], aID == [3]int{1, 2, 3})
	sID := aID[:]
	sID[1] = 8
	println(aID[1], len(sID), cap(sID))
`)
}

func (g *builder) copyAppend() {
	g.emit("", `
	sID := []int{1, 2, 3, 4}
	nID := copy(sID[1:], sID[0:3])
	println(nID, sID[0], sID[1], sID[2], sID[3])
	var baseID []int
	baseID = append(baseID, 1, 2, 3)
	tID := baseID[:2]
	tID[0] = 9
	println(baseID[0], len(tID))
	uID := append([]int{}, baseID...)
	uID[0] = 1
	println(baseID[0], uID[0], uID[2])
`)
}

func (g *builder) stringBytes() {
	g.emit("", `
	sID := "héllo"
	bID := []byte(sID)
	println(len(sID), len(bID), len([]rune(sID)))
	println(string(bID) == sID)
	rawID := []byte{0xff, 0x00}
	println(len(string(rawID)), rawID[0] == 0xff)
	backID := []byte(string(rawID))
	println(len(backID), backID[0] == 0xff)
`)
}

func (g *builder) structMap() {
	g.emit(`
type keyID struct {
	A int
	B string
}
`, `
	mID := map[keyID]int{{A: 1, B: "a"}: 2}
	mID[keyID{A: 1, B: "b"}] = 3
	println(mID[keyID{A: 1, B: "a"}], len(mID))
	delete(mID, keyID{A: 1, B: "b"})
	println(len(mID), mID[keyID{A: 1, B: "b"}])
`)
}

func (g *builder) shadow() {
	g.emit(`
func shadID(x int) int {
	x = x + 1
	{
		x := 10
		println(x)
	}
	return x
}
`, `
	xID := 1
	{
		xID := xID + 2
		println(xID)
	}
	println(xID, shadID(1))
`)
}

func (g *builder) floats() {
	g.emit("", `
	var fID float64 = 1.9
	println(int(fID))
	var gID float32 = -2.9
	println(int(gID))
	var zID float64
	nanID := zID / zID
	println(nanID == nanID)
	println(nanID != nanID)
	println(zID == 0)
`)
}

func (g *builder) widthWrap() {
	id := g.seq
	g.seq++
	switch g.r.IntN(6) {
	case 0:
		fmt.Fprintf(&g.b, "\tvar a%d int8 = 127\n\tvar one%d int8 = 1\n\tprintln(a%d + one%d)\n", id, id, id, id)
	case 1:
		fmt.Fprintf(&g.b, "\tvar a%d int16 = -32768\n\tvar one%d int16 = 1\n\tprintln(a%d - one%d)\n", id, id, id, id)
	case 2:
		fmt.Fprintf(&g.b, "\tvar a%d uint8 = 255\n\tvar one%d uint8 = 1\n\tprintln(a%d + one%d)\n", id, id, id, id)
	case 3:
		fmt.Fprintf(&g.b, "\tvar a%d uint16 = 0\n\tvar one%d uint16 = 1\n\tprintln(a%d - one%d)\n", id, id, id, id)
	case 4:
		fmt.Fprintf(&g.b, "\tvar a%d int32 = 2147483647\n\tvar one%d int32 = 1\n\tprintln(a%d + one%d)\n", id, id, id, id)
	default:
		fmt.Fprintf(&g.b, "\tvar wide%d int16 = 300\n\tprintln(int8(wide%d))\n\tvar neg%d int16 = -1\n\tprintln(uint8(neg%d))\n", id, id, id, id)
	}
}
