package main

func main() {
	_ = nil
}

// Error:
// main/assign_nil3.gno:4:2-9: use of untyped nil in assignment

// TypeCheckError:
// main/assign_nil3.gno:4:6: use of untyped nil in assignment to _ identifier
