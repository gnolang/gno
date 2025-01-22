package components

const SourceViewType ViewType = "source-view"

type SourceData struct {
	PkgPath     string
	Files       []string
	FileName    string
	FileSize    string
	FileLines   int
	FileCounter int
	FileSource  Component
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
	ComponentTOC Component
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
	}

	return NewTemplateView(SourceViewType, "renderSource", viewData)
}
