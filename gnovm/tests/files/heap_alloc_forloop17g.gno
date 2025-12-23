// package main

// func main() {
// 	var f func()
// OUTER:
// 	for i := 0; i < 3; {
// 		i := i
// 		println("enter, i: ", i)
// 		for j := 0; j < 2; j++ {
// 			if j == 1 {
// 				i := i  // i := .redefine_i, this is priortized.
// 				println("---continue")
// 				i++ // i

// 				func(){
// 					println(j)
// 				}()

// 				// copy out
// 				// .loopvar_j = .redefine_j
// 				// .loopvar_i = i
// 				continue OUTER
// 			}
// 		}

// 		i++

// 		f = func() {
// 			start := i
// 			println("capture, start: ", start)
// 		}
// 		f()
// 	}
// }

// // Preprocessed:
// // file{ package main; func main() { var f<!VPBlock(1,0)> func(); for .loopvar_i<!@VPBlock(1,0)> := (const (0 int)); .loopvar_i<*VPBlock(1,0)> < (const (3 int));  { .redefine_i<!~VPBlock(1,0)> := .loopvar_i<*VPBlock(2,0)>; (const (println func(...interface {})))((const ("---continue" string))); for .loopvar_j<!@VPBlock(1,0)> := (const (0 int)); .loopvar_j<*VPBlock(1,0)> < (const (2 int)); .loopvar_j<*VPBlock(1,0)>++ { .redefine_j<!~VPBlock(1,0)> := .loopvar_j<*VPBlock(2,0)>; if .redefine_j<VPBlock(2,0)> == (const (1 int)) { (const (println func(...interface {})))((const ("---continue" string))); .redefine_i<VPBlock(4,0)>++; func func(){ (const (println func(...interface {})))(.redefine_j<VPBlock(3,0)>) }(); continue OUTER<0,0,0> }; .redefine_i<VPBlock(4,0)>++; func func(){ (const (println func(...interface {})))(.redefine_j<VPBlock(3,0)>) }(); .loopvar_j<*VPBlock(2,0)> = .redefine_j<VPBlock(1,0)>; .loopvar_i<*VPBlock(4,0)> = .redefine_i<VPBlock(3,0)>; continue OUTER<0,0,0>; .loopvar_j<*VPBlock(2,0)> = .redefine_j<VPBlock(1,0)> } ; func func(){ (const (println func(...interface {})))(.redefine_j<VPBlock(3,0)>) }(); .loopvar_j<*VPInvalid(0)> = .redefine_j<VPInvalid(0)>; .loopvar_i<*VPInvalid(0)> = .redefine_i<VPInvalid(0)>; continue OUTER<0,0,0>; .loopvar_i<VPInvalid(0)> = .redefine_i<VPInvalid(0)> }  } }

// // Error:
// // main/heap_alloc_forloop17e.gno:6:2-32:3''': name .loopvar_j not declared
