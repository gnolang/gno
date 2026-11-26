package main

type S struct{ X int }

var p *int

var get func() int

// No defers, so `return r` of the named result r goes through
// doOpReturnAfterCopy. A pointer and a closure both reach r, so the result
// slot must keep r itself; copying it into the slot would leave p and get
// pointing at a discarded original.
func f() (r S) {
	p = &r.X
	get = func() int { return r.X }
	return r
}

func main() {
	f()
	*p = 5
	println(get())
}

// Output:
// 5
