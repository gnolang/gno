package page_builder

import (
	"strings"
)

type HTML struct {
    content []string
}

func New() *HTML {
    return &HTML{}
}

func (html *HTML) H1(text string) {
	html.content = append(html.content, "<h1>" + text + "</h1>")
}

func (html *HTML) H2(text string) {
	html.content = append(html.content, "<h2>" + text + "</h2>")
}

func (html *HTML) H3(text string) {
	html.content = append(html.content, "<h3>" + text + "</h3>")
}

func (html *HTML) P(text string) {
	html.content = append(html.content, "<p>" + text + "</p>")
}

func (html *HTML) UL(items []string) {
	var listItems []string
	for _, item := range items {
		listItems = append(listItems, "<li>" + item + "</li>")
	}
	html.content = append(html.content, "<ul>\n" + strings.Join(listItems, "\n") + "\n</ul>")
}

func (html *HTML) OL(items []string) {
	var listItems []string
	for _, item := range items {
		listItems = append(listItems, "<li>" + item + "</li>")
	}
	html.content = append(html.content, "<ol>\n" + strings.Join(listItems, "\n") + "\n</ol>")
}

// TODO: Implement code highlughting
func (html *HTML) CodeBlock(code string) {
	html.content = append(html.content, "<pre><code>" + code + "</code></pre>")
}

func (html *HTML) Image(src string, alt string) {
	html.content = append(html.content, `<img src="` + src + `" alt="` + alt + `" />`)
}

func (html *HTML) Link(href string, text string) {
	html.content = append(html.content, `<a href="` + href + `">` + text + `</a>`)
}

func (html *HTML) Bold(text string) string {
	return "<b>" + text + "</b>"
}

func (html *HTML) Italic(text string) string {
	return "<i>" + text + "</i>"
}

func (html *HTML) Strikethrough(text string) string {
	return "<s>" + text + "</s>"
}

func (html *HTML) BreakLine() {
	html.content = append(html.content, "<br>")
}

func (html *HTML) Render() string {
    return strings.Join(html.content, "\n")
}
