package main

import "filetests/extern/p1"

func main() { println("num:", p1.Uint32()) }

// Error:
// cannot directly modify readonly tainted object (w/o method): p<VPBlock(1,0)>.lo
