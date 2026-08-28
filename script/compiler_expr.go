package script

// ========== 表达式编译方法 ==========

// compileExpr 编译表达式
// 使用type switch分发，避免reflect.TypeOf和map查找开销
func (c *Compiler) compileExpr(expr Expr) error {
	switch e := expr.(type) {
	case *IdentExpr:
		return c.compileIdentExpr(e)
	case *LiteralExpr:
		return c.compileLiteralExpr(e)
	case *ArrayExpr:
		return c.compileArrayExpr(e)
	case *MapExpr:
		return c.compileMapExpr(e)
	case *BinaryExpr:
		return c.compileBinaryExpr(e)
	case *UnaryExpr:
		return c.compileUnaryExpr(e)
	case *IndexExpr:
		return c.compileIndexExpr(e)
	case *SliceExpr:
		return c.compileSliceExpr(e)
	case *CallExpr:
		return c.compileCallExpr(e)
	case *IfExpr:
		return c.compileIfExpr(e)
	}
	return NewCompileErrorFromPos(expr, "内部编译错误：无法识别的表达式类型")
}

// compileExprOrNil 编译表达式，如果表达式为nil则压入nil值
// 统一处理可能为nil的表达式编译场景
func (c *Compiler) compileExprOrNil(expr Expr) error {
	if expr == nil {
		c.emit(OpNil)
		return nil
	}
	return c.compileExpr(expr)
}

// compileExpr2 编译两个表达式（常用模式优化）
// 用于索引访问、Map键值对等场景
func (c *Compiler) compileExpr2(e1, e2 Expr) error {
	if err := c.compileExpr(e1); err != nil {
		return err
	}
	return c.compileExpr(e2)
}

// compileIdentExpr 编译标识符表达式
func (c *Compiler) compileIdentExpr(expr *IdentExpr) error {
	// 首先检查局部变量
	if idx, ok := c.localVars[expr.Name]; ok {
		c.emit1(OpLoadLocal, idx)
		return nil
	}

	// 检查是否是已定义的函数名(用于递归和函数间调用)
	if mainIdx, ok := c.functionConstIndices[expr.Name]; ok {
		// 本池创建占位常量, 编译收尾时统一修复为主池实际函数值
		localIdx := c.addConstant(NewValue(nil))
		c.pendingFunctionRefs[localIdx] = mainIdx
		c.emit1(OpConst, localIdx)
		return nil
	}

	// 未找到变量
	return NewCompileErrorFromPos(expr,
		"未定义的变量：'%s'（第%d行第%d列）。\n"+
			"→ 问题：使用了未声明的变量。\n"+
			"→ 原因：\n"+
			"  - 变量未声明就使用\n"+
			"  - 变量名拼写错误\n"+
			"  - 变量作用域问题（在作用域外访问）\n"+
			"→ 解决方法：\n"+
			"  - 使用 := 声明变量：x := 1\n"+
			"  - 从全局获取值：x := getBindValue(\"name\")\n"+
			"  - 检查变量名拼写是否正确\n"+
			"  - 确保在变量的作用域内使用",
		expr.Name, expr.Pos().Line, expr.Pos().Column)
}

// constructLiteral 根据字面量类型构造Value
func constructLiteral(litType LiteralType, value any) (Value, bool) {
	switch litType {
	case LiteralInt:
		return NewValue(value.(int)), true
	case LiteralFloat:
		return NewValue(value.(float64)), true
	case LiteralString:
		return NewValue(value.(string)), true
	case LiteralBool:
		return NewValue(value.(bool)), true
	case LiteralNil:
		return NewValue(nil), true
	}
	return Value{}, false
}

// compileLiteralExpr 编译字面量表达式
func (c *Compiler) compileLiteralExpr(expr *LiteralExpr) error {
	val, ok := constructLiteral(expr.Type, expr.Value)
	if !ok {
		return NewCompileErrorFromPos(expr,
			"不支持的字面量类型：%v（第%d行第%d列）。\n"+
				"→ 这是一个编译器内部错误，通常不应出现。\n"+
				"→ 支持的字面量类型：\n"+
				"  - LiteralInt：整数（如 123）\n"+
				"  - LiteralFloat：浮点数（如 3.14）\n"+
				"  - LiteralString：字符串（如 \"hello\"）\n"+
				"  - LiteralBool：布尔值（true 或 false）\n"+
				"  - LiteralNil：空值（nil）\n"+
				"→ 建议：如果看到此错误，请检查脚本语法或联系开发者",
			expr.Type, expr.Pos().Line, expr.Pos().Column)
	}

	idx := c.addConstant(val)
	c.emit1(OpConst, idx)
	return nil
}

