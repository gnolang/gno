package main

type T0 struct{ V int }
type T1 = struct{ A, B, C, D, E, F, G, H T0 }
type T2 = struct{ A, B, C, D, E, F, G, H T1 }
type T3 = struct{ A, B, C, D, E, F, G, H T2 }
type T4 = struct{ A, B, C, D, E, F, G, H T3 }

func main() {
	var v any = 1
	_, ok := v.(*T4)
	println(ok)
}

// Error:
// type too large: its identity passes 16384 bytes
