package mywebsite

import (
	"gno.land/p/demo/upgradeable"
)

// Global registry and function holders
var (
	registry   = upgradeable.New()
	renderFunc = upgradeable.NewStringFuncHolder(
		registry,
		"render",
		RenderV1, // Using a named package function
	)
)

// V1 implementation
func RenderV1(path string) string {
	return "Basic page for " + path
}

// V2 implementation with enhanced features
func RenderV2(path string) string {
	return "# Enhanced page for " + path
}

// Public function that uses the upgradeable implementation
func Render(path string) string {
	fn := renderFunc.Get()
	return fn(path)
}

// Admin function to upgrade the implementation
func UpgradeRender() error {
	return renderFunc.Update(RenderV2)
}

func UpgradeRenderTo(target interface{}) error {
	return renderFunc.Update(target)
}
