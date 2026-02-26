package main

func main() {
	var fns []func() int
	for i := 0; i < 5; i++ {
		f := func() int {
			return i
		}
		fns = append(fns, f)
	}
	for _, fn := range fns {
		println(fn())
	}
}

// go 1.22 loop var is not supported for now.

// Preprocessed:
// file{ package main; func main() { var fns<!VPBlock(1,0)> []func() .res.0 (const-type int); for i.loopvar<!~VPBlock(1,0)> := (const (0 int)); i.loopvar<~VPBlock(1,0)> < (const (5 int)); i.loopvar<~VPBlock(1,0)>++ { f<!VPBlock(1,1)> := func func() .res.0 (const-type int){ return i.loopvar<~VPBlock(1,1)> }<i.loopvar<()~VPBlock(1,0)>>; fns<VPBlock(2,0)> = (const (append func([]func() int, ...func() int) []func() int))(fns<VPBlock(2,0)>, f<VPBlock(1,1)>) }; for _<VPBlock(0,0)>, fn<!VPBlock(1,0)> := range fns<VPBlock(2,0)> { (const (println func(...interface {})))(fn<VPBlock(1,0)>()) } } }

// Output:
// 5
// 5
// 5
// 5
// 5
