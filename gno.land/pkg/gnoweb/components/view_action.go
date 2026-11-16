package components

import (
	"html/template"
	"strconv"
	"strings"

	// for error types
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

const HelpViewType ViewType = "help-view"

// HelpFunction pairs a doc.JSONFunc with its documentation already rendered
// as an HTML Component, so the template can embed it via `{{ render . }}`.
type HelpFunction struct {
	*doc.JSONFunc
	DocComponent Component
}

type HelpData struct {
	// Selected function
	SelectedFunc string
	SelectedArgs map[string]string
	SelectedSend string

	RealmName   string
	Functions   []HelpFunction
	ChainId     string
	Remote      string
	PkgPath     string
	PkgFullPath string
	Doc         Component
	Domain      string
	Origin      string // request scheme+host; makes help URLs shareable
	// Community is set on packages outside the trusted paths, the pages that
	// show the realm notice.
	Community bool
}

type HelpTocData struct {
	Icon  string
	Items []HelpTocItem
}

type HelpTocItem struct {
	Link string
	Text string
	// Icon replaces the list's icon for this item when set.
	Icon string
}

// helpEvalAnchor is the id of the expression evaluator below the functions.
const helpEvalAnchor = "expression-evaluator"

// Gas is the gas a printed gnokey command asks for and the fee it offers.
type Gas struct{ Wanted, Fee string }

var (
	// CallGas is what $help's maketx call commands ask for: room for any
	// function.
	CallGas = Gas{Wanted: "1_000_000_000", Fee: "1000000ugnot"}
	// RunGas is what a $run script asks for. The starter script names a
	// function rather than calling Render, whose gas is the package's (over
	// 400 million on some realms), and takes 1 to 9 million on the example
	// realms; the fee pays for 100 million at the chain's initial price
	// (1ugnot per 1000 gas), so raising the gas wanted that far needs no new
	// fee.
	RunGas = Gas{Wanted: "50_000_000", Fee: "100000ugnot"}
)

type CommandData struct {
	FuncName string
	PkgPath  string
	// ArgKeys links each -args slot to its param's inputs: a position on
	// $help, where names repeat or are missing ("_"), a name in a form.
	ArgKeys []string
	ChainId string
	Remote  string
}

type helpViewParams struct {
	HelpData
	Article      ArticleData
	ComponentTOC Component
}

func registerHelpFuncs(funcs template.FuncMap) {
	funcs["evalAnchor"] = func() string { return helpEvalAnchor }
	funcs["callGas"] = func() Gas { return CallGas }

	funcs["getSelectedArgValue"] = func(data HelpData, param *doc.JSONField) (string, error) {
		if data.SelectedArgs == nil {
			return "", nil
		}

		return data.SelectedArgs[param.Name], nil
	}

	funcs["buildHelpURL"] = func(data HelpData, fn HelpFunction) string {
		pkgPath := strings.TrimPrefix(data.PkgPath, data.Domain)
		var url strings.Builder
		url.WriteString(data.Origin + pkgPath + "$help&func=" + fn.Name)
		if len(fn.Params) > 0 {
			url.WriteString("&")
			for i, param := range fn.Params {
				if i > 0 {
					url.WriteString("&")
				}
				url.WriteString(param.Name + "=")
				if val, ok := data.SelectedArgs[param.Name]; ok {
					url.WriteString(val)
				}
			}
		}
		return url.String()
	}

	funcs["aiFuncAction"] = func(data HelpData, fn HelpFunction) *AIAction {
		return NewAIFuncAction(data.Origin, strings.TrimPrefix(data.PkgPath, data.Domain), fn.Name)
	}

	funcs["buildCommandData"] = func(data HelpData, fn HelpFunction) CommandData {
		argKeys := make([]string, len(fn.Params))
		for i := range fn.Params {
			argKeys[i] = strconv.Itoa(i)
		}

		return CommandData{
			FuncName: fn.Name,
			PkgPath:  data.PkgPath,
			ArgKeys:  argKeys,
			ChainId:  data.ChainId,
			Remote:   data.Remote,
		}
	}
}

func HelpView(data HelpData) *View {
	tocData := HelpTocData{
		Icon:  "code",
		Items: make([]HelpTocItem, len(data.Functions)),
	}

	for i, fn := range data.Functions {
		var sig strings.Builder
		sig.WriteString(fn.Name + "(")
		for j, param := range fn.Params {
			if j > 0 {
				sig.WriteString(", ")
			}
			sig.WriteString(param.Name)
		}
		sig.WriteString(")")

		tocData.Items[i] = HelpTocItem{
			Link: "#func-" + fn.Name,
			Text: sig.String(),
		}
	}
	// The evaluator stays below the functions; the list names it so a long
	// page does not hide it.
	tocData.Items = append(tocData.Items, HelpTocItem{
		Link: "#" + helpEvalAnchor,
		Text: "Expression evaluator",
		Icon: "eval",
	})

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
