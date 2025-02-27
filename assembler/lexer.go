package assembler

import (
	"strconv"
	"strings"
	"unicode"
)

type Lexer struct {
	input        string
	position     uint
	readPosition uint
	ch           byte
	line         uint
	columnn      uint
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:   input,
		line:    1,
		columnn: 1,
	}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() Token {
	var tok Token
	l.skipWhiteSpace()
	startColumn := l.columnn
	switch l.ch {
	case ':':
		tok = newToken(COLON, string(l.ch), l.line, l.columnn)
	case ';':
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
		return l.NextToken()
	case '(':
		tok = newToken(LPAREN, string(l.ch), l.line, l.columnn)
	case ')':
		tok = newToken(RPAREN, string(l.ch), l.line, l.columnn)
	case '{':
		tok = newToken(LBRACE, string(l.ch), l.line, l.columnn)
	case '}':
		tok = newToken(RBRACE, string(l.ch), l.line, l.columnn)
	case ',':
		tok = newToken(COMMA, string(l.ch), l.line, l.columnn)
	case '[':
		tok = newToken(LBRACKET, string(l.ch), l.line, l.columnn)
	case ']':
		tok = newToken(RBRACKET, string(l.ch), l.line, l.columnn)
	case '-':
		if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = newToken(ARROW, string(ch)+string(l.ch), l.line, startColumn)
		} else {
			tok = newToken(ILLEGAL, string(l.ch), l.line, startColumn)
		}
	case '"':
		tok = newToken(STRING, l.readString(), l.line, l.columnn)
	case 0:
		tok = newToken(EOF, "", l.line, l.columnn)
	default:
		if isLetter(l.ch) || l.ch == '.' {
			tok.Line = l.line
			tok.Column = l.columnn
			tok.Literal = l.readIdentifier()
			if instr, ok := instructions[tok.Literal]; ok {
				tok.Type = instr
			} else if keyword, ok := keywords[tok.Literal]; ok {
				tok.Type = keyword
			} else {
				tok.Type = IDENT
			}
			return tok
		} else if isDigit(l.ch) {
			return l.readNumber()
		} else {
			tok = newToken(ILLEGAL, string(l.ch), l.line, l.columnn)
		}
	}
	l.readChar()
	return tok
}

func (l *Lexer) readChar() {
	if l.readPosition >= uint(len(l.input)) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	if l.ch != '\n' {
		l.columnn++
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= uint(len(l.input)) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) skipWhiteSpace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' {
			l.line++
			l.columnn = 0
		}
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	pos := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' || l.ch == '.' {
		l.readChar()
	}
	return l.input[pos:l.position]
}

func (l *Lexer) readString() string {
	var result strings.Builder
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}

		if l.ch == '\\' {
			next := l.peekChar()
			switch next {
			case 'n':
				result.WriteByte('\n')
				l.readChar()
			case 't':
				result.WriteByte('\t')
				l.readChar()
			case '\\':
				result.WriteByte('\\')
				l.readChar()
			case '"':
				result.WriteByte('"')
				l.readChar()
			default:
				result.WriteByte('\\')
			}
		} else {
			result.WriteByte(l.ch)
		}
	}
	return result.String()
}

func (l *Lexer) readNumber() Token {
	pos := l.position
	isFloat := false
	for isDigit(l.ch) || l.ch == '.' {
		if l.ch == '.' {
			if isFloat {
				return Token{Type: ILLEGAL, Literal: "Invalid number format"}
			}
			isFloat = true
		}
		l.readChar()
	}
	numStr := l.input[pos:l.position]
	if isFloat {
		if _, err := strconv.ParseFloat(numStr, 32); err != nil {
			return Token{Type: ILLEGAL, Literal: "Invalid float format"}
		}
		return Token{Type: FLOAT, Literal: numStr}
	}
	if _, err := strconv.ParseInt(numStr, 10, 32); err != nil {
		return Token{Type: ILLEGAL, Literal: "Invalid integer format"}
	}
	return Token{Type: INT, Literal: numStr}
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch))
}

func isDigit(ch byte) bool {
	return unicode.IsDigit(rune(ch))
}
