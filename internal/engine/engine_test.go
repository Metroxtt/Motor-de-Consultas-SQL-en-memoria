package engine

import (
	"strings"
	"testing"

	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/catalog"
	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/parser"
)

const testCSV = `id,name,salary,active
1,Alice,50000,true
2,Bob,42000,true
3,Charlie,61000,false
4,Diana,48000,true
5,Eve,72000,false`

func loadTestTable(t *testing.T) *catalog.Table {
	t.Helper()
	table, err := catalog.LoadCSV("employees", strings.NewReader(testCSV), nil)
	if err != nil {
		t.Fatalf("LoadCSV error = %v", err)
	}
	return table
}

func TestScanOperator(t *testing.T) {
	table := loadTestTable(t)
	scan := NewScanOperator(table)
	defer scan.Close()

	count := 0
	for {
		row, err := scan.Next()
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		if row == nil {
			break
		}
		count++
	}

	if count != 5 {
		t.Errorf("Scan produjo %d filas, want 5", count)
	}
}

func TestFilterOperator(t *testing.T) {
	table := loadTestTable(t)
	scan := NewScanOperator(table)

	node, err := parser.Parse("SELECT * FROM t WHERE salary > 50000")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}

	filter := NewFilterOperator(scan, node.Where)
	defer filter.Close()

	count := 0
	for {
		row, err := filter.Next()
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		if row == nil {
			break
		}
		salary := row["salary"].(int64)
		if salary <= 50000 {
			t.Errorf("salary %d no cumple > 50000", salary)
		}
		count++
	}

	if count != 2 {
		t.Errorf("Filter produjo %d filas, want 2", count)
	}
}

func TestProjectOperator(t *testing.T) {
	table := loadTestTable(t)
	scan := NewScanOperator(table)

	node, err := parser.Parse("SELECT name, salary FROM t")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}

	project := NewProjectOperator(scan, node.Columns, table.Schema)
	defer project.Close()

	row, err := project.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if row == nil {
		t.Fatal("Next() retornó nil")
	}

	if _, ok := row["name"]; !ok {
		t.Error("row no tiene campo 'name'")
	}
	if _, ok := row["salary"]; !ok {
		t.Error("row no tiene campo 'salary'")
	}
	if _, ok := row["id"]; ok {
		t.Error("row tiene campo 'id' que no debía estar")
	}
}

func TestProjectStar(t *testing.T) {
	table := loadTestTable(t)
	scan := NewScanOperator(table)

	project := NewProjectOperator(scan, nil, table.Schema)
	defer project.Close()

	row, err := project.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}

	expected := []string{"id", "name", "salary", "active"}
	for _, col := range expected {
		if _, ok := row[col]; !ok {
			t.Errorf("row no tiene campo %q", col)
		}
	}
}

func TestFullPlan(t *testing.T) {
	table := loadTestTable(t)
	cat := catalog.NewCatalog()
	cat.AddTable(table)

	node, err := parser.Parse("SELECT name, salary FROM employees WHERE salary > 50000")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}

	plan, err := BuildPlan(node, cat)
	if err != nil {
		t.Fatalf("BuildPlan error = %v", err)
	}
	defer plan.Close()

	var results []catalog.Row
	for {
		row, err := plan.Next()
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		if row == nil {
			break
		}
		results = append(results, row)
	}

	if len(results) != 2 {
		t.Errorf("resultados = %d, want 2", len(results))
	}

	for _, row := range results {
		if _, ok := row["name"]; !ok {
			t.Error("resultado sin campo 'name'")
		}
		if _, ok := row["salary"]; !ok {
			t.Error("resultado sin campo 'salary'")
		}
	}
}

func TestBuildPlanTablaNoExiste(t *testing.T) {
	cat := catalog.NewCatalog()
	node, err := parser.Parse("SELECT * FROM noexiste")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}

	_, err = BuildPlan(node, cat)
	if err == nil {
		t.Error("BuildPlan esperaba error para tabla inexistente")
	}
}

func TestEvalComparisons(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		expect int
	}{
		{"igualdad", "SELECT * FROM employees WHERE id = 2", 1},
		{"desigualdad", "SELECT * FROM employees WHERE id <> 2", 4},
		{"menor", "SELECT * FROM employees WHERE id < 3", 2},
		{"mayor", "SELECT * FROM employees WHERE id > 3", 2},
		{"menor igual", "SELECT * FROM employees WHERE id <= 2", 2},
		{"mayor igual", "SELECT * FROM employees WHERE id >= 4", 2},
	}

	table := loadTestTable(t)
	cat := catalog.NewCatalog()
	cat.AddTable(table)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := parser.Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse error = %v", err)
			}

			plan, err := BuildPlan(node, cat)
			if err != nil {
				t.Fatalf("BuildPlan error = %v", err)
			}
			defer plan.Close()

			count := 0
			for {
				row, err := plan.Next()
				if err != nil {
					t.Fatalf("Next() error = %v", err)
				}
				if row == nil {
					break
				}
				count++
			}

			if count != tt.expect {
				t.Errorf("resultados = %d, want %d", count, tt.expect)
			}
		})
	}
}

func TestEvalANDOR(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		expect int
	}{
		{"AND", "SELECT * FROM employees WHERE salary > 45000 AND active = true", 2},
		{"OR", "SELECT * FROM employees WHERE id = 1 OR id = 5", 2},
		{"AND+OR", "SELECT * FROM employees WHERE (salary > 45000 AND active = true) OR id = 3", 3},
	}

	table := loadTestTable(t)
	cat := catalog.NewCatalog()
	cat.AddTable(table)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := parser.Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse error = %v", err)
			}

			plan, err := BuildPlan(node, cat)
			if err != nil {
				t.Fatalf("BuildPlan error = %v", err)
			}
			defer plan.Close()

			count := 0
			for {
				row, err := plan.Next()
				if err != nil {
					t.Fatalf("Next() error = %v", err)
				}
				if row == nil {
					break
				}
				count++
			}

			if count != tt.expect {
				t.Errorf("resultados = %d, want %d", count, tt.expect)
			}
		})
	}
}
