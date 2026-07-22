package engine

import (
	"fmt"

	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/catalog"
	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/parser"
)

type AggregateOperator struct {
	input     Operator
	groupBy   string
	aggs      []*parser.AggregateNode
	columns   []parser.Node
	groups    map[string][]catalog.Row
	groupKeys []string
	pos       int
}

func NewAggregateOperator(input Operator, groupBy string, aggs []*parser.AggregateNode, columns []parser.Node) *AggregateOperator {
	return &AggregateOperator{
		input:   input,
		groupBy: groupBy,
		aggs:    aggs,
		columns: columns,
	}
}

func (a *AggregateOperator) Next() (catalog.Row, error) {
	if a.groups == nil {
		if err := a.loadAndGroup(); err != nil {
			return nil, err
		}
	}

	if a.pos >= len(a.groupKeys) {
		return nil, nil
	}

	key := a.groupKeys[a.pos]
	rows := a.groups[key]
	a.pos++

	return a.evalGroup(key, rows)
}

func (a *AggregateOperator) Close() error {
	return a.input.Close()
}

func (a *AggregateOperator) loadAndGroup() error {
	a.groups = make(map[string][]catalog.Row)
	a.groupKeys = make([]string, 0)

	for {
		row, err := a.input.Next()
		if err != nil {
			return err
		}
		if row == nil {
			break
		}

		var key string
		if a.groupBy != "" {
			val, ok := row[a.groupBy]
			if !ok {
				return fmt.Errorf("columna GROUP BY %q no encontrada", a.groupBy)
			}
			key = fmt.Sprintf("%v", val)
		} else {
			key = "__all__"
		}

		if _, exists := a.groups[key]; !exists {
			a.groupKeys = append(a.groupKeys, key)
		}
		a.groups[key] = append(a.groups[key], row)
	}

	return nil
}

func (a *AggregateOperator) evalGroup(key string, rows []catalog.Row) (catalog.Row, error) {
	result := make(catalog.Row)

	if a.groupBy != "" {
		result[a.groupBy] = parseGroupKey(key)
	}

	for _, agg := range a.aggs {
		val := a.evalAgg(agg, rows)
		name := agg.Func + "(" + agg.Column + ")"
		result[name] = val
	}

	return result, nil
}

func (a *AggregateOperator) evalAgg(agg *parser.AggregateNode, rows []catalog.Row) interface{} {
	switch agg.Func {
	case "COUNT":
		return a.evalCount(agg.Column, rows)
	case "SUM":
		return a.evalSum(agg.Column, rows)
	case "AVG":
		return a.evalAvg(agg.Column, rows)
	case "MIN":
		return a.evalMin(agg.Column, rows)
	case "MAX":
		return a.evalMax(agg.Column, rows)
	}
	return nil
}

func (a *AggregateOperator) evalCount(col string, rows []catalog.Row) int64 {
	if col == "*" {
		return int64(len(rows))
	}

	var count int64
	for _, row := range rows {
		if row[col] != nil {
			count++
		}
	}
	return count
}

func (a *AggregateOperator) evalSum(col string, rows []catalog.Row) interface{} {
	var sum float64
	var hasValue bool

	for _, row := range rows {
		val := row[col]
		if val == nil {
			continue
		}
		if num, ok := toFloat64(val); ok {
			sum += num
			hasValue = true
		}
	}

	if !hasValue {
		return nil
	}
	if isInteger(rows, col) {
		return int64(sum)
	}
	return sum
}

func (a *AggregateOperator) evalAvg(col string, rows []catalog.Row) interface{} {
	var sum float64
	var count int64

	for _, row := range rows {
		val := row[col]
		if val == nil {
			continue
		}
		if num, ok := toFloat64(val); ok {
			sum += num
			count++
		}
	}

	if count == 0 {
		return nil
	}
	avg := sum / float64(count)
	if isInteger(rows, col) {
		return int64(avg)
	}
	return avg
}

func (a *AggregateOperator) evalMin(col string, rows []catalog.Row) interface{} {
	var minVal float64
	var hasValue bool

	for _, row := range rows {
		val := row[col]
		if val == nil {
			continue
		}
		if num, ok := toFloat64(val); ok {
			if !hasValue || num < minVal {
				minVal = num
				hasValue = true
			}
		}
	}

	if !hasValue {
		return nil
	}
	if isInteger(rows, col) {
		return int64(minVal)
	}
	return minVal
}

func (a *AggregateOperator) evalMax(col string, rows []catalog.Row) interface{} {
	var maxVal float64
	var hasValue bool

	for _, row := range rows {
		val := row[col]
		if val == nil {
			continue
		}
		if num, ok := toFloat64(val); ok {
			if !hasValue || num > maxVal {
				maxVal = num
				hasValue = true
			}
		}
	}

	if !hasValue {
		return nil
	}
	if isInteger(rows, col) {
		return int64(maxVal)
	}
	return maxVal
}

func isInteger(rows []catalog.Row, col string) bool {
	for _, row := range rows {
		if row[col] == nil {
			continue
		}
		switch row[col].(type) {
		case int64:
			continue
		default:
			return false
		}
	}
	return true
}

func parseGroupKey(key string) interface{} {
	if key == "<nil>" {
		return nil
	}
	return key
}
