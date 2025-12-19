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
// file{ package main; func Search(n (const-type int), f func(.arg_0 (const-type int)) .res.0 (const-type bool)) .res.0 (const-type int) { f<VPBlock(1,1)>((const (1 int))); return (const (0 int)) }; func main() { for .loopvar_x<!@VPBlock(1,0)> := (const (0 int)); .loopvar_x<*VPBlock(1,0)> < (const (2 int)); .loopvar_x<*VPBlock(1,0)>++ { .redefine_x<!~VPBlock(1,1)> := .loopvar_x<*VPBlock(2,0)>; count<!~VPBlock(1,0)> := (const (0 int)); (const (println func(...interface {})))((const (" first: count: " string)), count<~VPBlock(1,0)>); Search<VPBlock(5,0)>((const (1 int)), func func(i (const-type int)) .res.0 (const-type bool){ count<~VPBlock(1,2)>++; return (const-type bool)(i<VPBlock(1,0)> >= .redefine_x<~VPBlock(1,3)>) }<count<()~VPBlock(1,0)>, .redefine_x<()~VPBlock(1,1)>>); (const (println func(...interface {})))((const ("second: count: " string)), count<~VPBlock(1,0)>); .loopvar_x<*VPBlock(2,0)> = .redefine_x<~VPBlock(1,1)> }  } }

// Output:
//  first: count:  0
// second: count:  1
//  first: count:  0
// second: count:  1
