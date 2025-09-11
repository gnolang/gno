package main

func main() {
	var y, counter int
	var f []func()
	defer func() {
		for _, ff := range f {
			ff()
		}
	}()

LABEL_1:
	x := y
	if counter == 2 {
		counter = 0
		goto LABEL_2
	}
	f = append(f, func() { println(x) })
	y++
	counter++
	goto LABEL_1

LABEL_2:
	if counter == 2 {
		return
	}
	z := y
	f = append(f, func() { println(z) })
	y++
	counter++
	goto LABEL_2
}

// Preprocessed:
// file{ package main; func main() { var y<!VPBlock(1,0)>, counter<!VPBlock(1,1)> (const-type int); var f<!~VPBlock(1,2)> []func(); defer func func(){ for _<VPBlock(0,0)>, ff<!VPBlock(1,0)> := range f<~VPBlock(2,0)> { ff<VPBlock(1,0)>() } }<f<()~VPBlock(1,2)>>(); x<!~VPBlock(1,3)> := y<VPBlock(1,0)>; if counter<VPBlock(2,1)> == (const (2 int)) { counter<VPBlock(2,1)> = (const (0 int)); goto LABEL_2<1,0,9> }; f<~VPBlock(1,2)> = (const (append func([]func(), ...func()) []func()))(f<~VPBlock(1,2)>, func func(){ (const (println func(...interface {})))(x<~VPBlock(1,0)>) }<x<()~VPBlock(1,3)>>); y<VPBlock(1,0)>++; counter<VPBlock(1,1)>++; goto LABEL_1<0,0,3>; if counter<VPBlock(2,1)> == (const (2 int)) { return }; z<!~VPBlock(1,4)> := y<VPBlock(1,0)>; f<~VPBlock(1,2)> = (const (append func([]func(), ...func()) []func()))(f<~VPBlock(1,2)>, func func(){ (const (println func(...interface {})))(z<~VPBlock(1,0)>) }<z<()~VPBlock(1,4)>>); y<VPBlock(1,0)>++; counter<VPBlock(1,1)>++; goto LABEL_2<0,0,9> } }



// Output:
// 0
// 1
// 2
// 3
