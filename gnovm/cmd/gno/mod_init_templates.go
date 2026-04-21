// Template registry for gno init.
//
// To add a new template (e.g. a "dao" realm template):
//
//  1. Create template files under templates/<kind>/<name>/:
//     templates/realm/dao/source.gno.tmpl
//     templates/realm/dao/test.gno.tmpl       (optional)
//
//  2. Add an entry to the corresponding slice below (e.g. realmTemplates).
//     The Name field is what users see in the interactive menu and can pass
//     via --template.
//
// Template files use Go's text/template syntax with a single templateData
// struct providing {{.PkgName}} (derived from the module path's last segment).

package main

import (
	"bytes"
	"embed"
	"text/template"
)

//go:embed templates/realm/*/*.tmpl
var realmTemplatesFS embed.FS

//go:embed templates/package/*/*.tmpl
var packageTemplatesFS embed.FS

//go:embed templates/run/*/*.tmpl
var runTemplatesFS embed.FS

// templateData holds the data passed to template files.
type templateData struct {
	PkgName string // Package name, derived from the last segment of the module path
}

// initTemplate describes a single scaffold template option.
type initTemplate struct {
	Name        string   // Short name shown to user and accepted by --template (e.g. "basic", "dao")
	Description string   // One-line description shown in the interactive menu
	SourcePath  string   // Path within FS for the source .gno template
	TestPath    string   // Path within FS for the test .gno template (empty = no test file)
	FS          embed.FS // Embedded FS containing the template files
}

// realmTemplates lists available templates for realms.
var realmTemplates = []initTemplate{
	{
		Name:        "basic",
		Description: "minimal realm with a Render function",
		SourcePath:  "templates/realm/basic/source.gno.tmpl",
		TestPath:    "templates/realm/basic/test.gno.tmpl",
		FS:          realmTemplatesFS,
	},
}

// packageTemplates lists available templates for packages.
// Add new entries here to extend the template selection menu.
var packageTemplates = []initTemplate{
	{
		Name:        "basic",
		Description: "minimal package with a placeholder test",
		SourcePath:  "templates/package/basic/source.gno.tmpl",
		TestPath:    "templates/package/basic/test.gno.tmpl",
		FS:          packageTemplatesFS,
	},
}

// runTemplates lists available templates for main/run scripts.
// Add new entries here to extend the template selection menu.
var runTemplates = []initTemplate{
	{
		Name:        "basic",
		Description: "minimal main script for gnokey maketx run",
		SourcePath:  "templates/run/basic/source.gno.tmpl",
		FS:          runTemplatesFS,
	},
}

// renderTemplate parses and executes a template from the embedded FS.
func renderTemplate(fsys embed.FS, path string, data templateData) ([]byte, error) {
	tmpl, err := template.ParseFS(fsys, path)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
