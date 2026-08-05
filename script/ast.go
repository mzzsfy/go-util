package script

// ========== AST抽象语法树定义 ==========
// 本文件包含所有AST节点类型定义
// 包括：核心接口、表达式节点、语句节点、类型表达式

// ========== 核心接口 ==========

// Node AST节点接口
type Node interface {
	Pos() Position
	String() string
}

// Expr 表达式接口
type Expr interface {
	Node
	exprNode()
}

// Stmt 语句接口
type Stmt interface {
	Node
	stmtNode()
}

// TypeExpr 类型表达式接口
type TypeExpr interface {
	Node
	typeExprNode()
}

// Program 程序节点，AST的根节点
type Program struct {
	// Statements 顶层语句列表
	Statements []Stmt
}

func (p *Program) Pos() Position {
	if len(p.Statements) > 0 {
		return p.Statements[0].Pos()
	}
	return Position{Line: 1, Column: 1}
}

func (p *Program) String() string { return "[program]" }

// ========== 类型表达式节点 ==========

// BaseTypeExpr 基础类型表达式
// 表示基本数据类型的类型声明，如 int, float, string, bool
type BaseTypeExpr struct {
	Position
	// Name 类型名称，如 "int", "float", "string", "bool"
	Name string
}

func (e *BaseTypeExpr) typeExprNode()  {}
func (e *BaseTypeExpr) Pos() Position  { return e.Position }
func (e *BaseTypeExpr) String() string { return e.Name }

// ArrayTypeExpr 数组类型表达式
// 表示数组类型的类型声明
type ArrayTypeExpr struct {
	Position
	// ElemType 元素类型
	ElemType TypeExpr
}

func (e *ArrayTypeExpr) typeExprNode()  {}
func (e *ArrayTypeExpr) Pos() Position  { return e.Position }
func (e *ArrayTypeExpr) String() string { return "[]" + e.ElemType.String() }

// ========== 表达式节点 ==========

// IdentExpr 标识符表达式
// 表示变量名、函数名等标识符
type IdentExpr struct {
	Position
	// Name 标识符名称
	Name string
}

func (e *IdentExpr) exprNode()      {}
func (e *IdentExpr) Pos() Position  { return e.Position }
func (e *IdentExpr) String() string { return e.Name }

// LiteralType 字面量类型枚举
// 定义脚本支持的字面量值类型
type LiteralType int

const (
	// LiteralInt 整数字面量，如 123, 0xFF
	LiteralInt LiteralType = iota
	// LiteralFloat 浮点数字面量，如 3.14, 1.5
	LiteralFloat
	// LiteralString 字符串字面量，如 "hello"
	LiteralString
	// LiteralBool 布尔字面量，true 或 false
	LiteralBool
	// LiteralNil nil字面量
	LiteralNil
)

// LiteralExpr 字面量表达式
// 表示源码中的字面量值，如数字、字符串、布尔值
type LiteralExpr struct {
	Position
	// Type 字面量类型
	Type LiteralType
	// Value 字面量的实际值，类型由Type字段决定
	Value any
}

func (e *LiteralExpr) exprNode()      {}
func (e *LiteralExpr) Pos() Position  { return e.Position }
func (e *LiteralExpr) String() string { return formatValue(e.Value) }

// ArrayExpr 数组字面量表达式
// 表示数组字面量，如 [1, 2, 3]
type ArrayExpr struct {
	Position
	// Elements 数组元素列表
	Elements []Expr
}

func (e *ArrayExpr) exprNode()      {}
func (e *ArrayExpr) Pos() Position  { return e.Position }
func (e *ArrayExpr) String() string { return "[array]" }

// MapExpr Map字面量表达式
// 表示Map字面量，如 {"key": "value"}
type MapExpr struct {
	Position
	// Pairs 键值对列表
	Pairs []MapPair
}

func (e *MapExpr) exprNode()      {}
func (e *MapExpr) Pos() Position  { return e.Position }
func (e *MapExpr) String() string { return "[map]" }

