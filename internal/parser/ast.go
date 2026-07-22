package parser

type NodeType int

const (
	NodeSelect NodeType = iota
	NodeExpr
	NodeColumnRef
	NodeStringLit
	NodeNumberLit
	NodeBoolLit
	NodeNullLit
	NodeBinaryOp
	NodeComparison
)

type Node interface {
	Type() NodeType
}

type SelectNode struct {
	Columns []Node        // columnas a mostrar (* o lista)
	Table   string        // nombre de la tabla
	Where   Node          // condición WHERE (nil si no hay)
	OrderBy []OrderByItem // ordenamiento (nil si en caso no hay ORDERBY)
	Limit   *int          //limite de filas (nil en aso de que no haya limite)
}

func (n *SelectNode) Type() NodeType { return NodeSelect }

type ColumnRefNode struct {
	Name string
}

func (n *ColumnRefNode) Type() NodeType { return NodeColumnRef }

type StringLitNode struct {
	Value string
}

func (n *StringLitNode) Type() NodeType { return NodeStringLit }

type NumberLitNode struct {
	Value string
}

func (n *NumberLitNode) Type() NodeType { return NodeNumberLit }

type BoolLitNode struct {
	Value bool
}

func (n *BoolLitNode) Type() NodeType { return NodeBoolLit }

type NullLitNode struct{}

func (n *NullLitNode) Type() NodeType { return NodeNullLit }

type BinaryOpNode struct {
	Op    string // AND, OR
	Left  Node
	Right Node
}

func (n *BinaryOpNode) Type() NodeType { return NodeExpr }

type ComparisonNode struct {
	Op    string // =, <>, <, >, <=, >=
	Left  Node
	Right Node
}

func (n *ComparisonNode) Type() NodeType { return NodeComparison }

type OrderByItem struct {
	Column string
	Asc    bool
}
