package main

var s1 []*int

func forLoopRef() {
	defer func() {
		for i, e := range s1 {
			println(i, *e)
		}
	}()

 	for i := 0; i < 3; i++ {
		s1 = append(s1, &i)
	}
}

func main() {
	forLoopRef()
}

// Output:
// 0 0
// 1 1
// 2 2
