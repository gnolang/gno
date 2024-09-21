package main

func main() {
	_ = &(&struct{}{})
}

// Error:
// main/files/addressable_3d_err.gno:4:6: cannot take address of &(struct {  }{})
