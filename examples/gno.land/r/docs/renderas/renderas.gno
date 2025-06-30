// RenderAs demonstrates how a realm can provide viewer-specific rendering formats.
//
// When a client (like gnoweb) knows the viewer's identity, it will:
// 1. Try to call RenderAs(path, viewer) first if it exists
// 2. Fall back to Render(path) if RenderAs doesn't exist or viewer is unknown
//
// This allows realms to customize content per viewer while maintaining a default view.
// Common uses include showing customized data, or prefilling actions with user-specific data.
package renderas

// RenderAs provides a viewer-specific rendering format if viewer is provided.
func RenderAs(path, viewer string) string {
	return "Hello, " + viewer + "!"
}

// Render provides the default rendering format.
func Render(path string) string {
	return "Hello, World!"
}
