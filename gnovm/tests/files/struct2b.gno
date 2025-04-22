package main

type T struct {
	f int
	g int
	h int
}

func main() {
	_ = T{f: 1, g: 2, h: 3, f: 4, h: 5}
}

// Error:
// duplicate field name f in struct literal

// TypeCheckError:
// main/files/struct2b.gno:10:26: duplicate field name f in struct literal; main/files/struct2b.gno:10:32: duplicate field name h in struct literal
