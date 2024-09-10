package page_builder

import (
	"strings"
	"strconv"
)

type Div struct {
	content     []string
	numColumns  int
}

func New(numColumns int) *Div {
	return &Div{numColumns: numColumns}
}

func wrapInTag(tag, text string) string {
	return "<" + tag + ">" + text + "</" + tag + ">"
}

func (div *Div) H1(text string) {
	div.content = append(div.content, wrapInTag("h1", text))
}

func (div *Div) H2(text string) {
	div.content = append(div.content, wrapInTag("h2", text))
}

func (div *Div) H3(text string) {
	div.content = append(div.content, wrapInTag("h3", text))
}

func (div *Div) H4(text string) {
	div.content = append(div.content, wrapInTag("h4", text))
}

func (div *Div) H5(text string) {
	div.content = append(div.content, wrapInTag("h5", text))
}

func (div *Div) H6(text string) {
	div.content = append(div.content, wrapInTag("h6", text))
}

func (div *Div) P(text string) {
	div.content = append(div.content, wrapInTag("p", text))
}

func (div *Div) UL(items []string) {
	listItems := wrapListItems(items)
	div.content = append(div.content, "<ul>\n" + listItems + "\n</ul>")
}

func (div *Div) OL(items []string) {
	listItems := wrapListItems(items)
	div.content = append(div.content, "<ol>\n" + listItems + "\n</ol>")
}

func wrapListItems(items []string) string {
	var listItems []string
	for _, item := range items {
		listItems = append(listItems, wrapInTag("li", item))
	}
	return strings.Join(listItems, "\n")
}

func (div *Div) CodeBlock(code string) {
	div.content = append(div.content, "<pre><code>" + code + "</code></pre>")
}

func (div *Div) Image(src string, alt string) {
	div.content = append(div.content, `<img src="` + src + `" alt="` + alt + `" />`)
}

func (div *Div) Link(href string, text string) string {
	return `<a href="` + href + `">` + text + `</a>`
}

func (div *Div) Bold(text string) string {
	return wrapInTag("b", text)
}

func (div *Div) Italic(text string) string {
	return wrapInTag("i", text)
}

func (div *Div) Strikethrough(text string) string {
	return wrapInTag("s", text)
}

func (div *Div) HR() {
	div.content = append(div.content, "<hr>")
}

func (div *Div) Append(renderedHTML string) {
	div.content = append(div.content, renderedHTML)
}

func (div *Div) Render() string {
	var out strings.Builder

	if div.numColumns > 1 {
		out.WriteString("<div class='columns-" + strconv.Itoa(div.numColumns) + "'>\n")
	} else {
		out.WriteString("<div class='column'>\n")
	}

	for _, c := range div.content {
		out.WriteString(c + "\n")
	}

	out.WriteString("</div>")
	return out.String()
}
