package main

func main() {
	x := 10
	y := 20
	z := add(x, y)
	println("Result:", z)
	
	// Test with array
	arr := [3]int{1, 2, 3}
	for i, v := range arr {
		println("arr[", i, "] =", v)
	}
	
	// Test with slice
	slice := []string{"hello", "world"}
	slice = append(slice, "!")
	
	// Test with struct
	type Person struct {
		Name string
		Age  int
	}
	p := Person{Name: "Alice", Age: 30}
	println("Person:", p.Name, p.Age)
}

func add(a, b int) int {
	sum := a + b
	return sum
}