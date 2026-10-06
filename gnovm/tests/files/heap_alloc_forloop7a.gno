package main

func main() {
	var fns []func() int

	for i := 0; i < 2; i++ {
		x := i
		f := func() int {
			if true {
				return x
			}
			return 0
		}
		fns = append(fns, f)
	}
	for _, fn := range fns {
		println(fn())
	}
}

// Preprocessed:
// file{ package main; func main() { var fns<!VPBlock(1,0)> []func() .res.0 (const-type int); for i.loopvar<!VPBlock(1,0)> := (const (0 int)); i.loopvar<VPBlock(1,0)> < (const (2 int)); i.loopvar<VPBlock(1,0)>++ { x<!~VPBlock(1,1)> := i.loopvar<VPBlock(1,0)>; f<!VPBlock(1,2)> := func func() .res.0 (const-type int){ if (const (true bool)) { return x<~VPBlock(2,1)> }; return (const (0 int)) }<x<()~VPBlock(1,1)>>; fns<VPBlock(2,0)> = (const (append func([]func() int, ...func() int) []func() int))(fns<VPBlock(2,0)>, f<VPBlock(1,2)>) }; for _<VPBlock(0,0)>, fn<!VPBlock(1,0)> := range fns<VPBlock(2,0)> { (const (println func(...interface {})))(fn<VPBlock(1,0)>()) } } }

// Output:
// 0
// 1
