// eval/int32 is a evaluator for int32 expressions.
// This code is heavily forked from https://github.com/dengsgo/math-engine
// which is licensed under Apache 2.0:
// https://raw.githubusercontent.com/dengsgo/math-engine/298e2b57b7e7350d0f67bd036916efd5709abe25/LICENSE
package int32

import (
	"errors"
	"strconv"
	"strings"

	"gno.land/p/demo/ufmt"
)

const (
	Identifier = iota
	Number     // numbers
	Operator   // +, -, *, /, etc.
	Variable   // x, y, z, etc. (one-letter only)
)

type expression interface {
	String() string
}

type expressionRaw struct {
	expression string
	Type       int
	Flag       int
	Offset     int
}

type parser struct {
	Input  string
	ch     byte
	offset int
	err    error
}

type expressionNumber struct {
	Val int
	Str string
}

type expressionVariable struct {
	Val int
	Str string
}

type expressionOperation struct {
	Op string
	Lhs,
	Rhs expression
}

type ast struct {
	rawexpressions    []*expressionRaw
	source            string
	currentexpression *expressionRaw
	currentIndex      int
	depth             int
	err               error
}

// Parse takes an expression string, e.g. "1+2" and returns
// a parsed expression. If there is an error it will return.
func Parse(s string) (ar expression, err error) {
	toks, err := lexer(s)
	if err != nil {
		return
	}
	ast, err := newAST(toks, s)
	if err != nil {
		return
	}
	ar, err = ast.parseExpression()
	return
}

// Eval takes a parsed expression and a map of variables (or nil). The parsed
// expression is evaluated using any variables and returns the
// resulting int and/or error.
func Eval(expr expression, variables map[string]int) (res int, err error) {
	if err != nil {
		return
	}
	var l, r int
	switch expr.(type) {
	case expressionVariable:
		ast := expr.(expressionVariable)
		ok := false
		if variables != nil {
			res, ok = variables[ast.Str]
		}
		if !ok {
			err = ufmt.Errorf("variable '%s' not found", ast.Str)
		}
		return
	case expressionOperation:
		ast := expr.(expressionOperation)
		l, err = Eval(ast.Lhs, variables)
		if err != nil {
			return
		}
		r, err = Eval(ast.Rhs, variables)
		if err != nil {
			return
		}
		switch ast.Op {
		case "+":
			res = l + r
		case "-":
			res = l - r
		case "*":
			res = l * r
		case "/":
			if r == 0 {
				err = ufmt.Errorf("violation of arithmetic specification: a division by zero in Eval: [%d/%d]", l, r)
				return
			}
			res = l / r
		case "%":
			if r == 0 {
				res = 0
			} else {
				res = l % r
			}
		case "^":
			res = l ^ r
		case ">>":
			res = l >> r
		case "<<":
			res = l << r
		case ">":
			if l > r {
				res = 1
			} else {
				res = 0
			}
		case "<":
			if l < r {
				res = 1
			} else {
				res = 0
			}
		case "&":
			res = l & r
		case "|":
			res = l | r
		default:

		}
	case expressionNumber:
		res = expr.(expressionNumber).Val
	}

	return
}

func expressionError(s string, pos int) string {
	r := strings.Repeat("-", len(s)) + "\n"
	s += "\n"
	for i := 0; i < pos; i++ {
		s += " "
	}
	s += "^\n"
	return r + s + r
}

func (n expressionVariable) String() string {
	return ufmt.Sprintf(
		"expressionVariable: %s",
		n.Str,
	)
}

func (n expressionNumber) String() string {
	return ufmt.Sprintf(
		"expressionNumber: %s",
		n.Str,
	)
}

func (b expressionOperation) String() string {
	return ufmt.Sprintf(
		"expressionOperation: (%s %s %s)",
		b.Op,
		b.Lhs.String(),
		b.Rhs.String(),
	)
}

