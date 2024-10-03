package main

type Values map[string][]string

func (v Values) Set(key, value string) {
	v[key] = []string{value}
}

func main() {
	value1 := Values{}

	value1.Set("first", "v1")
	value1.Set("second", "v2")

	l := 0
	for k, v := range value1 {
		l += len(k) + len(v)
	}
	println(l)
}

// Output:
// 13
