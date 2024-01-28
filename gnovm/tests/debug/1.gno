package main

func main() {
	for i := 0; i < 2; i++ {
		l := []byte{'a', 'b', 'c', 'd'}
		if i == 0 {
			l[0] = 'f'
		}
		println(string(l))
	}
}

// Output:
// fbcd
// abcd