func newAST(toks []*expressionRaw, s string) (*ast, error) {
	a := &ast{
		rawexpressions: toks,
		source:         s,
	}
	if a.rawexpressions == nil || len(a.rawexpressions) == 0 {
		return a, errors.New("empty token")
	} else {
		a.currentIndex = 0
		a.currentexpression = a.rawexpressions[0]
	}
	return a, nil
}

func (a *ast) parseExpression() (expression, error) {
	a.depth++ // called depth
	lhs := a.parsePrimary()
	r := a.parseBinOpRHS(0, lhs)
	a.depth--
	if a.depth == 0 && a.currentIndex != len(a.rawexpressions) && a.err == nil {
		return r, ufmt.Errorf("bad expression, reaching the end or missing the operator\n%s",
			expressionError(a.source, a.currentexpression.Offset))
	}
	return r, nil
}

func (a *ast) getNextexpressionRaw() *expressionRaw {
	a.currentIndex++
	if a.currentIndex < len(a.rawexpressions) {
		a.currentexpression = a.rawexpressions[a.currentIndex]
		return a.currentexpression
	}
	return nil
}

func (a *ast) getTokPrecedence() int {
	switch a.currentexpression.expression {
	case "/", "%", "*":
		return 100
	case "<<", ">>":
		return 80
	case "+", "-":
		return 75
	case "<", ">":
		return 70
	case "&":
		return 60
	case "^":
		return 50
	case "|":
		return 40
	}
	return -1
}

func (a *ast) parseNumber() expressionNumber {
	f64, err := strconv.Atoi(a.currentexpression.expression)
	if err != nil {
		a.err = ufmt.Errorf("%v\nwant '(' or '0-9' but get '%s'\n%s",
			err.Error(),
			a.currentexpression.expression,
			expressionError(a.source, a.currentexpression.Offset))
		return expressionNumber{}
	}
	n := expressionNumber{
		Val: f64,
		Str: a.currentexpression.expression,
	}
	a.getNextexpressionRaw()
	return n
}

func (a *ast) parseVariable() expressionVariable {
	n := expressionVariable{
		Val: 0,
		Str: a.currentexpression.expression,
	}
	a.getNextexpressionRaw()
	return n
}

func (a *ast) parsePrimary() expression {
	switch a.currentexpression.Type {
	case Variable:
		return a.parseVariable()
	case Number:
		return a.parseNumber()
	case Operator:
		if a.currentexpression.expression == "(" {
			t := a.getNextexpressionRaw()
			if t == nil {
				a.err = ufmt.Errorf("want '(' or '0-9' but get EOF\n%s",
					expressionError(a.source, a.currentexpression.Offset))
				return nil
			}
			e, _ := a.parseExpression()
			if e == nil {
				return nil
			}
			if a.currentexpression.expression != ")" {
				a.err = ufmt.Errorf("want ')' but get %s\n%s",
					a.currentexpression.expression,
					expressionError(a.source, a.currentexpression.Offset))
				return nil
			}
			a.getNextexpressionRaw()
			return e
		} else if a.currentexpression.expression == "-" {
			if a.getNextexpressionRaw() == nil {
				a.err = ufmt.Errorf("want '0-9' but get '-'\n%s",
					expressionError(a.source, a.currentexpression.Offset))
				return nil
			}
			bin := expressionOperation{
				Op:  "-",
				Lhs: expressionNumber{},
				Rhs: a.parsePrimary(),
			}
			return bin
		} else {
			return a.parseNumber()
		}
	default:
		return nil
	}
}

