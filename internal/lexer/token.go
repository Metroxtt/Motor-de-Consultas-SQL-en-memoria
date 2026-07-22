package lexer

import "fmt"

type TokenType int

const (
	// Literales
	TokenIdent  TokenType = iota // nombre de columna/tabla
	TokenNumber                  // 42, 3.14
	TokenString                  // 'texto'

	// Palabras clave
	TokenSelect // SELECT
	TokenFrom   // FROM
	TokenWhere  // WHERE
	TokenAnd    // AND
	TokenOr     // OR
	TokenGroup  // GROUP
	TokenBy     // BY

	// Operadores
	TokenEq     // =
	TokenNeq    // <>
	TokenLt     // <
	TokenGt     // >
	TokenLe     // <=
	TokenGe     // >=
	TokenStar   // *
	TokenLParen // (
	TokenRParen // )
	TokenComma  // ,

	// Especiales
	TokenEOF // fin de entrada
)

var tokenNames = map[TokenType]string{
	TokenIdent: "IDENT", TokenNumber: "NUMBER", TokenString: "STRING",
	TokenSelect: "SELECT", TokenFrom: "FROM", TokenWhere: "WHERE",
	TokenAnd: "AND", TokenOr: "OR", TokenGroup: "GROUP", TokenBy: "BY",
	TokenEq: "=", TokenNeq: "<>", TokenLt: "<", TokenGt: ">",
	TokenLe: "<=", TokenGe: ">=", TokenStar: "*",
	TokenLParen: "(", TokenRParen: ")", TokenComma: ",",
	TokenEOF: "EOF",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return "UNKNOWN"
}

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

func (t Token) String() string {
	if t.Literal != "" {
		return fmt.Sprintf("%s(%q)@%d:%d", t.Type, t.Literal, t.Line, t.Column)
	}
	return fmt.Sprintf("%s@%d:%d", t.Type, t.Line, t.Column)
}
