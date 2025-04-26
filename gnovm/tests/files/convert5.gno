package main

func main() {
	var ints []int
	ints = append(ints, nil, nil)
	println(ints)
}

// Error:
// main/files/convert5.gno:3:1: cannot convert (const (undefined)) to IntKind

// TypeCheckError:
// main/files/convert5.gno:5:22: cannot use nil as int value in argument to append; main/files/convert5.gno:5:27: cannot use nil as int value in argument to append
