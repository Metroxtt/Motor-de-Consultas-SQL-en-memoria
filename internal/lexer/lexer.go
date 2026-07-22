package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

var keywords = map[string]TokenType{
	"SELECT": TokenSelect, "FROM": TokenFrom,
	"WHERE": TokenWhere, "AND": TokenAnd, "OR": TokenOr,
	"GROUP": TokenGroup, "BY": TokenBy,
}

type Lexer struct {
	input  string
	pos    int
	line   int
	col    int
	errors []string
}

func New(input string) *Lexer {
	return &Lexer{input: input, line: 1, col: 1}
}

func (l *Lexer) Errors() []string {
	return l.errors
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return rune(l.input[l.pos])
}

func (l *Lexer) advance() rune {
	ch := rune(l.input[l.pos])
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		ch := rune(l.input[l.pos])
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			l.advance()
		} else {
			break
		}
	}
}

func (l *Lexer) readIdent() string {
	start := l.pos - 1
	for l.pos < len(l.input) {
		ch := rune(l.input[l.pos])
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
			l.advance()
		} else {
			break
		}
	}
	return l.input[start:l.pos]
}

func (l *Lexer) readNumber() string {
	start := l.pos - 1
	hasDot := false
	for l.pos < len(l.input) {
		ch := rune(l.input[l.pos])
		if unicode.IsDigit(ch) {
			l.advance()
		} else if ch == '.' && !hasDot {
			hasDot = true
			l.advance()
		} else {
			break
		}
	}
	return l.input[start:l.pos]
}

func (l *Lexer) readString() string {
	start := l.pos
	for l.pos < len(l.input) {
		ch := l.advance()
		if ch == '\'' {
			return l.input[start : l.pos-1]
		}
	}
	l.errors = append(l.errors, fmt.Sprintf("cadena sin cerrar en línea %d", l.line))
	return l.input[start:]
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return Token{Type: TokenEOF, Line: l.line, Column: l.col}
	}

	line, col := l.line, l.col
	ch := l.advance()

	switch {
	case unicode.IsLetter(ch) || ch == '_':
		word := l.readIdent()
		upper := strings.ToUpper(word)
		if tokType, ok := keywords[upper]; ok {
			return Token{Type: tokType, Literal: upper, Line: line, Column: col}
		}
		return Token{Type: TokenIdent, Literal: word, Line: line, Column: col}

	case unicode.IsDigit(ch):
		num := l.readNumber()
		return Token{Type: TokenNumber, Literal: num, Line: line, Column: col}

	case ch == '\'':
		str := l.readString()
		return Token{Type: TokenString, Literal: str, Line: line, Column: col}

	case ch == '*':
		return Token{Type: TokenStar, Line: line, Column: col}
	case ch == '(':
		return Token{Type: TokenLParen, Line: line, Column: col}
	case ch == ')':
		return Token{Type: TokenRParen, Line: line, Column: col}
	case ch == ',':
		return Token{Type: TokenComma, Line: line, Column: col}

	case ch == '=':
		return Token{Type: TokenEq, Line: line, Column: col}

	case ch == '<':
		if l.peek() == '>' {
			l.advance()
			return Token{Type: TokenNeq, Line: line, Column: col}
		}
		if l.peek() == '=' {
			l.advance()
			return Token{Type: TokenLe, Line: line, Column: col}
		}
		return Token{Type: TokenLt, Line: line, Column: col}

	case ch == '>':
		if l.peek() == '=' {
			l.advance()
			return Token{Type: TokenGe, Line: line, Column: col}
		}
		return Token{Type: TokenGt, Line: line, Column: col}

	default:
		l.errors = append(l.errors, fmt.Sprintf("carácter inesperado %q en línea %d, columna %d", ch, line, col))
		return Token{Type: TokenEOF, Line: line, Column: col}
	}
}

func Tokenize(input string) ([]Token, []string) {
	lexer := New(input)
	var tokens []Token
	for {
		tok := lexer.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			break
		}
	}
	return tokens, lexer.Errors()
}
