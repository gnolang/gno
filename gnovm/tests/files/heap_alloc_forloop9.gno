package main

import "fmt"

// recursive closure does not capture
func main() {
	var fns []func(int) int
	var recursiveFunc func(int) int

	for i := 0; i < 3; i++ {
		recursiveFunc = func(num int) int {
			x := i
			println("value of x: ", x)
			if num <= 0 {
				return 1
			}
			return num * recursiveFunc(num-1)
		}
		fns = append(fns, recursiveFunc)
	}

	for i, r := range fns {
		result := r(i)
		fmt.Printf("Factorial of %d is: %d\n", i, result)
	}
}

// go 1.22 loop var is not supported for now.

// Preprocessed:
// file{ package main; import fmt fmt; func main() { var fns<!VPBlock(1,0)> []func(.arg_0 (const-type int))  (const-type int); var recursiveFunc<!VPBlock(1,1)> func(.arg_0 (const-type int))  (const-type int); for i<!~VPBlock(1,0)> := (const (0 int)); i<~VPBlock(1,0)> < (const (3 int)); i<~VPBlock(1,0)>++ { recursiveFunc<VPBlock(2,1)> = func func(num (const-type int))  (const-type int){ x<!VPBlock(1,1)> := i<~VPBlock(1,3)>; (const (println func(xs ...interface{})()))((const ("value of x: " string)), x<VPBlock(1,1)>); if num<VPBlock(2,0)> <= (const (0 int)) { return (const (1 int)) }; return num<VPBlock(1,0)> * recursiveFunc<VPBlock(3,1)>(num<VPBlock(1,0)> - (const (1 int))) }<i<()~VPBlock(1,0)>>; fns<VPBlock(2,0)> = (const (append func(x []func(.arg_0 int)( int),args ...func(.arg_0 int)( int))(res []func(.arg_0 int)( int))))(fns<VPBlock(2,0)>, recursiveFunc<VPBlock(2,1)>) }; for i<!VPBlock(1,0)>, r<!VPBlock(1,1)> := range fns<VPBlock(2,0)> { result<!VPBlock(1,2)> := r<VPBlock(1,1)>(i<VPBlock(1,0)>); fmt<VPBlock(3,0)>.Printf((const ("Factorial of %d is: %d\n" string)), (const-type gonative{interface {}})(i<VPBlock(1,0)>), (const-type gonative{interface {}})(result<VPBlock(1,2)>)) } } }

// Output:
// value of x:  3
// Factorial of 0 is: 1
// value of x:  3
// value of x:  3
// Factorial of 1 is: 1
// value of x:  3
// value of x:  3
// value of x:  3
// Factorial of 2 is: 2
