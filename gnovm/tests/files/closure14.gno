package main

import "fmt"

func foo() (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf("%s", v)
			}
		}
	}()

	panic("xxx")
}

func main() {
	err := foo()
	println(err.Error())
}

// Output:
// xxx
