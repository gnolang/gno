package events

import (
	"fmt"
	"html/template"
	"strings"
)

var tmplFuncs = template.FuncMap{
	"jsEventsArray": func(events []EventType) string {
		var b strings.Builder
		b.WriteString("[")
		for i, v := range events {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%q", v))
		}
		b.WriteString("]")
		return b.String()
	},
}
