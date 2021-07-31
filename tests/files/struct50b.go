package main

type Node struct {
	Name  string
	Child []Node
}

func main() {
	a := Node{Name: "hello"}
	a.Child = append([]Node{}, Node{Name: "world"})
	println(a)
	a.Child[0].Child = append([]Node{}, Node{Name: "sunshine"})
	println(a)
}

// Output:
// struct{("hello" string),(slice[(struct{("world" string),(undefined)} main.Node)] []main.Node)}
// struct{("hello" string),(slice[(struct{("world" string),(slice[(struct{("sunshine" string),(undefined)} main.Node)] []main.Node)} main.Node)] []main.Node)}
