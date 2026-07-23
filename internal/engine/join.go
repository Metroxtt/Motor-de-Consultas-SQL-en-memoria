package engine

import (
	"fmt"
	"strings"

	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/catalog"
	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/parser"
)

// joinRows fusiona dos filas en una nueva prefijando con el nombre de la tabla.
func joinRows(leftName string, left catalog.Row, rightName string, right catalog.Row) catalog.Row {
	result := make(catalog.Row, len(left)+len(right))
	for k, v := range left {
		if !strings.Contains(k, ".") && leftName != "" {
			result[leftName+"."+k] = v
		} else {
			result[k] = v
		}
	}
	for k, v := range right {
		if !strings.Contains(k, ".") && rightName != "" {
			result[rightName+"."+k] = v
		} else {
			result[k] = v
		}
	}
	return result
}

// ==========================================
// NestedLoopJoinOperator
// ==========================================

type NestedLoopJoinOperator struct {
	left        Operator
	right       Operator
	leftTable   string
	rightTable  string
	onCondition parser.Node

	rightRows   []catalog.Row
	leftRow     catalog.Row
	rightIdx    int
	initialized bool
}

func NewNestedLoopJoinOperator(left, right Operator, leftTable, rightTable string, onCondition parser.Node) *NestedLoopJoinOperator {
	return &NestedLoopJoinOperator{
		left:        left,
		right:       right,
		leftTable:   leftTable,
		rightTable:  rightTable,
		onCondition: onCondition,
	}
}

func (op *NestedLoopJoinOperator) init() error {
	for {
		rRow, err := op.right.Next()
		if err != nil {
			return err
		}
		if rRow == nil {
			break
		}
		op.rightRows = append(op.rightRows, rRow)
	}
	op.initialized = true
	return nil
}

func (op *NestedLoopJoinOperator) Next() (catalog.Row, error) {
	if !op.initialized {
		if err := op.init(); err != nil {
			return nil, err
		}
	}

	for {
		if op.leftRow == nil {
			lRow, err := op.left.Next()
			if err != nil {
				return nil, err
			}
			if lRow == nil {
				return nil, nil // Fin del lado izquierdo
			}
			op.leftRow = lRow
			op.rightIdx = 0
		}

		for op.rightIdx < len(op.rightRows) {
			rRow := op.rightRows[op.rightIdx]
			op.rightIdx++

			merged := joinRows(op.leftTable, op.leftRow, op.rightTable, rRow)

			// Evaluar ON condition
			val, err := EvalExpr(op.onCondition, merged)
			if err != nil {
				return nil, err
			}

			if ok, _ := val.(bool); ok {
				return merged, nil
			}
		}

		// Hemos agotado la derecha para esta fila de la izquierda
		op.leftRow = nil
	}
}

func (op *NestedLoopJoinOperator) Close() error {
	if err := op.left.Close(); err != nil {
		return err
	}
	return op.right.Close()
}

// ==========================================
// HashJoinOperator
// ==========================================

type HashJoinOperator struct {
	left        Operator
	right       Operator
	leftTable   string
	rightTable  string
	onCondition parser.Node

	leftColName  string
	rightColName string

	hashTable    map[interface{}][]catalog.Row
	leftRow      catalog.Row
	rightMatches []catalog.Row
	matchIdx     int
	initialized  bool
}

func NewHashJoinOperator(left, right Operator, leftTable, rightTable string, onCondition parser.Node) *HashJoinOperator {
	return &HashJoinOperator{
		left:        left,
		right:       right,
		leftTable:   leftTable,
		rightTable:  rightTable,
		onCondition: onCondition,
	}
}

func getColNameWithoutPrefix(colName string) string {
	parts := strings.Split(colName, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return colName
}

func (op *HashJoinOperator) extractColumns() error {
	comp, ok := op.onCondition.(*parser.ComparisonNode)
	if !ok || comp.Op != "=" {
		return fmt.Errorf("HashJoin requiere una condicion de igualdad (T1.col = T2.col)")
	}

	lCol, okL := comp.Left.(*parser.ColumnRefNode)
	rCol, okR := comp.Right.(*parser.ColumnRefNode)

	if !okL || !okR {
		return fmt.Errorf("HashJoin requiere referencias a columnas en ambos lados del ON")
	}

	// Guardamos los nombres originales para buscar en la fila ya pre-fixada
	// o buscar en la fila del ScanOperator que no tiene prefijos
	op.leftColName = getColNameWithoutPrefix(lCol.Name)
	op.rightColName = getColNameWithoutPrefix(rCol.Name)

	return nil
}

func (op *HashJoinOperator) init() error {
	if err := op.extractColumns(); err != nil {
		return err
	}

	op.hashTable = make(map[interface{}][]catalog.Row)

	for {
		rRow, err := op.right.Next()
		if err != nil {
			return err
		}
		if rRow == nil {
			break
		}

		// Intentar obtener el valor de rightColName
		keyVal, ok := rRow[op.rightColName]
		if !ok {
			// Es posible que el orden de ON t1.col = t2.col esté invertido en el AST
			keyVal, ok = rRow[op.leftColName]
			if ok {
				// Intercambiamos left y right
				op.leftColName, op.rightColName = op.rightColName, op.leftColName
			} else {
				return fmt.Errorf("no se encontro columna de JOIN %s ni %s en la tabla derecha", op.rightColName, op.leftColName)
			}
		}

		op.hashTable[keyVal] = append(op.hashTable[keyVal], rRow)
	}

	op.initialized = true
	return nil
}

func (op *HashJoinOperator) Next() (catalog.Row, error) {
	if !op.initialized {
		if err := op.init(); err != nil {
			return nil, err
		}
	}

	for {
		if op.matchIdx < len(op.rightMatches) {
			rRow := op.rightMatches[op.matchIdx]
			op.matchIdx++
			return joinRows(op.leftTable, op.leftRow, op.rightTable, rRow), nil
		}

		lRow, err := op.left.Next()
		if err != nil {
			return nil, err
		}
		if lRow == nil {
			return nil, nil // EOF
		}

		op.leftRow = lRow
		keyVal, ok := lRow[op.leftColName]
		if !ok {
			return nil, fmt.Errorf("no se encontro la columna %s en la tabla izquierda", op.leftColName)
		}

		op.rightMatches = op.hashTable[keyVal]
		op.matchIdx = 0
	}
}

func (op *HashJoinOperator) Close() error {
	if err := op.left.Close(); err != nil {
		return err
	}
	return op.right.Close()
}
