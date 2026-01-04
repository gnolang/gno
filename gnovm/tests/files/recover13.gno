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

// TypeCheckError:
// main/recover13.gno:11:13: invalid argument: index -1 (constant of type int) must not be negative
