package gnolang

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func contains(list []string, item string) bool {
	for _, i := range list {
		if i == item {
			return true
		}
	}
	return false
}

func endsWith(item string, suffixes []string) bool {
	for _, i := range suffixes {
		if strings.HasSuffix(item, i) {
			return true
		}
	}
	return false
}

// TODO(hariom): move to better place
func ParseMemMod(dir string) *std.MemMod {
	var memMod std.MemMod
	gm, err := gnomod.ParseAt(dir)
	if err != nil {
		// TODO(hariom): return error instead
		panic(fmt.Sprintf("error parsing gno.mod at: %q", dir))
	}
	// TODO(hariom): make sure requires are accurate in gno.mod
	var requires []*std.Requirements
	for _, req := range gm.Require {
		requires = append(requires, &std.Requirements{
			Path:    req.Mod.Path,
			Version: req.Mod.Version,
		})
	}

	memMod.Requires = requires
	memMod.ImportPath = gm.Module.Mod.Path
	memMod.Version = gm.Module.Mod.Version

	return &memMod
}
