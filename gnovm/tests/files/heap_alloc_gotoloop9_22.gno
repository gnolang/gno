package main

func main() {
	var y, counter int
	var f []func()
	defer func() {
		for _, ff := range f {
			ff()
		}
	}()

	for i := 0; i < 2; i++ {
		counter = 0
	LABEL_1:
		if counter == 2 {
			continue
		}
		x := y
		f = append(f, func() { println(x) })
		y++
		counter++
		goto LABEL_1
	}
}

// Preprocessed:
// file{ package main; func main() { var y<!VPBlock(1,0)>, counter<!VPBlock(1,1)> (const-type int); var f<!~VPBlock(1,2)> []func(); defer func func(){ for _<VPBlock(0,0)>, ff<!VPBlock(1,0)> := range f<~VPBlock(2,0)> { ff<VPBlock(1,0)>() } }<f<()~VPBlock(1,2)>>(); for .loopvar_i<!VPBlock(1,0)> := (const (0 int)); .loopvar_i<VPBlock(1,0)> < (const (2 int)); .loopvar_i<VPBlock(1,0)>++ { counter<VPBlock(2,1)> = (const (0 int)); if counter<VPBlock(3,1)> == (const (2 int)) { continue }; x<!~VPBlock(1,1)> := y<VPBlock(2,0)>; f<~VPBlock(2,2)> = (const (append func([]func(), ...func()) []func()))(f<~VPBlock(2,2)>, func func(){ (const (println func(...interface {})))(x<~VPBlock(1,0)>) }<x<()~VPBlock(1,1)>>); y<VPBlock(2,0)>++; counter<VPBlock(2,1)>++; goto LABEL_1<0,0,1> } } }

// Output:
// 0
// 1
// 2
// 3
