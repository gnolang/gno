package main

func BalanceOf(i int) int {
	return 100
}

func main() {
	bar := 100
	for i, tc := range []struct {
		name     string
		expected interface{}
		fn       func() interface{}
	}{
		{"BalanceOf(admin, tid1)", 100, func() interface{} { return BalanceOf(bar) }},
	} {
		println("hey")
		println(i)
		println(tc)
	}
}

// Output:
// hey
// 0
// struct{("BalanceOf(admin, tid1)" string),(100 int),( func()( interface{}))}
