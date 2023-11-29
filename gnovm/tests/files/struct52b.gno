package main

type Node struct {
	Name  string
	Child map[string]Node
}

func main() {
	a := Node{Name: "hello", Child: map[string]Node{}}
	a.Child["1"] = Node{Name: "world", Child: map[string]Node{}}
	println(a)
	a.Child["1"].Child["1"] = Node{Name: "sunshine", Child: map[string]Node{}}
	println(a)
}

// Output:
// (struct{("hello" string),(map{("1" string):(struct{("world" string),(map{} map[string]main.Node)} main.Node)} map[string]main.Node)} main.Node)
// (struct{("hello" string),(map{("1" string):(struct{("world" string),(map{("1" string):(struct{("sunshine" string),(map{} map[string]main.Node)} main.Node)} map[string]main.Node)} main.Node)} map[string]main.Node)} main.Node)