// MapPair Map键值对
// 表示Map中的一个键值对
type MapPair struct {
	// Key 键表达式
	Key Expr
	// Value 值表达式
	Value Expr
}

// BinaryExpr 二元运算表达式
// 表示二元运算，如 a + b, a == b
type BinaryExpr struct {
	Position
	// Left 左操作数
	Left Expr
	// Operator 运算符，如 "+", "-", "==", "<"
	Operator string
	// Right 右操作数
	Right Expr
}

func (e *BinaryExpr) exprNode()     {}
func (e *BinaryExpr) Pos() Position { return e.Position }
func (e *BinaryExpr) String() string {
	return "(" + e.Left.String() + " " + e.Operator + " " + e.Right.String() + ")"
}

// UnaryExpr 一元运算表达式
// 表示一元运算，如 -a, !b
type UnaryExpr struct {
	Position
	// Operator 运算符，如 "-", "!"
	Operator string
	// Operand 操作数
	Operand Expr
}

func (e *UnaryExpr) exprNode()      {}
func (e *UnaryExpr) Pos() Position  { return e.Position }
func (e *UnaryExpr) String() string { return e.Operator + e.Operand.String() }

// IndexExpr 索引访问表达式
// 表示索引访问，如 arr[i], map["key"]
type IndexExpr struct {
	Position
	// Object 被访问的对象
	Object Expr
	// Index 索引表达式
	Index Expr
}

func (e *IndexExpr) exprNode()      {}
func (e *IndexExpr) Pos() Position  { return e.Position }
func (e *IndexExpr) String() string { return e.Object.String() + "[" + e.Index.String() + "]" }

// SliceExpr 切片表达式
// 表示切片操作，如 arr[start:end]
type SliceExpr struct {
	Position
	// Object 被切片的对象
	Object Expr
	// Start 起始索引，nil表示从开头
	Start Expr
	// End 结束索引，nil表示到末尾
	End Expr
}

func (e *SliceExpr) exprNode()      {}
func (e *SliceExpr) Pos() Position  { return e.Position }
func (e *SliceExpr) String() string { return "[slice]" }

// CallExpr 函数调用表达式
// 表示函数调用，如 func(a, b)
type CallExpr struct {
	Position
	// Func 被调用的函数表达式
	Func Expr
	// Args 参数列表
	Args []Expr
}

func (e *CallExpr) exprNode()      {}
func (e *CallExpr) Pos() Position  { return e.Position }
func (e *CallExpr) String() string { return "[call]" }

// IfExpr if表达式
// 表示if表达式，有返回值
type IfExpr struct {
	Position
	// Condition 条件表达式
	Condition Expr
	// Then 条件为真时执行的表达式
	Then Expr
	// Else 条件为假时执行的表达式，可能为nil
	Else Expr
}

func (e *IfExpr) exprNode()      {}
func (e *IfExpr) Pos() Position  { return e.Position }
func (e *IfExpr) String() string { return "[if-expr]" }

// ========== 语句节点 ==========

// VarDeclStmt 变量声明语句
// 表示变量声明，如 x := 1
type VarDeclStmt struct {
	Position
	// Name 变量名
	Name string
	// TypeAnnot 类型注解，可能为nil
	TypeAnnot TypeExpr
	// Value 初始值表达式
	Value Expr
}

func (s *VarDeclStmt) stmtNode()      {}
func (s *VarDeclStmt) Pos() Position  { return s.Position }
func (s *VarDeclStmt) String() string { return s.Name + " := " + s.Value.String() }

// ExprStmt 表达式语句
// 表示作为语句的表达式
type ExprStmt struct {
	Position
	// Expr 表达式
	Expr Expr
}

func (s *ExprStmt) stmtNode()      {}
func (s *ExprStmt) Pos() Position  { return s.Position }
func (s *ExprStmt) String() string { return s.Expr.String() }

// BlockStmt 块语句
// 表示由大括号包围的语句序列
type BlockStmt struct {
	Position
	// Statements 块内的语句列表
	Statements []Stmt
}

func (s *BlockStmt) stmtNode()      {}
func (s *BlockStmt) Pos() Position  { return s.Position }
func (s *BlockStmt) String() string { return "[block]" }

