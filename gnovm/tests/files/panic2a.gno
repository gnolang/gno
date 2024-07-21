package main

func p(i int) {
	if i == 200 {
		panic("here")
	}
	p(i + 1)
}

func main() {
	p(0)
}

// Stacktrace:
// panic here
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:5
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// ...74 frame(s) elided...
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + int)
//     main/files/panic2a.gno:7
// p<VPBlock(3,0)>(int)
//     main/files/panic2a.gno:7
// main<VPBlock(1,1)>()
//     main/files/panic2a.gno:11

// Error:
// here
