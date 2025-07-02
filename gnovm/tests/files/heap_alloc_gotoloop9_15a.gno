package main

var y, counter int
var f []func()

func main() {
LABEL_1:
	x := y
	if counter == 2 {
		counter = 0
		bar()
		for _, ff := range f { // XXX, why defer on this not work
			ff()
		}
		return
	}
	f = append(f, func() { println(x) })
	y++
	counter++
	goto LABEL_1
}

func bar() {
	println("---bar---")
LABEL_2:
	if counter == 2 {
		println("---end---")
		return
	}
	z := y
	f = append(f, func() { println(z) })
	y++
	counter++
	goto LABEL_2
}

// Preprocessed:
// file{ package main; var y<!~VPBlock(2,0)>, counter<!~VPBlock(2,1)> (const-type int); var f<!~VPBlock(2,2)> []func(); func main() { x<!~VPBlock(1,0)> := y<~VPBlock(3,0)>; if counter<~VPBlock(4,1)> == (const (2 int)) { counter<~VPBlock(4,1)> = (const (0 int)); bar<VPBlock(4,4)>(); for _<VPBlock(0,0)>, ff<!VPBlock(1,0)> := range f<~VPBlock(5,2)> { ff<VPBlock(1,0)>() }; return }; f<~VPBlock(3,2)> = (const (append func([]func(), ...func()) []func()))(f<~VPBlock(3,2)>, func func(){ (const (println func(...interface {})))(x<~VPBlock(1,0)>) }<x<()~VPBlock(1,0)>>); y<~VPBlock(3,0)>++; counter<~VPBlock(3,1)>++; goto LABEL_1<0,0,0> }; func bar() { (const (println func(...interface {})))((const ("---bar---" string))); if counter<~VPBlock(4,1)> == (const (2 int)) { (const (println func(...interface {})))((const ("---end---" string))); return }; z<!~VPBlock(1,0)> := y<~VPBlock(3,0)>; f<~VPBlock(3,2)> = (const (append func([]func(), ...func()) []func()))(f<~VPBlock(3,2)>, func func(){ (const (println func(...interface {})))(z<~VPBlock(1,0)>) }<z<()~VPBlock(1,0)>>); y<~VPBlock(3,0)>++; counter<~VPBlock(3,1)>++; goto LABEL_2<0,0,1> } }



// Output:
// ---bar---
// ---end---
// 0
// 1
// 2
// 3
