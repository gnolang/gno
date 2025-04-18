package gnomark

import (
	"strings"
)

// TODO: build an interface and registry for parsers and renderers
// make this dyamic (after POC proving call out to gno will work
var (
	KindPetriNet = "PetriNetBlock"

	webHost = &WebHost{
		Base: "https://cdn.jsdelivr.net/gh/pflow-xyz/pflow-app@",
		Tag:  "0.2.1",
		Path: "/static/",
	}
)

// pflowBlock represents a custom block for pflow.
type pflowBlock struct {
	JSONContent string
}

// pflowParser parses the custom <petri-net> block.
type pflowParser struct{}

var startTag = "<petri-net>"
var endTag = "</petri-net>"

func (p *pflowParser) Parse(input string) *pflowBlock {
	lines := strings.Split(input, "\n")
	var block *pflowBlock
	for _, line := range lines {
		if strings.HasPrefix(line, startTag) {
			block = &pflowBlock{}
			continue
		}
		if strings.HasPrefix(line, endTag) {
			break
		}
		if block != nil {
			block.JSONContent += line + "\n"
		}
	}
	return block
}

// pflowRenderer renders the pflow block as HTML.
type pflowRenderer struct{}

func (r *pflowRenderer) Render(block *pflowBlock) string {
	if block == nil {
		return ""
	}
	return Render(block.JSONContent)
}

// WebHost represents the CDN configuration.
type WebHost struct {
	Base string
	Tag  string
	Path string
}

func (h *WebHost) Cdn() string {
	return h.Base + h.Tag + h.Path
}

// Render generates the HTML for the pflow block.
func Render(source string) string {
	return templateHtml("{CDN}", webHost.Cdn(), source)
}

func templateHtml(key, value, source string) string {
	out := strings.ReplaceAll(htmlContent, key, value)
	return strings.ReplaceAll(out, "{SOURCE}", source)
}

var htmlContent = `
    <style type="text/css">
        @import url("{CDN}pflow.css");
    </style>
    <svg id="svgCanvas" width="100%" height="100%"  xmlns="http://www.w3.org/2000/svg">
        <foreignObject height="100%" width="100%" x="0" y="0">
            <object id="svgObject" type="image/svg+xml" data="{CDN}model.svg"></object>
        </foreignObject>
        <g id="toolbar">
            <g id="status" transform="translate(5, 607)" >
                <rect x="0" y="0" width="140" height="20" fill="#fff" rx="5" ry="5"/>
                <text x="10" y="15">Status: Ready</text>
            </g>
            <g id="playBtn" transform="translate(148, 604)" >
                <circle cx="12" cy="11" r="20" fill="transparent" stroke="transparent" stroke-width="2" />
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2m0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8m-2.5-3.5 7-4.5-7-4.5z"></path>
            </g>
        </g>
        <g id="history" transform="translate(5, 605)" ></g>
        <foreignObject height="50%" width="98%" x="0" y="635">
            <textarea id="source">{SOURCE}</textarea>
        </foreignObject>
    </svg>
    <script src="{CDN}pflow.js"></script>
`
