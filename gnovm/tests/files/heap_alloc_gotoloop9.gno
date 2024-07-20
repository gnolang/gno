package main

func main() {
	c := 0
	closures := []func(){}
	i := 1
loop2:
	closures = append(closures, func() {
		println(i)
	})
	c += 1
	if c < 10 {
		goto loop2
	}
	for _, cl := range closures {
		cl()
	}
}

// This one doesn't because the goto stmt doesn't go back far enough.

// Output:
// 1
// 1
// 1
// 1
// 1
// 1
// 1
// 1
// 1
// 1
