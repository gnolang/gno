package main

func Search(n int, f func(int) bool) int {
	f(1)
	return 0
}

// TODO: identify this pattern, optimize.
func main() {
	for x := 0; x < 2; x++ {
		count := 0
		println(" first: count: ", count)
		Search(1, func(i int) bool { count++; return i >= x })
		println("second: count: ", count)
	}
}

// Preprocessed:
// file{ package main; func Search(n (const-type int), f func(.arg_0 (const-type int))  (const-type bool))  (const-type int) { f<VPBlock(1,1)>((const (1 int))); return (const (0 int)) }; func main() { for x<!~VPBlock(1,0)> := (const (0 int)); x<~VPBlock(1,0)> < (const (2 int)); x<~VPBlock(1,0)>++ { count<!~VPBlock(1,1)> := (const (0 int)); (const (println func(xs ...interface{})()))((const (" first: count: " string)), count<~VPBlock(1,1)>); Search<VPBlock(4,0)>((const (1 int)), func func(i (const-type int))  (const-type bool){ count<~VPBlock(1,2)>++; return (const-type bool)(i<VPBlock(1,0)> >= x<~VPBlock(1,3)>) }<count<()~VPBlock(1,1)>, x<()~VPBlock(1,0)>>); (const (println func(xs ...interface{})()))((const ("second: count: " string)), count<~VPBlock(1,1)>) } } }

// Output:
// first: count:  0
// second: count:  1
//  first: count:  0
// second: count:  1
