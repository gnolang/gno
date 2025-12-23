// package main

// func main() {
// 	var f func()
// OUTER:
// 	for i := 0; i < 3; {
// 		// .redefine_i := .loopvar_i
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
// // file{ package main; func main() { var f<!VPBlock(1,0)> func(); for .loopvar_i<!@VPBlock(1,0)> := (const (0 int)); .loopvar_i<*VPBlock(1,0)> < (const (3 int));  { .redefine_i<!~VPBlock(1,0)> := .loopvar_i<*VPBlock(2,0)>; (const (println func(...interface {})))((const ("enter, i: " string)), .redefine_i<~VPBlock(1,0)>); for .loopvar_j<!@VPBlock(1,0)> := (const (0 int)); .loopvar_j<*VPBlock(1,0)> < (const (2 int)); .loopvar_j<*VPBlock(1,0)>++ { .redefine_j<!~VPBlock(1,0)> := .loopvar_j<*VPBlock(2,0)>; if .redefine_j<~VPBlock(2,0)> == (const (1 int)) { i<!VPBlock(1,0)> := .redefine_i<~VPBlock(4,0)>; (const (println func(...interface {})))((const ("---continue" string))); i<VPBlock(1,0)>++; func func(){ (const (println func(...interface {})))(.redefine_j<~VPBlock(1,0)>) }<.redefine_j<()~VPBlock(2,0)>>(); .loopvar_i<*VPBlock(5,0)> = i<VPBlock(1,0)>; .loopvar_j<*VPBlock(3,0)> = .redefine_j<~VPBlock(2,0)>; .loopvar_i<*VPBlock(5,0)> = i<VPBlock(1,0)>; .loopvar_j<*VPBlock(3,0)> = .redefine_j<~VPBlock(2,0)>; continue OUTER<0,0,0> }; .loopvar_j<*VPBlock(2,0)> = .redefine_j<~VPBlock(1,0)> } ; .redefine_i<~VPBlock(1,0)>++; f<VPBlock(3,0)> = func func(){ start<!VPBlock(1,0)> := .redefine_i<~VPBlock(1,1)>; (const (println func(...interface {})))((const ("capture, start: " string)), start<VPBlock(1,0)>) }<.redefine_i<()~VPBlock(1,0)>>; f<VPBlock(3,0)>(); .loopvar_i<*VPBlock(2,0)> = .redefine_i<~VPBlock(1,0)> }  } }

// // Output:
// // enter, i:  0
// // ---continue
// // 1
// // enter, i:  1
// // ---continue
// // 1
// // enter, i:  2
// // ---continue
// // 1
