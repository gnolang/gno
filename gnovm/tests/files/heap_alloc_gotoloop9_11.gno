package main

func main() {
	var y, counter int
	var f []func()
	defer func() {
		for _, ff := range f {
			ff()
		}
	}()

	// this is actually an implicit for loop
LABEL_1:
	if counter == 2 {
		return
	}
	var _, x = 0, y
	f = append(f, func() { println(x) })
	y++
	counter++
	goto LABEL_1
}

// Preprocessed:
// file{ package main; func main() { var y<!VPBlock(1,0)>, counter<!VPBlock(1,1)> (const-type int); var f<!~VPBlock(1,2)> []func(); defer func func(){ for _<VPBlock(0,0)>, ff<!VPBlock(1,0)> := range f<~VPBlock(2,0)> { ff<VPBlock(1,0)>() } }<f<()~VPBlock(1,2)>>(); if counter<VPBlock(2,1)> == (const (2 int)) { return }; var _<VPBlock(0,0)>, x<!~VPBlock(1,3)> = (const (0 int)), y<VPBlock(1,0)>; f<~VPBlock(1,2)> = (const (append func([]func(), ...func()) []func()))(f<~VPBlock(1,2)>, func func(){ (const (println func(...interface {})))(x<~VPBlock(1,0)>) }<x<()~VPBlock(1,3)>>); y<VPBlock(1,0)>++; counter<VPBlock(1,1)>++; goto LABEL_1<0,0,3> } }



// Output:
// 0
// 1