// compileExprList 编译表达式列表
// 用于数组和Map等需要编译多个表达式的场景
func (c *Compiler) compileExprList(exprs []Expr) error {
	for _, expr := range exprs {
		if err := c.compileExpr(expr); err != nil {
			return err
		}
	}
	return nil
}

// compileArrayExpr 编译数组表达式
func (c *Compiler) compileArrayExpr(expr *ArrayExpr) error {
	if err := c.compileExprList(expr.Elements); err != nil {
		return err
	}
	c.emit1(OpArrayNew, len(expr.Elements))
	return nil
}

// compileMapExpr 编译Map表达式
func (c *Compiler) compileMapExpr(expr *MapExpr) error {
	for _, pair := range expr.Pairs {
		if err := c.compileExpr2(pair.Key, pair.Value); err != nil {
			return err
		}
	}
	c.emit1(OpMapNew, len(expr.Pairs))
	return nil
}

// compileBinaryExpr 编译二元表达式
// 处理算术、比较、逻辑等二元运算
// 对于&&和||运算符，使用短路求值优化
func (c *Compiler) compileBinaryExpr(expr *BinaryExpr) error {
	// 特殊处理索引赋值：arr[index] = value 或 arr[index] := value
	if expr.Operator == "=" || expr.Operator == ":=" {
		if indexExpr, ok := expr.Left.(*IndexExpr); ok {
			return c.compileIndexAssignment(indexExpr, expr.Right)
		}
	}

	// 复合赋值运算符: x += y, arr[i] -= z 等
	if isCompoundAssign(expr.Operator) {
		return c.compileCompoundAssign(expr)
	}

	if err := c.compileExpr(expr.Left); err != nil {
		return err
	}

	if isShortCircuitOp(expr.Operator) {
		return c.compileShortCircuit(expr)
	}

	if err := c.compileExpr(expr.Right); err != nil {
		return err
	}

	return c.compileBinaryOp(expr)
}

// isCompoundAssign 判断是否为复合赋值运算符
func isCompoundAssign(op string) bool {
	switch op {
	case "+=", "-=", "*=", "/=":
		return true
	}
	return false
}

// compileCompoundAssign 编译复合赋值表达式
// 语义: left op= right 等价于 left = left op right
func (c *Compiler) compileCompoundAssign(expr *BinaryExpr) error {
	baseOp := expr.Operator[:len(expr.Operator)-1]

	switch left := expr.Left.(type) {
	case *IdentExpr:
		return c.compileCompoundIdentAssign(left, expr.Right, baseOp)
	case *IndexExpr:
		return c.compileCompoundIndexAssign(left, expr.Right, baseOp)
	default:
		return c.binaryOpError(expr)
	}
}

// compileCompoundIdentAssign 编译变量复合赋值 x op= right
// 栈布局: -> result (保留结果作为表达式值)
func (c *Compiler) compileCompoundIdentAssign(ident *IdentExpr, right Expr, baseOp string) error {
	idx, ok := c.localVars[ident.Name]
	if !ok {
		// 变量未声明, 通过compileExpr触发标准错误
		return c.compileIdentExpr(ident)
	}
	c.emit1(OpLoadLocal, idx)
	if err := c.compileExpr(right); err != nil {
		return err
	}
	op, _ := lookupBinaryOp(baseOp)
	c.emit(op)
	// 复制结果: 一份存储, 一份留在栈上作为表达式值
	c.emit(OpDup)
	c.emit1(OpStoreLocal, idx)
	return nil
}

