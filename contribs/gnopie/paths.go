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
	PathFile                      // gno.land/r/foo/bar/file.gno
	PathAddress                   // g1abc...xyz (bech32 address)
	PathUser                      // gno.land/u/moul (gnoweb user URL)
)

type GnoPath struct {
	Raw        string
	Domain     string
	PkgPath    string
	Symbol     string
	File       string // e.g., "admin.gno" for file paths
	Address    string // e.g., "g1abc...xyz" for address paths
	RenderPath string // e.g., "p/monthly-dev-17" from gnoweb URL `:` separator
	Args       []string
	Kind       PathKind
}

func (p *GnoPath) IsPublic() bool {
	return p.Symbol != "" && p.Symbol[0] >= 'A' && p.Symbol[0] <= 'Z'
}

func ParsePath(input string) (*GnoPath, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty path")
	}

	// Strip https:// or http:// prefix (allow pasting gnoweb URLs)
	for _, prefix := range []string{"https://", "http://"} {
		if strings.HasPrefix(input, prefix) {
			input = input[len(prefix):]
			break
		}
	}

	// Extract URL fragment (#...) — may contain function name (e.g., #func-AdminSetAddr)
	var fragment string
	if hashIdx := strings.Index(input, "#"); hashIdx >= 0 {
		fragment = input[hashIdx+1:]
		input = input[:hashIdx]
	}

	// Strip trailing slash
	input = strings.TrimRight(input, "/")

	p := &GnoPath{Raw: input}

	// Detect gnoweb modifiers ($source, $help, $funcs)
	// e.g., gno.land/r/foo/bar$source or gno.land/r/foo/bar$source&file=admin.gno
	if dollarIdx := strings.Index(input, "$"); dollarIdx > 0 {
		modPart := input[dollarIdx+1:]
		input = input[:dollarIdx]
		p.Raw = input

		// Parse modifier and optional &key=value params
		modifier := modPart
		if ampIdx := strings.Index(modPart, "&"); ampIdx >= 0 {
			modifier = modPart[:ampIdx]
			params := modPart[ampIdx+1:]
			// Parse file= param
			for _, param := range strings.Split(params, "&") {
				if strings.HasPrefix(param, "file=") {
					p.File = param[5:]
				}
			}
		}

		switch modifier {
		case "source":
			p.RenderPath = "$source"
		case "help", "funcs":
			p.RenderPath = "$" + modifier
		}
	}

	// Handle #func-Name fragments (from $help URLs)
	if fragment != "" && strings.HasPrefix(fragment, "func-") {
		p.Symbol = fragment[5:] // strip "func-" prefix
	}

	// Detect bech32 addresses (g1...)
	if strings.HasPrefix(input, "g1") && !strings.Contains(input, "/") && !strings.Contains(input, ".") {
		p.Address = input
		p.Kind = PathAddress
		return p, nil
	}

	slashIdx := strings.Index(input, "/")
	if slashIdx < 0 {
		p.Domain = input
		p.Kind = PathNetwork
		return p, nil
	}

	p.Domain = input[:slashIdx]
	rest := input[slashIdx:]

	// Handle gnoweb user URLs: /u/moul or /u/g1abc...
	if strings.HasPrefix(rest, "/u/") {
		username := rest[3:]
		p.Symbol = username // store username in Symbol
		p.Kind = PathUser
		return p, nil
	}

	// Handle gnoweb render path separator (:)
	// e.g., /r/gnoland/blog:p/monthly-dev-17
	if colonIdx := strings.Index(rest, ":"); colonIdx > 0 {
		p.RenderPath = rest[colonIdx+1:]
		rest = rest[:colonIdx]
		p.PkgPath = p.Domain + rest
		p.Kind = PathPackage
		return p, nil
	}

	// Check for call expression: Func(args...)
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

	// Check for dot-separated symbol (Pkg.Symbol)
	pkgPath, symbol := splitSymbol(p.Domain, rest)
	if symbol != "" {
		p.PkgPath = pkgPath
		p.Symbol = symbol
		p.Kind = PathSymbol
		return p, nil
	}

	fullPath := p.Domain + rest

	// Check if last segment is a file
	if lastSlash := strings.LastIndex(rest, "/"); lastSlash > 0 {
		lastSeg := rest[lastSlash+1:]
		if isFileExtension(lastSeg) {
			p.PkgPath = p.Domain + rest[:lastSlash]
			p.File = lastSeg
			p.Kind = PathFile
			return p, nil
		}
	}

	p.PkgPath = fullPath
	parts := strings.Split(rest, "/")
	if len(parts) <= 3 {
		p.Kind = PathNamespace
	} else {
		p.Kind = PathPackage
	}
	return p, nil
}

// fileExtensions that should be treated as file paths, not symbol references.
var fileExtensions = []string{".gno", ".toml", ".md", ".txt", ".json"}

func isFileExtension(s string) bool {
	for _, ext := range fileExtensions {
		if strings.HasSuffix(s, ext) {
			return true
		}
	}
	return false
}

func splitSymbol(domain, pathPart string) (pkgPath, symbol string) {
	lastSlash := strings.LastIndex(pathPart, "/")
	if lastSlash < 0 {
		return "", ""
	}
	lastSegment := pathPart[lastSlash+1:]
	// Don't treat file extensions as symbol separators
	if isFileExtension(lastSegment) {
		return "", ""
	}
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
