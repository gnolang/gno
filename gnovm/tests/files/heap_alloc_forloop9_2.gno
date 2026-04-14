package main

func main() {
	var fns []func() int

	println("start for loop")
	for i := 0; i < 2; i++ {
		defer func() {
			println("defer")
			for _, fn := range fns {
				println(fn())
			}
		}()

		x := i
		f := func() int {
			return x
		}

		fns = append(fns, f)
	}
	println("end for loop")
}

// Preprocessed:
// file{ package main; func main() { var fns<!~VPBlock(1,0)> []func() .res.0 (const-type int); (const (println func(...interface {})))((const ("start for loop" string))); for i.loopvar<!VPBlock(1,0)> := (const (0 int)); i.loopvar<VPBlock(1,0)> < (const (2 int)); i.loopvar<VPBlock(1,0)>++ { defer func func(){ (const (println func(...interface {})))((const ("defer" string))); for _<VPBlock(0,0)>, fn<!VPBlock(1,0)> := range fns<~VPBlock(2,0)> { (const (println func(...interface {})))(fn<VPBlock(1,0)>()) } }<fns<()~VPBlock(2,0)>>(); x<!~VPBlock(1,1)> := i.loopvar<VPBlock(1,0)>; f<!VPBlock(1,2)> := func func() .res.0 (const-type int){ return x<~VPBlock(1,1)> }<x<()~VPBlock(1,1)>>; fns<~VPBlock(2,0)> = (const (append func([]func() int, ...func() int) []func() int))(fns<~VPBlock(2,0)>, f<VPBlock(1,2)>) }; (const (println func(...interface {})))((const ("end for loop" string))) } }

// Output:
// start for loop
// end for loop
// defer
// 0
// 1
// defer
// 0
// 1
