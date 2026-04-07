package main

import (
	"fmt"
	"strings"
)

type PathKind int

const (
	PathNetwork   PathKind = iota // gno.land
	PathNamespace                 // gno.land/r/foo
	PathPackage                   // gno.land/r/foo/bar
	PathSymbol                    // gno.land/r/foo/bar.Blah or gno.land/r/foo/bar:baz
	PathCall                      // gno.land/r/foo/bar.Blah("arg1","arg2")
)

type GnoPath struct {
	Raw     string
	Domain  string
	PkgPath string
	Symbol  string
	Args    []string
	Kind    PathKind
}

func (p *GnoPath) IsPublic() bool {
	return p.Symbol != "" && p.Symbol[0] >= 'A' && p.Symbol[0] <= 'Z'
}

func ParsePath(input string) (*GnoPath, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty path")
	}

	p := &GnoPath{Raw: input}
	slashIdx := strings.Index(input, "/")
	if slashIdx < 0 {
		p.Domain = input
		p.Kind = PathNetwork
		return p, nil
	}

	p.Domain = input[:slashIdx]
	rest := input[slashIdx:]

	if callStart := strings.Index(rest, "("); callStart >= 0 {
		pkgPath, symbol := splitSymbol(p.Domain, rest[:callStart])
		if symbol == "" {
			return nil, fmt.Errorf("call expression requires a function name")
		}
		p.PkgPath = pkgPath
		p.Symbol = symbol
		args, err := parseCallArgs(rest[callStart:])
		if err != nil {
			return nil, err
		}
		p.Args = args
		p.Kind = PathCall
		return p, nil
	}

	pkgPath, symbol := splitSymbol(p.Domain, rest)
	if symbol != "" {
		p.PkgPath = pkgPath
		p.Symbol = symbol
		p.Kind = PathSymbol
		return p, nil
	}

	p.PkgPath = p.Domain + rest
	parts := strings.Split(rest, "/")
	if len(parts) <= 3 {
		p.Kind = PathNamespace
	} else {
		p.Kind = PathPackage
	}
	return p, nil
}

func splitSymbol(domain, pathPart string) (pkgPath, symbol string) {
	if colonIdx := strings.LastIndex(pathPart, ":"); colonIdx > 0 {
		return domain + pathPart[:colonIdx], pathPart[colonIdx+1:]
	}
	lastSlash := strings.LastIndex(pathPart, "/")
	if lastSlash < 0 {
		return "", ""
	}
	lastSegment := pathPart[lastSlash+1:]
	if dotIdx := strings.Index(lastSegment, "."); dotIdx > 0 {
		fullDotIdx := lastSlash + 1 + dotIdx
		return domain + pathPart[:fullDotIdx], pathPart[fullDotIdx+1:]
	}
	return "", ""
}

func parseCallArgs(s string) ([]string, error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "(") || !strings.HasSuffix(s, ")") {
		return nil, fmt.Errorf("invalid call syntax: %s", s)
	}
	inner := strings.TrimSpace(s[1 : len(s)-1])
	if inner == "" {
		return nil, nil
	}

	var args []string
	var cur strings.Builder
	inQuote := false
	quoteChar := byte(0)
	hadQuote := false // track if current arg had quotes (empty string is valid)

	for i := 0; i < len(inner); i++ {
		ch := inner[i]
		if ch == '\\' && inQuote && i+1 < len(inner) {
			cur.WriteByte(ch)
			i++
			cur.WriteByte(inner[i])
			continue
		}
		if (ch == '"' || ch == '\'') && !inQuote {
			inQuote = true
			quoteChar = ch
			hadQuote = true
			continue
		}
		if inQuote && ch == quoteChar {
			inQuote = false
			continue
		}
		if ch == ',' && !inQuote {
			args = append(args, strings.TrimSpace(cur.String()))
			cur.Reset()
			hadQuote = false
			continue
		}
		cur.WriteByte(ch)
	}
	if inQuote {
		return nil, fmt.Errorf("unterminated string in arguments")
	}
	last := strings.TrimSpace(cur.String())
	if last != "" || hadQuote {
		args = append(args, last)
	}
	return args, nil
}
