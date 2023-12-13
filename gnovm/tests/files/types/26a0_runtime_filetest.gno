package main

func main() {
	m := map[string]bool{"foo": true}
	m["foo"]++
}

// Error:
// main/files/types/26a0_runtime_filetest.gno:5: operator ++ not defined on: bool
