package main

func main() {
	var f func()
	for i := 0; i < 1; {
		// right place to inject...
		// i := i // .redefine_i = .loopvar_i
		println("enter, i: ", i) // 0
		i++ // .redefine_i
		println("after inc, i: ", i) // 1

		f = func(){
			start := i // .redefine_i, value is 1
			println("capture, start: ", start)
		}
		f()
		i++ // .redefine_i
		// Copy back
		// .loopvar_i = .redefine_i
	}
}

// Preprocessed:
// file{ package main; func main() { var f<!VPBlock(1,0)> func(); for .loopvar_i<!@VPBlock(1,0)> := (const (0 int)); .loopvar_i<*VPBlock(1,0)> < (const (1 int));  { .redefine_i<!~VPBlock(1,1)> := .loopvar_i<VPBlock(1,0)>; (const (println func(...interface {})))((const ("enter, i: " string)), .redefine_i<~VPBlock(1,1)>); .redefine_i<~VPBlock(1,1)>++; (const (println func(...interface {})))((const ("after inc, i: " string)), .redefine_i<~VPBlock(1,1)>); f<VPBlock(2,0)> = func func(){ start<!VPBlock(1,0)> := .redefine_i<~VPBlock(1,1)>; (const (println func(...interface {})))((const ("capture, start: " string)), start<VPBlock(1,0)>) }<.redefine_i<()~VPBlock(1,1)>>; f<VPBlock(2,0)>(); .redefine_i<~VPBlock(1,1)>++; .loopvar_i<!VPBlock(1,0)> := .redefine_i<~VPBlock(1,1)> } } }

// Output:
// enter, i:  0
// after inc, i:  1
// capture, start:  1
