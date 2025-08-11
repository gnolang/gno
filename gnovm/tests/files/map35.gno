package main

import "strconv"

type stringMap map[*int]int

var (
	a        int = 1
	accounts     = stringMap{}
	ptr          = &a
)

func Register() {
	accounts[ptr] = 1
}

func Set(v int) {
	*ptr = v
}

func Render(s string) string {
	for k := range accounts {
		s += "key:" + strconv.Itoa(*k)
	}
	return s
}

func main() {
	Register()
	Set(4)
	println(Render(""))
}

// Output:
// key:4
