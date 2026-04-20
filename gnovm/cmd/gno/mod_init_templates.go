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
	PkgName string
}

// initTemplate describes a single scaffold template option.
type initTemplate struct {
	Name        string   // short name shown to user (e.g. "basic", "dao")
	Description string   // one-line description
	SourcePath  string   // path within the embedded FS for the source template
	TestPath    string   // path within the embedded FS for the test template
	FS          embed.FS // the embedded FS containing the templates
}

// realmTemplates lists available templates for realms.
// Add new entries here to extend the template selection menu.
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
