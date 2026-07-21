package engine

import "github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/catalog"

type Operator interface {
	Next() (catalog.Row, error)
	Close() error
}
