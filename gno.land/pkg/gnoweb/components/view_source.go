package components

import (
	"strings"
)

const SourceViewType ViewType = "source-view"

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

// IsReadme reports whether the current file is README.md.
func (d SourceData) IsReadme() bool {
	return d.FileName == "README.md"
}

// WrappedSource returns a Component: raw for README.md, or code_wrapper otherwise.
func (d SourceData) WrappedSource() Component {
	if d.IsReadme() {
		return d.FileSource
	}
	return NewTemplateComponent("ui/code_wrapper", d.FileSource)
}

// ArticleClasses returns the CSS classes based on file type.
func (d SourceData) ArticleClasses() string {
	if d.IsReadme() {
		return "realm-view bg-light px-4 pt-6 pb-4 rounded lg:col-span-7"
	}
	return "source-view col-span-1 lg:col-span-7 lg:row-start-2 pb-24 text-gray-900"
}

type SourceTocData struct {
	Icon         string
	ReadmeFile   SourceTocItem
	GnoFiles     []SourceTocItem
	GnoTestFiles []SourceTocItem
}

type SourceTocItem struct {
	Link string
	Text string
}

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
		case SourceData{FileName: file}.IsReadme():
			tocData.ReadmeFile = item

		case strings.HasSuffix(file, "_test.gno") || strings.HasSuffix(file, "_filetest.gno"):
			tocData.GnoTestFiles = append(tocData.GnoTestFiles, item)

		case strings.HasSuffix(file, ".gno"):
			tocData.GnoFiles = append(tocData.GnoFiles, item)
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
