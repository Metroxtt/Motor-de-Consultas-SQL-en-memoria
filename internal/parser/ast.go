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
	NodeJoin
)

type Node interface {
	Type() NodeType
}

type JoinType int

const (
	InnerJoin JoinType = iota
)

type JoinNode struct {
	JoinType    JoinType
	RightTable  string
	OnCondition Node
}

func (n *JoinNode) Type() NodeType { return NodeJoin }

type SelectNode struct {
	Columns []Node      // columnas a mostrar (* o lista)
	Table   string      // nombre de la tabla
	Joins   []*JoinNode // lista de JOINs
	Where   Node        // condición WHERE (nil si no hay)
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
