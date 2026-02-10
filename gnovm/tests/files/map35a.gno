package main

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

func main() {
	Register()
	Set(4)
	println(accounts)
	println(key)
}

// Output:
// (map{(1 int):(1 int)} main.stringMap)
// 4
