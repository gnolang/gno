// param: 4 8 16

package main

func main() {
	for i := 0; i < {{ .N }}; i++ {
		fib({{ .Param }})
	}
}

func fib(n int) int {
	if n < 2 {
		return 1
	}
	return fib(n-1) + fib(n-2)
}
