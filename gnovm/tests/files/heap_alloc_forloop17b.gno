package main

func main() {
	var f func()
	// .loopvar_i
	for i := 0; i < 5;{
		// .redefine_i = .loopvar_i
		println("enter, i: ", i)

		// XXX, only search type with LoopvarUse.

		// search with base 'i', looking for 
		// 'i'(iff explicitly defined i:=i), 
		// '.redefine_i', '.loopvar_i'
		// if found, set name accordingly(final).
		// if not found any, panic.
		if i == 2 { // .redefine_i
			println("---continue")
			i++ // .redefine_i, => 3
			// .loopvar_i = .redefine_i // copy out
			continue
		}

		// shadow
		// .loopvar_i
		// start search with 'i',
		// will find .loopvar_i, then use it.
		for i:=1; i<3; i++{
			// no inject
			println(i) // .loopvar_i
		}

		for i<6{ // .redefine_i
			i++ // .redefine_i
		}

		f = func() {
			start := i // .redefine_i
			println("capture, start: ", start)
		}

		f()
		// .loopvar_i = .redefine_i // copy out
	}
}

// Preprocessed:
// file{ package main; func main() { var f<!VPBlock(1,0)> func(); for .loopvar_i<!@VPBlock(1,0)> := (const (0 int)); .loopvar_i<*VPBlock(1,0)> < (const (5 int));  { .redefine_i<!~VPBlock(1,0)> := .loopvar_i<*VPBlock(2,0)>; (const (println func(...interface {})))((const ("enter, i: " string)), .redefine_i<~VPBlock(1,0)>); if .redefine_i<~VPBlock(2,0)> == (const (2 int)) { (const (println func(...interface {})))((const ("---continue" string))); .redefine_i<~VPBlock(2,0)>++; .loopvar_i<*VPBlock(3,0)> = .redefine_i<~VPBlock(2,0)>; continue }; for .loopvar_i<!@VPBlock(1,0)> := (const (1 int)); .loopvar_i<*VPBlock(1,0)> < (const (3 int)); .loopvar_i<*VPBlock(1,0)>++ { (const (println func(...interface {})))(.loopvar_i<*VPBlock(2,0)>) } ; for ; .redefine_i<~VPBlock(2,0)> < (const (6 int));  { .redefine_i<~VPBlock(3,0)>++ } ; f<VPBlock(3,0)> = func func(){ start<!VPBlock(1,0)> := .redefine_i<~VPBlock(1,1)>; (const (println func(...interface {})))((const ("capture, start: " string)), start<VPBlock(1,0)>) }<.redefine_i<()~VPBlock(1,0)>>; f<VPBlock(3,0)>(); .loopvar_i<*VPBlock(2,0)> = .redefine_i<~VPBlock(1,0)> }  } }

// Output:
// enter, i:  0
// 1
// 2
// capture, start:  6
