package main

type f func()

var fs []f

func forLoopClosure() {
	defer func() {
		for _, f := range fs {
			f()
		}
	}()

	for i := 0; i < 3; i++ {
		fs = append(fs, func() { println(i) })
	}
}

func main() {
	forLoopClosure()
}

// go 1.22 loop var is not supported for now.

// Preprocessed:
// file{ package main; type f (const-type main.f); var fs<!~VPBlock(2,1)> []typeval{main.f}; func forLoopClosure() { defer func func(){ for _<VPBlock(0,0)>, f<!VPBlock(1,0)> := range (const (ref(main) package{})).fs { f<VPBlock(1,0)>() } }(); for .loopvar_i<!~VPBlock(1,0)> := (const (0 int)); .loopvar_i<~VPBlock(1,0)> < (const (3 int)); .loopvar_i<~VPBlock(1,0)>++ fs<~VPBlock(4,1)> = (const (append func([]main.f, ...main.f) []main.f))(fs<~VPBlock(4,1)>, (const-type main.f)(func func(){ (const (println func(...interface {})))(.loopvar_i<~VPBlock(1,0)>) }<.loopvar_i<()~VPBlock(1,0)>>)); .loopvar_i<*~VPBlock(1,0)> = .loopvar_i<~VPBlock(1,0)>  }; func main() { forLoopClosure<VPBlock(3,2)>() } }

// Output:
// 0
// 1
// 2
