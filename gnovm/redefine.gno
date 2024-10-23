package main

type foo struct {
	name string
	num  int
}

// var ff = []foo{foo{name: "a", num: 0}, foo{name: "b", num: 1}}
var ss = []int{1, 2, 3}

func main() {
	for _, s := range ss {
		s := s
		println(s)
	}
}
