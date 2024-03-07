package main

type FooS struct {
}

type BarS FooS

func main() {
	var _ *BarS = (*BarS)(nil)
}
