package main


func main() {
    defer func() {
    		r := recover()
    		println("recover:", r)
    }()

    arr := []int{1, 2, 3}
    _ = arr[-1:] // Panics because of negative index
}

// Output:
// recover: invalid slice index -1 (index must be non-negative)
