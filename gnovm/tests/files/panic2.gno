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
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:5
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// ...74 frame(s) elided...
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>(i<VPBlock(1,0)> + (const (1 int)))
//     main/files/panic2.gno:7
// p<VPBlock(3,0)>((const (0 int)))
//     main/files/panic2.gno:7
// main()
//     main/files/panic2.gno:11

// Error:
// here
