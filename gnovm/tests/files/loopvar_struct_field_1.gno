package main

// Struct composite literal: field names (i, v) match range loop var names.
// Field names must NOT be renamed to i.loopvar / v.loopvar.
// Value expressions (i, v) must be renamed correctly.
type S struct {
	i int
	v string
}

func main() {
	var fns []func() S
	for i, v := range []string{"a", "b", "c"} {
		i := i
		v := v
		fns = append(fns, func() S { return S{i: i, v: v} })
	}
	for _, fn := range fns {
		s := fn()
		println(s.i, s.v)
	}
}

// Preprocessed:
// file{ package main; type S (const-type main.S); func main() { var fns<!VPBlock(1,0)> []func() .res.0 typeval{main.S}; for i<!~VPBlock(1,0)>, v<!~VPBlock(1,1)> := range (const-type []string){(const ("a" string)), (const ("b" string)), (const ("c" string))} { i<!~VPBlock(1,0)> := i<~VPBlock(1,0)>; v<!~VPBlock(1,1)> := v<~VPBlock(1,1)>; fns<VPBlock(2,0)> = (const (append func([]func() main.S, ...func() main.S) []func() main.S))(fns<VPBlock(2,0)>, func func() .res.0 typeval{main.S}{ return (const-type main.S){i<VPField(0,0,i)>: i<~VPBlock(1,1)>, v<VPField(0,1,v)>: v<~VPBlock(1,2)>} }<i<()~VPBlock(1,0)>, v<()~VPBlock(1,1)>>) }; for _<VPBlock(0,0)>, fn<!VPBlock(1,0)> := range fns<VPBlock(2,0)> { s<!VPBlock(1,1)> := fn<VPBlock(1,0)>(); (const (println func(...interface {})))(s<VPBlock(1,1)>.i, s<VPBlock(1,1)>.v) } } }

// Output:
// 0 a
// 1 b
// 2 c
