package main

func level3() {
	panic("level3")
	defer func() {
		if r := recover(); r != nil {
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
			println("recover level3 in level1")
		}
	}()
	level2()
	panic("level1")
}

func main() {
	level1()
}

// Output:
// recover level3 in level2
// recover level3 in level1
