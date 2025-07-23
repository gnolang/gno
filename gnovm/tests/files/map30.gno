package main

func main() {
	var v map[string]string = nil
	for k, v := range v {
		println(k, v)
	}
	for v := range v {
		println(v)
	}
	println("done") // should not panic
}

// Output:
// done
