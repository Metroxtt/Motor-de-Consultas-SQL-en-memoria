package catalog

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

type Table struct {
	Name   string
	Schema Schema
	Rows   []Row
}

type Catalog struct {
	tables map[string]*Table
}

func NewCatalog() *Catalog {
	return &Catalog{tables: make(map[string]*Table)}
}

func (c *Catalog) AddTable(t *Table) {
	c.tables[strings.ToLower(t.Name)] = t
}

func (c *Catalog) GetTable(name string) (*Table, bool) {
	t, ok := c.tables[strings.ToLower(name)]
	return t, ok
}

func (c *Catalog) TableNames() []string {
	names := make([]string, 0, len(c.tables))
	for name := range c.tables {
		names = append(names, name)
	}
	return names
}

func (c *Catalog) TableCount() int {
	return len(c.tables)
}

func LoadCSV(tableName string, r io.Reader, headerTypes map[string]TypeTag) (*Table, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV for table %q: %w", tableName, err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV for table %q is empty", tableName)
	}

	header := records[0]
	for i, h := range header {
		header[i] = strings.TrimSpace(h)
	}

	schemaCols := make([]Column, len(header))
	if headerTypes != nil {
		for i, h := range header {
			tag, ok := headerTypes[h]
			if !ok {
				tag = TypeText
			}
			schemaCols[i] = Column{Name: h, Type: tag}
		}
	} else {
		transposed := make([][]string, len(header))
		for _, row := range records[1:] {
			for i, val := range row {
				if i < len(transposed) {
					transposed[i] = append(transposed[i], val)
				}
			}
		}
		for i, h := range header {
			schemaCols[i] = Column{Name: h, Type: InferType(transposed[i])}
		}
	}

	schema := NewSchema(schemaCols)
	rows := make([]Row, 0, len(records)-1)

	for lineNum, record := range records[1:] {
		row := make(Row, len(header))
		for i, raw := range record {
			if i >= len(header) {
				break
			}
			val, err := ParseValue(raw, schemaCols[i].Type)
			if err != nil {
				return nil, fmt.Errorf("table %q, row %d, column %q: %w", tableName, lineNum+2, header[i], err)
			}
			row[header[i]] = val
		}
		rows = append(rows, row)
	}

	return &Table{Name: tableName, Schema: schema, Rows: rows}, nil
}

func LoadCSVFile(tableName, filepath string, headerTypes map[string]TypeTag) (*Table, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("opening CSV file %q: %w", filepath, err)
	}
	defer f.Close()
	return LoadCSV(tableName, f, headerTypes)
}
