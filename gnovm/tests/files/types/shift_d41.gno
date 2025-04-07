package main

func main() {
	a := 2
	s := map[string]float32{"k": 1 << a}
	println(s["k"])
}

// Error:
// main/files/types/shift_d41.gno:5:7: operator << not defined on: Float32Kind
