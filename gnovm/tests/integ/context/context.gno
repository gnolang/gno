package runtests

import (
	"chain/runtime"
	"fmt"
)

func Context() {
	// This requires a Context to work; it will fail ugly if the Context isn't available.
	fmt.Printf("Context worked: %d\n", runtime.ChainHeight())
}
