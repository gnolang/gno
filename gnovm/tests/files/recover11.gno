package main

func level3() {
	panic("level3")
	defer func() {
		if r3 := recover(); r3 != nil {
			println("recover level3 in level3")
		}
	}()
}

func level2() {
	defer func() {
		if r := recover(); r != nil {
			println("recover level3 in level2")
		}
	}()
	defer level3()
	panic("level2")
}

func level1() {
	defer func() {
		if r := recover(); r != nil {
			println("recover level1 in level1")
		}
	}()
	level2()
	panic("level1")
}

func main() {
	level1()
}

// Error:
// panic: level2
//	panic: level3[recovered]

// Output:
// recover level3 in level2
// recover level1 in level1
