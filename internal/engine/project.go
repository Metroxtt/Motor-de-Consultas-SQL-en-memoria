package engine

import (
	"fmt"

	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/catalog"
	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/parser"
)

type ProjectOperator struct {
	input   Operator
	columns []parser.Node
	allCols bool
	schema  catalog.Schema
}

func NewProjectOperator(input Operator, columns []parser.Node, schema catalog.Schema) *ProjectOperator {
	return &ProjectOperator{
		input:   input,
		columns: columns,
		allCols: columns == nil,
		schema:  schema,
	}
}

func (p *ProjectOperator) Next() (catalog.Row, error) {
	row, err := p.input.Next()
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}

	if p.allCols {
		return row, nil
	}

	result := make(catalog.Row, len(p.columns))
	for _, col := range p.columns {
		var name string
		var val interface{}

		if agg, ok := col.(*parser.AggregateNode); ok {
			name = agg.Func + "(" + agg.Column + ")"
			val = row[name]
		} else {
			var err error
			val, err = EvalExpr(col, row)
			if err != nil {
				return nil, fmt.Errorf("error en proyección: %w", err)
			}
			if ref, ok := col.(*parser.ColumnRefNode); ok {
				name = ref.Name
			} else {
				name = fmt.Sprintf("%v", col)
			}
		}
		result[name] = val
	}

	return result, nil
}

func (p *ProjectOperator) Close() error {
	return p.input.Close()
}
