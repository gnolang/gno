package main

func main() {
	var fns []func() int
	s := "hello"
	for i, _ := range s {
		x := i
		f := func() int {
			return x
		}
		fns = append(fns, f)
	}
	for _, fn := range fns {
		println(fn())
	}
}

// Preprocessed:
// file{ package main; func main() { var fns<!VPBlock(1,0)> []func() .res.0 (const-type int); s<!VPBlock(1,1)> := (const ("hello" string)); for i.loopvar<!VPBlock(1,0)>, _<VPBlock(0,0)> := range s<VPBlock(2,1)> { x<!~VPBlock(1,1)> := i.loopvar<VPBlock(1,0)>; f<!VPBlock(1,2)> := func func() .res.0 (const-type int){ return x<~VPBlock(1,1)> }<x<()~VPBlock(1,1)>>; fns<VPBlock(2,0)> = (const (append func([]func() int, ...func() int) []func() int))(fns<VPBlock(2,0)>, f<VPBlock(1,2)>) }; for _<VPBlock(0,0)>, fn<!VPBlock(1,0)> := range fns<VPBlock(2,0)> { (const (println func(...interface {})))(fn<VPBlock(1,0)>()) } } }

// Output:
// 0
// 1
// 2
// 3
// 4
