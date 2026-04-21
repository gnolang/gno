// Template registry for gno init.
//
// To add a new template (e.g. a "dao" realm template):
//
//  1. Create a directory under templates/<kind>/<name>/ with .tmpl files.
//     Filenames ending in .tmpl are processed as Go text/templates;
//     the output filename is the stem with .tmpl stripped, also run
//     through the template engine (so {{.PkgName}}.gno.tmpl produces
//     <pkgName>.gno).
//
//     Example layout for a "dao" realm template:
//       templates/realm/dao/{{.PkgName}}.gno.tmpl
//       templates/realm/dao/{{.PkgName}}_test.gno.tmpl
//       templates/realm/dao/state.gno.tmpl
//       templates/realm/dao/helpers.gno.tmpl
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
	"io/fs"
	"path/filepath"
	"strings"
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

// initTemplate describes a single scaffold template directory.
type initTemplate struct {
	Name        string   // Short name shown to user and accepted by --template (e.g. "basic", "dao")
	Description string   // One-line description shown in the interactive menu
	Dir         string   // Directory path within FS (e.g. "templates/realm/basic")
	FS          embed.FS // Embedded FS containing the template directory
}

// realmTemplates lists available templates for realms.
var realmTemplates = []initTemplate{
	{
		Name:        "basic",
		Description: "minimal realm with a Render function",
		Dir:         "templates/realm/basic",
		FS:          realmTemplatesFS,
	},
}

// packageTemplates lists available templates for packages.
var packageTemplates = []initTemplate{
	{
		Name:        "basic",
		Description: "minimal package with a placeholder test",
		Dir:         "templates/package/basic",
		FS:          packageTemplatesFS,
	},
}

// runTemplates lists available templates for main/run scripts.
var runTemplates = []initTemplate{
	{
		Name:        "basic",
		Description: "minimal main script for gnokey maketx run",
		Dir:         "templates/run/basic",
		FS:          runTemplatesFS,
	},
}

// renderTemplateDir walks a template directory and renders all .tmpl files.
// For each file, the .tmpl suffix is stripped; both the resulting filename
// and the file contents are executed as Go text/templates with data.
// Returns a map of output filename → rendered content.
func renderTemplateDir(fsys embed.FS, dir string, data templateData) (map[string][]byte, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}

	files := make(map[string][]byte, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".tmpl") {
			continue
		}

		// Output filename = stem with .tmpl stripped, also templated
		outName, err := renderString(strings.TrimSuffix(name, ".tmpl"), data)
		if err != nil {
			return nil, err
		}

		// Render file contents
		path := filepath.Join(dir, name)
		content, err := renderTemplateFile(fsys, path, data)
		if err != nil {
			return nil, err
		}

		files[outName] = content
	}

	return files, nil
}

// renderTemplateFile parses and executes a single template file from the embedded FS.
func renderTemplateFile(fsys embed.FS, path string, data templateData) ([]byte, error) {
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

// renderString executes a string as a Go text/template with the given data.
func renderString(s string, data templateData) (string, error) {
	tmpl, err := template.New("").Parse(s)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
