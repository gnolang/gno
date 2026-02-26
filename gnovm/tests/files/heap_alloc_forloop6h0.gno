package main

func main() {
	var fns []func() int
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			f := func() int {
				return i + j
			}
			fns = append(fns, f)
		}
	}
	for _, fn := range fns {
		println(fn())
	}
}

// Preprocessed:
// file{ package main; func main() { var fns<!VPBlock(1,0)> []func() .res.0 (const-type int); for i.loopvar<!~VPBlock(1,0)> := (const (0 int)); i.loopvar<~VPBlock(1,0)> < (const (2 int)); i.loopvar<~VPBlock(1,0)>++ { for j.loopvar<!~VPBlock(1,0)> := (const (0 int)); j.loopvar<~VPBlock(1,0)> < (const (2 int)); j.loopvar<~VPBlock(1,0)>++ { f<!VPBlock(1,1)> := func func() .res.0 (const-type int){ return i.loopvar<~VPBlock(1,1)> + j.loopvar<~VPBlock(1,2)> }<i.loopvar<()~VPBlock(2,0)>, j.loopvar<()~VPBlock(1,0)>>; fns<VPBlock(3,0)> = (const (append func([]func() int, ...func() int) []func() int))(fns<VPBlock(3,0)>, f<VPBlock(1,1)>) } }; for _<VPBlock(0,0)>, fn<!VPBlock(1,0)> := range fns<VPBlock(2,0)> { (const (println func(...interface {})))(fn<VPBlock(1,0)>()) } } }

// Output:
// 2
// 3
// 3
// 4
