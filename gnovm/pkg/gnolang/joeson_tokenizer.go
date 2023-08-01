package gnolang

// Imagine a .go source code in a file.
// We are interested in the
// operation of normalizing that source
// code, to simplify parsing (similar
// to what go does):
//
// +------+                 +------+
// | ===  |                 | ===; |
// | ==   |   tokenize      | ==;  |
// | ==   | --------------> | ==;  |
// +------+                 +------+
//   .go                tokenized source
// (lines)                   (tokens)
//

import (
	"errors"
	"fmt"
	goscanner "go/scanner"
	"go/token"
	"regexp"
	"strings"

	j "github.com/grepsuzette/joeson"
)

type scannerError struct {
	pos token.Position
	msg string
}

func (se scannerError) Error() string {
	return fmt.Sprintf("there was an error at %s: %s", se.pos.String(), se.msg)
}

var scannerErrors []error

func tokenize(source string) (*j.TokenStream, error) {
	var scan goscanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(source))
	scan.Init(file, []byte(source), ferror, 0 /*goscanner.ScanComments*/)
	if scan.ErrorCount > 0 {
		if scan.ErrorCount != len(scannerErrors) {
			panic("assert") // errors must have been collected
		}
		return nil, errors.Join(scannerErrors...)
	}
	tokens := []j.Token{}
	workOffset := 0
	prev := ""

	// Go lexer adds an automatic semicolon when the line's last token is:
	// * an identifier
	// * an integer, floating-point, imaginary, rune, or string literal
	// * one of the keywords break, continue, fallthrough, or return
	// * one of the operators and delimiters ++, --, ), ], or }
	var b strings.Builder
	mustInsertSpaceAfter := regexp.MustCompile("[a-zA-Z0-9_=]$")
	for {
		pos, tok, lit := scan.Scan()
		if tok == token.EOF {
			break
		}
		s := ""
		tokStr := tok.String()
		if tokStr == ";" && lit == "\n" {
			s = ";\n"
		} else if lit != "" {
			if mustInsertSpaceAfter.MatchString(prev) {
				s = " " + lit
			} else {
				s = lit
			}
		} else {
			// fmt.Printf("%s\t%s\t%q\n", fset.Position(pos), tok, lit)
			// switch tokStr {
			// case "(", ")", "[", "]":
			// 	s = tokStr
			// default:
			// 	s = tokStr + sep
			// }
			if mustInsertSpaceAfter.MatchString(prev) &&
				(tok.IsOperator() && tok == token.COMMA) {
				s = " " + tokStr
			} else {
				s = tokStr
			}
		}
		workOffset += len(prev)
		prev = s
		tokens = append(tokens, j.Token{s, int(pos), workOffset})
		b.WriteString(s)
	}
	return j.NewTokenStream(source, tokens), nil
}

func ferror(pos token.Position, msg string) {
	scannerErrors = append(scannerErrors, scannerError{pos, msg})
}
