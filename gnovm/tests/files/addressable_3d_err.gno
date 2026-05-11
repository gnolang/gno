package main

func main() {
	_ = &(&struct{}{})
}

// Error:
// main/addressable_3d_err.gno:4:6-20: invalid operation: cannot take address of &((const-type struct{}){}) (value of type *struct{})

// TypeCheckError:
// main/addressable_3d_err.gno:4:7: invalid operation: cannot take address of (&struct{}{}) (value of type *struct{})
