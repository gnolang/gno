package main

var length = 30

func f() any {
    return make([]int, length)
}

func main() {
    a := f()
    println(len(a.([]int)))
}

// Output:
// 30