package main

func main() {
	for i := 0; i < 2; i++ {
		l := []byte("lead")
		if i == 0 {
			l[0] = 'f'
		}
		println(string(l))
	}
}

// Output:
// fead
// lead
