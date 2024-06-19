package main

func main() {
	m := map[string]bool{"foo": true}
	m["foo"]++
}

// Error:
// main/files/types/runtime_a0.gno:5: operator ++ not defined on: BoolKind
