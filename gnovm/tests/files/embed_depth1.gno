package main

// A pointer embed does not build S2 before S1 is sealed, so at that point
// S1's depth reads as 1. The check after the whole group is built sees 9.

type S1 struct{ *S2 }
type S2 struct{ *S3 }
type S3 struct{ *S4 }
type S4 struct{ *S5 }
type S5 struct{ *S6 }
type S6 struct{ *S7 }
type S7 struct{ *S8 }
type S8 struct{ *S9 }
type S9 struct{ *S10 }
type S10 struct{ x int }

func main() {}

// Error:
// main/embed_depth1.gno:6:6-22: type S1 embed depth 9 exceeds max 8
