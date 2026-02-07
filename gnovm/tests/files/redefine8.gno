package main

import "fmt"

func main() {
    for i := 0; i < 3; i++ {
        i := i // shadow loop variable
        func() {
            i += 1
            println(i)
        }()
    }
    fmt.Println("Hello, 世界")
}

// Preprocessed:
// file{ package main; import fmt fmt; func main() { for .loopvar_i<!@VPBlock(1,0)> := (const (0 int)); .loopvar_i<*VPBlock(1,0)> < (const (3 int)); .loopvar_i<*VPBlock(1,0)>++ { i<!~VPBlock(1,0)> := .loopvar_i<*VPBlock(2,0)>; func func(){ i<~VPBlock(1,0)> += (const (1 int)); (const (println func(...interface {})))(i<~VPBlock(1,0)>) }<i<()~VPBlock(1,0)>>() } ; (const (ref(fmt) package{})).Println((const ("Hello, 世界" string))) } }

// Output:
// 1
// 2
// 3
// Hello, 世界
