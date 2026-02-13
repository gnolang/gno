package main

func main() {
	var f func()
	// .loopvar_i
	for i := 0; i < 5;{
		println("enter, i: ", i)

		if i == 2 {
			println("---continue")
			i++
			// Copy out
			continue
		}

		// shadow
		// .loopvar_i
		// start search with 'i',
		// will find .loopvar_i, then use it.
		for i:=1; i<3; i++{
			println(i) // .loopvar_i
		}

		for i<6{
			i++ 
		}

		f = func() {
			start := i 
			println("capture, start: ", start)
		}

		f()
		// Copy out
	}
}

// Preprocessed:
// file{ package main; func main() { var f<!VPBlock(1,0)> func(); for .loopvar_i<!~VPBlock(1,0)> := (const (0 int)); .loopvar_i<~VPBlock(1,0)> < (const (5 int));  { (const (println func(...interface {})))((const ("enter, i: " string)), .loopvar_i<~VPBlock(1,0)>); if .loopvar_i<~VPBlock(2,0)> == (const (2 int)) { (const (println func(...interface {})))((const ("---continue" string))); .loopvar_i<~VPBlock(2,0)>++; .loopvar_i<@!~VPBlock(2,0)> = .loopvar_i<~VPBlock(2,0)>; continue }; for .loopvar_i<!VPBlock(1,0)> := (const (1 int)); .loopvar_i<VPBlock(1,0)> < (const (3 int)); .loopvar_i<VPBlock(1,0)>++ { (const (println func(...interface {})))(.loopvar_i<VPBlock(1,0)>) }; for ; .loopvar_i<~VPBlock(2,0)> < (const (6 int));  { .loopvar_i<~VPBlock(2,0)>++ }; f<VPBlock(2,0)> = func func(){ start<!VPBlock(1,0)> := .loopvar_i<~VPBlock(1,1)>; (const (println func(...interface {})))((const ("capture, start: " string)), start<VPBlock(1,0)>) }<.loopvar_i<()~VPBlock(1,0)>>; f<VPBlock(2,0)>(); .loopvar_i<@!~VPBlock(1,0)> = .loopvar_i<~VPBlock(1,0)> } } }

// Output:
// enter, i:  0
// 1
// 2
// capture, start:  6
