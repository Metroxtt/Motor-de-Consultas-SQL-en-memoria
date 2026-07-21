package catalog

import (
	"strings"
	"testing"
)

func TestInferType(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		expected TypeTag
	}{
		{"all integers", []string{"1", "2", "100", "-5"}, TypeInteger},
		{"all decimals", []string{"1.5", "2.7", "3.14"}, TypeDecimal},
		{"integers and decimals", []string{"1", "2.5", "3"}, TypeDecimal},
		{"all booleans", []string{"true", "false", "true"}, TypeBool},
		{"0/1 are integers", []string{"0", "1", "0"}, TypeInteger},
		{"all text", []string{"hello", "world", "foo"}, TypeText},
		{"mixed types", []string{"hello", "123", "true"}, TypeText},
		{"empty values", []string{"", "", ""}, TypeText},
		{"null values", []string{"null", "NULL", ""}, TypeText},
		{"single integer", []string{"42"}, TypeInteger},
		{"negative integers", []string{"-1", "-100"}, TypeInteger},
		{"yes/no booleans", []string{"yes", "no", "yes"}, TypeBool},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InferType(tt.values)
			if got != tt.expected {
				t.Errorf("InferType(%v) = %v, want %v", tt.values, got, tt.expected)
			}
		})
	}
}

func TestParseValue(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		tag     TypeTag
		want    interface{}
		wantErr bool
	}{
		{"integer ok", "42", TypeInteger, int64(42), false},
		{"integer invalid", "abc", TypeInteger, nil, true},
		{"decimal ok", "3.14", TypeDecimal, 3.14, false},
		{"decimal invalid", "abc", TypeDecimal, nil, true},
		{"bool true", "true", TypeBool, true, false},
		{"bool false", "false", TypeBool, false, false},
		{"bool invalid", "maybe", TypeBool, nil, true},
		{"text", "hello", TypeText, "hello", false},
		{"empty returns nil", "", TypeText, nil, false},
		{"null returns nil", "null", TypeInteger, nil, false},
		{"NULL returns nil", "NULL", TypeDecimal, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseValue(tt.raw, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseValue(%q, %v) error = %v, wantErr %v", tt.raw, tt.tag, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseValue(%q, %v) = %v (%T), want %v (%T)", tt.raw, tt.tag, got, got, tt.want, tt.want)
			}
		})
	}
}

func TestLoadCSV(t *testing.T) {
	csvData := `id,name,score,active
1,Alice,95.5,true
2,Bob,88.0,false
3,,null,true`

	t.Run("inferencia de tipos", func(t *testing.T) {
		table, err := LoadCSV("test", strings.NewReader(csvData), nil)
		if err != nil {
			t.Fatalf("LoadCSV() error = %v", err)
		}
		if table.Name != "test" {
			t.Errorf("table.Name = %q, want %q", table.Name, "test")
		}
		if len(table.Schema.Columns) != 4 {
			t.Fatalf("len(columns) = %d, want 4", len(table.Schema.Columns))
		}

		expected := []struct {
			name string
			typ  TypeTag
		}{
			{"id", TypeInteger},
			{"name", TypeText},
			{"score", TypeDecimal},
			{"active", TypeBool},
		}
		for i, e := range expected {
			if table.Schema.Columns[i].Name != e.name {
				t.Errorf("column[%d].Name = %q, want %q", i, table.Schema.Columns[i].Name, e.name)
			}
			if table.Schema.Columns[i].Type != e.typ {
				t.Errorf("column[%d].Type = %v, want %v", i, table.Schema.Columns[i].Type, e.typ)
			}
		}
	})

	t.Run("filas cargadas", func(t *testing.T) {
		table, err := LoadCSV("test", strings.NewReader(csvData), nil)
		if err != nil {
			t.Fatalf("LoadCSV() error = %v", err)
		}
		if len(table.Rows) != 3 {
			t.Fatalf("len(rows) = %d, want 3", len(table.Rows))
		}

		if table.Rows[0]["id"] != int64(1) {
			t.Errorf("row[0][\"id\"] = %v (%T), want int64(1)", table.Rows[0]["id"], table.Rows[0]["id"])
		}
		if table.Rows[0]["name"] != "Alice" {
			t.Errorf("row[0][\"name\"] = %v, want \"Alice\"", table.Rows[0]["name"])
		}
		if table.Rows[1]["active"] != false {
			t.Errorf("row[1][\"active\"] = %v, want false", table.Rows[1]["active"])
		}
	})

	t.Run("valores nulos", func(t *testing.T) {
		table, err := LoadCSV("test", strings.NewReader(csvData), nil)
		if err != nil {
			t.Fatalf("LoadCSV() error = %v", err)
		}
		if table.Rows[2]["name"] != nil {
			t.Errorf("row[2][\"name\"] = %v, want nil", table.Rows[2]["name"])
		}
		if table.Rows[2]["score"] != nil {
			t.Errorf("row[2][\"score\"] = %v, want nil", table.Rows[2]["score"])
		}
	})

	t.Run("tipos explicitos", func(t *testing.T) {
		headerTypes := map[string]TypeTag{
			"id":    TypeInteger,
			"name":  TypeText,
			"score": TypeDecimal,
		}
		table, err := LoadCSV("test", strings.NewReader(csvData), headerTypes)
		if err != nil {
			t.Fatalf("LoadCSV() error = %v", err)
		}
		if table.Schema.Columns[3].Type != TypeText {
			t.Errorf("unspecified column type = %v, want TypeText", table.Schema.Columns[3].Type)
		}
	})
}

