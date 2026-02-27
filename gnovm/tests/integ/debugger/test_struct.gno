package main

type Person struct {
	Name string
	Age  int
	Tags []string
}

func main() {
	p := Person{
		Name: "Alice",
		Age:  30,
		Tags: []string{"developer", "gopher"},
	}
	println(p.Name) // Set breakpoint here
}