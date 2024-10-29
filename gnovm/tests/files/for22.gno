package main

func main() {
    for i := 0; i < 10; i++ {
        if i == 1 {
            _ = func() int {
                fallthrough
                return 11
            }()
        }
        println(i)
    }
    println("wat???")
}

// Error:
// fallthrough statement out of place