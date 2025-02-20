package main

type S struct {
	i int
}

func main() {
	println(&new(S))
}

// Error:
// main/files/addressable_10b_err.gno:8:10: cannot take address of (const (new func(t type{})( *main.S)))(S<VPBlock(3,0)>)
