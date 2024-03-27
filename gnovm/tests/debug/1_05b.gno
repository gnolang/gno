package main

func main() {
	var y, counter int
	var f []func()
	defer func() {
		for _, ff := range f {
			ff()
		}
	}()

	// NOTE: in this case, the reProcess flow executed in trans_leave forStmt and reBuildBody
	// can this be merged?
	{
		for i := 0; i < 2; i++ { // wrapped in for loop
			counter = 0
		LABEL_1:
			if counter == 2 {
				continue
			}
			x := y
			f = append(f, func() { println(x) })
			y++
			counter++
			goto LABEL_1
		}
	}
}

// Output:
// 0
// 1
// 2
// 3
