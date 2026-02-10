package main

func main() {
	for i := 0; i < 3; i++ {
		println("outer: ", i)
	inner:
		for j := 0; j < 5; j++ {
			println("   inner: ", j)
			if j == 2 {
				break // breaks inner
			}
			if j == 4 { // not reachable
				panic("!!!!!!!!")
				break inner
			}
		}
	}
}

// Output:
// outer:  0
//    inner:  0
//    inner:  1
//    inner:  2
// outer:  1
//    inner:  0
//    inner:  1
//    inner:  2
// outer:  2
//    inner:  0
//    inner:  1
//    inner:  2
