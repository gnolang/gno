package main

// NOTE: instead of patching the vm code, this should be handled by an
// integration of `gno tool transpile --gobuild`, which uses `go build` under
// the hood and already detects unused variables.
func main() {
	var x int
}

// Error:
// x declared but not used
