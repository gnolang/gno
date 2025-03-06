package main

func main() {
	a := 1.0
	const c1 = len([1]int{int(a)})
	println(c1)
	nestedArr := [1]int{func() int { println("hey!!"); return 1 }()}
	const b = len(nestedArr)
	println(b)
}

// Output:
// 1
// hey!!
// 1
