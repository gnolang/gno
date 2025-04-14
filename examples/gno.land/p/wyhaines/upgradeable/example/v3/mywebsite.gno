package mywebsitev3

import (
	"std"
	"strconv"

	v1 "gno.land/r/wyhaines/v1/mywebsite" // Import the v1 realm
)

// Package information
const (
	Version = 3
)

func init() {
	// Directly upgrade the function in the v1 realm's Registry
	err := v1.Registry.RegisterFunction("render", RenderV3)
	if err != nil {
		std.Emit("UpgradeError", "error", err.Error())
	} else {
		std.Emit("FunctionUpgraded", "function", "render", "version", strconv.Itoa(Version))
	}
}

// V3 implementation with more enhancements
func RenderV3(path string) string {
	return "# V3 Enhanced page for " + path + " with advanced formatting"
}

// We can provide our own public API that uses the v1 function
// (which now uses the v3 implementation)
func Render(path string) string {
	return v1.Render(path)
}

// Get version info
func GetVersion() int {
	return Version
}
