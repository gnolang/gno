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
		fs = append(fs, func() {
			z := i
			println(z)
		})
	}
}

func main() {
	forLoopClosure()
}

// This does make 'i' NameExprTypeHeapDefine.
// You can tell by the preprocess printout of i<!~...> and i<()~...>.

// Output:
// 3
// 3
// 3
