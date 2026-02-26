package main

import "fmt"

func main() {
    for i := 0; i < 3; i++ {
        i := i // shadow loop variable
        func() {
            i += 1
        }()
    }
    fmt.Println("Hello, 世界")
}

// Preprocessed:
// file{ package main; import fmt fmt; func main() { for i.loopvar<!VPBlock(1,0)> := (const (0 int)); i.loopvar<VPBlock(1,0)> < (const (3 int)); i.loopvar<VPBlock(1,0)>++ { i<!~VPBlock(1,1)> := i.loopvar<VPBlock(1,0)>; func func(){ i<~VPBlock(1,0)> += (const (1 int)) }<i<()~VPBlock(1,1)>>() }; (const (ref(fmt) package{})).Println((const ("Hello, 世界" string))) } }

// Output:
// Hello, 世界
