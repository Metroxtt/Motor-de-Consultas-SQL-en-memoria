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

func TestParseCountStar(t *testing.T) {
	node, err := Parse("SELECT COUNT(*) FROM employees")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(node.Columns) != 1 {
		t.Fatalf("len(Columns) = %d, want 1", len(node.Columns))
	}

	agg, ok := node.Columns[0].(*AggregateNode)
	if !ok {
		t.Fatalf("Columns[0] is %T, want *AggregateNode", node.Columns[0])
	}

	if agg.Func != "COUNT" {
		t.Errorf("Func = %q, want %q", agg.Func, "COUNT")
	}
	if agg.Column != "*" {
		t.Errorf("Column = %q, want %q", agg.Column, "*")
	}
}

func TestParseAggregateFunctions(t *testing.T) {
	tests := []struct {
		query    string
		wantFunc string
		wantCol  string
	}{
		{"SELECT COUNT(*) FROM t", "COUNT", "*"},
		{"SELECT COUNT(id) FROM t", "COUNT", "id"},
		{"SELECT SUM(salary) FROM t", "SUM", "salary"},
		{"SELECT AVG(salary) FROM t", "AVG", "salary"},
		{"SELECT MIN(salary) FROM t", "MIN", "salary"},
		{"SELECT MAX(salary) FROM t", "MAX", "salary"},
	}

	for _, tt := range tests {
		t.Run(tt.wantFunc, func(t *testing.T) {
			node, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", tt.query, err)
			}

			if len(node.Columns) != 1 {
				t.Fatalf("len(Columns) = %d, want 1", len(node.Columns))
			}

			agg, ok := node.Columns[0].(*AggregateNode)
			if !ok {
				t.Fatalf("Columns[0] is %T, want *AggregateNode", node.Columns[0])
			}

			if agg.Func != tt.wantFunc {
				t.Errorf("Func = %q, want %q", agg.Func, tt.wantFunc)
			}
			if agg.Column != tt.wantCol {
				t.Errorf("Column = %q, want %q", agg.Column, tt.wantCol)
			}
		})
	}
}

func TestParseMultipleAggregates(t *testing.T) {
	node, err := Parse("SELECT COUNT(*), SUM(salary), AVG(salary) FROM t")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(node.Columns) != 3 {
		t.Fatalf("len(Columns) = %d, want 3", len(node.Columns))
	}

	expectedFuncs := []string{"COUNT", "SUM", "AVG"}
	for i, col := range node.Columns {
		agg, ok := col.(*AggregateNode)
		if !ok {
			t.Fatalf("Columns[%d] is %T, want *AggregateNode", i, col)
		}
		if agg.Func != expectedFuncs[i] {
			t.Errorf("Columns[%d].Func = %q, want %q", i, agg.Func, expectedFuncs[i])
		}
	}
}

func TestParseGroupBy(t *testing.T) {
	node, err := Parse("SELECT active, COUNT(*) FROM employees GROUP BY active")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if node.GroupBy != "active" {
		t.Errorf("GroupBy = %q, want %q", node.GroupBy, "active")
	}

	if len(node.Columns) != 2 {
		t.Fatalf("len(Columns) = %d, want 2", len(node.Columns))
	}

	if _, ok := node.Columns[0].(*ColumnRefNode); !ok {
		t.Errorf("Columns[0] is %T, want *ColumnRefNode", node.Columns[0])
	}
	if _, ok := node.Columns[1].(*AggregateNode); !ok {
		t.Errorf("Columns[1] is %T, want *AggregateNode", node.Columns[1])
	}
}

func TestParseGroupByWithWhere(t *testing.T) {
	node, err := Parse("SELECT active, COUNT(*) FROM employees WHERE salary > 50000 GROUP BY active")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if node.Where == nil {
		t.Fatal("Where = nil, want comparación")
	}

	if node.GroupBy != "active" {
		t.Errorf("GroupBy = %q, want %q", node.GroupBy, "active")
	}
}

func TestParseGroupByErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"GROUP BY sin columna", "SELECT COUNT(*) FROM t GROUP BY"},
		{"GROUP BY sin BY", "SELECT COUNT(*) FROM t GROUP"},
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

func TestParseAggregateErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"COUNT sin parentesis", "SELECT COUNT FROM t"},
		{"COUNT sin columna", "SELECT COUNT() FROM t"},
		{"COUNT sin cerrar", "SELECT COUNT(* FROM t"},
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
