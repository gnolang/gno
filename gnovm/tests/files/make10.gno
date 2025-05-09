package main

var length = 2.1

func f() any {
    return make([]int, length)
}

func main() {
    a := f()
    println(len(a.([]int)))
}

// Error:
// main/files/make10.gno:6:12: invalid argument: index length (variable of type float64) must be integer

// TypeCheckError:
// main/files/make10.gno:6:24: invalid argument: index length (variable of type float64) must be integer