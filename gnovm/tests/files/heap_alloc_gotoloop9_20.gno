package main

func main() {
	var y, counter int
	var f []func() (int, func() int)
	defer func() {
		for _, ff := range f {
			n, f := ff()
			println(n + f())
		}
	}()

LABEL_1:
	if counter == 2 {
		return
	}
	x := y
	z := y
	f = append(f, func() (int, func() int) {
		return z, func() int { return x }
	})
	y++
	counter++
	goto LABEL_1
}

// Preprocessed:
// file{ package main; func main() { var y<!VPBlock(1,0)>, counter<!VPBlock(1,1)> (const-type int); var f<!~VPBlock(1,2)> []func() .res.0 (const-type int), .res.1 func() .res.0 (const-type int); defer func func(){ for _<VPBlock(0,0)>, ff<!VPBlock(1,0)> := range f<~VPBlock(2,0)> { n<!VPBlock(1,1)>, f<!VPBlock(1,2)> := ff<VPBlock(1,0)>(); (const (println func(...interface {})))(n<VPBlock(1,1)> + f<VPBlock(1,2)>()) } }<f<()~VPBlock(1,2)>>(); if counter<VPBlock(2,1)> == (const (2 int)) { return }; x<!~VPBlock(1,3)> := y<VPBlock(1,0)>; z<!~VPBlock(1,4)> := y<VPBlock(1,0)>; f<~VPBlock(1,2)> = (const (append func([]func() (int, func() int), ...func() (int, func() int)) []func() (int, func() int)))(f<~VPBlock(1,2)>, func func() .res.0 (const-type int), .res.1 func() .res.0 (const-type int){ return z<~VPBlock(1,2)>, func func() .res.0 (const-type int){ return x<~VPBlock(1,1)> }<x<()~VPBlock(1,3)>> }<z<()~VPBlock(1,4)>, x<()~VPBlock(1,3)>>); y<VPBlock(1,0)>++; counter<VPBlock(1,1)>++; goto LABEL_1<0,0,3> } }



// Output:
// 0
// 2
