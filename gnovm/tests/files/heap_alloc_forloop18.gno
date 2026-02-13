package main

func main() {
    // Both i and j are "loop variables" defined in the Header Scope
    for i, j := 0, 10; i < j; i, j = i+1, j-1 {
        func() {
            println(i, j) // Captures the iteration copy of i AND j
        }()
    }
}

// Preprocessed:
// file{ package main; func main() { for .loopvar_i<!VPBlock(1,0)>, .loopvar_j<!VPBlock(1,1)> := (const (0 int)), (const (10 int)); .loopvar_i<VPBlock(1,0)> < .loopvar_j<VPBlock(1,1)>; .loopvar_i<VPBlock(1,0)>, .loopvar_j<VPBlock(1,1)> = .loopvar_i<VPBlock(1,0)> + (const (1 int)), .loopvar_j<VPBlock(1,1)> - (const (1 int)) { .redefine_j<!~VPBlock(1,1)> := .loopvar_j<VPBlock(2,1)>; .redefine_i<!~VPBlock(1,0)> := .loopvar_i<VPBlock(2,0)>; func func(){ (const (println func(...interface {})))(.redefine_i<~VPBlock(1,0)>, .redefine_j<~VPBlock(1,1)>) }<.redefine_i<()~VPBlock(1,0)>, .redefine_j<()~VPBlock(1,1)>>(); .loopvar_i<VPBlock(2,0)> = .redefine_i<~VPBlock(1,0)>; .loopvar_j<VPBlock(2,1)> = .redefine_j<~VPBlock(1,1)> }  } }

// Output:
// 0 10
// 1 9
// 2 8
// 3 7
// 4 6
