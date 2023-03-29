package main

type nat3 []Word
type Word uint


func (n nat3) add3() bool {
	return true
}

func main() {
	items := map[string]nat3{}

	n := []Word{0}

	items["test"] = n
	r := items["test"]

	println(r.add3())
}

// Output:
// true
