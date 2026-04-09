package components

import (
	"go/token"

	"github.com/gnolang/gno/gnovm/pkg/doc"
)

const EvalViewType ViewType = "eval-view"

type EvalFuncInfo struct {
	Name      string
	Signature string
	Doc       string
	Params    []*doc.JSONField
}

type EvalData struct {
	Remote    string
	PkgPath   string
	Domain    string
	Functions []EvalFuncInfo
}

type EvalTocData struct {
	Icon  string
	Items []HelpTocItem
}

type evalViewParams struct {
	EvalData
	Article      ArticleData
	ComponentTOC Component
}

func EvalView(data EvalData) *View {
	tocData := EvalTocData{
		Icon:  "code",
		Items: make([]HelpTocItem, len(data.Functions)),
	}

	for i, fn := range data.Functions {
		tocData.Items[i] = HelpTocItem{
			Link: "#func-" + fn.Name,
			Text: fn.Name,
		}
	}

	toc := NewTemplateComponent("ui/toc_generic", tocData)
	content := NewTemplateComponent("renderEvalContent", data)

	viewData := evalViewParams{
		EvalData: data,
		Article: ArticleData{
			ComponentContent: content,
			Classes:          "c-eval-view",
		},
		ComponentTOC: toc,
	}

	return NewTemplateView(EvalViewType, "renderEval", viewData)
}

// BuildEvalFuncs extracts public non-method functions suitable for eval.
func BuildEvalFuncs(jdoc *doc.JSONDocumentation) []EvalFuncInfo {
	funcs := make([]EvalFuncInfo, 0, len(jdoc.Funcs))
	for _, fn := range jdoc.Funcs {
		if fn.Type != "" || !token.IsExported(fn.Name) {
			continue
		}
		// Skip crossing functions (they mutate state)
		if fn.Crossing {
			continue
		}

		params := fn.Params
		if len(params) >= 1 && params[0].Type == "realm" {
			params = params[1:]
		}

		funcs = append(funcs, EvalFuncInfo{
			Name:      fn.Name,
			Signature: fn.Signature,
			Doc:       fn.Doc,
			Params:    params,
		})
	}
	return funcs
}
