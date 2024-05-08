package recur

type A struct {
	X B
}

func Do() {
	var p, q A
	println(p == q)
}
