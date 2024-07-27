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
// file{ package main; func main() { var fns<!VPBlock(1,0)> []func()  (const-type int); for i<!~VPBlock(1,0)> := (const (0 int)); i<~VPBlock(1,0)> < (const (2 int)); i<~VPBlock(1,0)>++ { for j<!~VPBlock(1,0)> := (const (0 int)); j<~VPBlock(1,0)> < (const (2 int)); j<~VPBlock(1,0)>++ { f<!VPBlock(1,1)> := func func()  (const-type int){ return i<~VPBlock(1,1)> + j<~VPBlock(1,2)> }<i<()~VPBlock(2,0)>, j<()~VPBlock(1,0)>>; fns<VPBlock(3,0)> = (const (append func(x []func()( int),args ...func()( int))(res []func()( int))))(fns<VPBlock(3,0)>, f<VPBlock(1,1)>) } }; for _<VPBlock(0,0)>, fn<VPBlock(1,0)> := range fns<VPBlock(2,0)> { (const (println func(xs ...interface{})()))(fn<VPBlock(1,0)>()) } } }

// Output:
// 4
// 4
// 4
// 4
