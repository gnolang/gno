// Blank key with real value over map and string, in the assign form.
package main

func main() {
	m := map[string]int{"a": 1}
	var v int
	for _, v = range m {
	}
	println(v)

	var r rune
	for _, r = range "ab" {
	}
	println(string(r))

	for _ = range m {
	}
	for _ = range "ab" {
	}
	println("ok")
}

// Output:
// 1
// b
// ok
