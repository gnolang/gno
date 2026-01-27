package main

func main() {
	var fns []func() int
	for i := 0; i < 5; i++ {
		var x int
		f := func() int {
			return x
		}
		x = i
		fns = append(fns, f)
	}
	for _, fn := range fns {
		println(fn())
	}
}

// Preprocessed:
// file{ package main; func main() { var fns<!VPBlock(1,0)> []func() .res.0 (const-type int); for .loopvar_i<!VPBlock(1,0)> := (const (0 int)); .loopvar_i<VPBlock(1,0)> < (const (5 int)); .loopvar_i<VPBlock(1,0)>++ var x<!~VPBlock(1,1)> (const-type int); f<!VPBlock(1,2)> := func func() .res.0 (const-type int){ return x<~VPBlock(1,1)> }<x<()~VPBlock(1,1)>>; x<~VPBlock(1,1)> = .loopvar_i<VPBlock(1,0)>; fns<VPBlock(2,0)> = (const (append func([]func() int, ...func() int) []func() int))(fns<VPBlock(2,0)>, f<VPBlock(1,2)>) ; for _<VPBlock(0,0)>, fn<!VPBlock(1,0)> := range fns<VPBlock(2,0)> { (const (println func(...interface {})))(fn<VPBlock(1,0)>()) } } }

// Output:
// 0
// 1
// 2
// 3
// 4
