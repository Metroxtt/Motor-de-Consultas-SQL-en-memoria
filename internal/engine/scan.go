package engine

import (
	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/catalog"
)

type ScanOperator struct {
	table *catalog.Table
	pos   int
}

func NewScanOperator(table *catalog.Table) *ScanOperator {
	return &ScanOperator{table: table, pos: 0}
}

func (s *ScanOperator) Next() (catalog.Row, error) {
	if s.pos >= len(s.table.Rows) {
		return nil, nil
	}
	row := s.table.Rows[s.pos]
	s.pos++
	return row, nil
}

func (s *ScanOperator) Close() error {
	return nil
}
