package main

func main() {
	defer func() {
		println(recover())
	}()
	defer func() {
		defer func() {
			recover()
		}()
		defer recover()
		panic("3")
	}()
	defer func() {
		defer func() {
			recover()
		}()
		defer recover()
		panic("2")
	}()
	panic("1")
}

// Output:
// 1
