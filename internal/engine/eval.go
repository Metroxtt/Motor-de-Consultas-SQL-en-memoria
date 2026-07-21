package engine

import (
	"fmt"
	"strings"

	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/catalog"
	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/parser"
)

func EvalExpr(node parser.Node, row catalog.Row) (interface{}, error) {
	if node == nil {
		return nil, nil
	}

	switch n := node.(type) {
	case *parser.ColumnRefNode:
		val, ok := row[n.Name]
		if !ok {
			return nil, fmt.Errorf("columna %q no encontrada", n.Name)
		}
		return val, nil

	case *parser.NumberLitNode:
		return catalog.ParseValue(n.Value, catalog.TypeDecimal)

	case *parser.StringLitNode:
		return n.Value, nil

	case *parser.BoolLitNode:
		return n.Value, nil

	case *parser.NullLitNode:
		return nil, nil

	case *parser.ComparisonNode:
		return evalComparison(n, row)

	case *parser.BinaryOpNode:
		return evalBinaryOp(n, row)

	default:
		return nil, fmt.Errorf("tipo de nodo no soportado: %T", node)
	}
}

func evalComparison(n *parser.ComparisonNode, row catalog.Row) (interface{}, error) {
	left, err := EvalExpr(n.Left, row)
	if err != nil {
		return nil, err
	}
	right, err := EvalExpr(n.Right, row)
	if err != nil {
		return nil, err
	}

	if left == nil || right == nil {
		return nil, nil
	}

	lNum, lIsNum := toFloat64(left)
	rNum, rIsNum := toFloat64(right)

	if lIsNum && rIsNum {
		switch n.Op {
		case "=":
			return lNum == rNum, nil
		case "<>":
			return lNum != rNum, nil
		case "<":
			return lNum < rNum, nil
		case ">":
			return lNum > rNum, nil
		case "<=":
			return lNum <= rNum, nil
		case ">=":
			return lNum >= rNum, nil
		}
	}

	lStr := fmt.Sprintf("%v", left)
	rStr := fmt.Sprintf("%v", right)

	switch n.Op {
	case "=":
		return strings.EqualFold(lStr, rStr), nil
	case "<>":
		return !strings.EqualFold(lStr, rStr), nil
	case "<":
		return lStr < rStr, nil
	case ">":
		return lStr > rStr, nil
	case "<=":
		return lStr <= rStr, nil
	case ">=":
		return lStr >= rStr, nil
	}

	return nil, fmt.Errorf("operador no soportado: %s", n.Op)
}

func evalBinaryOp(n *parser.BinaryOpNode, row catalog.Row) (interface{}, error) {
	left, err := EvalExpr(n.Left, row)
	if err != nil {
		return nil, err
	}
	right, err := EvalExpr(n.Right, row)
	if err != nil {
		return nil, err
	}

	leftBool, leftOk := toBool(left)
	rightBool, rightOk := toBool(right)

	if !leftOk || !rightOk {
		return nil, fmt.Errorf("AND/OR requieren valores booleanos")
	}

	switch n.Op {
	case "AND":
		return leftBool && rightBool, nil
	case "OR":
		return leftBool || rightBool, nil
	}

	return nil, fmt.Errorf("operador binario no soportado: %s", n.Op)
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int64:
		return float64(val), true
	case float64:
		return val, true
	case int:
		return float64(val), true
	default:
		return 0, false
	}
}

func toBool(v interface{}) (bool, bool) {
	switch val := v.(type) {
	case bool:
		return val, true
	default:
		return false, false
	}
}
