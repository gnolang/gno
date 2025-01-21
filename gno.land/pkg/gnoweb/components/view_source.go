package components

import (
	"bytes"
	"html/template"
	"io"
)

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
	TOC         template.HTML
}

type SourceTocData struct {
	Icon  string
	Items []SourceTocItem
}

type SourceTocItem struct {
	Link string
	Text string
}

func RenderSourceComponent(w io.Writer, data SourceData) error {
	var tocBuf, contentBuf bytes.Buffer

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

	if err := tmpl.ExecuteTemplate(&tocBuf, "layout/toc_list", tocData); err != nil {
		return err
	}

	// Générer le contenu avec wrapper
	if err := tmpl.ExecuteTemplate(&contentBuf, "renderSourceContent", data.FileSource); err != nil {
		return err
	}

	viewData := SourceViewData{
		Article: ArticleData{
			Content: template.HTML(contentBuf.String()), //nolint:gosec
			Classes: "source-content col-span-1 lg:col-span-7 lg:row-start-2 pb-24 text-gray-900",
		},
		TOC:         template.HTML(tocBuf.String()), //nolint:gosec
		Files:       data.Files,
		FileName:    data.FileName,
		FileSize:    data.FileSize,
		FileLines:   data.FileLines,
		FileCounter: data.FileCounter,
		PkgPath:     data.PkgPath,
	}
	return tmpl.ExecuteTemplate(w, "renderSource", viewData)
}
