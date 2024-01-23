package main

import (
	"errors"
)

func foo() (err error) {
	y := 1
	defer func() {
		if r := recover(); r != nil {
			switch y {
			case 1:
				err = errors.New("xxx")
			default:
				err = nil
			}
		}
	}()
	return
}

func main() {
	foo()
}
