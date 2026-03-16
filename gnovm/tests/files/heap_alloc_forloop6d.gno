package main

func main() {
	x := 1
	y := 3
	f := func() int {
		x += y
		return x
	}
	println(f())
}

// Preprocessed:
// file{ package main; func main() { x<!~VPBlock(1,0)> := (const (1 int)); y<!~VPBlock(1,1)> := (const (3 int)); f<!VPBlock(1,2)> := func func() .res.0 (const-type int){ x<~VPBlock(1,2)> += y<~VPBlock(1,1)>; return x<~VPBlock(1,2)> }<y<()~VPBlock(1,1)>, x<()~VPBlock(1,0)>>; (const (println func(...interface {})))(f<VPBlock(1,2)>()) } }

// Output:
// 4
