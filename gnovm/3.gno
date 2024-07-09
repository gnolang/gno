package main


func main() {
  type S struct {
	  T S
  }
	var a, b S
	println(a == b)
}


