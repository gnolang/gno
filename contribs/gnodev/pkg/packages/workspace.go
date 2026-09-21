package packages

import (
	vmpackages "github.com/gnolang/gno/gnovm/pkg/packages"
)

// FindWorkspace resolves the loader root for start, or "" when start is in
// neither a gnowork.toml workspace nor a gnomod.toml package directory (the
// caller then falls back to discovery mode). It delegates to gnovm's loader
// context so the workspace gnodev eager-loads is, by construction, one gnovm's
// own Load can satisfy — the two cannot drift on which markers define a root.
func FindWorkspace(start string) string {
	root, err := vmpackages.FindLoaderRoot(start)
	if err != nil {
		return ""
	}
	return root
}
