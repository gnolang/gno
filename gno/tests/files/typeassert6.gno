package main

type A interface {
	Do(s string)
}

func main() {
	var v interface{}
	v = 9

	if _, ok := v.(A); !ok {
		println(v)
	}

	vp := new(int)
	*vp = 99
	v = vp
	if _, ok := v.(A); !ok {
		println(*(v.(*int)))
	}
}

// Output:
// 9
// 99
