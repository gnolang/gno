package main

func main() {

	var fs []func()
	defer func() {
		for _, f := range fs {
			f()
		}
	}()

	for i := 0; i < 2; i++ {
		x := i
		for j := 0; j < 2; j++ {
			y := j
			fs = append(fs, func() { println(x + y) })
		}
	}
}

// Output:
// 0
// 1
// 1
// 2
