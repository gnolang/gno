// Sample target for the debugger fallthrough test.
package main

func main() {
	switch 1 {
	case 1:
		y := 5
		_ = y
		fallthrough
	case 2:
		println("two")
	}
}
