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

// go 1.22 loop var is not supported for now.

// Preprocessed:
// file{ package main; func main() { var fns<!VPBlock(1,0)> []func() .res.0 (const-type int); for .loopvar_i<!VPBlock(1,0)> := (const (0 int)); .loopvar_i<VPBlock(1,0)> < (const (2 int)); .loopvar_i<VPBlock(1,0)>++ { .redefine_i<!~VPBlock(1,0)> := .loopvar_i<VPBlock(2,0)>; for .loopvar_j<!VPBlock(1,0)> := (const (0 int)); .loopvar_j<VPBlock(1,0)> < (const (2 int)); .loopvar_j<VPBlock(1,0)>++ { .redefine_j<!~VPBlock(1,1)> := .loopvar_j<VPBlock(2,0)>; f<!VPBlock(1,0)> := func func() .res.0 (const-type int){ return .redefine_i<~VPBlock(1,1)> + .redefine_j<~VPBlock(1,2)> }<.redefine_i<()~VPBlock(3,0)>, .redefine_j<()~VPBlock(1,1)>>; fns<VPBlock(5,0)> = (const (append func([]func() int, ...func() int) []func() int))(fns<VPBlock(5,0)>, f<VPBlock(1,0)>); .loopvar_j<VPBlock(2,0)> = .redefine_j<~VPBlock(1,1)> } ; .loopvar_i<VPBlock(2,0)> = .redefine_i<~VPBlock(1,0)> } ; for _<VPBlock(0,0)>, fn<!VPBlock(1,0)> := range fns<VPBlock(2,0)> { (const (println func(...interface {})))(fn<VPBlock(1,0)>()) } } }

// Output:
// 0
// 1
// 1
// 2
