package engine

import (
	"fmt"
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

func TestCountStar(t *testing.T) {
	table := loadTestTable(t)
	cat := catalog.NewCatalog()
	cat.AddTable(table)

	node, err := parser.Parse("SELECT COUNT(*) FROM employees")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}

	plan, err := BuildPlan(node, cat)
	if err != nil {
		t.Fatalf("BuildPlan error = %v", err)
	}
	defer plan.Close()

	row, err := plan.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if row == nil {
		t.Fatal("Next() retornó nil")
	}

	count, ok := row["COUNT(*)"].(int64)
	if !ok {
		t.Fatalf("COUNT(*) es %T, want int64", row["COUNT(*)"])
	}
	if count != 5 {
		t.Errorf("COUNT(*) = %d, want 5", count)
	}
}

func TestSumSalary(t *testing.T) {
	table := loadTestTable(t)
	cat := catalog.NewCatalog()
	cat.AddTable(table)

	node, err := parser.Parse("SELECT SUM(salary) FROM employees")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}

	plan, err := BuildPlan(node, cat)
	if err != nil {
		t.Fatalf("BuildPlan error = %v", err)
	}
	defer plan.Close()

	row, err := plan.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}

	sum, ok := row["SUM(salary)"].(int64)
	if !ok {
		t.Fatalf("SUM(salary) es %T, want int64", row["SUM(salary)"])
	}
	// 50000 + 42000 + 61000 + 48000 + 72000 = 273000
	if sum != 273000 {
		t.Errorf("SUM(salary) = %d, want 273000", sum)
	}
}

func TestAvgSalary(t *testing.T) {
	table := loadTestTable(t)
	cat := catalog.NewCatalog()
	cat.AddTable(table)

	node, err := parser.Parse("SELECT AVG(salary) FROM employees")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}

	plan, err := BuildPlan(node, cat)
	if err != nil {
		t.Fatalf("BuildPlan error = %v", err)
	}
	defer plan.Close()

	row, err := plan.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}

	avg, ok := row["AVG(salary)"].(int64)
	if !ok {
		t.Fatalf("AVG(salary) es %T, want int64", row["AVG(salary)"])
	}
	// 273000 / 5 = 54600
	if avg != 54600 {
		t.Errorf("AVG(salary) = %d, want 54600", avg)
	}
}

func TestMinMaxSalary(t *testing.T) {
	table := loadTestTable(t)
	cat := catalog.NewCatalog()
	cat.AddTable(table)

	node, err := parser.Parse("SELECT MIN(salary), MAX(salary) FROM employees")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}

	plan, err := BuildPlan(node, cat)
	if err != nil {
		t.Fatalf("BuildPlan error = %v", err)
	}
	defer plan.Close()

	row, err := plan.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}

	min, ok := row["MIN(salary)"].(int64)
	if !ok {
		t.Fatalf("MIN(salary) es %T, want int64", row["MIN(salary)"])
	}
	if min != 42000 {
		t.Errorf("MIN(salary) = %d, want 42000", min)
	}

	max, ok := row["MAX(salary)"].(int64)
	if !ok {
		t.Fatalf("MAX(salary) es %T, want int64", row["MAX(salary)"])
	}
	if max != 72000 {
		t.Errorf("MAX(salary) = %d, want 72000", max)
	}
}

func TestGroupByActive(t *testing.T) {
	table := loadTestTable(t)
	cat := catalog.NewCatalog()
	cat.AddTable(table)

	node, err := parser.Parse("SELECT active, COUNT(*) FROM employees GROUP BY active")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}

	plan, err := BuildPlan(node, cat)
	if err != nil {
		t.Fatalf("BuildPlan error = %v", err)
	}
	defer plan.Close()

	groups := make(map[string]int64)
	for {
		row, err := plan.Next()
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		if row == nil {
			break
		}
		key := fmt.Sprintf("%v", row["active"])
		count := row["COUNT(*)"].(int64)
		groups[key] = count
	}

	// true: Alice, Bob, Diana (3)
	// false: Charlie, Eve (2)
	if groups["true"] != 3 {
		t.Errorf("GROUP BY active, COUNT(*) para true = %d, want 3", groups["true"])
	}
	if groups["false"] != 2 {
		t.Errorf("GROUP BY active, COUNT(*) para false = %d, want 2", groups["false"])
	}
}

func TestGroupByWithAvg(t *testing.T) {
	table := loadTestTable(t)
	cat := catalog.NewCatalog()
	cat.AddTable(table)

	node, err := parser.Parse("SELECT active, AVG(salary) FROM employees GROUP BY active")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}

	plan, err := BuildPlan(node, cat)
	if err != nil {
		t.Fatalf("BuildPlan error = %v", err)
	}
	defer plan.Close()

	groups := make(map[string]int64)
	for {
		row, err := plan.Next()
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		if row == nil {
			break
		}
		key := fmt.Sprintf("%v", row["active"])
		avg := row["AVG(salary)"].(int64)
		groups[key] = avg
	}

	// true: (50000 + 42000 + 48000) / 3 = 46666
	// false: (61000 + 72000) / 2 = 66500
	if groups["true"] != 46666 {
		t.Errorf("AVG(salary) para active=true = %d, want 46666", groups["true"])
	}
	if groups["false"] != 66500 {
		t.Errorf("AVG(salary) para active=false = %d, want 66500", groups["false"])
	}
}

func TestCountWithWhere(t *testing.T) {
	table := loadTestTable(t)
	cat := catalog.NewCatalog()
	cat.AddTable(table)

	node, err := parser.Parse("SELECT COUNT(*) FROM employees WHERE salary > 50000")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}

	plan, err := BuildPlan(node, cat)
	if err != nil {
		t.Fatalf("BuildPlan error = %v", err)
	}
	defer plan.Close()

	row, err := plan.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}

	count := row["COUNT(*)"].(int64)
	// Charlie (61000) y Eve (72000)
	if count != 2 {
		t.Errorf("COUNT(*) con WHERE = %d, want 2", count)
	}
}

func TestMultipleAggregates(t *testing.T) {
	table := loadTestTable(t)
	cat := catalog.NewCatalog()
	cat.AddTable(table)

	node, err := parser.Parse("SELECT COUNT(*), SUM(salary), AVG(salary) FROM employees")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}

	plan, err := BuildPlan(node, cat)
	if err != nil {
		t.Fatalf("BuildPlan error = %v", err)
	}
	defer plan.Close()

	row, err := plan.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}

	count := row["COUNT(*)"].(int64)
	sum := row["SUM(salary)"].(int64)
	avg := row["AVG(salary)"].(int64)

	if count != 5 {
		t.Errorf("COUNT(*) = %d, want 5", count)
	}
	if sum != 273000 {
		t.Errorf("SUM(salary) = %d, want 273000", sum)
	}
	if avg != 54600 {
		t.Errorf("AVG(salary) = %d, want 54600", avg)
	}
}
