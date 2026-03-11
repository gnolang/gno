package main

func main() {
	for i, v := range []int{10, 20, 30} {
		i = i+1
		v = v+1
		println("inner i:", i) 
		println("inner v:", v) 
	}
}

// Output:
// inner i: 1
// inner v: 11
// inner i: 2
// inner v: 21
// inner i: 3
// inner v: 31

// Preprocessed:
// file{ package main; func main() { for i<!VPBlock(1,0)>, v<!VPBlock(1,1)> := range (const-type []int){(const (10 int)), (const (20 int)), (const (30 int))} { i<VPBlock(1,0)> = i<VPBlock(1,0)> + (const (1 int)); v<VPBlock(1,1)> = v<VPBlock(1,1)> + (const (1 int)); (const (println func(...interface {})))((const ("inner i:" string)), i<VPBlock(1,0)>); (const (println func(...interface {})))((const ("inner v:" string)), v<VPBlock(1,1)>) } } }
