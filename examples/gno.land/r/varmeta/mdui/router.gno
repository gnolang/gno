package mdui

import "strings"

// Render entry point to route to different pages based on path
func Render(path string) string {
	switch {
	case strings.Contains(path, "components"):
		return RenderComponentsPage()
	case strings.Contains(path, "examples"):
		return RenderExamplesPage()
	case strings.Contains(path, "builder-guide"):
		return RenderBuilderGuidePage()
	default:
		return RenderHomePage()
	}
}