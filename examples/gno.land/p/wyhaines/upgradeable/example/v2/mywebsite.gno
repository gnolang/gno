package mywebsitev2

import (
	"std"

	"gno.land/p/demo/ufmt"
	v1site "gno.land/r/wyhaines/v1/mywebsite"
)

func RenderV3(path string) string {
	return "# V3 Enhanced page for " + path
}

func Render(path string) string {
	return v1site.Render(path)
}

// Automatically apply upgrades
func init() {
	std.Emit("Upgrading mywebsite Render to V3")
	err := v1site.UpgradeRenderTo(RenderV3)
	if err != nil {
		std.Emit(ufmt.Sprintf("%v", err))
	}
}
