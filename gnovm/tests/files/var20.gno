package main

func r() int {
	return 1
}

func main() {
	var a, b, c = r()
}

// Error:
// main/files/var20.gno:7: should not happen:
// --- preprocess stack ---
// stack 0: func main() { var a<VPBlock(1,0)>, b<VPBlock(1,1)>, c<VPBlock(1,2)> = r<VPBlock(3,0)>() }
// ------------------------
