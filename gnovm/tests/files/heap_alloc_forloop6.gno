package main

func main() {
	var fns []func() int

	for i := 0; i < 3; i++ {
		z := i
		f := func() int {
			return z
		}
		fns = append(fns, f)
	}

	for _, fn := range fns {
		println(fn())
	}
}

// Preprocessed:
// file{ package main; func main() { var fns<!VPBlock(1,0)> []func() .res.0 (const-type int); for i.loopvar<!VPBlock(1,0)> := (const (0 int)); i.loopvar<VPBlock(1,0)> < (const (3 int)); i.loopvar<VPBlock(1,0)>++ { z<!~VPBlock(1,1)> := i.loopvar<VPBlock(1,0)>; f<!VPBlock(1,2)> := func func() .res.0 (const-type int){ return z<~VPBlock(1,1)> }<z<()~VPBlock(1,1)>>; fns<VPBlock(2,0)> = (const (append func([]func() int, ...func() int) []func() int))(fns<VPBlock(2,0)>, f<VPBlock(1,2)>) }; for _<VPBlock(0,0)>, fn<!VPBlock(1,0)> := range fns<VPBlock(2,0)> { (const (println func(...interface {})))(fn<VPBlock(1,0)>()) } } }

// Output:
// 0
// 1
// 2