// compileCompoundIndexAssign 编译索引复合赋值 arr[i] op= right
// 栈布局: obj, index, (obj[index] op right) -> OpStoreIndex
func (c *Compiler) compileCompoundIndexAssign(indexExpr *IndexExpr, right Expr, baseOp string) error {
	// 编译 obj, index 供 OpStoreIndex 使用
	if err := c.compileExpr(indexExpr.Object); err != nil {
		return err
	}
	if err := c.compileExpr(indexExpr.Index); err != nil {
		return err
	}
	// 再编译 obj[index] 供读取当前值
	if err := c.compileExpr(indexExpr.Object); err != nil {
		return err
	}
	if err := c.compileExpr(indexExpr.Index); err != nil {
		return err
	}
	c.emit(OpIndex)
	// 编译右值并运算
	if err := c.compileExpr(right); err != nil {
		return err
	}
	op, _ := lookupBinaryOp(baseOp)
	c.emit(op)
	c.emit(OpStoreIndex)
	return nil
}

// compileIndexAssignment 编译索引赋值表达式
// 栈布局：obj, index, value -> (无)
func (c *Compiler) compileIndexAssignment(indexExpr *IndexExpr, valueExpr Expr) error {
	// 编译对象表达式（如 arr）
	if err := c.compileExpr(indexExpr.Object); err != nil {
		return err
	}
	// 编译索引表达式（如 0）
	if err := c.compileExpr(indexExpr.Index); err != nil {
		return err
	}
	// 编译值表达式（如 100）
	if err := c.compileExpr(valueExpr); err != nil {
		return err
	}
	// 生成索引赋值指令
	c.emit(OpStoreIndex)
	return nil
}

// compileBinaryOp 编译二元运算符
func (c *Compiler) compileBinaryOp(expr *BinaryExpr) error {
	op, ok := lookupBinaryOp(expr.Operator)
	if !ok {
		return c.binaryOpError(expr)
	}
	c.emit(op)
	return nil
}

// binaryOpError 生成二元运算符错误
func (c *Compiler) binaryOpError(expr *BinaryExpr) error {
	return NewCompileErrorFromPos(expr,
		"不支持的运算符：'%s'（第%d行第%d列）。\n"+
			"→ 问题：该运算符不被当前版本支持。\n"+
			"→ 支持的运算符：\n"+
			"  - 算术运算：+（加）、-（减）、*（乘）、/（除）、%%（取模）\n"+
			"  - 比较运算：==（等于）、!=（不等）、<（小于）、<=（小于等于）、>（大于）、>=（大于等于）\n"+
			"  - 逻辑运算：&&（与）、||（或）\n"+
			"  - 位运算：&（位与）、|（位或）、^（位异或）、<<（左移）、>>（右移）\n"+
			"→ 建议：检查运算符是否正确，或使用等效的表达式",
		expr.Operator, expr.Pos().Line, expr.Pos().Column)
}

// compileUnaryExpr 编译一元表达式
func (c *Compiler) compileUnaryExpr(expr *UnaryExpr) error {
	if err := c.compileExpr(expr.Operand); err != nil {
		return err
	}

	opcode, ok := lookupUnaryOp(expr.Operator)
	if !ok {
		return NewCompileErrorFromPos(expr,
			"无效的一元运算符：'%s'（第%d行第%d列）。\n"+
				"→ 问题：该运算符不能作为一元运算符使用。\n"+
				"→ 支持的一元运算符：\n"+
				"  - -（取负）：对数值取负，如 -123, -3.14\n"+
				"  - !（逻辑非）：对布尔值取反，如 !true == false\n"+
				"  - ^（位取反）：对整数按位取反，如 ^5\n"+
				"→ 示例：x := -5  // 正确\n"+
				"→ 错误：x := +5  // 不支持一元+\n"+
				"→ 建议：检查运算符是否正确，或使用等效的表达式",
			expr.Operator, expr.Pos().Line, expr.Pos().Column)
	}
	c.emit(opcode)
	return nil
}

// compileIndexExpr 编译索引表达式
func (c *Compiler) compileIndexExpr(expr *IndexExpr) error {
	if err := c.compileExpr2(expr.Object, expr.Index); err != nil {
		return err
	}
	c.emit(OpIndex)
	return nil
}

// compileSliceBound 编译切片边界表达式
// 如果表达式为nil，使用预定义的默认索引
func (c *Compiler) compileSliceBound(expr Expr, defaultIdx int) error {
	if expr != nil {
		return c.compileExpr(expr)
	}
	c.emit1(OpConst, defaultIdx)
	return nil
}

