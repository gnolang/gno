package main

import "strconv"

type stringMap map[int]int

var (
	key      int = 1
	accounts     = stringMap{}
)

func Register() {
	accounts[key] = 1
}

func Set(v int) {
	key = v
}

func Render(s string) string {
	for k := range accounts {
		s += "key:" + strconv.Itoa(k)
	}
	return s
}

func main() {
	Register()
	Set(4)
	println(Render(""))
}

// Output:
// key:1
