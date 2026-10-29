package run

import (
	"fmt"
	"go/token"
	"html/template"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// RunData is the render payload for templates/page.html.
type RunData struct {
	PkgPath string // full path, e.g. "gno.land/r/demo/boards"
	Remote  string // e.g. "https://rpc.gno.land:443"
	ChainId string // e.g. "portal-loop"
	Gas     components.Gas
	// Script is the script the editor opens with, and ScriptHTML the same
	// script highlighted, shown until the editor loads (or without JS).
	Script     string
	ScriptHTML template.HTML
}

// PkgAlias returns the name an unaliased import of PkgPath binds: the last
// path element, skipping a version suffix ("gno.land/r/foo/v2" gives "foo").
func (d RunData) PkgAlias() string {
	return gno.LastPathElement(d.PkgPath)
}

// starterScript is a script that runs as generated against the package
// whose functions funcs lists: it prints Render("") when the package has a
// Render, and otherwise names one of its functions, since an import left
// unused does not compile.
func starterScript(pkgPath string, funcs []*doc.JSONFunc) string {
	alias := RunData{PkgPath: pkgPath}.PkgAlias()
	call, main := starterCall(alias, funcs), "main()"
	switch {
	case call == "":
		return fmt.Sprintf("package main\n\nimport _ %q\n\nfunc main() {\n\t// Name the import above to call the package.\n}\n", pkgPath)
	case strings.Contains(call, "cross(cur)"):
		// A crossing call needs the realm main runs in.
		main = "main(cur realm)"
	}
	return fmt.Sprintf("package main\n\nimport %q\n\nfunc %s {\n\t// Call %s functions here, e.g.:\n\t%s\n}\n", pkgPath, main, alias, call)
}

// starterCall is the starter script's one statement: a call to Render, or
// else a reference to the first exported function, or else nothing.
func starterCall(alias string, funcs []*doc.JSONFunc) string {
	first := ""
	for _, fn := range funcs {
		if fn.Type != "" || !token.IsExported(fn.Name) {
			continue
		}
		switch {
		case fn.Name == "Render" && isRender(fn, "string"):
			return fmt.Sprintf("println(%s.Render(\"\"))", alias)
		case fn.Name == "Render" && isRender(fn, "realm", "string"):
			return fmt.Sprintf("println(%s.Render(cross(cur), \"\"))", alias)
		case first == "":
			first = fmt.Sprintf("_ = %s.%s", alias, fn.Name)
		}
	}
	return first
}

// isRender reports whether fn takes params of the given types and returns a
// string, the shapes Render is called with.
func isRender(fn *doc.JSONFunc, params ...string) bool {
	if len(fn.Params) != len(params) || len(fn.Results) != 1 || fn.Results[0].Type != "string" {
		return false
	}
	for i, p := range fn.Params {
		if p.Type != params[i] {
			return false
		}
	}
	return true
}
