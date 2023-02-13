package main

type Values map[string][]string

func (v Values) Set(key, value string) {
	v[key] = []string{value}
}

func main() {
	value1 := Values{}

	value1.Set("first", "v1")
	value1.Set("second", "v2")

	_, ok := value1["first"]

	println(ok)
}

// Output:
// true
