package main

func main() {
	var f func()
	for i := 0; i < 5; {

		println("enter, i: ", i)

		if i == 2 {
			println("---continue")
			i++ // .redefine_i
			// .loopvar_i = .redefine_i
			continue
		}

		//i++
		println("after inc, i: ", i) // 1

		f = func() {
			start := i
			println("capture, start: ", start)
		}
		f()
	}
}
