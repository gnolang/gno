package main

type matrix struct {
	e []int
}

func (a matrix) equal() bool {
	for _ = range a.e {
	}
	for _, _ = range a.e {
	}
	for range a.e {
	}
	return true
}

func main() {
	var a matrix
	var i interface{}
	i = true && a.equal()
	println(i)
}

// Output:
// true
