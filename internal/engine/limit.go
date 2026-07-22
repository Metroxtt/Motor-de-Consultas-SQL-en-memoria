package engine

import "github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/catalog"

type LimitOperator struct {
	input Operator
	limit int
	count int
}

func NewLimitOperator(input Operator, limit int) *LimitOperator {
	return &LimitOperator{input: input, limit: limit}
}

func (l *LimitOperator) Next() (catalog.Row, error) {
	if l.count >= l.limit {
		return nil, nil
	}
	row, err := l.input.Next()
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	l.count++
	return row, nil
}

func (l *LimitOperator) Close() error {
	return l.input.Close()
}
