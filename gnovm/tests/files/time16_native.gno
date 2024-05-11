package main

import (
	"fmt"
	"time"
)

func main() {
	var a int64 = 2
	fmt.Println(time.Second * a)
}

// Error:
// main/files/time16_native.gno:10: incompatible types in binary expression: time<VPBlock(2,1)>.Second MUL a<VPBlock(1,0)>
