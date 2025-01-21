package components

import (
	"bytes"
	"html/template"
	"io"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm" // for error types
)

type HelpData struct {
	// Selected function
	SelectedFunc string
	SelectedArgs map[string]string

	RealmName string
	Functions []vm.FunctionSignature
	ChainId   string
	Remote    string
	PkgPath   string
}

type HelpViewData struct {
	HelpData
	Article ArticleData
	TOC     template.HTML
}

func registerHelpFuncs(funcs template.FuncMap) {
	funcs["helpFuncSignature"] = func(fsig vm.FunctionSignature) (string, error) {
		var fsigStr strings.Builder

		fsigStr.WriteString(fsig.FuncName)
		fsigStr.WriteRune('(')
		for i, param := range fsig.Params {
			if i > 0 {
				fsigStr.WriteString(", ")
			}
			fsigStr.WriteString(param.Name)
		}
		fsigStr.WriteRune(')')

		return fsigStr.String(), nil
	}

	funcs["getSelectedArgValue"] = func(data HelpData, param vm.NamedType) (string, error) {
		if data.SelectedArgs == nil {
			return "", nil
		}

		return data.SelectedArgs[param.Name], nil
	}
}

func RenderHelpComponent(w io.Writer, data HelpData) error {
	var contentBuf, tocBuf bytes.Buffer

	// Generate the content
	if err := tmpl.ExecuteTemplate(&contentBuf, "renderHelpContent", data); err != nil {
		return err
	}

	// Generate the ToC
	if err := tmpl.ExecuteTemplate(&tocBuf, "renderHelpToc", data); err != nil {
		return err
	}

	viewData := HelpViewData{
		HelpData: data,
		Article: ArticleData{
			Content: template.HTML(contentBuf.String()),
			Classes: "help-content",
		},
		TOC: template.HTML(tocBuf.String()),
	}

	return tmpl.ExecuteTemplate(w, "renderHelp", viewData)
}
