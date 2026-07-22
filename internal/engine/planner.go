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

	if hasAggregates(node) {
		aggs := collectAggregates(node.Columns)
		op = NewAggregateOperator(op, node.GroupBy, aggs, node.Columns)
	}

	op = NewProjectOperator(op, node.Columns, table.Schema)

	return op, nil
}

func hasAggregates(node *parser.SelectNode) bool {
	if node.GroupBy != "" {
		return true
	}
	for _, col := range node.Columns {
		if _, ok := col.(*parser.AggregateNode); ok {
			return true
		}
	}
	return false
}

func collectAggregates(columns []parser.Node) []*parser.AggregateNode {
	var aggs []*parser.AggregateNode
	for _, col := range columns {
		if agg, ok := col.(*parser.AggregateNode); ok {
			aggs = append(aggs, agg)
		}
	}
	return aggs
}