func TestLoadCVErrores(t *testing.T) {
	tests := []struct {
		name string
		csv  string
	}{
		{"csv vacío", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadCSV("test", strings.NewReader(tt.csv), nil)
			if err == nil {
				t.Error("LoadCSV() esperaba error, obtuvo nil")
			}
		})
	}
}

func TestLoadCSVHeaderOnly(t *testing.T) {
	csv := "id,name"
	table, err := LoadCSV("test", strings.NewReader(csv), nil)
	if err != nil {
		t.Fatalf("LoadCSV() error = %v", err)
	}
	if len(table.Rows) != 0 {
		t.Errorf("len(rows) = %d, want 0", len(table.Rows))
	}
	if len(table.Schema.Columns) != 2 {
		t.Errorf("len(columns) = %d, want 2", len(table.Schema.Columns))
	}
}

func TestSchema(t *testing.T) {
	schema := NewSchema([]Column{
		{Name: "id", Type: TypeInteger},
		{Name: "Name", Type: TypeText},
	})

	t.Run("HasColumn", func(t *testing.T) {
		if !schema.HasColumn("id") {
			t.Error("HasColumn(\"id\") = false, want true")
		}
		if !schema.HasColumn("NAME") {
			t.Error("HasColumn(\"NAME\") = false, want true (case insensitive)")
		}
		if schema.HasColumn("missing") {
			t.Error("HasColumn(\"missing\") = true, want false")
		}
	})

	t.Run("ColumnIndex", func(t *testing.T) {
		idx, ok := schema.ColumnIndex("name")
		if !ok || idx != 1 {
			t.Errorf("ColumnIndex(\"name\") = (%d, %v), want (1, true)", idx, ok)
		}
	})

	t.Run("ColumnByName", func(t *testing.T) {
		col, ok := schema.ColumnByName("ID")
		if !ok || col.Type != TypeInteger {
			t.Errorf("ColumnByName(\"ID\") = (%v, %v), want {Name:id Type:INTEGER}, true", col, ok)
		}
	})
}

func TestCatalog(t *testing.T) {
	c := NewCatalog()

	csvData := `id,name
1,Alice
2,Bob`

	table, err := LoadCSV("users", strings.NewReader(csvData), nil)
	if err != nil {
		t.Fatalf("LoadCSV() error = %v", err)
	}

	c.AddTable(table)

	t.Run("GetTable", func(t *testing.T) {
		got, ok := c.GetTable("USERS")
		if !ok || got.Name != "users" {
			t.Errorf("GetTable(\"USERS\") = (%v, %v), want users, true", got, ok)
		}
	})

	t.Run("GetTable missing", func(t *testing.T) {
		_, ok := c.GetTable("missing")
		if ok {
			t.Error("GetTable(\"missing\") = ok, want false")
		}
	})

	t.Run("TableNames", func(t *testing.T) {
		names := c.TableNames()
		if len(names) != 1 || names[0] != "users" {
			t.Errorf("TableNames() = %v, want [users]", names)
		}
	})

	t.Run("TableCount", func(t *testing.T) {
		if c.TableCount() != 1 {
			t.Errorf("TableCount() = %d, want 1", c.TableCount())
		}
	})
}

func TestLoadCSVFile(t *testing.T) {
	_, err := LoadCSVFile("employees", "../../data/employees.csv", nil)
	if err != nil {
		t.Fatalf("LoadCSVFile() error = %v", err)
	}
}

func TestLoadCSVFileNotFound(t *testing.T) {
	_, err := LoadCSVFile("missing", "../../data/nonexistent.csv", nil)
	if err == nil {
		t.Error("LoadCSVFile() esperaba error, obtuvo nil")
	}
}
