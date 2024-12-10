package mdui

import (
	"gno.land/p/varmeta/mdui"
)

// Constants for common paths
const (
	BasePath         = "/r/varmeta/mdui"
	ComponentsPath   = "/r/varmeta/mdui:components"
	ExamplesPath     = "/r/varmeta/mdui:examples"
	BuilderGuidePath = "/r/varmeta/mdui:builder-guide"
)

// Common Navbar Links
var NavbarLinks = []mdui.NavLink{
	{"Home", BasePath},
	{"Components", ComponentsPath},
	{"Examples", ExamplesPath},
	{"Builder Guide", BuilderGuidePath},
}
