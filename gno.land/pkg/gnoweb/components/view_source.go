package components

import (
	"fmt"
)

const SourceViewType ViewType = "source-view"

type DisplayMode int

const (
	ModeCode DisplayMode = iota
	ModeMarkdown
)

type SourceData struct {
	PkgPath      string
	Mode         DisplayMode
	Files        []string
	FileName     string
	FileSize     string
	FileLines    int
	FileCounter  int
	FileDownload string
	FileSource   Component
	IsMarkdown   bool
}

type SourceTocData struct {
	Icon  string
	Items []SourceTocItem
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
	Mode         DisplayMode
	IsMarkdown   bool
}

func SourceView(data SourceData) *View {
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

	toc := NewTemplateComponent("ui/toc_generic", tocData)
	var content Component
	if data.Mode == ModeCode {
		content = NewTemplateComponent("ui/code_wrapper", data.FileSource)
	} else {
		content = data.FileSource
	}
	viewData := sourceViewParams{
		Article: ArticleData{
			ComponentContent: content,
			Classes: fmt.Sprintf("%s col-span-1 lg:col-span-7 lg:row-start-2 mb-24 text-gray-900",
				map[DisplayMode]string{ModeCode: "source-view", ModeMarkdown: "md-view bg-light rounded px-4"}[data.Mode]),
		},
		ComponentTOC: toc,
		Files:        data.Files,
		FileName:     data.FileName,
		FileSize:     data.FileSize,
		FileLines:    data.FileLines,
		FileCounter:  data.FileCounter,
		PkgPath:      data.PkgPath,
		FileDownload: data.FileDownload,
		Mode:         data.Mode,
		IsMarkdown:   data.IsMarkdown,
	}

	return NewTemplateView(SourceViewType, "renderSource", viewData)
}
