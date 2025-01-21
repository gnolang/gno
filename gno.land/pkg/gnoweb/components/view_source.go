package components

import (
	"html/template"
)

const SourceViewType ViewType = "source"

type SourceData struct {
	PkgPath     string
	Files       []string
	FileName    string
	FileSize    string
	FileLines   int
	FileCounter int
	FileSource  template.HTML
}

type SourceViewData struct {
	Article     ArticleData
	Files       []string
	FileName    string
	FileSize    string
	FileLines   int
	FileCounter int
	PkgPath     string
	TOC         Component
}

type SourceTocData struct {
	Icon  string
	Items []SourceTocItem
}

type SourceTocItem struct {
	Link string
	Text string
}

func RenderSourceView(data SourceData) *View {
	tocData := SourceTocData{
		Icon:  "file",
		Items: make([]SourceTocItem, len(data.Files)),
	}

	for i, file := range data.Files {
		tocData.Items[i] = SourceTocItem{
			Link: data.PkgPath + "$source&file=" + file,
			Text: file,
		}
	}

	toc := NewTemplateComponent("layout/toc_list", tocData)
	content := NewTemplateComponent("renderSourceContent", data.FileSource)
	viewData := SourceViewData{
		Article: ArticleData{
			Content: content,
			Classes: "source-view col-span-1 lg:col-span-7 lg:row-start-2 pb-24 text-gray-900",
		},
		TOC:         toc,
		Files:       data.Files,
		FileName:    data.FileName,
		FileSize:    data.FileSize,
		FileLines:   data.FileLines,
		FileCounter: data.FileCounter,
		PkgPath:     data.PkgPath,
	}

	return NewTemplateView(SourceViewType, "renderSource", viewData)
}
