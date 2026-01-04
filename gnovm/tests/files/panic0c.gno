package main

type S struct {
	s string
}

func f(
	s string,
	b bool,
	by byte,
	it int,
	it8 int8,
	it16 int16,
	it32 int32,
	it64 int64,
	uit uint,
	uit8 uint8,
	uit16 uint16,
	uit32 uint32,
	uit64 uint64,
	ft32 float32,
	ft64 float64,
	strs []string,
	st S,
	m map[string]string,
	t func(s string) string,
) {
	panic("wtf")
}

func main() {
	strs := []string{"a", "b"}
	st := S{"c"}
	m := map[string]string{"d": "gg", "test": "test"}
	t := func(s string) string {
		return s
	}

	const s string = "a"

	const b bool = true

	const by byte = 0

	const it int = 1
	const it8 int8 = 1
	const it16 int16 = 1
	const it32 int32 = 1
	const it64 int64 = 1
	const uit uint = 1
	const uit8 uint8 = 1
	const uit16 uint16 = 1
	const uit32 uint32 = 1
	const uit64 uint64 = 1
	const ft32 float32 = 1
	const ft64 float64 = 1
	f(
		s,
		b,
		by,
		it,
		it8,
		it16,
		it32,
		it64,
		uit,
		uit8,
		uit16,
		uit32,
		uit64,
		ft32,
		ft64,
		strs,
		st,
		m,
		t)
}

// Stacktrace:
// panic: wtf
// f<VPBlock(3,1)>(a,true,0,1,1,1,1,1,1,1,1,1,1,1,1,strs<VPBlock(1,0)>,st<VPBlock(1,1)>,m<VPBlock(1,2)>,t<VPBlock(1,3)>)
//     main/panic0c.gno:28
// main<VPBlock(1,2)>()
//     main/panic0c.gno:57

// Error:
// wtf
