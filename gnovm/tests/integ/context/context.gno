package runtests

import (
	"std"
)

func Context() {
	// This requires a Context to work; it will fail ugly if the Context isn't available.
	println("Context worked: " + string(std.ChainHeight()))
}
