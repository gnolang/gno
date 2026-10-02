package components

import (
	"regexp"
	"strings"
)

const (
	SourceViewType ViewType = "source-view"
	ReadmeFileName string   = "README.md"
)

var ReLicenseFileName *regexp.Regexp = regexp.MustCompile(`(?i)^licen[cs]e(.md|.txt)?$`)

// SourceData holds data for rendering a source code view.
type SourceData struct {
	PkgPath      string
	Files        []string
	FileName     string
	FileSize     string
	FileLines    int
	FileCounter  int
	FileDownload string
	FileSource   Component
}

// WrappedSource wraps the rendered file content in an element carrying the
// copy target. README.md is rendered markdown (not code), so it uses a plain
// wrapper; every other file uses the code frame. Both expose data-copy-target
// so the header Copy button has something to read.
func (d SourceData) WrappedSource() Component {
	if d.FileName == ReadmeFileName {
		return NewTemplateComponent("ui/readme_wrapper", d.FileSource)
	}
	return NewTemplateComponent("ui/code_wrapper", d.FileSource)
}

// ArticleClasses returns the CSS classes based on file type.
func (d SourceData) ArticleClasses() string {
	if d.FileName == ReadmeFileName {
		return "c-readme-view"
	}
	return "c-source-view"
}

type SourceTocData struct {
	Icon         string
	ReadmeFile   SourceTocItem
	LicenseFile  SourceTocItem
	GnoFiles     []SourceTocItem
	GnoTestFiles []SourceTocItem
	TomlFiles    []SourceTocItem
}

// SourceTocItem represents an item in the source view table of contents.
type SourceTocItem struct {
	Link string
	Text string
}

// sourceViewParams holds parameters for rendering the source view template.
type sourceViewParams struct {
	Article      ArticleData
	Files        []string
	FileName     string
	FileSize     string
	FileLines    int
	FileCounter  int
	PkgPath      string
	FileDownload string
	ComponentTOC Component
}

// SourceView creates a new View for displaying source code and its table of contents.
func SourceView(data SourceData) *View {
	tocData := SourceTocData{
		Icon: "file",
	}

	for _, file := range data.Files {
		item := SourceTocItem{
			Link: data.PkgPath + "$source&file=" + file,
			Text: file,
		}

		switch {
		case file == ReadmeFileName:
			tocData.ReadmeFile = item

		case strings.HasSuffix(file, "_test.gno") || strings.HasSuffix(file, "_filetest.gno"):
			tocData.GnoTestFiles = append(tocData.GnoTestFiles, item)

		case strings.HasSuffix(file, ".gno"):
			tocData.GnoFiles = append(tocData.GnoFiles, item)

		case strings.HasSuffix(file, ".toml"):
			tocData.TomlFiles = append(tocData.TomlFiles, item)
		}

		if ReLicenseFileName.MatchString(file) {
			tocData.LicenseFile = item
		}
	}

	toc := NewTemplateComponent("ui/toc_source", tocData)
	content := data.WrappedSource()
	viewData := sourceViewParams{
		Article: ArticleData{
			ComponentContent: content,
			Classes:          data.ArticleClasses(),
		},
		ComponentTOC: toc,
		Files:        data.Files,
		FileName:     data.FileName,
		FileSize:     data.FileSize,
		FileLines:    data.FileLines,
		FileCounter:  data.FileCounter,
		PkgPath:      data.PkgPath,
		FileDownload: data.FileDownload,
	}

	return NewTemplateView(SourceViewType, "renderSource", viewData)
}
