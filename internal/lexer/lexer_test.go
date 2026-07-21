package lexer

import (
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTypes []TokenType
		wantErr   bool
	}{
		{
			name:      "SELECT simple con *",
			input:     "SELECT * FROM employees",
			wantTypes: []TokenType{TokenSelect, TokenStar, TokenFrom, TokenIdent, TokenEOF},
		},
		{
			name:      "SELECT con columnas",
			input:     "SELECT name, age FROM users",
			wantTypes: []TokenType{TokenSelect, TokenIdent, TokenComma, TokenIdent, TokenFrom, TokenIdent, TokenEOF},
		},
		{
			name:      "WHERE con igualdad",
			input:     "SELECT name FROM users WHERE age = 25",
			wantTypes: []TokenType{TokenSelect, TokenIdent, TokenFrom, TokenIdent, TokenWhere, TokenIdent, TokenEq, TokenNumber, TokenEOF},
		},
		{
			name:      "WHERE con comparaciones",
			input:     "SELECT * FROM t WHERE a <> 1 AND b > 10 OR c <= 5",
			wantTypes: []TokenType{TokenSelect, TokenStar, TokenFrom, TokenIdent, TokenWhere, TokenIdent, TokenNeq, TokenNumber, TokenAnd, TokenIdent, TokenGt, TokenNumber, TokenOr, TokenIdent, TokenLe, TokenNumber, TokenEOF},
		},
		{
			name:      "parentesis",
			input:     "SELECT * FROM t WHERE (a = 1 OR b = 2)",
			wantTypes: []TokenType{TokenSelect, TokenStar, TokenFrom, TokenIdent, TokenWhere, TokenLParen, TokenIdent, TokenEq, TokenNumber, TokenOr, TokenIdent, TokenEq, TokenNumber, TokenRParen, TokenEOF},
		},
		{
			name:      "cadena de texto",
			input:     "SELECT * FROM t WHERE name = 'Alice'",
			wantTypes: []TokenType{TokenSelect, TokenStar, TokenFrom, TokenIdent, TokenWhere, TokenIdent, TokenEq, TokenString, TokenEOF},
		},
		{
			name:    "cadena sin cerrar",
			input:   "SELECT * FROM t WHERE name = 'Alice",
			wantErr: true,
		},
		{
			name:    "carácter inesperado",
			input:   "SELECT @ FROM t",
			wantErr: true,
		},
		{
			name:      "decimal",
			input:     "SELECT * FROM t WHERE price = 3.14",
			wantTypes: []TokenType{TokenSelect, TokenStar, TokenFrom, TokenIdent, TokenWhere, TokenIdent, TokenEq, TokenNumber, TokenEOF},
		},
		{
			name:      "case insensitive keywords",
			input:     "select * from employees where age > 25",
			wantTypes: []TokenType{TokenSelect, TokenStar, TokenFrom, TokenIdent, TokenWhere, TokenIdent, TokenGt, TokenNumber, TokenEOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, errs := Tokenize(tt.input)

			if tt.wantErr {
				if len(errs) == 0 {
					t.Error("Tokenize() esperaba errores, no obtuvo ninguno")
				}
				return
			}

			if len(errs) > 0 {
				t.Errorf("Tokenize() errores inesperados: %v", errs)
				return
			}

			if len(tokens) != len(tt.wantTypes) {
				t.Errorf("cantidad de tokens = %d, want %d", len(tokens), len(tt.wantTypes))
				for i, tok := range tokens {
					t.Logf("  [%d] %s", i, tok)
				}
				return
			}

			for i, tok := range tokens {
				if tok.Type != tt.wantTypes[i] {
					t.Errorf("token[%d].Type = %s, want %s (literal: %q)", i, tok.Type, tt.wantTypes[i], tok.Literal)
				}
			}
		})
	}
}

func TestTokenPosition(t *testing.T) {
	tokens, _ := Tokenize("SELECT name FROM users")

	if tokens[0].Line != 1 || tokens[0].Column != 1 {
		t.Errorf("SELECT pos = %d:%d, want 1:1", tokens[0].Line, tokens[0].Column)
	}

	if tokens[1].Column != 8 {
		t.Errorf("name column = %d, want 8", tokens[1].Column)
	}
}

func TestKeywords(t *testing.T) {
	keywords := map[string]TokenType{
		"SELECT": TokenSelect, "FROM": TokenFrom,
		"WHERE": TokenWhere, "AND": TokenAnd, "OR": TokenOr,
	}

	for word, expectedType := range keywords {
		tokens, _ := Tokenize(word)
		if tokens[0].Type != expectedType {
			t.Errorf("Tokenize(%q) = %s, want %s", word, tokens[0].Type, expectedType)
		}
	}
}
