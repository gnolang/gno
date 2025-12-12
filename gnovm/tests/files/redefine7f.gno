package main

func main() {
	for i := 0; i < 3; i++ {
		for j := i; j < 3; j++ {
			println(j)
		}
	}
}

// Preprocessed:
// file{ package main; func main() { for .loopvar_i<!@VPBlock(1,0)> := (const (0 int)); .loopvar_i<*VPBlock(1,0)> < (const (3 int)); .loopvar_i<*VPBlock(1,0)>++ { for .loopvar_j<!@VPBlock(1,0)> := .loopvar_i<*VPBlock(2,0)>; .loopvar_j<*VPBlock(1,0)> < (const (3 int)); .loopvar_j<*VPBlock(1,0)>++ { (const (println func(...interface {})))(.loopvar_j<*VPBlock(1,0)>) } } } }

// Output:
// 0
// 1
// 2
// 1
// 2
// 2
