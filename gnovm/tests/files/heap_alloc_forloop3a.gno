package main

type f func()

var fs []f

func forLoopClosure() {
	defer func() {
		for _, f := range fs {
			f()
		}
	}()

	for i := 0; i < 3; i++ {
		x := i
		println(x)
		z := i
		fs = append(fs, func() { println(z) })
	}
}

func main() {
	forLoopClosure()
}

// This does make 'z' NameExprTypeHeapDefine.
// You can tell by the preprocess printout of z<!~...> and z<()~...>.

// Output:
// 0
// 1
// 2
// 0
// 1
// 2
