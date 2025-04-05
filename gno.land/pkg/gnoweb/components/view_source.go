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

type SourceTocData struct {
	Icon         string
	Items        []SourceTocItem
	RegularFiles []SourceTocItem
	TestFiles    []SourceTocItem
	NonGnoFiles  []SourceTocItem
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
		Icon:  "file",
		Items: make([]SourceTocItem, len(data.Files)),
	}

	for i, file := range data.Files {
		item := SourceTocItem{
			Link: data.PkgPath + "$source&file=" + file,
			Text: file,
		}

		tocData.Items[i] = item

		if file == "README.md" {
			tocData.RegularFiles = append(tocData.RegularFiles, item)
		} else if strings.Contains(file, "_test.") || strings.HasSuffix(file, "test.gno") || strings.HasSuffix(file, "_filetest.gno") {
			tocData.TestFiles = append(tocData.TestFiles, item)
		} else if !strings.HasSuffix(file, ".gno") {
			tocData.NonGnoFiles = append(tocData.NonGnoFiles, item)
		} else {
			tocData.RegularFiles = append(tocData.RegularFiles, item)
		}
	}

	toc := NewTemplateComponent("ui/toc_source", tocData)
	content := NewTemplateComponent("ui/code_wrapper", data.FileSource)
	viewData := sourceViewParams{
		Article: ArticleData{
			ComponentContent: content,
			Classes:          "source-view col-span-1 lg:col-span-7 lg:row-start-2 pb-24 text-gray-900",
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
