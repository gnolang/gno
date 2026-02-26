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
		fs = append(fs, func() {
			z := i
			println(z)
		})
	}
}

func main() {
	forLoopClosure()
}

// Preprocessed:
// file{ package main; type f (const-type main.f); var fs<!~VPBlock(2,1)> []typeval{main.f}; func forLoopClosure() { defer func func(){ for _<VPBlock(0,0)>, f<!VPBlock(1,0)> := range (const (ref(main) package{})).fs { f<VPBlock(1,0)>() } }(); for i.loopvar<!~VPBlock(1,0)> := (const (0 int)); i.loopvar<~VPBlock(1,0)> < (const (3 int)); i.loopvar<~VPBlock(1,0)>++ { fs<~VPBlock(4,1)> = (const (append func([]main.f, ...main.f) []main.f))(fs<~VPBlock(4,1)>, (const-type main.f)(func func(){ z<!VPBlock(1,0)> := i.loopvar<~VPBlock(1,1)>; (const (println func(...interface {})))(z<VPBlock(1,0)>) }<i.loopvar<()~VPBlock(1,0)>>)) } }; func main() { forLoopClosure<VPBlock(3,2)>() } }

// Output:
// 1
// 2
// 3
