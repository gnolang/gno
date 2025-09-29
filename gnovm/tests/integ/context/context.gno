package runtests

import (
	"chain/runtime"
	"strconv"
)

func Context() {
	// This requires a Context to work; it will fail ugly if the Context isn't available.
	println("Context worked: " + strconv.Itoa(int(runtime.ChainHeight())))
}
