package main

func main() {
	var f func()
	for i := 0; i < 3;{
		// .redefine_i := .loopvar_i
		println("enter, i: ", i)

		if i == 1 {
			println("---continue")
			i++ // .redefine_i, => 3
			// .loopvar_i = .redefine_i
			continue
		}

		// println("---after continue...")

		i++
		// println("after inc, i: ", i) // 1

		f = func() {
			start := i
			println("capture, start: ", start)
		}
		f()
	}
}

// Preprocessed:
// file{ package main; func main() { var f<!VPBlock(1,0)> func(); for .loopvar_i<!~VPBlock(1,0)> := (const (0 int)); .loopvar_i<~VPBlock(1,0)> < (const (3 int));  (const (println func(...interface {})))((const ("enter, i: " string)), .loopvar_i<~VPBlock(1,0)>); if .loopvar_i<~VPBlock(2,0)> == (const (1 int)) { (const (println func(...interface {})))((const ("---continue" string))); .loopvar_i<~VPBlock(2,0)>++; .loopvar_i<*~VPBlock(2,0)> = .loopvar_i<~VPBlock(2,0)>; continue }; .loopvar_i<~VPBlock(1,0)>++; f<VPBlock(2,0)> = func func(){ start<!VPBlock(1,0)> := .loopvar_i<~VPBlock(1,1)>; (const (println func(...interface {})))((const ("capture, start: " string)), start<VPBlock(1,0)>) }<.loopvar_i<()~VPBlock(1,0)>>; f<VPBlock(2,0)>(); .loopvar_i<*~VPBlock(1,0)> = .loopvar_i<~VPBlock(1,0)>  } }

// Output:
// enter, i:  0
// capture, start:  1
// enter, i:  1
// ---continue
// enter, i:  2
// capture, start:  3
