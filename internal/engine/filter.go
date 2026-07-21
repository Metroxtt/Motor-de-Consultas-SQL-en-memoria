package engine

import (
	"fmt"

	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/catalog"
	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/parser"
)

type FilterOperator struct {
	input Operator
	where parser.Node
}

func NewFilterOperator(input Operator, where parser.Node) *FilterOperator {
	return &FilterOperator{input: input, where: where}
}

func (f *FilterOperator) Next() (catalog.Row, error) {
	for {
		row, err := f.input.Next()
		if err != nil {
			return nil, err
		}
		if row == nil {
			return nil, nil
		}

		result, err := EvalExpr(f.where, row)
		if err != nil {
			return nil, fmt.Errorf("error evaluando WHERE: %w", err)
		}

		if result == nil {
			continue
		}

		if boolVal, ok := result.(bool); ok && boolVal {
			return row, nil
		}
	}
}

func (f *FilterOperator) Close() error {
	return f.input.Close()
}
