package main

func main() {
	println([]byte("a") == nil) // lx: (const (slice[0x61] []uint8)), rx: (const (undefined)), cmp = 0
}

// Output:
// false
