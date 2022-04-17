package main

func main() {
  var foo int
  foo = 2

  type foo struct{}
  var bar foo
  println(bar)
}

// Error:
// files/redeclaration3.gno:7:8: foo redeclared in this block
