package main

func main() {
	x := 1
	y := 2
	f := func() {
		x, y = y, x
	}
	f()
	println(x, y)
}

// Preprocessed:
// file{ package main; func main() { x<!~VPBlock(1,0)> := (const (1 int)); y<!~VPBlock(1,1)> := (const (2 int)); f<!VPBlock(1,2)> := func func(){ x<~VPBlock(1,1)>, y<~VPBlock(1,0)> = y<~VPBlock(1,0)>, x<~VPBlock(1,1)> }<y<()~VPBlock(1,1)>, x<()~VPBlock(1,0)>>; f<VPBlock(1,2)>(); (const (println func(...interface {})))(x<~VPBlock(1,0)>, y<~VPBlock(1,1)>) } }

// Output:
// 2 1
