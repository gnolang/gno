package components

import (
	"bytes"
	"encoding/base64"
	"html/template"
	"image/png"
	"strings"

	// for error types
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
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
}

type HelpTocData struct {
	Icon  string
	Items []HelpTocItem
}

type HelpTocItem struct {
	Link string
	Text string
}

type CommandData struct {
	FuncName   string
	PkgPath    string
	ParamNames []string
	ChainId    string
	Remote     string
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

	funcs["buildHelpQR"] = func(data HelpData, fn HelpFunction) string {
		// create and scale the QR code
		buildHelpURL := funcs["buildHelpURL"].(func(HelpData, HelpFunction) string)
		url := buildHelpURL(data, fn)
		qrCode, _ := qr.Encode(url, qr.M, qr.Auto)
		qrCode, _ = barcode.Scale(qrCode, 200, 200)

		// encode the QR code as png in base64
		buf := new(bytes.Buffer)
		png.Encode(buf, qrCode)
		return base64.StdEncoding.EncodeToString(buf.Bytes())
	}

	funcs["buildCommandData"] = func(data HelpData, fn HelpFunction) CommandData {
		// Extract parameter names
		paramNames := make([]string, len(fn.Params))

		for i, param := range fn.Params {
			paramNames[i] = param.Name
		}

		return CommandData{
			FuncName:   fn.Name,
			PkgPath:    data.PkgPath,
			ParamNames: paramNames,
			ChainId:    data.ChainId,
			Remote:     data.Remote,
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
