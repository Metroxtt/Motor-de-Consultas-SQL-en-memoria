package engine

import (
	"fmt"

	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/catalog"
	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/parser"
)

func BuildPlan(node *parser.SelectNode, cat *catalog.Catalog) (Operator, error) {
	table, ok := cat.GetTable(node.Table)
	if !ok {
		return nil, fmt.Errorf("tabla %q no encontrada", node.Table)
	}

	var op Operator = NewScanOperator(table)

	if node.Where != nil {
		op = NewFilterOperator(op, node.Where)
	}

	op = NewProjectOperator(op, node.Columns, table.Schema)

	if len(node.OrderBy) > 0 {
		op = NewOrderOperator(op, node.OrderBy)
	}

	if node.Limit != nil {
		op = NewLimitOperator(op, *node.Limit)
	}
	return op, nil
}
