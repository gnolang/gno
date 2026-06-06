package main

func main() {
	for i:=0; i<3; i++{
		println("start: ", i)
		{
			i := i+1
			println("shadow: ", i) // should not rename.
		}
		println("end: ", i)
	}
}

// Preprocessed:
// file{ package main; func main() { for i.loopvar<!VPBlock(1,0)> := (const (0 int)); i.loopvar<VPBlock(1,0)> < (const (3 int)); i.loopvar<VPBlock(1,0)>++ { (const (println func(...interface {})))((const ("start: " string)), i.loopvar<VPBlock(1,0)>); { i<!VPBlock(1,0)> := i.loopvar<VPBlock(2,0)> + (const (1 int)); (const (println func(...interface {})))((const ("shadow: " string)), i<VPBlock(1,0)>) }; (const (println func(...interface {})))((const ("end: " string)), i.loopvar<VPBlock(1,0)>) } } }

// Output:
// start:  0
// shadow:  1
// end:  0
// start:  1
// shadow:  2
// end:  1
// start:  2
// shadow:  3
// end:  2
