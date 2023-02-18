package doc

import (
	"go/doc/comment"
	"html/template"
	"regexp"
)

var TemplateFuncs = template.FuncMap{
	"comment": CommentTplFunc,
}

func CommentTplFunc(s string) template.HTML {
	var p comment.Parser
	doc := p.Parse(s)
	var pr comment.Printer
	pr.DocLinkBaseURL = "/p/demo"

	re := regexp.MustCompile(`(?s)<pre>(.*?)</pre>`)
	output := re.ReplaceAllString(string(pr.HTML(doc)), "<pre><code>$1</code></pre>")

	return template.HTML(output)
}
