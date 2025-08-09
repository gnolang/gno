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
	i := 0
	for k := range accounts {
		s += "" + strconv.Itoa(i) + "key.num:" + strconv.Itoa(k)
		i++
	}
	return s
}

func main() {
	Register()
	Set(4)
	println(Render(""))
}

// Output:
// 0key.num:1
