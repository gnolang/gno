package main

func main() {
	var f func()
	for i := 0; i < 1; {
		println("enter, i: ", i) // 0
		i++
		println("after inc, i: ", i) // 1

		f = func(){
			start := i 
			println("capture, start: ", start)
		}
		f()
		i++ 
		// Copy back
		// .loopvar_i<NameExprTypeLoopVarHeapDefine> = .loopvar_i
	}
}

// Preprocessed:
// file{ package main; func main() { var f<!VPBlock(1,0)> func(); for .loopvar_i<!~VPBlock(1,0)> := (const (0 int)); .loopvar_i<~VPBlock(1,0)> < (const (1 int));  (const (println func(...interface {})))((const ("enter, i: " string)), .loopvar_i<~VPBlock(1,0)>); .loopvar_i<~VPBlock(1,0)>++; (const (println func(...interface {})))((const ("after inc, i: " string)), .loopvar_i<~VPBlock(1,0)>); f<VPBlock(2,0)> = func func(){ start<!VPBlock(1,0)> := .loopvar_i<~VPBlock(1,1)>; (const (println func(...interface {})))((const ("capture, start: " string)), start<VPBlock(1,0)>) }<.loopvar_i<()~VPBlock(1,0)>>; f<VPBlock(2,0)>(); .loopvar_i<~VPBlock(1,0)>++; .loopvar_i<*~VPBlock(1,0)> = .loopvar_i<~VPBlock(1,0)>  } }

// Output:
// enter, i:  0
// after inc, i:  1
// capture, start:  1
