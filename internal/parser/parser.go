package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/lexer"
)

type Parser struct {
	tokens []lexer.Token
	pos    int
	errors []string
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peek() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() lexer.Token {
	tok := p.tokens[p.pos]
	p.pos++
	return tok
}

func (p *Parser) expect(tt lexer.TokenType) (lexer.Token, error) {
	tok := p.peek()
	if tok.Type != tt {
		return tok, fmt.Errorf("se esperaba %s pero se encontró %s en línea %d, columna %d",
			tt, tok.Type, tok.Line, tok.Column)
	}
	return p.advance(), nil
}

func (p *Parser) Parse() (*SelectNode, error) {
	if err := p.parseSelect(); err != nil {
		return nil, err
	}

	node := &SelectNode{}

	tok := p.peek()
	if tok.Type == lexer.TokenStar {
		p.advance()
		node.Columns = nil
	} else {
		cols, err := p.parseColumnList()
		if err != nil {
			return nil, err
		}
		node.Columns = cols
	}

	if _, err := p.expect(lexer.TokenFrom); err != nil {
		return nil, err
	}

	tableTok, err := p.expect(lexer.TokenIdent)
	if err != nil {
		return nil, err
	}
	node.Table = tableTok.Literal

	if p.peek().Type == lexer.TokenWhere {
		p.advance()
		expr, err := p.parseOrExpr()
		if err != nil {
			return nil, err
		}
		node.Where = expr
	}

	if p.peek().Type != lexer.TokenEOF {
		return nil, fmt.Errorf("token inesperado %s en línea %d, columna %d",
			p.peek().Type, p.peek().Line, p.peek().Column)
	}

	return node, nil
}

func (p *Parser) parseSelect() error {
	_, err := p.expect(lexer.TokenSelect)
	return err
}

func (p *Parser) parseColumnList() ([]Node, error) {
	var columns []Node

	tok := p.peek()
	if tok.Type == lexer.TokenStar {
		p.advance()
		return nil, nil
	}

	col, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	columns = append(columns, col)

	for p.peek().Type == lexer.TokenComma {
		p.advance()
		col, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		columns = append(columns, col)
	}

	return columns, nil
}

func (p *Parser) parseOrExpr() (Node, error) {
	left, err := p.parseAndExpr()
	if err != nil {
		return nil, err
	}

	for p.peek().Type == lexer.TokenOr {
		p.advance()
		right, err := p.parseAndExpr()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Op: "OR", Left: left, Right: right}
	}

	return left, nil
}

func (p *Parser) parseAndExpr() (Node, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	for p.peek().Type == lexer.TokenAnd {
		p.advance()
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Op: "AND", Left: left, Right: right}
	}

	return left, nil
}

func (p *Parser) parseExpr() (Node, error) {
	return p.parseOrExpr()
}

func (p *Parser) parseComparison() (Node, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	tok := p.peek()
	switch tok.Type {
	case lexer.TokenEq, lexer.TokenNeq, lexer.TokenLt, lexer.TokenGt, lexer.TokenLe, lexer.TokenGe:
		p.advance()
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &ComparisonNode{Op: tok.Type.String(), Left: left, Right: right}, nil
	}

	return left, nil
}

func (p *Parser) parsePrimary() (Node, error) {
	tok := p.peek()

	switch tok.Type {
	case lexer.TokenLParen:
		p.advance()
		expr, err := p.parseOrExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(lexer.TokenRParen); err != nil {
			return nil, err
		}
		return expr, nil

	case lexer.TokenIdent:
		lower := strings.ToUpper(tok.Literal)
		if lower == "TRUE" || lower == "FALSE" {
			p.advance()
			return &BoolLitNode{Value: lower == "TRUE"}, nil
		}
		if lower == "NULL" {
			p.advance()
			return &NullLitNode{}, nil
		}
		if lower == "COUNT" || lower == "SUM" || lower == "AVG" || lower == "MIN" || lower == "MAX" {
			return p.parseAggregate(lower)
		}
		p.advance()
		return &ColumnRefNode{Name: tok.Literal}, nil

	case lexer.TokenNumber:
		p.advance()
		return &NumberLitNode{Value: tok.Literal}, nil

	case lexer.TokenString:
		p.advance()
		return &StringLitNode{Value: tok.Literal}, nil

	case lexer.TokenSelect:
		subQuery, err := p.Parse()
		if err != nil {
			return nil, err
		}
		return subQuery, nil

	default:
		return nil, fmt.Errorf("expresión inesperada %s en línea %d, columna %d",
			tok.Type, tok.Line, tok.Column)
	}
}

func (p *Parser) parseAggregate(funcName string) (Node, error) {
	p.advance() // consumimos el nombre de la función (COUNT, SUM, etc.)

	if _, err := p.expect(lexer.TokenLParen); err != nil {
		return nil, fmt.Errorf("se esperaba ( después de %s", funcName)
	}

	tok := p.peek()
	var col string
	if tok.Type == lexer.TokenStar {
		p.advance()
		col = "*"
	} else if tok.Type == lexer.TokenIdent {
		col = tok.Literal
		p.advance()
	} else {
		return nil, fmt.Errorf("se esperaba columna o * dentro de %s()", funcName)
	}

	if _, err := p.expect(lexer.TokenRParen); err != nil {
		return nil, fmt.Errorf("se esperaba ) después de %s(%s", funcName, col)
	}

	return &AggregateNode{Func: funcName, Column: col}, nil
}

func Parse(input string) (*SelectNode, error) {
	tokens, lexErrors := lexer.Tokenize(input)
	if len(lexErrors) > 0 {
		return nil, fmt.Errorf("errores de lexer: %v", lexErrors)
	}

	parser := New(tokens)
	node, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	if len(parser.Errors()) > 0 {
		return nil, fmt.Errorf("errores de parser: %v", parser.Errors())
	}

	return node, nil
}

func formatError(err error) string {
	return err.Error()
}

func parseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
