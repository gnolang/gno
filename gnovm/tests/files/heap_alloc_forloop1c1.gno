package main

var f func()
func forLoopClosure() {
	for i := 0; i < 10; i++ {
		i := i
		f = func(){
			println(i)
		}
	}
	f()
}

func main() {
	forLoopClosure()
}

// Preprocessed:
// file{ package main; var f<!~VPBlock(2,0)> func(); func forLoopClosure() { for .loopvar_i<!VPBlock(1,0)> := (const (0 int)); .loopvar_i<VPBlock(1,0)> < (const (10 int)); .loopvar_i<VPBlock(1,0)>++ i<!~VPBlock(1,1)> := .loopvar_i<VPBlock(1,0)>; f<~VPBlock(4,0)> = func func(){ (const (println func(...interface {})))(i<~VPBlock(1,0)>) }<i<()~VPBlock(1,1)>> ; f<~VPBlock(3,0)>() }; func main() { forLoopClosure<VPBlock(3,1)>() } }

// Output:
// 9
