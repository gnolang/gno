package components

import (
	"html/template"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm" // for error types
)

const HelpViewType ViewType = "help-view"

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
	TOC     Component
}

type HelpTocData struct {
	Icon  string
	Items []HelpTocItem
}

type HelpTocItem struct {
	Link string
	Text string
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

func RenderHelpView(data HelpData) *View {
	funcMap := template.FuncMap{}
	registerHelpFuncs(funcMap)

	tocData := HelpTocData{
		Icon:  "code",
		Items: make([]HelpTocItem, len(data.Functions)),
	}

	for i, fn := range data.Functions {
		sig := fn.FuncName + "("
		for j, param := range fn.Params {
			if j > 0 {
				sig += ", "
			}
			sig += param.Name
		}
		sig += ")"

		tocData.Items[i] = HelpTocItem{
			Link: "#func-" + fn.FuncName,
			Text: sig,
		}
	}

	toc := NewTemplateComponent("layout/toc_list", tocData)
	content := NewTemplateComponent("renderHelpContent", data)
	viewData := HelpViewData{
		HelpData: data,
		Article: ArticleData{
			Content: content,
			Classes: "",
		},
		TOC: toc,
	}

	return NewTemplateView(HelpViewType, "renderHelp", viewData)
}
