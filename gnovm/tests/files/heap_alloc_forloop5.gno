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

// This does make 'i' NameExprTypeHeapDefine.
// You can tell by the preprocess printout of i<!~...> and i<()~...>.

// Preprocessed:
// file{ package main; type f (const-type main.f); var fs<!VPBlock(2,1)> []f<VPBlock(2,0)>; func forLoopClosure() { defer func func(){ for _<!VPBlock(0,0)>, f<!VPBlock(1,0)> := range fs<VPBlock(5,1)> { f<VPBlock(1,0)>() } }(); for i<!~VPBlock(1,0)> := (const (0 int)); i<~VPBlock(1,0)> < (const (3 int)); i<{}~VPBlock(1,0)>++ { fs<VPBlock(4,1)> = (const (append func(x []main.f,args ...main.f)(res []main.f)))(fs<VPBlock(4,1)>, (const-type main.f)(func func(){ z<!VPBlock(1,0)> := i<~VPBlock(1,0)>; (const (println func(xs ...interface{})()))(z<VPBlock(1,0)>) }<i<()~VPBlock(1,0)>>)) } }; func main() { forLoopClosure<VPBlock(3,2)>() } }

// Output:
// 0
// 1
// 2
