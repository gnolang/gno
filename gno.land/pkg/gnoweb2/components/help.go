package components

import (
	"html/template"
	"io"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm" // for error types
)

type HelpData struct {
	RealmName string
	Functions []vm.FunctionSignature
	ChainId string
	Remote string
	PkgPath string
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
}


func RenderHelpComponent(w io.Writer, data HelpData) error {
	return tmpl.ExecuteTemplate(w, "renderHelp", data)
}
