package main

func main() {
	a := 0
	// short-circuit evaluation
	if i := 0; i>0 && i/a == 1 {
		println("ok")
	}else{
		println("fail")
	}
}

// Output:
// fail
