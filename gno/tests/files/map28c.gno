package main

type Values map[string][]string

func (v Values) Set(key, value string) {
	v[key] = []string{value}
}

func main() {
	value1 := Values{}

	value1.Set("first", "v1")
	value1.Set("second", "v2")

	// Index case: v, ok := x[k], x is map.
	_, ok := value1["first"]

	println(ok)
}

// Output:
// true
