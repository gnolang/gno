package main

func main() {
	m := map[string]bool{"foo": true}
	m["foo"]++
}

// Error:
// main/files/types/runtime_a0.gno:5:2: operator ++ not defined on: BoolKind

// TypeCheckError:
// main/files/types/runtime_a0.gno:5:2: invalid operation: m["foo"]++ (non-numeric type bool)
