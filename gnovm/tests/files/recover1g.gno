package main

func level3() {
	panic("level3")
	defer func() {
		if r3 := recover(); r3 != nil {
			println("recover level3 in level3")
		}
	}()
}

func main() {
	level3()
}

// Error:
// panic: level3
