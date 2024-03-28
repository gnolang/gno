package main

// the vp of the arg of the outer closure should be (2,0), 2 is loopDepth
// the vp of the x of inner closure should be (2,0)
func main() {
	var y int
	var f []func()

	defer func() {
		for _, ff := range f {
			ff()
		}
	}()

	for i := 0; i < 3; i++ {
		x := y
		{
			f = append(f, func() { println(x) })
		}
		y++
	}
}

// Output:
// 0
// 1
// 2
