package mathml

import (
	"errors"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

type TokenKind int
type lexerState int

const (
	lxBegin lexerState = iota
	lxEnd
	lxContinue
	lxSpace
	lxWasBackslash
	lxCommand
	lxNumber
	lxFence
	lxComment
	lxMacroArg
)
const (
	tokWhitespace TokenKind = 1 << iota
	tokComment
	tokCommand
	tokEscaped
	tokNumber
	tokLetter
	tokChar
	tokOpen
	tokClose
	tokMiddle
	tokCurly
	tokEnv
	tokFence
	tokSubsup
	tokMacroarg
	tokBadmacro
	tokReserved
	tokBigness1
	tokBigness2
	tokBigness3
	tokBigness4
	tokInfix
	tokStarSuffix
	tokNull = 0
)

var (
	brace_match_map = map[string]string{
		"(": ")",
		"{": "}",
		"[": "]",
		")": "(",
		"}": "{",
		"]": "[",
	}
	char_open     = []rune("([{")
	char_close    = []rune(")]}")
	char_reserved = []rune(`#$%^&_{}~\`)
)

func init() {
	logger = log.New(os.Stderr, "MathML: ", log.LstdFlags)
}

type Token struct {
	Kind        TokenKind
	MatchOffset int // offset from current index to matching paren, brace, etc.
	Value       string
	start       int // the index of the beginning of this token in the untokenized rune slice
	end         int // the index immediately after the end of this token in the untokenized rune slice
}

func getToken(tex []rune, start int) (Token, int) {
	var state lexerState
	var kind TokenKind
	// A capacity of 24 is reasonable. Most commands, numbers, etc are not more than 24 chars in length, and setting
	// this capacity grants a huge speedup by avoiding extra allocations.
	result := make([]rune, 0, 24)
	var idx int
	for idx = start; idx < len(tex); idx++ {
		r := tex[idx]
		switch state {
		case lxEnd:
			return Token{Kind: kind, Value: string(result), start: start, end: idx}, idx
		case lxBegin:
			switch {
			case unicode.IsLetter(r):
				state = lxEnd
				kind = tokLetter
				result = append(result, r)
			case unicode.IsNumber(r):
				state = lxNumber
				kind = tokNumber
				result = append(result, r)
			case r == '\\':
				state = lxWasBackslash
			case r == '{':
				state = lxEnd
				kind = tokCurly | tokOpen
				result = append(result, r)
			case r == '}':
				state = lxEnd
				kind = tokCurly | tokClose
				result = append(result, r)
			case slices.Contains(char_open, r):
				state = lxEnd
				kind = tokOpen
				result = append(result, r)
			case slices.Contains(char_close, r):
				state = lxEnd
				kind = tokClose
				result = append(result, r)
			case r == '^' || r == '_':
				state = lxEnd
				kind = tokSubsup
				result = append(result, r)
			case r == '%':
				state = lxComment
				kind = tokComment
			case r == '#':
				state = lxMacroArg
				kind = tokMacroarg
			case slices.Contains(char_reserved, r):
				state = lxEnd
				kind = tokReserved
				result = append(result, r)
			case unicode.IsSpace(r):
				state = lxSpace
				kind = tokWhitespace
				result = append(result, ' ')
			case r == '|':
				state = lxEnd
				kind = tokLetter
				result = append(result, r)
			default:
				state = lxEnd
				kind = tokChar
				result = append(result, r)
			}
		case lxComment:
			switch r {
			case '\n':
				state = lxEnd
				result = append(result, r)
			default:
				result = append(result, r)
			}
		case lxSpace:
			switch {
			case !unicode.IsSpace(r):
				return Token{Kind: kind, Value: string(result), start: start, end: idx}, idx
			}
		case lxNumber:
			switch {
			case r == '.':
				if idx < len(tex)-1 && !unicode.IsNumber(tex[idx+1]) {
					return Token{Kind: kind, Value: string(result), start: start, end: idx}, idx
				}
				result = append(result, r)
			case !unicode.IsNumber(r):
				return Token{Kind: kind, Value: string(result), start: start, end: idx}, idx
			default:
				result = append(result, r)
			}
		case lxMacroArg:
			if unicode.IsNumber(r) {
				result = append(result, r)
				state = lxEnd
			} else {
				return Token{Kind: kind, Value: "#", start: start, end: idx}, idx
			}
		case lxWasBackslash:
			switch {
			case r == '|':
				state = lxEnd
				kind = tokFence | tokEscaped
				result = append(result, r)
			case slices.Contains(char_open, r):
				state = lxEnd
				kind = tokOpen | tokEscaped | tokFence
				result = append(result, r)
			case slices.Contains(char_close, r):
				state = lxEnd
				kind = tokClose | tokEscaped | tokFence
				result = append(result, r)
			case slices.Contains(char_reserved, r):
				state = lxEnd
				kind = tokChar | tokEscaped
				result = append(result, r)
			case unicode.IsSpace(r):
				state = lxEnd
				kind = tokCommand
				result = append(result, ' ')
			case unicode.IsLetter(r):
				state = lxCommand
				kind = tokCommand
				result = append(result, r)
			default:
				state = lxEnd
				kind = tokCommand
				result = append(result, r)
			}
		case lxCommand:
			switch {
			case r == '*': // the asterisk should only occur at the end of a command.
				state = lxEnd
				kind |= tokStarSuffix
			case !unicode.IsLetter(r):
				val := string(result)
				return Token{Kind: kind, Value: val, start: start, end: idx}, idx
			default:
				result = append(result, r)
			}
		}
	}
	return Token{Kind: kind, Value: string(result)}, idx
}

type ExprKind int

const (
	expr_single_tok ExprKind = 1 << iota
	expr_options
	expr_fenced
	expr_group
	expr_environment
	expr_whitespace
)

// Get the next single token or expression enclosed in brackets. Return the index immediately after the end of the
// returned expression. Example:
//
//	\frac{a^2+b^2}{c+d}
//	     │╰──┬──╯╰─ final position returned
//	     │   ╰───── slice of tokens returned
//	     ╰───────── idx (initial position)
func GetNextExpr(tokens []Token, idx int) ([]Token, int, ExprKind) {
	var result []Token
	kind := expr_single_tok
	// an expression may contain whitespace, but never start with whitespace
	for idx < len(tokens) && tokens[idx].Kind&(tokComment|tokWhitespace) > 0 {
		idx++
	}
	if idx >= len(tokens) {
		return nil, idx, kind
	}
	if tokens[idx].MatchOffset > 0 && tokens[idx].Kind&tokEscaped == 0 {

		switch tokens[idx].Value {
		case "{":
			kind = expr_group
		case "[":
			kind = expr_options
		default:
			kind = expr_fenced
		}
		end := idx + tokens[idx].MatchOffset
		result = tokens[idx+1 : end]
		idx = end
	} else {
		result = []Token{tokens[idx]}
	}
	return result, idx, kind
}

type TokenBuffer struct {
	Expr []Token // The current sub-expression
	idx  int     // The index in the current sub-expression
	jump int
}

type TokenBufferErr struct {
	code int
	err  error
}

const (
	tbEndErr = iota + 1
	tbIsExprErr
	tbIsSingleErr
)

var (
	ErrTokenBufferEnd    = errors.New("end of TokenBuffer")
	ErrTokenBufferExpr   = errors.New("expected token, got expression")
	ErrTokenBufferSingle = errors.New("expected expression, got token")
)

// Error the TokenBufferErr
func (e *TokenBufferErr) Error() string {
	return e.err.Error()
}

// Unwrap the TokenBufferErr
func (e *TokenBufferErr) Unwrap() error {
	return e.err
}

// Create a new TokenBuffer
func NewTokenBuffer(t []Token) *TokenBuffer {
	return &TokenBuffer{Expr: t, idx: 0}
}

// Check if the TokenBuffer is empty
func (b *TokenBuffer) Empty() bool {
	if b.idx >= len(b.Expr) {
		return true
	}
	temp := b.idx
	// an expression may contain whitespace, but never start with whitespace
	for b.idx < len(b.Expr) && b.Expr[b.idx].Kind&(tokComment|tokWhitespace) > 0 {
		b.idx++
	}
	if b.idx >= len(b.Expr) {
		return true
	}
	b.idx = temp
	return false
}

// Advance the TokenBuffer by one token
func (b *TokenBuffer) Advance() {
	b.idx++
	b.jump = 1
}

func (b *TokenBuffer) GetNextToken(skipWhitespace ...bool) (Token, error) {
	var result Token
	temp := b.idx
	for b.idx < len(b.Expr) && b.Expr[b.idx].Kind&tokComment > 0 {
		b.idx++
	}
	if skipWhitespace == nil || skipWhitespace[0] {
		for b.idx < len(b.Expr) && b.Expr[b.idx].Kind&tokWhitespace > 0 {
			b.idx++
		}
	}
	if b.idx >= len(b.Expr) {
		b.idx = temp
		return result, &TokenBufferErr{tbEndErr, ErrTokenBufferEnd}
	}
	if b.Expr[b.idx].Kind&(tokEscaped|tokCurly|tokOpen) == (tokCurly | tokOpen) {
		b.idx = temp
		return result, &TokenBufferErr{tbIsExprErr, ErrTokenBufferExpr}
	}
	result = b.Expr[b.idx]
	b.idx++
	b.jump = b.idx - temp
	return result, nil
}

// Get the next expression enclosed in {curly braces}
func (b *TokenBuffer) GetNextExpr() (*TokenBuffer, error) {
	temp := b.idx
	var result *TokenBuffer
	// an expression may contain whitespace, but never start with whitespace
	for b.idx < len(b.Expr) && b.Expr[b.idx].Kind&(tokComment|tokWhitespace) > 0 {
		b.idx++
	}
	if b.idx >= len(b.Expr) {
		b.idx = temp
		return nil, &TokenBufferErr{tbEndErr, ErrTokenBufferEnd}
	}
	if b.Expr[b.idx].MatchOffset > 0 && b.Expr[b.idx].Kind&tokEscaped == 0 && b.Expr[b.idx].Value == "{" {
		end := b.idx + b.Expr[b.idx].MatchOffset
		result = NewTokenBuffer(b.Expr[b.idx+1 : end])
		b.idx = end + 1
	} else {
		b.idx = temp
		return nil, &TokenBufferErr{tbIsSingleErr, ErrTokenBufferSingle}
	}
	b.jump = b.idx - temp
	return result, nil
}

// Extract the tokens strictly between [square brackets]. skipWhitespace is true by default.
func (b *TokenBuffer) GetOptions(skipWhitespace ...bool) (*TokenBuffer, error) {
	temp := b.idx
	var result *TokenBuffer
	// an expression may contain whitespace, but never start with whitespace
	if len(skipWhitespace) < 1 || skipWhitespace[0] {
		for b.idx < len(b.Expr) && b.Expr[b.idx].Kind&(tokComment|tokWhitespace) > 0 {
			b.idx++
		}
	}
	if b.idx >= len(b.Expr) {
		b.idx = temp
		return nil, &TokenBufferErr{tbEndErr, ErrTokenBufferEnd}
	}
	if b.Expr[b.idx].MatchOffset > 0 && b.Expr[b.idx].Kind&tokEscaped == 0 && b.Expr[b.idx].Value == "[" {
		end := b.idx + b.Expr[b.idx].MatchOffset
		result = NewTokenBuffer(b.Expr[b.idx+1 : end])
		b.idx = end + 1 // Don't parse closing "]"
	} else {
		b.idx = temp
		return nil, &TokenBufferErr{}
	}
	b.jump = b.idx - temp
	return result, nil
}

// Get tokens until (but not including) the condition f evaluates as true, or the end of the token buffer is reached
func (b *TokenBuffer) GetUntil(f func(Token) bool) *TokenBuffer {
	start := b.idx
	for b.idx < len(b.Expr) && !f(b.Expr[b.idx]) {
		b.idx++
	}
	b.jump = b.idx - start
	return NewTokenBuffer(b.Expr[start:b.idx])
}

// Get the next n tokens
func (b *TokenBuffer) GetNextN(n int, skipWhitespace ...bool) (*TokenBuffer, error) {
	if b.idx+n > len(b.Expr) {
		return NewTokenBuffer(b.Expr[b.idx:len(b.Expr)]), &TokenBufferErr{tbEndErr, ErrTokenBufferEnd}
	}
	for b.idx < len(b.Expr) && b.Expr[b.idx].Kind&tokComment > 0 {
		b.idx++
	}
	start := b.idx
	if skipWhitespace != nil && skipWhitespace[0] {
		for b.idx < len(b.Expr) && b.Expr[b.idx].Kind&(tokComment|tokWhitespace) > 0 {
			b.idx++
		}
		start = b.idx
	}
	b.idx += n
	b.jump = n
	return NewTokenBuffer(b.Expr[start:b.idx]), nil
}

// Unget the last n tokens
func (b *TokenBuffer) Unget() {
	b.idx -= b.jump
}

// Tokenize the tex string
func tokenize(tex []rune) ([]Token, error) {
	var tok Token
	tokens := make([]Token, 0)
	idx := 0
	for idx < len(tex) {
		tok, idx = getToken(tex, idx)
		switch tok.Value {
		case "over", "choose", "atop":
			tok.Kind |= tokInfix
		}
		tokens = append(tokens, tok)
	}
	return postProcessTokens(tokens)
}

// Stringify the tokens
func StringifyTokens(toks []Token) string {
	var sb strings.Builder
	for _, t := range toks {
		sb.WriteString(t.Value)
	}
	return sb.String()
}
func stringifyTokensHtml(toks []Token) string {
	var sb strings.Builder
	for _, t := range toks {
		if t.Value == " " {
			sb.WriteString("&nbsp;")
		} else {
			sb.WriteString(t.Value)
		}
	}
	return sb.String()
}

// MismatchedBraceError is an error that occurs when a brace is mismatched
type MismatchedBraceError struct {
	kind    string
	context string
	pos     int
}

// Create a new MismatchedBraceError
func newMismatchedBraceError(kind string, context string, pos int) MismatchedBraceError {
	return MismatchedBraceError{kind, context, pos}
}

// Error the MismatchedBraceError
func (e MismatchedBraceError) Error() string {
	var sb strings.Builder
	sb.WriteString("mismatched ")
	sb.WriteString(e.kind)
	sb.WriteString(" at position ")
	sb.WriteString(strconv.FormatInt(int64(e.pos), 10))
	if e.context != "" {
		sb.WriteString(e.context)
	}
	return sb.String()
}

// ErrorContext is a helper function to create a context string for an error
func errorContext(t Token, context string) string {
	var sb strings.Builder
	sb.WriteRune('\n')
	sb.WriteString(context)
	sb.WriteRune('\n')
	toklen := len(t.Value)
	if len(context)-toklen <= 4 {
		sb.WriteString(strings.Repeat(" ", max(0, len(context)-toklen)))
		sb.WriteString(strings.Repeat("^", toklen))
		sb.WriteString("HERE")
	} else {
		sb.WriteString(strings.Repeat(" ", max(0, len(context)-toklen-4)))
		sb.WriteString("HERE")
		sb.WriteString(strings.Repeat("^", toklen))
	}
	sb.WriteRune('\n')
	return sb.String()
}

// Find matching {curly braces}
func matchBracesCritical(tokens []Token, kind TokenKind) error {
	s := newStack[int]()
	contextLength := 16
	for i, t := range tokens {
		if t.Kind&(tokOpen|kind) == tokOpen|kind {
			s.Push(i)
		} else if t.Kind&(tokClose|kind) == tokClose|kind {
			if s.empty() {
				var k string
				if t.Kind&tokCurly > 0 {
					k = "curly brace"
				}
				if t.Kind&tokEnv > 0 {
					k = "environment (" + t.Value + ")"
				}
				context := errorContext(t, StringifyTokens(tokens[max(0, i-contextLength):min(i+contextLength, len(tokens))]))
				return newMismatchedBraceError(k, "<pre>"+context+"</pre>", i)
			}
			mate := tokens[s.Peek()]
			if kind == tokEnv && mate.Value != t.Value {
				context := errorContext(t, StringifyTokens(tokens[max(0, i-contextLength):min(i+contextLength, len(tokens))]))
				return newMismatchedBraceError("environment ("+mate.Value+")", "<pre>"+context+"</pre>", i)
			}
			if (mate.Kind&t.Kind)&kind > 0 {
				pos := s.Pop()
				tokens[i].MatchOffset = pos - i
				tokens[pos].MatchOffset = i - pos
			}
		}
	}
	if !s.empty() {
		pos := s.Pop()
		t := tokens[pos]
		var kind string
		if t.Kind&tokCurly > 0 {
			kind = "curly brace"
		}
		if t.Kind&tokEnv > 0 {
			kind = "environment (" + t.Value + ")"
		}
		context := errorContext(t, StringifyTokens(tokens[max(0, pos-contextLength):min(pos+contextLength, len(tokens))]))
		return newMismatchedBraceError(kind, "<pre>"+context+"</pre>", pos)
	}
	return nil
}

// MatchBracesLazy is a helper function to match braces lazily
func matchBracesLazy(tokens []Token) {
	s := newStack[int]()
	for i, t := range tokens {
		if t.MatchOffset != 0 {
			// Critical regions have already been taken care of.
			continue
		}
		if t.Kind&tokOpen > 0 {
			s.Push(i)
			continue
		}
		if t.Kind&tokClose > 0 {
			if s.empty() {
				continue
			}
			mate := tokens[s.Peek()]
			if (t.Kind&mate.Kind)&tokFence > 0 || brace_match_map[mate.Value] == t.Value {
				pos := s.Pop()
				tokens[i].MatchOffset = pos - i
				tokens[pos].MatchOffset = i - pos
			}
		}
	}
}

// FixFences is a helper function to fix fences
func fixFences(toks []Token) []Token {
	out := make([]Token, 0, len(toks))
	var i int
	var temp Token
	bigLevel := func(s string) TokenKind {
		switch s {
		case "big":
			return tokBigness1
		case "Big":
			return tokBigness2
		case "bigg":
			return tokBigness3
		case "Bigg":
			return tokBigness4
		}
		return tokNull
	}
	for i < len(toks) {
		if i == len(toks)-1 {
			out = append(out, toks[i])
			break
		}
		temp = toks[i]
		nextval := toks[i+1].Value

		switch val := toks[i].Value; val {
		case "left":
			i++
			temp = toks[i]
			if nextval == "." {
				temp.Value = ""
				temp.Kind = tokNull
			} else {
				temp.Value = nextval
			}
			temp.Kind |= tokFence | tokOpen
			temp.Kind &= ^(tokMiddle | tokClose)
		case "middle":
			i++
			temp = toks[i]
			if nextval == "." {
				temp.Value = ""
				temp.Kind = tokNull
			} else {
				temp.Value = nextval
			}
			temp.Kind |= tokFence | tokMiddle
			temp.Kind &= ^(tokOpen | tokClose)
		case "right":
			i++
			temp = toks[i]
			if nextval == "." {
				temp.Value = ""
				temp.Kind = tokNull
			} else {
				temp.Value = nextval
			}
			temp.Kind |= tokFence | tokClose
			temp.Kind &= ^(tokOpen | tokMiddle)
		case "big", "Big", "bigg", "Bigg":
			i++
			temp = toks[i]
			temp.Kind |= bigLevel(val)
			temp.Kind &= ^(tokOpen | tokClose | tokFence)
		case "bigl", "Bigl", "biggl", "Biggl":
			i++
			temp = toks[i]
			temp.Kind |= tokOpen | bigLevel(val[:len(val)-1])
			temp.Kind &= ^tokFence
		case "bigr", "Bigr", "biggr", "Biggr":
			i++
			temp = toks[i]
			temp.Kind |= tokClose | bigLevel(val[:len(val)-1])
			temp.Kind &= ^tokFence
		}
		out = append(out, temp)
		i++
	}
	return out
}

// PostProcessTokens is a helper function to post-process the tokens
func postProcessTokens(toks []Token) ([]Token, error) {
	toks = fixFences(toks)
	err := matchBracesCritical(toks, tokCurly)
	if err != nil {
		return toks, err
	}
	out := make([]Token, 0, len(toks))
	var i int
	var temp Token
	var name []Token
	for i < len(toks) {
		temp = toks[i]
		temp.MatchOffset = 0
		switch toks[i].Value {
		case "begin":
			name, i, _ = GetNextExpr(toks, i+1)
			temp.Value = StringifyTokens(name)
			temp.Kind = tokEnv | tokOpen
		case "end":
			name, i, _ = GetNextExpr(toks, i+1)
			temp.Value = StringifyTokens(name)
			temp.Kind = tokEnv | tokClose
		}
		out = append(out, temp)
		i++
	}
	err = matchBracesCritical(out, tokEnv)
	if err != nil {
		return out, err
	}
	// Indicies could have changed after processing environments!!
	err = matchBracesCritical(out, tokCurly)
	if err != nil {
		return out, err
	}
	matchBracesLazy(out)
	return out, nil
}
