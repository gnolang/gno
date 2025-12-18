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
// xxx

// Output: