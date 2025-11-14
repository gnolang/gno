package components

import (
	"bytes"
	"html/template"
	"strings"

	// for error types
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

const HelpViewType ViewType = "help-view"

type HelpData struct {
	// Selected function
	SelectedFunc string
	SelectedArgs map[string]string
	SelectedSend string

	RealmName   string
	Functions   []*doc.JSONFunc
	ChainId     string
	Remote      string
	PkgPath     string
	PkgFullPath string
	Doc         string
	Domain      string
}

type HelpTocData struct {
	Icon  string
	Items []HelpTocItem
}

type HelpTocItem struct {
	Link string
	Text string
}

type helpViewParams struct {
	HelpData
	Article      ArticleData
	ComponentTOC Component
}

func registerHelpFuncs(funcs template.FuncMap) {
	funcs["getSelectedArgValue"] = func(data HelpData, param *doc.JSONField) (string, error) {
		if data.SelectedArgs == nil {
			return "", nil
		}

		return data.SelectedArgs[param.Name], nil
	}

	funcs["buildHelpURL"] = func(data HelpData, fn *doc.JSONFunc) string {
		pkgPath := strings.TrimPrefix(data.PkgPath, data.Domain)
		url := data.Domain + pkgPath + "$help&func=" + fn.Name
		if len(fn.Params) > 0 {
			url += "&"
			for i, param := range fn.Params {
				if i > 0 {
					url += "&"
				}
				url += param.Name + "="
				if val, ok := data.SelectedArgs[param.Name]; ok {
					url += val
				}
			}
		}
		return url
	}

	// Render command block using the single utility function
	funcs["renderCommandBlock"] = func(funcName, funcSig, pkgPath, chainId, remote, selectedSend string, params []*doc.JSONField) template.HTML {
		// Convert doc.JSONField to vm.NamedType
		vmParams := make([]vm.NamedType, len(params))
		for i, param := range params {
			vmParams[i] = vm.NamedType{Name: param.Name, Type: param.Type}
		}

		data := markdown.CommandBlockData{
			FuncName:     funcName,
			FuncSig:      funcSig,
			Params:       vmParams,
			PkgPath:      pkgPath,
			ChainId:      chainId,
			Remote:       remote,
			SelectedSend: selectedSend,
			Prefix:       "function",
		}

		var buf bytes.Buffer
		if err := markdown.RenderCommandBlock(&buf, data); err != nil {
			return template.HTML("<!-- Error rendering command block: " + err.Error() + " -->")
		}

		return template.HTML(buf.String())
	}
}

func HelpView(data HelpData) *View {
	tocData := HelpTocData{
		Icon:  "code",
		Items: make([]HelpTocItem, len(data.Functions)),
	}

	for i, fn := range data.Functions {
		sig := fn.Name + "("
		for j, param := range fn.Params {
			if j > 0 {
				sig += ", "
			}
			sig += param.Name
		}
		sig += ")"

		tocData.Items[i] = HelpTocItem{
			Link: "#func-" + fn.Name,
			Text: sig,
		}
	}

	toc := NewTemplateComponent("ui/toc_generic", tocData)
	content := NewTemplateComponent("ui/help_function", data)
	viewData := helpViewParams{
		HelpData: data,
		Article: ArticleData{
			ComponentContent: content,
			Classes:          "",
		},
		ComponentTOC: toc,
	}

	return NewTemplateView(HelpViewType, "renderHelp", viewData)
}
