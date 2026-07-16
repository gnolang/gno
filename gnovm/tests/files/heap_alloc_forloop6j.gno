package main

func main() {
	a := 10
	b := 20
	f := func() {
		a, b = b+1, a
	}
	f()
	println(a, b) // 21 10
}

// Preprocessed:
// file{ package main; func main() { a<!~VPBlock(1,0)> := (const (10 int)); b<!~VPBlock(1,1)> := (const (20 int)); f<!VPBlock(1,2)> := func func(){ a<~VPBlock(1,1)>, b<~VPBlock(1,0)> = b<~VPBlock(1,0)> + (const (1 int)), a<~VPBlock(1,1)> }<b<()~VPBlock(1,1)>, a<()~VPBlock(1,0)>>; f<VPBlock(1,2)>(); (const (println func(...interface {})))(a<~VPBlock(1,0)>, b<~VPBlock(1,1)>) } }

// Output:
// 21 10
