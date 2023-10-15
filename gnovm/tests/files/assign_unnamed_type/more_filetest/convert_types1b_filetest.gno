package main

type nat []int

func main() {
  var a nat
  var b []int

  a = (nat)([]int{0})
  b = ([]int)(nat{1})
  println(a)
	println(b)

  a = (nat)(b)
  b = ([]int)(a)
  println(a)
	println(b)

}
