package main

type A interface {
	Do(s string)
}

func test1() {
	defer func() {
		if r := recover(); r != nil {
			println(r)
		}
	}()

	var v any
	v = 9
	_ = v.(A)
}

func test2() {
	defer func() {
		if r := recover(); r != nil {
			println(r)
		}
	}()

	var v any
	vp := new(int)
	*vp = 99
	v = vp
	_ = v.(A)
}

func main() {
	test1()
	test2()
}

// Output:
// int doesn't implement interface {Do func(string)} (missing method Do)
// *int doesn't implement interface {Do func(string)} (missing method Do)
