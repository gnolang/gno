package mywebsite

import (
	"gno.land/p/wyhaines/upgradeable"
)

// Package information
const (
	PackagePath = "gno.land/r/wyhaines/v1/mywebsite"
	Version     = 1
)

// Global registry and function holders
// Notice Registry is public (capitalized) so other realms can access it
var (
	Registry   = upgradeable.New() // Public, accessible from other realms
	renderFunc = upgradeable.NewStringFuncHolder(
		Registry,
		"render",
		RenderV1,
	)
)

// V1 implementation
func RenderV1(path string) string {
	return "Basic page for " + path
}

// Public function that uses the upgradeable implementation
func Render(path string) string {
	fn := renderFunc.Get()
	return fn(path)
}

// Public function to check current version
func GetVersion() int {
	return Version
}
