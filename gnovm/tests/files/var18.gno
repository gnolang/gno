package main

func main() {
	a, b := 1
}

// Error:
// main/files/var18.gno:3: should not happen:
// --- preprocess stack ---
// stack 0: func main() { var a<VPBlock(1,0)>, b<VPBlock(1,1)> = (const (1 <untyped> bigint)) }
// ------------------------
