package main

type T struct {
	a int
}

func (tv T) Mv(a int) int { return 0 }

func main() {
	const t = T.Mv
}

// Error:
// main/files/const39.gno:10:8: T<VPBlock(3,0)>.Mv (variable of type main.T) is not constant