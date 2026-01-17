package lint

import (
	"regexp"
	"strings"
)

var nolintPattern = regexp.MustCompile(`^//\s*nolint(?::([A-Za-z0-9,]+))?`)

type NolintDirective struct {
	Line  int
	Rules []string
}

type NolintParser struct {
	byLine map[int]NolintDirective
}

func NewNolintParser(source string) *NolintParser {
	p := &NolintParser{
		byLine: make(map[int]NolintDirective),
	}
	p.parse(source)
	return p
}

func (p *NolintParser) parse(source string) {
	lines := strings.Split(source, "\n")

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		if !strings.HasPrefix(trimmed, "//") {
			continue
		}

		if match := nolintPattern.FindStringSubmatch(trimmed); match != nil {
			p.addDirective(lineNum, match)
		}
	}
}

func (p *NolintParser) addDirective(lineNum int, match []string) {
	directive := NolintDirective{
		Line: lineNum,
	}

	if match[1] != "" {
		rules := strings.Split(match[1], ",")
		for i, r := range rules {
			rules[i] = strings.TrimSpace(r)
		}
		directive.Rules = rules
	}

	p.byLine[lineNum] = directive
}

func (p *NolintParser) IsSuppressed(line int, ruleID string) bool {
	d, ok := p.byLine[line-1]
	if !ok {
		return false
	}
	return p.matchesRule(d, ruleID)
}

func (p *NolintParser) matchesRule(d NolintDirective, ruleID string) bool {
	if len(d.Rules) == 0 {
		return true
	}

	for _, rule := range d.Rules {
		if rule == ruleID {
			return true
		}
	}
	return false
}
