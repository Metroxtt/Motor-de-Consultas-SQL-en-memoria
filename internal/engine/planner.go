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

	if len(node.Joins) > 0 {
		for _, join := range node.Joins {
			rTable, ok := cat.GetTable(join.RightTable)
			if !ok {
				return nil, fmt.Errorf("tabla derecha de join %q no encontrada", join.RightTable)
			}
			rOp := NewScanOperator(rTable)
			
			// Por defecto usamos HashJoin ya que es más rápido O(N+M).
			// Para usar NestedLoopJoin sería: op = NewNestedLoopJoinOperator(op, rOp, join.OnCondition)
			op = NewHashJoinOperator(op, rOp, join.OnCondition)
		}
	}

	if node.Where != nil {
		op = NewFilterOperator(op, node.Where)
	}

	op = NewProjectOperator(op, node.Columns, table.Schema)

	return op, nil
}
