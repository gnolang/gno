package main

type time int

var time string

func main() {
  time = "hello"
  println(time)
}

// Error:
// files/redeclaration-global0.gno:5:5: time redeclared in this block
//	previous declaration at files/redeclaration-global0.gno:3:6
