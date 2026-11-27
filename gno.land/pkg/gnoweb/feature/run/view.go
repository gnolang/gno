package run

import (
	"fmt"
	"go/token"
	"html/template"

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

// starterScript is a script that runs as generated against the package
// whose functions funcs lists, for a gas that does not depend on the
// package: it names one of its functions, since an import left unused does
// not compile, and shows a call to Render, which costs what the package
// makes it cost, in a comment.
func starterScript(pkgPath string, funcs []*doc.JSONFunc) string {
	alias := gno.LastPathElement(pkgPath)
	ref, call, crossing := starterRef(alias, funcs)
	if ref == "" {
		return fmt.Sprintf("package main\n\nimport _ %q\n\nfunc main() {\n\t// Name the import above to call the package.\n}\n", pkgPath)
	}
	main := "main()"
	if crossing {
		// Uncommented, a crossing call needs the realm main runs in.
		main = "main(cur realm)"
	}
	if call != "" {
		ref = "// " + call + "\n\t" + ref
	}
	return fmt.Sprintf("package main\n\nimport %q\n\nfunc %s {\n\t// Call %s functions here, e.g.:\n\t%s\n}\n", pkgPath, main, alias, ref)
}

// starterRef is the starter script's one statement, a reference to Render
// or else to the first exported function, and the call to Render it shows
// when there is one; crossing is true when that call crosses into the realm.
func starterRef(alias string, funcs []*doc.JSONFunc) (ref, call string, crossing bool) {
	for _, fn := range funcs {
		if fn.Type != "" || !token.IsExported(fn.Name) {
			continue
		}
		switch {
		case fn.Name == "Render" && isRender(fn, "string"):
			return "_ = " + alias + ".Render", fmt.Sprintf("println(%s.Render(\"\"))", alias), false
		case fn.Name == "Render" && isRender(fn, "realm", "string"):
			return "_ = " + alias + ".Render", fmt.Sprintf("println(%s.Render(cross(cur), \"\"))", alias), true
		case ref == "":
			ref = "_ = " + alias + "." + fn.Name
		}
	}
	return ref, "", false
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
