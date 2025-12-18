package main

func main() {
	for i := 0; i < 3; i++ {
		func() {
            i += 1
        }()
		println(i)
	}
}

// Preprocessed:
// file{ package main; func main() { for .loopvar_i<!@VPBlock(1,0)> := (const (0 int)); .loopvar_i<*VPBlock(1,0)> < (const (3 int)); .loopvar_i<*VPBlock(1,0)>++ { .redefine_i<!~VPBlock(1,1)> := .loopvar_i<VPBlock(1,0)>; func func(){ .redefine_i<~VPBlock(1,0)> += (const (1 int)) }<.redefine_i<()~VPBlock(1,1)>>(); (const (println func(...interface {})))(.redefine_i<~VPBlock(1,1)>); .loopvar_i<!VPBlock(1,0)> := .redefine_i<~VPBlock(1,1)> } } }

// Output:
// 1
// 3
