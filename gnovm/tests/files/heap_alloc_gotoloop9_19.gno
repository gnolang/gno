package main

func main() {
	var y, counter int
	var f []func() func() int
	defer func() {
		for _, ff := range f {
			println(ff()())
		}
	}()

LABEL_1:
	if counter == 2 {
		return
	}
	x := y
	f = append(f, func() func() int {
		return func() int { return x }
	})
	y++
	counter++
	goto LABEL_1
}

// Preprocessed:
// file{ package main; func main() { var y<!VPBlock(1,0)>, counter<!VPBlock(1,1)> (const-type int); var f<!VPBlock(1,2)> []func()  func()  (const-type int); defer func func(){ for _<VPBlock(0,0)>, ff<VPBlock(1,0)> := range f<VPBlock(3,2)> { (const (println func(xs ...interface{})()))(ff<VPBlock(1,0)>()()) } }(); if counter<VPBlock(2,1)> == (const (2 int)) { return }; x<!~VPBlock(1,3)> := y<VPBlock(1,0)>; f<VPBlock(1,2)> = (const (append func(x []func()( func()( int)),args ...func()( func()( int)))(res []func()( func()( int)))))(f<VPBlock(1,2)>, func func()  func()  (const-type int){ return func func()  (const-type int){ return x<~VPBlock(2,1)> } }<x<()~VPBlock(1,3)>>); y<VPBlock(1,0)>++; counter<VPBlock(1,1)>++; goto LABEL_1<0,3> } }

// Output:
// 0
// 1
