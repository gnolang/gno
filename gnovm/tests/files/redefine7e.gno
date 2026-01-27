package main

func main() {
	for i := 0; i < 3; i++ {
		i := i
		func() {
            i += 1
			println(i)
        }()
	}
}

// Preprocessed:
// file{ package main; func main() { for .loopvar_i<!@VPBlock(1,0)> := (const (0 int)); .loopvar_i<*VPBlock(1,0)> < (const (3 int)); .loopvar_i<*VPBlock(1,0)>++ i<!~VPBlock(1,1)> := .loopvar_i<*VPBlock(1,0)>; func func(){ i<~VPBlock(1,0)> += (const (1 int)); (const (println func(...interface {})))(i<~VPBlock(1,0)>) }<i<()~VPBlock(1,1)>>()  } }

// Output:
// 1
// 2
// 3
