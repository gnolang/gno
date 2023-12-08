package main

type Int int8

func (i Int) Inc() Int {
	i++
	return i
}

// right is typed const, can not use as its correspondence declared type
func main() {
	var a Int
	a = Int(int8(0))
	a = a.Inc()
	println(a)
}

// Output:
// (1 main.Int)