// compileSliceExpr 编译切片表达式
// 栈布局：obj, start, end -> result
// 当 end 为 nil 时，使用特殊值 -1 表示需要使用对象长度
func (c *Compiler) compileSliceExpr(expr *SliceExpr) error {
	if err := c.compileExpr(expr.Object); err != nil {
		return err
	}
	if err := c.compileSliceBound(expr.Start, c.sliceStartIdx); err != nil {
		return err
	}
	if err := c.compileSliceBound(expr.End, c.sliceEndIdx); err != nil {
		return err
	}
	c.emit(OpSlice)
	return nil
}

// compileCallExpr 编译函数调用表达式
// 使用分发表模式分发到不同类型的调用编译器
func (c *Compiler) compileCallExpr(expr *CallExpr) error {
	callTy, name, extIdx := c.classifyCallExpr(expr)

	switch callTy {
	case callBuiltin:
		return c.compileBuiltinCall(name, expr.Args)
	case callExternal:
		return c.compileExternalCall(expr.Args, extIdx)
	default:
		return c.compileUserDefinedCall(expr)
	}
}

// compileExternalCall 编译外部函数调用
func (c *Compiler) compileExternalCall(args []Expr, extIdx int) error {
	if err := c.compileArgs(args); err != nil {
		return err
	}
	c.emit2(OpCallBind, extIdx, len(args))
	return nil
}

// callType 函数调用类型
type callType int

const (
	// callBuiltin 内置函数
	callBuiltin callType = iota
	// callExternal 外部函数
	callExternal
	// callUserDefined 用户定义函数
	callUserDefined
)

// classifyCallExpr 分类函数调用类型
func (c *Compiler) classifyCallExpr(expr *CallExpr) (callType, string, int) {
	ident, ok := expr.Func.(*IdentExpr)
	if !ok {
		return callUserDefined, "", -1
	}

	if c.isBuiltin(ident.Name) {
		return callBuiltin, ident.Name, -1
	}

	if extIdx := c.getExternalIndex(ident.Name); extIdx >= 0 {
		return callExternal, ident.Name, extIdx
	}

	return callUserDefined, "", -1
}

// compileUserDefinedCall 编译用户定义函数调用
func (c *Compiler) compileUserDefinedCall(expr *CallExpr) error {
	if err := c.compileExpr(expr.Func); err != nil {
		return err
	}
	if err := c.compileArgs(expr.Args); err != nil {
		return err
	}
	c.emit1(OpCall, len(expr.Args))
	return nil
}

// compileIfExpr 编译if表达式
func (c *Compiler) compileIfExpr(expr *IfExpr) error {
	// 编译条件
	if err := c.compileExpr(expr.Condition); err != nil {
		return err
	}

	jumpIfFalsePos := c.emit1(OpJumpIfFalse, 0)

	// then分支
	if err := c.compileExpr(expr.Then); err != nil {
		return err
	}

	jumpPos := c.emit1(OpJump, 0)
	c.patchJump(jumpIfFalsePos)

	// else分支
	if err := c.compileExprOrNil(expr.Else); err != nil {
		return err
	}

	c.patchJump(jumpPos)

	return nil
}

// compileBlockExpr 编译块表达式
// 与 compileBlockStmt 不同，保留最后一个表达式的值在栈上
// 用于 if 语句需要返回值的场景
func (c *Compiler) compileBlockExpr(block *BlockStmt) error {
	if block == nil || len(block.Statements) == 0 {
		c.emit(OpNil)
		return nil
	}

	// 编译除最后一个语句外的所有语句
	for i := 0; i < len(block.Statements)-1; i++ {
		if err := c.compileStmt(block.Statements[i]); err != nil {
			return err
		}
	}

	// 特殊处理最后一个语句：如果是表达式语句，保留其值
	lastStmt := block.Statements[len(block.Statements)-1]
	if exprStmt, ok := lastStmt.(*ExprStmt); ok {
		return c.compileExpr(exprStmt.Expr)
	}

	// IfStmt作为最后语句: 使用值版编译保留块值
	if ifStmt, ok := lastStmt.(*IfStmt); ok {
		return c.compileIfStmtValue(ifStmt)
	}

	// 如果不是表达式语句，正常编译然后压入 nil
	if err := c.compileStmt(lastStmt); err != nil {
		return err
	}
	c.emit(OpNil)
	return nil
}

// compileArgs 编译参数列表
// 复用compileExprList实现
func (c *Compiler) compileArgs(args []Expr) error {
	return c.compileExprList(args)
}
