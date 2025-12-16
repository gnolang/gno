package main

var fns []func()

func main() {
	for i := 0; i < 3; i++ {
		f := func() {
            i += 1
			println(i)
        }
		fns = append(fns, f)
		// println(i)
	}
	
	for _, ff := range fns{
		ff()
	}
}

// Preprocessed:
// file{ package main; var fns<!~VPBlock(2,0)> []func(); func main() { for .loopvar_i<!@VPBlock(1,0)> := (const (0 int)); .loopvar_i<*VPBlock(1,0)> < (const (3 int)); .loopvar_i<*VPBlock(1,0)>++ { .redefine_i<!~VPBlock(1,2)> := .loopvar_i<VPBlock(1,0)>; f<!VPBlock(1,1)> := func func(){ .redefine_i<~VPBlock(1,0)> += (const (1 int)); (const (println func(...interface {})))(.redefine_i<~VPBlock(1,0)>) }<.redefine_i<()~VPBlock(1,2)>>; fns<~VPBlock(4,0)> = (const (append func([]func(), ...func()) []func()))(fns<~VPBlock(4,0)>, f<VPBlock(1,1)>) }; for _<VPBlock(0,0)>, ff<!VPBlock(1,0)> := range fns<~VPBlock(4,0)> { ff<VPBlock(1,0)>() } } }

// Output:
// 1
// 2
// 3
