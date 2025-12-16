package main

func main() {
	var fns []func() int

	for i := 0; i < 2; i++ {
		x := i
		y := 1
		f := func() int {
			switch y {
			case 1:
				x += 1
			default:
				x += 0
			}
			return x
		}
		fns = append(fns, f)
	}
	for _, fn := range fns {
		println(fn())
	}
}

// Preprocessed:
// file{ package main; func main() { var fns<!VPBlock(1,0)> []func() .res.0 (const-type int); for .loopvar_i<!@VPBlock(1,0)> := (const (0 int)); .loopvar_i<*VPBlock(1,0)> < (const (2 int)); .loopvar_i<*VPBlock(1,0)>++ { x<!~VPBlock(1,1)> := .loopvar_i<*VPBlock(1,0)>; y<!~VPBlock(1,2)> := (const (1 int)); f<!VPBlock(1,3)> := func func() .res.0 (const-type int){ switch y<~VPBlock(2,1)> { case (const (1 int)): x<~VPBlock(2,2)> += (const (1 int)); default: x<~VPBlock(2,2)> += (const (0 int)) }; return x<~VPBlock(1,2)> }<y<()~VPBlock(1,2)>, x<()~VPBlock(1,1)>>; fns<VPBlock(2,0)> = (const (append func([]func() int, ...func() int) []func() int))(fns<VPBlock(2,0)>, f<VPBlock(1,3)>) }; for _<VPBlock(0,0)>, fn<!VPBlock(1,0)> := range fns<VPBlock(2,0)> { (const (println func(...interface {})))(fn<VPBlock(1,0)>()) } } }

// Output:
// 1
// 2
