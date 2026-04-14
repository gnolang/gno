package main

func main() {
	_ = make(int)
}

// Error:
// main/varg_12.gno:4:6-15: invalid argument: cannot make int; type must be slice, map

// TypeCheckError:
// main/varg_12.gno:4:11: invalid argument: cannot make int; type must be slice, map, or channel
