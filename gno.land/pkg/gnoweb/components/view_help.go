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
		url := pkgPath + "$help&func=" + fn.Name
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

	funcs["buildHelpQR"] = func(data HelpData, fn *doc.JSONFunc) string {
		// create and scale the QR code
		buildHelpURL := funcs["buildHelpURL"].(func(HelpData, *doc.JSONFunc) string)
		qrCode, _ := qr.Encode(buildHelpURL(data, fn), qr.M, qr.Auto)
		qrCode, _ = barcode.Scale(qrCode, 200, 200)

		// encode the QR code as png in base64
		buf := new(bytes.Buffer)
		png.Encode(buf, qrCode)
		return base64.StdEncoding.EncodeToString(buf.Bytes())
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
