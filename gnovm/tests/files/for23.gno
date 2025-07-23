package main

func main() {
    for i := 0; i < 10; i++ {
        if i == 1 {
            _ = func() int {
                break
                return 11
            }()
        }
        println(i)
    }
    println("wat???")
}

// Error:
// main/for23.gno:7:17-22: break statement out of place

// TypeCheckError:
// main/for23.gno:7:17: break not in for, switch, or select statement