func (a *ast) parseBinOpRHS(execPrec int, lhs expression) expression {
	for {
		tokPrec := a.getTokPrecedence()
		if tokPrec < execPrec {
			return lhs
		}
		binOp := a.currentexpression.expression
		if a.getNextexpressionRaw() == nil {
			a.err = ufmt.Errorf("want '(' or '0-9' but get EOF\n%s",
				expressionError(a.source, a.currentexpression.Offset))
			return nil
		}
		rhs := a.parsePrimary()
		if rhs == nil {
			return nil
		}
		nextPrec := a.getTokPrecedence()
		if tokPrec < nextPrec {
			rhs = a.parseBinOpRHS(tokPrec+1, rhs)
			if rhs == nil {
				return nil
			}
		}
		lhs = expressionOperation{
			Op:  binOp,
			Lhs: lhs,
			Rhs: rhs,
		}
	}
}

func lexer(s string) ([]*expressionRaw, error) {
	p := &parser{
		Input: s,
		err:   nil,
		ch:    s[0],
	}
	toks := p.parse()
	if p.err != nil {
		return nil, p.err
	}
	return toks, nil
}

func (p *parser) parse() []*expressionRaw {
	toks := make([]*expressionRaw, 0)
	for {
		tok := p.nextTok()
		if tok == nil {
			break
		}
		toks = append(toks, tok)
	}
	return toks
}

func (p *parser) nextTok() *expressionRaw {
	if p.offset >= len(p.Input) || p.err != nil {
		return nil
	}
	var err error
	for p.isWhitespace(p.ch) && err == nil {
		err = p.nextCh()
	}
	start := p.offset
	var tok *expressionRaw
	switch p.ch {
	case
		'(',
		')',
		'+',
		'-',
		'*',
		'/',
		'^',
		'&',
		'|',
		'%':
		tok = &expressionRaw{
			expression: string(p.ch),
			Type:       Operator,
		}
		tok.Offset = start
		err = p.nextCh()
	case '>', '<':
		tokS := string(p.ch)
		bb, be := p.nextChPeek()
		if be == nil && string(bb) == tokS {
			tokS += string(p.ch)
		}
		tok = &expressionRaw{
			expression: tokS,
			Type:       Operator,
		}
		tok.Offset = start
		if len(tokS) > 1 {
			p.nextCh()
		}
		err = p.nextCh()
	case
		'0',
		'1',
		'2',
		'3',
		'4',
		'5',
		'6',
		'7',
		'8',
		'9':
		for p.isDigitNum(p.ch) && p.nextCh() == nil {
			if (p.ch == '-' || p.ch == '+') && p.Input[p.offset-1] != 'e' {
				break
			}
		}
		tok = &expressionRaw{
			expression: strings.ReplaceAll(p.Input[start:p.offset], "_", ""),
			Type:       Number,
		}
		tok.Offset = start
	default:
		if p.isChar(p.ch) {
			tok = &expressionRaw{
				expression: string(p.ch),
				Type:       Variable,
			}
			tok.Offset = start
			err = p.nextCh()
		} else if p.ch != ' ' {
			p.err = ufmt.Errorf("symbol error: unknown '%v', pos [%v:]\n%s",
				string(p.ch),
				start,
				expressionError(p.Input, start))
		}
	}
	return tok
}

func (p *parser) nextChPeek() (byte, error) {
	offset := p.offset + 1
	if offset < len(p.Input) {
		return p.Input[offset], nil
	}
	return byte(0), errors.New("no byte")
}

func (p *parser) nextCh() error {
	p.offset++
	if p.offset < len(p.Input) {
		p.ch = p.Input[p.offset]
		return nil
	}
	return errors.New("EOF")
}

func (p *parser) isWhitespace(c byte) bool {
	return c == ' ' ||
		c == '\t' ||
		c == '\n' ||
		c == '\v' ||
		c == '\f' ||
		c == '\r'
}

func (p *parser) isDigitNum(c byte) bool {
	return '0' <= c && c <= '9' || c == '.' || c == '_' || c == 'e' || c == '-' || c == '+'
}

func (p *parser) isChar(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

func (p *parser) isWordChar(c byte) bool {
	return p.isChar(c) || '0' <= c && c <= '9'
}
