package main

func main() {
	var counter int
	var f []func()
	defer func() {
		for _, ff := range f {
			ff()
		}
	}()

	f1 := func() {
	LABEL_1:
		if counter == 2 {
			return
		}

		func() {
			x := counter
			f = append(f, func() { println(x) })
		}()
		counter++
		goto LABEL_1
	}

	f1()
}

// Preprocessed:
// file{ package main; func main() { var counter<!VPBlock(1,0)> (const-type int); var f<!VPBlock(1,1)> []func(); defer func func(){ for _<VPBlock(0,0)>, ff<VPBlock(1,0)> := range f<VPBlock(3,1)> { ff<VPBlock(1,0)>() } }(); f1<!VPBlock(1,2)> := func func(){ if counter<VPBlock(3,0)> == (const (2 int)) { return }; func func(){ x<!VPBlock(1,0)> := counter<VPBlock(3,0)>; f<VPBlock(3,1)> = (const (append func(x []func()(),args ...func()())(res []func()())))(f<VPBlock(3,1)>, func func(){ (const (println func(xs ...interface{})()))(x<VPBlock(2,0)>) }) }(); counter<VPBlock(2,0)>++; goto LABEL_1<0,0> }; f1<VPBlock(1,2)>() } }

// Output:
// 0
// 1
