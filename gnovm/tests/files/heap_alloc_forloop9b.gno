package main

func main() {
	var y int
	var f []func()
	defer func() {
		for _, ff := range f {
			ff()
		}
	}()

	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			x := y
			f = append(f, func() { println(x) })
			y++
		}
	}
}

// Preprocessed:
// file{ package main; func main() { var y<!VPBlock(1,0)> (const-type int); var f<!~VPBlock(1,1)> []func(); defer func func(){ for _<VPBlock(0,0)>, ff<!VPBlock(1,0)> := range f<~VPBlock(2,0)> { ff<VPBlock(1,0)>() } }<f<()~VPBlock(1,1)>>(); for .loopvar_i<!VPBlock(1,0)> := (const (0 int)); .loopvar_i<VPBlock(1,0)> < (const (2 int)); .loopvar_i<VPBlock(1,0)>++ { for .loopvar_j<!VPBlock(1,0)> := (const (0 int)); .loopvar_j<VPBlock(1,0)> < (const (2 int)); .loopvar_j<VPBlock(1,0)>++ { x<!~VPBlock(1,1)> := y<VPBlock(3,0)>; f<~VPBlock(3,1)> = (const (append func([]func(), ...func()) []func()))(f<~VPBlock(3,1)>, func func(){ (const (println func(...interface {})))(x<~VPBlock(1,0)>) }<x<()~VPBlock(1,1)>>); y<VPBlock(3,0)>++ } } } }

// Output:
// 0
// 1
// 2
// 3