// IfStmt if条件语句
// 表示if条件语句
type IfStmt struct {
	Position
	// Condition 条件表达式
	Condition Expr
	// Then 条件为真时执行的块
	Then *BlockStmt
	// Else 条件为假时执行的语句，可能是nil或BlockStmt或另一个IfStmt
	Else Stmt
}

func (s *IfStmt) stmtNode()      {}
func (s *IfStmt) Pos() Position  { return s.Position }
func (s *IfStmt) String() string { return "[if]" }

// ForMode for循环模式
// 定义不同类型的for循环结构
type ForMode int

const (
	// ForCount 计数循环: for i := n
	ForCount ForMode = iota
	// ForWhile 条件循环: for cond
	ForWhile
	// ForStandard 标准for循环: for init; cond; post
	ForStandard
	// ForRange 遍历循环: for k := arr
	ForRange
)

// ForStmt for循环语句
// 表示for循环，支持多种模式
type ForStmt struct {
	Position
	// Init 初始化语句，用于ForStandard模式
	Init Stmt
	// Cond 循环条件
	Cond Expr
	// Post 后置语句，用于ForStandard模式
	Post Stmt
	// Body 循环体
	Body *BlockStmt
	// Mode 循环模式
	Mode ForMode
	// RangeValueVar range双变量模式的value变量名
	// 非空时表示 for k, v := range arr 中的 v
	RangeValueVar string
}

func (s *ForStmt) stmtNode()      {}
func (s *ForStmt) Pos() Position  { return s.Position }
func (s *ForStmt) String() string { return "[for]" }

// Param 函数参数
// 表示函数的参数定义
type Param struct {
	// Name 参数名
	Name string
	// TypeAnnot 参数类型注解，可能为nil
	TypeAnnot TypeExpr
}

// FuncDeclStmt 函数定义语句
// 表示函数定义
type FuncDeclStmt struct {
	Position
	// Name 函数名
	Name string
	// Params 参数列表
	Params []Param
	// Return 返回类型注解，可能为nil
	Return TypeExpr
	// Body 函数体
	Body *BlockStmt
}

func (s *FuncDeclStmt) stmtNode()      {}
func (s *FuncDeclStmt) Pos() Position  { return s.Position }
func (s *FuncDeclStmt) String() string { return "fn " + s.Name }

// ReturnStmt return语句
// 表示函数返回
type ReturnStmt struct {
	Position
	// Value 返回值表达式，可能为nil
	Value Expr
}

func (s *ReturnStmt) stmtNode()      {}
func (s *ReturnStmt) Pos() Position  { return s.Position }
func (s *ReturnStmt) String() string { return "[return]" }

// BreakStmt break语句
// 表示循环中的break
type BreakStmt struct{ Position }

func (s *BreakStmt) stmtNode()      {}
func (s *BreakStmt) Pos() Position  { return s.Position }
func (s *BreakStmt) String() string { return "break" }

// ContinueStmt continue语句
// 表示循环中的continue
type ContinueStmt struct{ Position }

func (s *ContinueStmt) stmtNode()      {}
func (s *ContinueStmt) Pos() Position  { return s.Position }
func (s *ContinueStmt) String() string { return "continue" }

// ThrowStmt throw语句
// 表示抛出异常
type ThrowStmt struct {
	Position
	// Value 要抛出的值
	Value Expr
}

func (s *ThrowStmt) stmtNode()      {}
func (s *ThrowStmt) Pos() Position  { return s.Position }
func (s *ThrowStmt) String() string { return "[throw]" }

// DefDirectiveStmt 外部函数声明指令
// 表示#fn指令，用于声明外部函数
type DefDirectiveStmt struct {
	Position
	// Name 函数名
	Name string
	// Params 参数列表
	Params []Param
	// Return 返回类型注解
	Return TypeExpr
}

func (s *DefDirectiveStmt) stmtNode()      {}
func (s *DefDirectiveStmt) Pos() Position  { return s.Position }
func (s *DefDirectiveStmt) String() string { return "#fn " + s.Name }
