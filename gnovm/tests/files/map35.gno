package main

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

func main() {
	Register()
	Set(4)
	println(accounts)
}

// Output:
// (map{(&(4 int) *int):(1 int)} main.stringMap)
