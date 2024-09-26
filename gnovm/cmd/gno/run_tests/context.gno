package runtests

import (
	"fmt"
        "std"
)

func Context() {
	// This requires a Context to work; it will fail ugly if the Context isn't available.
	fmt.Printf("Context worked: %d\n", std.GetHeight())
}
