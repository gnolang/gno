// package main

// func initloopvar(p *int) {
// 	*p = 0
// }

// func incloopvar(p *int) {
// 	*p += 1
// }

// func main() {
// 	for i := 0; i < 5; i++ {
// 		println("outer i: ", i)

// 		for initloopvar(&i); i < 2; incloopvar(&i) {
// 			println("i inner: ", i)
// 		}

// 		println("end inner, i: ", i)
// 	}
// }

// Error:
// redefine9f.gno:24:58: expected 'package', found 'EOF'
