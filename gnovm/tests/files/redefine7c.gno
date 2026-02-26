package main

func main() {
	for i, v := range []int{1, 2, 3} {
		i := i
		println(i, v)
	}
}

// Preprocessed:
// file{ package main; func main() { for i<!VPBlock(1,0)>, v<!VPBlock(1,1)> := range (const-type []int){(const (1 int)), (const (2 int)), (const (3 int))} { i<!VPBlock(1,0)> := i<VPBlock(1,0)>; (const (println func(...interface {})))(i<VPBlock(1,0)>, v<VPBlock(1,1)>) } } }

// Output:
// 0 1
// 1 2
// 2 3
