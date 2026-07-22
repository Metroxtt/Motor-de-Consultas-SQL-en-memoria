package parser

import (
	"testing"
)

func TestParseSelectStar(t *testing.T) {
	node, err := Parse("SELECT * FROM employees")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if node.Table != "employees" {
		t.Errorf("Table = %q, want %q", node.Table, "employees")
	}

	if node.Columns != nil {
		t.Errorf("Columns = %v, want nil (para *)", node.Columns)
	}

	if node.Where != nil {
		t.Errorf("Where = %v, want nil", node.Where)
	}
}

func TestParseSelectColumns(t *testing.T) {
	node, err := Parse("SELECT name, age, salary FROM users")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if node.Table != "users" {
		t.Errorf("Table = %q, want %q", node.Table, "users")
	}

	if len(node.Columns) != 3 {
		t.Fatalf("len(Columns) = %d, want 3", len(node.Columns))
	}

	expected := []string{"name", "age", "salary"}
	for i, col := range node.Columns {
		ref, ok := col.(*ColumnRefNode)
		if !ok {
			t.Fatalf("Columns[%d] is %T, want *ColumnRefNode", i, col)
		}
		if ref.Name != expected[i] {
			t.Errorf("Columns[%d].Name = %q, want %q", i, ref.Name, expected[i])
		}
	}
}

func TestParseWhereSimple(t *testing.T) {
	node, err := Parse("SELECT * FROM users WHERE age = 25")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if node.Where == nil {
		t.Fatal("Where = nil, want comparación")
	}

	comp, ok := node.Where.(*ComparisonNode)
	if !ok {
		t.Fatalf("Where is %T, want *ComparisonNode", node.Where)
	}

	if comp.Op != "=" {
		t.Errorf("Op = %q, want %q", comp.Op, "=")
	}
}

func TestParseWhereAND(t *testing.T) {
	node, err := Parse("SELECT * FROM t WHERE a > 1 AND b = 'x'")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	bin, ok := node.Where.(*BinaryOpNode)
	if !ok {
		t.Fatalf("Where is %T, want *BinaryOpNode", node.Where)
	}

	if bin.Op != "AND" {
		t.Errorf("Op = %q, want %q", bin.Op, "AND")
	}

	if bin.Left.Type() != NodeComparison {
		t.Errorf("Left type = %v, want NodeComparison", bin.Left.Type())
	}

	if bin.Right.Type() != NodeComparison {
		t.Errorf("Right type = %v, want NodeComparison", bin.Right.Type())
	}
}

func TestParseWhereOR(t *testing.T) {
	node, err := Parse("SELECT * FROM t WHERE a = 1 OR b = 2")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	bin, ok := node.Where.(*BinaryOpNode)
	if !ok {
		t.Fatalf("Where is %T, want *BinaryOpNode", node.Where)
	}

	if bin.Op != "OR" {
		t.Errorf("Op = %q, want %q", bin.Op, "OR")
	}
}

func TestParseWhereComplex(t *testing.T) {
	node, err := Parse("SELECT * FROM t WHERE (a > 1 OR b < 2) AND c = 'hello'")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	bin, ok := node.Where.(*BinaryOpNode)
	if !ok {
		t.Fatalf("Where is %T, want *BinaryOpNode", node.Where)
	}

	if bin.Op != "AND" {
		t.Errorf("Op = %q, want %q", bin.Op, "AND")
	}

	// El lado izquierdo es OR (entre paréntesis)
	orNode, ok := bin.Left.(*BinaryOpNode)
	if !ok {
		t.Fatalf("Left is %T, want *BinaryOpNode (OR)", bin.Left)
	}
	if orNode.Op != "OR" {
		t.Errorf("Left.Op = %q, want %q", orNode.Op, "OR")
	}
}

func TestParseComparisons(t *testing.T) {
	ops := []struct {
		op   string
		want string
	}{
		{"=", "="}, {"<>", "<>"}, {"<", "<"}, {">", ">"}, {"<=", "<="}, {">=", ">="},
	}

	for _, tt := range ops {
		query := "SELECT * FROM t WHERE x " + tt.op + " 1"
		node, err := Parse(query)
		if err != nil {
			t.Errorf("Parse(%q) error = %v", query, err)
			continue
		}
		comp, ok := node.Where.(*ComparisonNode)
		if !ok {
			t.Errorf("Parse(%q) Where is %T, want *ComparisonNode", query, node.Where)
			continue
		}
		if comp.Op != tt.want {
			t.Errorf("Op = %q, want %q", comp.Op, tt.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"sin SELECT", "FROM users"},
		{"sin FROM", "SELECT * WHERE x = 1"},
		{"sin tabla", "SELECT * FROM"},
		{"token extra", "SELECT * FROM users extra"},
		{"WHERE incompleto", "SELECT * FROM users WHERE"},
		{"parentesis sin cerrar", "SELECT * FROM t WHERE (a = 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if err == nil {
				t.Errorf("Parse(%q) esperaba error, obtuvo nil", tt.input)
			}
		})
	}
}

