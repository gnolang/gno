package main

type S struct {
	i int
}

func main() {
	println(&new(S))
}

// Error:
// main/addressable_10b_err.gno:8:10-17: invalid operation: cannot take address of (const (new func(type{}) *main.S))(typeval{main.S}) (value of type *main.S)

// TypeCheckError:
// main/addressable_10b_err.gno:8:11: invalid operation: cannot take address of new(S) (value of type *S)
