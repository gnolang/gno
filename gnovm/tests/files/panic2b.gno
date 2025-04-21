package main

func p(i int) {
	defer func() {
		panic("here")
	}()
	if i == 4 {
		panic("here")
	}
	p(i + 1)
}

func main() {
	p(0)
}

// Stacktrace:
// panic: here
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + 1)
//     main/files/panic2b.gno:8
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + 1)
//     main/files/panic2b.gno:10
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + 1)
//     main/files/panic2b.gno:10
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + 1)
//     main/files/panic2b.gno:10
// p<VPBlock(3,0)>(0)
//     main/files/panic2b.gno:10
// main<VPBlock(1,1)>()
//     main/files/panic2b.gno:14
// ... 4 panic(s) elided ...
// panic: here
// defer func(){ ... }()
//     main/files/panic2b.gno:5
// p<VPBlock(3,0)>(0)
//     main/files/panic2b.gno:4
// main<VPBlock(1,1)>()
//     main/files/panic2b.gno:14

// Error:
// here
// 	here
// 	here
// 	here
// 	here
// 	here
