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

	var v interface{}
	v = 9
	_ = v.(A)
}

func test2() {
	defer func() {
		if r := recover(); r != nil {
			println(r)
		}
	}()

	var v interface{}
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
// int doesn't implement interface{Do func(string)()}
// *int doesn't implement interface{Do func(string)()}
