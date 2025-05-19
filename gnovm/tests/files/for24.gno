package main

func main() {
    for i := 0; i < 10; i++ {
        if i == 1 {
            _ = func() int {
                if true {
                    break
                }
                return 11
            }()
        }
        println(i)
    }
    println("wat???")
}

// Error:
// main/for24.gno:8:21-26: break statement out of place

// TypeCheckError:
// main/for24.gno:8:21: break not in for, switch, or select statement
