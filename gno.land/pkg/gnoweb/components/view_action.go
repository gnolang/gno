package components

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"image/png"
	"strings"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"

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

// helpQRSize is the rendered QR edge in pixels: large enough to scan from a
// laptop screen, small enough that the data URI stays a few KB.
const helpQRSize = 256

// buildHelpURL is the function's help page with its args pinned:
// `$help&func=Name&p1=v1&...`. The Execute form's action, the anchor link and
// the QR all resolve to this one URL.
func buildHelpURL(data HelpData, fn HelpFunction) string {
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

// buildHelpQR renders that URL as a QR, embedded as a data URI so the page
// needs no JS encoder and works offline. The URL already carries the submitted
// args, so the code cannot drift from what the form holds.
func buildHelpQR(data HelpData, fn HelpFunction) (template.URL, error) {
	code, err := qr.Encode(buildHelpURL(data, fn), qr.M, qr.Auto)
	if err != nil {
		return "", fmt.Errorf("unable to encode help QR: %w", err)
	}
	scaled, err := barcode.Scale(code, helpQRSize, helpQRSize)
	if err != nil {
		return "", fmt.Errorf("unable to scale help QR: %w", err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, scaled); err != nil {
		return "", fmt.Errorf("unable to encode help QR PNG: %w", err)
	}
	// template.URL: html/template rewrites an unknown scheme in src to
	// "#ZgotmplZ", and data: is one.
	return template.URL("data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())), nil //nolint:gosec // self-generated PNG data URI
}

func registerHelpFuncs(funcs template.FuncMap) {
	funcs["getSelectedArgValue"] = func(data HelpData, param *doc.JSONField) (string, error) {
		if data.SelectedArgs == nil {
			return "", nil
		}

		return data.SelectedArgs[param.Name], nil
	}

	funcs["buildHelpURL"] = buildHelpURL
	funcs["buildHelpQR"] = buildHelpQR

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