func TestParseStringLiterals(t *testing.T) {
	node, err := Parse("SELECT * FROM t WHERE name = 'Alice'")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	comp, ok := node.Where.(*ComparisonNode)
	if !ok {
		t.Fatalf("Where is %T, want *ComparisonNode", node.Where)
	}

	str, ok := comp.Right.(*StringLitNode)
	if !ok {
		t.Fatalf("Right is %T, want *StringLitNode", comp.Right)
	}

	if str.Value != "Alice" {
		t.Errorf("Value = %q, want %q", str.Value, "Alice")
	}
}

func TestParseNumberLiterals(t *testing.T) {
	node, err := Parse("SELECT * FROM t WHERE price = 3.14")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	comp := node.Where.(*ComparisonNode)
	num, ok := comp.Right.(*NumberLitNode)
	if !ok {
		t.Fatalf("Right is %T, want *NumberLitNode", comp.Right)
	}

	if num.Value != "3.14" {
		t.Errorf("Value = %q, want %q", num.Value, "3.14")
	}
}

func TestParseOrderBy(t *testing.T) {
	node, err := Parse("SELECT * FROM t ORDER BY name ASC")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(node.OrderBy) != 1 {
		t.Fatalf("len(OrderBy) = %d, want 1", len(node.OrderBy))
	}
	if node.OrderBy[0].Column != "name" {
		t.Errorf("OrderBy[0].Column = %q, want %q", node.OrderBy[0].Column, "name")
	}
	if !node.OrderBy[0].Asc {
		t.Error("OrderBy[0].Asc = false, want true (ASC)")
	}
}
func TestParseOrderByDesc(t *testing.T) {
	node, err := Parse("SELECT * FROM t ORDER BY salary DESC")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(node.OrderBy) != 1 {
		t.Fatalf("len(OrderBy) = %d, want 1", len(node.OrderBy))
	}
	if node.OrderBy[0].Column != "salary" {
		t.Errorf("OrderBy[0].Column = %q, want %q", node.OrderBy[0].Column, "salary")
	}
	if node.OrderBy[0].Asc {
		t.Error("OrderBy[0].Asc = true, want false (DESC)")
	}
}
func TestParseOrderByMultiple(t *testing.T) {
	node, err := Parse("SELECT * FROM t ORDER BY age ASC, name DESC")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(node.OrderBy) != 2 {
		t.Fatalf("len(OrderBy) = %d, want 2", len(node.OrderBy))
	}
	if node.OrderBy[0].Column != "age" || !node.OrderBy[0].Asc {
		t.Errorf("OrderBy[0] = {Column:%q, Asc:%v}, want {age, true}", node.OrderBy[0].Column, node.OrderBy[0].Asc)
	}
	if node.OrderBy[1].Column != "name" || node.OrderBy[1].Asc {
		t.Errorf("OrderBy[1] = {Column:%q, Asc:%v}, want {name, false}", node.OrderBy[1].Column, node.OrderBy[1].Asc)
	}
}

func TestParseLimit(t *testing.T) {
	node, err := Parse("SELECT * FROM t LIMIT 5")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if node.Limit == nil {
		t.Fatal("Limit = nil, want 5")
	}
	if *node.Limit != 5 {
		t.Errorf("*Limit = %d, want 5", *node.Limit)
	}
}

func TestParseOrderByAndLimit(t *testing.T) {
	node, err := Parse("SELECT * FROM t WHERE age > 20 ORDER BY name DESC LIMIT 3")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if node.Where == nil {
		t.Fatal("Where = nil, want comparación")
	}
	if len(node.OrderBy) != 1 {
		t.Fatalf("len(OrderBy) = %d, want 1", len(node.OrderBy))
	}
	if node.OrderBy[0].Column != "name" || node.OrderBy[0].Asc {
		t.Errorf("OrderBy[0] = {Column:%q, Asc:%v}, want {name, false}", node.OrderBy[0].Column, node.OrderBy[0].Asc)
	}
	if node.Limit == nil || *node.Limit != 3 {
		t.Errorf("Limit = %v, want 3", node.Limit)
	}
}

func TestParseLimitSinOrderBy(t *testing.T) {
	node, err := Parse("SELECT * FROM t LIMIT 2")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if node.OrderBy != nil {
		t.Errorf("OrderBy = %v, want nil", node.OrderBy)
	}
	if node.Limit == nil || *node.Limit != 2 {
		t.Errorf("Limit = %v, want 2", node.Limit)
	}
}
