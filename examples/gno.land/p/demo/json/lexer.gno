package json

type Lexer struct {
	data  []byte
	token byte
	pos   int
	depth int
}

func New(data []byte) *Lexer {
	return &Lexer{
		data: data,
	}
}

func (l *Lexer) findTokenStart(token byte) int {
	for i := len(l.data) - 1; i >= 0; i-- {
		switch l.data[i] {
		case token:
			return i
		case '[', '{':
			l.depth++
			return 0
		}
	}

	return 0
}

func (l *Lexer) tokenEnd() int {
	for i, tok := range l.data {
		switch tok {
		case ' ', '\n', '\r', '\t', ',', '}', ']':
			return i
		}
	}

	return len(l.data)
}

func (l *Lexer) tokenStart() int {
	for i := len(l.data) - 1; i >= 0; i-- {
		switch l.data[i] {
		case '\n', '\r', '\t', ',', '{', '[':
			return i
		}
	}

	return 0
}

func (l *Lexer) nextToken() int {
	for i, b := range l.data {
		switch b {
		case '{', '[':
			l.depth++
			continue
		case '}', ']':
			l.depth--
			continue
		default:
			return i
		}
	}

	return -1
}

func (l *Lexer) findLastTokenPosition() int {
	for i := len(l.data) - 1; i >= 0; i-- {
		switch l.data[i] {
		case ' ', '\n', '\r', '\t':
			continue
		default:
			return i
		}
	}

	return -1
}
