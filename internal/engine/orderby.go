package engine

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/catalog"
	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/parser"
)

type OrderOperator struct {
	input  Operator
	items  []parser.OrderByItem
	rows   []catalog.Row // buffer de todas las filas
	pos    int           // posición actual en el buffer
	loaded bool          // ya se cargaron las filas
}

func NewOrderOperator(input Operator, items []parser.OrderByItem) *OrderOperator {
	return &OrderOperator{input: input, items: items}
}

func (o *OrderOperator) Next() (catalog.Row, error) {
	if !o.loaded {
		if err := o.loadAll(); err != nil {
			return nil, err
		}
		o.loaded = true
	}
	if o.pos >= len(o.rows) {
		return nil, nil
	}
	row := o.rows[o.pos]
	o.pos++
	return row, nil
}
func (o *OrderOperator) loadAll() error {
	for {
		row, err := o.input.Next()
		if err != nil {
			return err
		}
		if row == nil {
			break
		}
		o.rows = append(o.rows, row)
	}
	if len(o.rows) > 0 {
		for _, item := range o.items {
			if _, ok := o.rows[0][item.Column]; !ok {
				return fmt.Errorf("columna %q no encontrada para ORDER BY", item.Column)
			}
		}
	}
	sort.Slice(o.rows, func(i, j int) bool {
		for _, item := range o.items {
			vi := o.rows[i][item.Column]
			vj := o.rows[j][item.Column]

			cmp := compareValues(vi, vj)
			if cmp != 0 {
				if item.Asc {
					return cmp < 0
				}
				return cmp > 0
			}
		}
		return false
	})

	return nil
}
func (o *OrderOperator) Close() error {
	return o.input.Close()
}

func compareValues(a, b interface{}) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	aNum, aOk := toFloat64(a)
	bNum, bOk := toFloat64(b)
	if aOk && bOk {
		if aNum < bNum {
			return -1
		}
		if aNum > bNum {
			return 1
		}
		return 0
	}

	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.Compare(aStr, bStr)
}
