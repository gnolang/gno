package components

import (
	"bytes"
	"go/parser"
	"go/token"
	"html/template"
	"io"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

// licenseSignatures orders detections most-specific-first to avoid false matches.
var licenseSignatures = []struct {
	Kind string
	RE   *regexp.Regexp
}{
	{"MIT", regexp.MustCompile(`(?i)^\s*(the )?mit license`)},
	{"Apache-2.0", regexp.MustCompile(`(?i)apache license\s*,?\s*version 2\.0`)},
	{"AGPL-3.0", regexp.MustCompile(`(?i)gnu affero general public license.*version 3`)},
	{"GPL-3.0", regexp.MustCompile(`(?i)gnu general public license.*version 3`)},
	{"LGPL", regexp.MustCompile(`(?i)gnu lesser general public license`)},
	{"BSD-3-Clause", regexp.MustCompile(`(?i)redistribution and use.*with or without modification[\s\S]*3\.\s*neither`)},
	{"BSD-2-Clause", regexp.MustCompile(`(?i)redistribution and use.*with or without modification`)},
	{"ISC", regexp.MustCompile(`(?i)isc license`)},
	{"MPL-2.0", regexp.MustCompile(`(?i)mozilla public license.*version 2\.0`)},
	{"Unlicense", regexp.MustCompile(`(?i)this is free and unencumbered software`)},
}

var spdxRE = regexp.MustCompile(`(?i)SPDX-License-Identifier:\s*([^\s]+)`)

// deriveLicense returns the first recognized license file.
// Content is read up to 4 KB to bound regex work and avoid ReDoS surface.
// If the file exists but content lookup fails, FileName is set and Kind is empty.
func deriveLicense(files []string, fileContent func(string) ([]byte, bool)) License {
	var licenseFile string
	for _, f := range files {
		if ReLicenseFileName.MatchString(f) {
			licenseFile = f
			break
		}
	}
	if licenseFile == "" {
		return License{}
	}

	body, ok := fileContent(licenseFile)
	if !ok || len(body) == 0 {
		return License{FileName: licenseFile}
	}
	sample := body
	if len(sample) > 4096 {
		sample = sample[:4096]
	}

	if m := spdxRE.FindSubmatch(sample); len(m) == 2 {
		return License{Kind: string(m[1]), FileName: licenseFile}
	}
	for _, sig := range licenseSignatures {
		if sig.RE.Match(sample) {
			return License{Kind: sig.Kind, FileName: licenseFile}
		}
	}
	return License{FileName: licenseFile}
}

// parseImports extracts deduplicated import paths from .gno sources.
// Uses parser.ImportsOnly which stops after the imports block and tolerates
// .gno-only syntax that would otherwise fail a full parse.
func parseImports(sources map[string][]byte, domain string) []ImportLink {
	if len(sources) == 0 {
		return nil
	}

	seen := map[string]string{} // path -> kind
	fset := token.NewFileSet()
	for _, content := range sources {
		f, err := parser.ParseFile(fset, "", content, parser.ImportsOnly)
		if err != nil {
			continue
		}
		for _, imp := range f.Imports {
			p := strings.Trim(imp.Path.Value, `"`)
			if _, ok := seen[p]; !ok {
				seen[p] = classifyImport(p)
			}
		}
	}
	if len(seen) == 0 {
		return nil
	}

	out := make([]ImportLink, 0, len(seen))
	for p, kind := range seen {
		out = append(out, ImportLink{
			Path: p,
			Kind: kind,
			Link: buildImportLink(p, kind, domain),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

func classifyImport(p string) string {
	switch {
	case strings.HasPrefix(p, "gno.land/p/"):
		return "package"
	case strings.HasPrefix(p, "gno.land/r/"):
		return "realm"
	case strings.Contains(p, "."):
		return "external"
	default:
		return "stdlib"
	}
}

func buildImportLink(p, kind, domain string) string {
	if kind == "package" || kind == "realm" {
		return strings.TrimPrefix(p, domain)
	}
	return ""
}

// computeStats derives numeric counters from the file list and qdoc payload.
func computeStats(files []string, jdoc *doc.JSONDocumentation, imports []ImportLink) PackageStats {
	s := PackageStats{
		FileCount:   len(files),
		ImportCount: len(imports),
	}
	for _, f := range files {
		switch {
		case strings.HasSuffix(f, "_test.gno"), strings.HasSuffix(f, "_filetest.gno"):
			s.TestCount++
			s.GnoFileCount++
		case strings.HasSuffix(f, ".gno"):
			s.GnoFileCount++
		}
	}
	if jdoc == nil {
		return s
	}
	for _, fn := range jdoc.Funcs {
		s.FuncCount++
		if fn.Type == "" && token.IsExported(fn.Name) {
			s.ExportedFunc++
		}
		if fn.Crossing {
			s.CrossingCount++
		}
	}
	s.TypeCount = len(jdoc.Types)
	for _, v := range jdoc.Values {
		if v.Const {
			s.ConstCount++
		} else {
			s.VarCount++
		}
	}
	return s
}

// deriveQuality flags presence indicators used by the quality UI section.
func deriveQuality(files []string, jdoc *doc.JSONDocumentation) PackageQuality {
	q := PackageQuality{SourceVerified: true}
	for _, f := range files {
		if f == "README.md" {
			q.HasReadme = true
		}
		if ReLicenseFileName.MatchString(f) {
			q.HasLicense = true
		}
		if strings.HasSuffix(f, "_test.gno") || strings.HasSuffix(f, "_filetest.gno") {
			q.HasTests = true
		}
	}
	if jdoc != nil && strings.TrimSpace(jdoc.PackageDoc) != "" {
		q.HasPkgDoc = true
	}
	return q
}

// extractSynopsis returns the first line of packageDoc, capped at 120 runes.
func extractSynopsis(packageDoc string) string {
	trimmed := strings.TrimSpace(packageDoc)
	if trimmed == "" {
		return ""
	}
	first := strings.SplitN(trimmed, "\n", 2)[0]
	runes := []rune(first)
	if len(runes) > 120 {
		return string(runes[:117]) + "..."
	}
	return first
}

// buildFileLinks turns file names into URL-backed entries for the Files section.
func buildFileLinks(pkgPath string, files []string) []FileLink {
	out := make([]FileLink, 0, len(files))
	for _, f := range files {
		link := pkgPath + "$source&file=" + f
		entry := FileLink{Name: f, Link: link}
		switch {
		case strings.HasSuffix(f, "_test.gno"), strings.HasSuffix(f, "_filetest.gno"):
			entry.IsTest = true
		case f == "README.md":
			entry.IsReadme = true
		}
		if ReLicenseFileName.MatchString(f) {
			entry.IsLicense = true
		}
		out = append(out, entry)
	}
	return out
}

// buildSubpackages keeps direct children only (one level deep) and drops self.
// Input paths are expected to be domain-relative (e.g. "/r/demo/foo/bar").
func buildSubpackages(self string, paths []string) []SubpackageLink {
	out := make([]SubpackageLink, 0)
	seen := map[string]bool{self: true}
	selfPrefix := self + "/"
	for _, p := range paths {
		p = strings.TrimSuffix(p, "/")
		if seen[p] {
			continue
		}
		seen[p] = true
		rel := strings.TrimPrefix(p, selfPrefix)
		if rel == p || rel == "" || strings.Contains(rel, "/") {
			continue
		}
		out = append(out, SubpackageLink{Name: rel, Path: p})
	}
	return out
}

// buildOverviewTOC builds the hierarchical table-of-contents items used by the sidebar.
func buildOverviewTOC(quality PackageQuality, funcs []FuncEntry, types []TypeEntry, values []ValueGroup) []*TocItem {
	var toc []*TocItem
	if quality.HasPkgDoc {
		toc = append(toc, &TocItem{Title: "Overview", ID: "overview"})
	}
	if quality.HasReadme {
		toc = append(toc, &TocItem{Title: "README", ID: "readme"})
	}
	hasConst, hasVar := false, false
	for _, v := range values {
		if v.Kind == "const" {
			hasConst = true
		}
		if v.Kind == "var" {
			hasVar = true
		}
	}
	if hasConst {
		toc = append(toc, &TocItem{Title: "Constants", ID: "constants"})
	}
	if hasVar {
		toc = append(toc, &TocItem{Title: "Variables", ID: "variables"})
	}
	if len(funcs) > 0 {
		item := &TocItem{Title: "Functions", ID: "functions"}
		for _, fn := range funcs {
			item.Items = append(item.Items, &TocItem{Title: fn.Name, ID: fn.AnchorID})
		}
		toc = append(toc, item)
	}
	if len(types) > 0 {
		item := &TocItem{Title: "Types", ID: "types"}
		for _, t := range types {
			ti := &TocItem{Title: t.Name, ID: t.AnchorID}
			for _, m := range t.Methods {
				ti.Items = append(ti.Items, &TocItem{Title: m.Name, ID: m.AnchorID})
			}
			item.Items = append(item.Items, ti)
		}
		toc = append(toc, item)
	}
	toc = append(toc, &TocItem{Title: "Files", ID: "files"})
	return toc
}

// deriveInfo builds the sidebar metadata for a package.
func deriveInfo(gnourl *weburl.GnoURL, _ []string, gnomod []byte) PackageInfo {
	info := PackageInfo{
		Namespace:   gnourl.Namespace(),
		PackagePath: gnourl.Path,
		PackageType: packageTypeOf(gnourl),
	}
	if v, ok := parseGnoVersion(gnomod); ok {
		info.GnoVersion = v
	}
	return info
}

func packageTypeOf(gnourl *weburl.GnoURL) string {
	switch {
	case gnourl.IsRealm():
		return "realm"
	case gnourl.IsPure():
		return "pure"
	}
	return ""
}

var gnoVersionRE = regexp.MustCompile(`(?m)^\s*gno\s*=\s*"([^"]+)"`)

// parseGnoVersion reads the gno version directive from gnomod.toml content.
func parseGnoVersion(gnomod []byte) (string, bool) {
	if len(gnomod) == 0 {
		return "", false
	}
	m := gnoVersionRE.FindSubmatch(gnomod)
	if len(m) != 2 {
		return "", false
	}
	return string(m[1]), true
}

// buildSymbols splits jdoc.Funcs into top-level functions and methods
// attached to their receiver TypeEntry. Unexported entries are dropped.
// Methods whose receiver isn't declared in jdoc.Types are silently skipped.
func buildSymbols(jdoc *doc.JSONDocumentation, render DocRenderer, pkgPath string) ([]FuncEntry, []TypeEntry) {
	if jdoc == nil {
		return nil, nil
	}

	typeTable := make(map[string]*TypeEntry, len(jdoc.Types))
	for _, t := range jdoc.Types {
		if !token.IsExported(t.Name) {
			continue
		}
		typeTable[t.Name] = &TypeEntry{
			Name:               t.Name,
			Signature:          t.Type,
			SignatureComponent: renderSignature(t.Type, render),
			Kind:               t.Kind,
			Doc:                renderDocString(t.Doc, render),
			AnchorID:           "type-" + t.Name,
			SourceURL:          buildSourceURL(pkgPath, t.File, t.Line),
		}
	}

	var topFuncs []FuncEntry
	for _, fn := range jdoc.Funcs {
		if !token.IsExported(fn.Name) {
			continue
		}
		entry := FuncEntry{
			Name:               fn.Name,
			Signature:          fn.Signature,
			SignatureComponent: renderSignature(fn.Signature, render),
			Doc:                renderDocString(fn.Doc, render),
			Crossing:           fn.Crossing,
			Receiver:           fn.Type,
			IsMethod:           fn.Type != "",
			AnchorID:           symbolAnchor(fn),
			SourceURL:          buildSourceURL(pkgPath, fn.File, fn.Line),
		}
		if fn.Type == "" && fn.Name != "Render" {
			entry.ActionURL = pkgPath + "$help&func=" + fn.Name
		}
		if fn.Type == "" {
			topFuncs = append(topFuncs, entry)
			continue
		}
		if t, ok := typeTable[fn.Type]; ok {
			t.Methods = append(t.Methods, entry)
		}
	}

	out := make([]TypeEntry, 0, len(jdoc.Types))
	for _, t := range jdoc.Types {
		if entry := typeTable[t.Name]; entry != nil {
			out = append(out, *entry)
		}
	}
	return topFuncs, out
}

// isExportedValueDecl reports whether a value declaration group contains at
// least one exported name.
func isExportedValueDecl(v *doc.JSONValueDecl) bool {
	for _, vv := range v.Values {
		if token.IsExported(vv.Name) {
			return true
		}
	}
	return false
}

// buildSourceURL returns the deep link to a symbol's declaration site in the
// source view. Returns "" if file/line is not known — callers must check
// before emitting links.
func buildSourceURL(pkgPath, file string, line int) string {
	if file == "" || line <= 0 {
		return ""
	}
	return pkgPath + "$source&file=" + file + "#L" + strconv.Itoa(line)
}

// symbolAnchor returns a stable anchor id for a function or method.
func symbolAnchor(fn *doc.JSONFunc) string {
	if fn.Type != "" {
		return "method-" + fn.Type + "-" + fn.Name
	}
	return "func-" + fn.Name
}

// renderSignature syntax-highlights a Gno signature snippet as HTML.
// Returns nil for empty signatures; falls back to HTML-escaped text on renderer failure.
func renderSignature(sig string, render DocRenderer) Component {
	if strings.TrimSpace(sig) == "" {
		return nil
	}
	var buf bytes.Buffer
	if err := render.RenderSource(&buf, "sig.gno", []byte(sig)); err != nil {
		return rawHTMLComponent(template.HTMLEscapeString(sig))
	}
	return rawHTMLComponent(buf.String())
}

// renderDocString renders a doc string through the DocRenderer into a Component.
// A nil or empty doc returns nil so the template can use {{ with .Doc }}.
// Renderer failures degrade to a plain HTML-escaped string.
func renderDocString(src string, render DocRenderer) Component {
	if strings.TrimSpace(src) == "" {
		return nil
	}
	var buf bytes.Buffer
	if err := render.RenderDocumentation(&buf, []byte(src)); err != nil {
		escaped := template.HTMLEscapeString(src)
		return rawHTMLComponent(escaped)
	}
	return rawHTMLComponent(buf.String())
}

// rawHTMLComponent wraps pre-rendered HTML (safe by construction) as a Component.
type rawHTMLComponent string

func (s rawHTMLComponent) Render(w io.Writer) error {
	_, err := io.WriteString(w, string(s))
	return err
}

// buildValues flattens jdoc.Values preserving source order and groups names.
// Unexported declarations are dropped (jdoc may include unexported symbols).
func buildValues(jdoc *doc.JSONDocumentation, render DocRenderer, pkgPath string) []ValueGroup {
	if jdoc == nil || len(jdoc.Values) == 0 {
		return nil
	}
	out := make([]ValueGroup, 0, len(jdoc.Values))
	for i, v := range jdoc.Values {
		if !isExportedValueDecl(v) {
			continue
		}
		kind := "var"
		if v.Const {
			kind = "const"
		}
		names := make([]string, 0, len(v.Values))
		for _, vv := range v.Values {
			names = append(names, vv.Name)
		}
		anchorID := "value-"
		if len(v.Values) > 0 {
			anchorID += v.Values[0].Name
		} else {
			anchorID += kind + "-" + strconv.Itoa(i)
		}
		out = append(out, ValueGroup{
			Kind:               kind,
			Names:              strings.Join(names, ", "),
			SignatureComponent: renderSignature(v.Signature, render),
			Doc:                renderDocString(v.Doc, render),
			AnchorID:           anchorID,
			SourceURL:          buildSourceURL(pkgPath, v.File, v.Line),
		})
	}
	return out
}

// OverviewInput aggregates the data required to build an OverviewData.
type OverviewInput struct {
	URL         *weburl.GnoURL
	Files       []string
	Doc         *doc.JSONDocumentation
	Sources     map[string][]byte
	Subpaths    []string
	Readme      Component
	Domain      string
	DocRenderer DocRenderer
}

// BuildOverview is pure: given fetched inputs, it returns the rendered OverviewData.
func BuildOverview(in OverviewInput) OverviewData {
	info := deriveInfo(in.URL, in.Files, in.Sources["gnomod.toml"])
	info.License = deriveLicense(in.Files, func(name string) ([]byte, bool) {
		v, ok := in.Sources[name]
		return v, ok
	})
	quality := deriveQuality(in.Files, in.Doc)
	imports := parseImports(filterNonTestSources(in.Sources), in.Domain)
	funcs, types := buildSymbols(in.Doc, in.DocRenderer, in.URL.Path)
	values := buildValues(in.Doc, in.DocRenderer, in.URL.Path)
	stats := computeStats(in.Files, in.Doc, imports)
	files := buildFileLinks(in.URL.Path, in.Files)
	subpacks := buildSubpackages(in.URL.Path, in.Subpaths)
	toc := buildOverviewTOC(quality, funcs, types, values)

	pkgDocSynopsis := ""
	var pkgDocComp Component
	if in.Doc != nil {
		pkgDocSynopsis = extractSynopsis(in.Doc.PackageDoc)
		pkgDocComp = renderDocString(in.Doc.PackageDoc, in.DocRenderer)
	}

	var bugs []string
	if in.Doc != nil {
		bugs = in.Doc.Bugs
	}

	return OverviewData{
		PkgPath:      in.URL.Path,
		Title:        path.Base(in.URL.Path),
		Synopsis:     pkgDocSynopsis,
		PackageDoc:   pkgDocComp,
		Readme:       in.Readme,
		Info:         info,
		Stats:        stats,
		Quality:      quality,
		Funcs:        funcs,
		Types:        types,
		Values:       values,
		Imports:      imports,
		Files:        files,
		Subpackages:  subpacks,
		Bugs:         bugs,
		ComponentTOC: NewTemplateComponent("ui/toc_realm", &RealmTOCData{Items: toc}),
	}
}

// filterNonTestSources returns the subset of sources used for import parsing.
func filterNonTestSources(sources map[string][]byte) map[string][]byte {
	if len(sources) == 0 {
		return nil
	}
	out := make(map[string][]byte, len(sources))
	for name, body := range sources {
		if strings.HasSuffix(name, ".gno") &&
			!strings.HasSuffix(name, "_test.gno") &&
			!strings.HasSuffix(name, "_filetest.gno") {
			out[name] = body
		}
	}
	return out
}
