package main

func main() {
	for i := 0; i < 2; i++{
		println("i: ", i)
	}
}

// Preprocessed:
// file{ package main; func main() { for .loopvar_i<!@VPBlock(1,0)> := (const (0 int)); .loopvar_i<*VPBlock(1,0)> < (const (2 int)); .loopvar_i<*VPBlock(1,0)>++ { (const (println func(...interface {})))((const ("i: " string)), .loopvar_i<*VPBlock(2,0)>) }  } }

// Output:
// i:  0
// i:  1
