package main

func main() {
	for i:=0; i<3; i++{
		println("start: ", i)
		{
			println("shadow: ", i) 
		}
		println("end: ", i)
	}
}

// Preprocessed:
// file{ package main; func main() { for .loopvar_i<!VPBlock(1,0)> := (const (0 int)); .loopvar_i<VPBlock(1,0)> < (const (3 int)); .loopvar_i<VPBlock(1,0)>++ { (const (println func(...interface {})))((const ("start: " string)), .loopvar_i<VPBlock(1,0)>); { (const (println func(...interface {})))((const ("shadow: " string)), .loopvar_i<VPBlock(2,0)>) }; (const (println func(...interface {})))((const ("end: " string)), .loopvar_i<VPBlock(1,0)>) } } }

// Output:
// start:  0
// shadow:  0
// end:  0
// start:  1
// shadow:  1
// end:  1
// start:  2
// shadow:  2
// end:  2
