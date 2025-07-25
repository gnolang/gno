package main

import "fmt"

func trymake() func() string {
	if s := "hello"; true {
		return func() string {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("Recovered, r: ", r)
				}
				fmt.Println("Recovered, s: ", s)
			}()

			defer func() {
				s = "!!!"
				panic(s)
			}()

			println(s) // hello
			return s
		}
	}
	return func() string { return "else..." }
}

func main() {
	s1 := trymake()()
	println("return value: ", s1)
}

// Output:
// hello
// Recovered, r:  !!!
// Recovered, s:  !!!
// return value:  hello
