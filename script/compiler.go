package script

import (
	"math"
)

// ========== 编译器核心结构 ==========
// Compiler 将AST编译为字节码
// 负责将语法树转换为虚拟机可执行的字节码指令序列
// 支持常量池、局部变量、外部函数等特性
type Compiler struct {
	// constants 常量池，存储字面量和编译时常量
	constants []Value
	// mainConstants 主常量池引用,用于函数编译时访问外部常量
	mainConstants *[]Value
	// instructions 指令序列，存储生成的字节码
	instructions []Instruction
	// functions 用户定义的函数列表
	functions []*CompiledFunction
	// externals 外部函数声明列表
	externals []ExternalFunc
	// localVars 局部变量名到索引的映射
	localVars map[string]int
	// externalCache 外部函数索引缓存，用于O(1)查找
	externalCache map[string]int
	// constantCache 常量索引缓存，用于去重
	constantCache map[cacheKey]int
	// functionConstIndices 函数常量索引缓存，用于递归调用
	functionConstIndices map[string]int
	// pendingFunctionRefs 待修复的函数引用：函数体常量池索引 -> 主常量池索引
	// 用于处理函数体内引用其他函数的情况，这些引用需要在函数编译完成后更新
	pendingFunctionRefs map[int]int
	// sliceStartIdx 切片起始索引常量，预定义避免重复创建
	sliceStartIdx int
	// sliceEndIdx 切片结束索引常量，预定义避免重复创建
	sliceEndIdx int
	// loopBreaks 保存当前循环的 break 跳转位置列表（嵌套循环使用二维切片）
	loopBreaks [][]int
	// loopContinues 保存当前循环的 continue 跳转位置列表
	loopContinues [][]int
	// loopStarts 保存当前循环的开始位置（用于 continue）
	loopStarts []int
	// inTypedDeclaration 标记当前是否在类型安全声明（:=>）中
	inTypedDeclaration bool
}

// cacheKey 常量缓存键，用于常量池去重
// 通过类型和值的哈希来唯一标识常量
type cacheKey struct {
	// typ 值类型
	typ ValueType
	// data 存储值的哈希或直接值
	data uint64
}

// NewCompiler 创建编译器实例
// 初始化所有内部数据结构并预定义常用常量
func NewCompiler() *Compiler {
	c := &Compiler{
		localVars:     make(map[string]int),
		externalCache: make(map[string]int),
		constantCache: make(map[cacheKey]int),
		constants:     make([]Value, 0, 32),
		instructions:  make([]Instruction, 0, 64),
		functions:     make([]*CompiledFunction, 0, 4),
		externals:     make([]ExternalFunc, 0, 4),
	}
	// 预定义切片常用常量
	c.sliceStartIdx = c.addConstant(NewValue(0))
	c.sliceEndIdx = c.addConstant(NewValue(SliceEndDefault))
	return c
}

// Compile 编译AST到字节码
// 参数program为解析后的语法树根节点
// 返回编译后的脚本和可能的错误
func (c *Compiler) Compile(program *Program) (*CompiledScript, error) {
	// 编译所有语句，除了最后一个
	for i := 0; i < len(program.Statements)-1; i++ {
		if err := c.compileStmt(program.Statements[i]); err != nil {
			return nil, err
		}
	}

	// 编译最后一个语句
	if err := c.compileLastStatement(program.Statements); err != nil {
		return nil, err
	}

	// 创建主函数
	return c.createCompiledScript(), nil
}

// compileLastStatement 编译最后一个语句
// 特殊处理：对于表达式语句和if语句保留值在栈上作为返回值
// 对于其他非表达式语句，压入nil作为默认返回值
func (c *Compiler) compileLastStatement(statements []Stmt) error {
	if len(statements) == 0 {
		c.emit(OpNil)
		return nil
	}

	lastStmt := statements[len(statements)-1]

	// 表达式语句：直接编译表达式，保留值在栈上
	if exprStmt, ok := lastStmt.(*ExprStmt); ok {
		return c.compileExpr(exprStmt.Expr)
	}

	// if语句: 使用值版编译, 保留最后一个表达式的值作为返回值
	if ifStmt, ok := lastStmt.(*IfStmt); ok {
		return c.compileIfStmtValue(ifStmt)
	}

	// 其他语句：编译后压入 nil
	if err := c.compileStmt(lastStmt); err != nil {
		return err
	}
	c.emit(OpNil)
	return nil
}

// createCompiledScript 创建编译后的脚本对象
// 将编译结果封装为CompiledScript，包含主函数和所有依赖
func (c *Compiler) createCompiledScript() *CompiledScript {
	main := &CompiledFunction{
		Name:         "main",
		Instructions: c.instructions,
		Constants:    c.constants,
		NumLocals:    len(c.localVars),
		NumParams:    0,
	}

	return &CompiledScript{
		Functions: c.functions,
		Main:      main,
		Externals: c.externals,
		Constants: c.constants,
	}
}

// compileStmt 编译语句
// 使用type switch分发，避免reflect.TypeOf和map查找开销
func (c *Compiler) compileStmt(stmt Stmt) error {
	switch s := stmt.(type) {
	case *VarDeclStmt:
		return c.compileVarDecl(s)
	case *ExprStmt:
		return c.compileExprStmt(s)
	case *BlockStmt:
		return c.compileBlockStmt(s)
	case *IfStmt:
		return c.compileIfStmt(s)
	case *ForStmt:
		return c.compileForStmt(s)
	case *FuncDeclStmt:
		return c.compileFuncDecl(s)
	case *ReturnStmt:
		return c.compileReturnStmt(s)
	case *BreakStmt:
		return c.compileBreakStmt(s)
	case *ContinueStmt:
		return c.compileContinueStmt(s)
	case *ThrowStmt:
		return c.compileThrowStmt(s)
	case *DefDirectiveStmt:
		return c.compileDefDirective(s)
	}
	return NewCompileErrorFromPos(stmt, "内部编译错误：无法识别的语句类型")
}

// emit 生成无参数指令
func (c *Compiler) emit(op OpCode) int {
	c.instructions = append(c.instructions, Instruction{Op: op})
	return len(c.instructions) - 1
}

// emit1 生成单参数指令
func (c *Compiler) emit1(op OpCode, arg int) int {
	c.instructions = append(c.instructions, Instruction{Op: op, ArgCount: 1, Args: [2]int{arg}})
	return len(c.instructions) - 1
}

// emit2 生成双参数指令
func (c *Compiler) emit2(op OpCode, arg1, arg2 int) int {
	c.instructions = append(c.instructions, Instruction{Op: op, ArgCount: 2, Args: [2]int{arg1, arg2}})
	return len(c.instructions) - 1
}

// patchJump 修补跳转指令的目标地址
// 用于在代码生成后回填跳转目标位置
func (c *Compiler) patchJump(pos int) {
	offset := len(c.instructions)
	c.instructions[pos].Args[0] = offset
	c.instructions[pos].ArgCount = 1
}

// addConstant 添加常量到常量池
// 返回常量在池中的索引，供后续指令引用
// 对可哈希类型进行去重优化
func (c *Compiler) addConstant(value Value) int {
	key := makeCacheKey(value)
	if idx, ok := c.tryGetCachedConstant(key); ok {
		return idx
	}
	return c.addNewConstant(value, key)
}

// tryGetCachedConstant 尝试从缓存获取常量索引
// 如果缓存命中返回 (idx, true)，否则返回 (0, false)
func (c *Compiler) tryGetCachedConstant(key cacheKey) (int, bool) {
	if key.typ == TypeNil {
		return 0, false
	}
	idx, ok := c.constantCache[key]
	return idx, ok
}

// addNewConstant 添加新常量到常量池
// 不进行缓存检查，仅添加并更新缓存
func (c *Compiler) addNewConstant(value Value, key cacheKey) int {
	idx := len(c.constants)
	c.constants = append(c.constants, value)
	if key.typ != TypeNil {
		c.constantCache[key] = idx
	}
	return idx
}

// makeCacheKey 为常量值创建缓存键
func makeCacheKey(v Value) cacheKey {
	switch v.Type {
	case TypeInt:
		return cacheKey{typ: TypeInt, data: uint64(v.Int())}
	case TypeString:
		h := uint64(0)
		for _, ch := range v.String() {
			h = h*31 + uint64(ch)
		}
		return cacheKey{typ: TypeString, data: h}
	case TypeBool:
		if v.Bool() {
			return cacheKey{typ: TypeBool, data: 1}
		}
		return cacheKey{typ: TypeBool, data: 0}
	case TypeFloat:
		return cacheKey{typ: TypeFloat, data: math.Float64bits(v.Float())}
	}
	return cacheKey{typ: TypeNil}
}

// ========== 运算符查找 ==========

// lookupBinaryOp 二元运算符到操作码查找
func lookupBinaryOp(op string) (OpCode, bool) {
	switch op {
	// 算术运算符
	case "+":
		return OpAdd, true
	case "-":
		return OpSub, true
	case "*":
		return OpMul, true
	case "/":
		return OpDiv, true
	case "%":
		return OpMod, true
	// 比较运算符
	case "==":
		return OpEqual, true
	case "!=":
		return OpNotEqual, true
	case "<":
		return OpLess, true
	case "<=":
		return OpLessEq, true
	case ">":
		return OpGreater, true
	case ">=":
		return OpGreaterEq, true
	// 位运算符
	case "&":
		return OpBitAnd, true
	case "|":
		return OpBitOr, true
	case "^":
		return OpBitXor, true
	case "<<":
		return OpLShift, true
	case ">>":
		return OpRShift, true
	}
	return 0, false
}

// isShortCircuitOp 判断是否为短路求值运算符
func isShortCircuitOp(op string) bool {
	return op == "&&" || op == "||"
}

// lookupUnaryOp 一元运算符到操作码查找
func lookupUnaryOp(op string) (OpCode, bool) {
	switch op {
	case "-":
		return OpNeg, true
	case "!":
		return OpNot, true
	case "^":
		return OpBitNot, true
	}
	return 0, false
}

// shortCircuitJumpOp 短路求值跳转操作码
func shortCircuitJumpOp(op string) OpCode {
	if op == "||" {
		return OpJumpIfTrue
	}
	return OpJumpIfFalse
}

// compileShortCircuit 编译短路求值表达式
func (c *Compiler) compileShortCircuit(expr *BinaryExpr) error {
	jumpOp := shortCircuitJumpOp(expr.Operator)

	// 复制左侧值用于短路判断
	c.emit(OpDup)
	jumpPos := c.emit1(jumpOp, 0)

	// 如果不跳转，弹出复制的值
	c.emit(OpPop)

	// 编译右侧表达式
	if err := c.compileExpr(expr.Right); err != nil {
		return err
	}

	// 修补跳转目标地址
	c.patchJump(jumpPos)
	return nil
}

// ========== 循环管理方法 ==========
// 用于处理 break 和 continue 语句的跳转目标回填

// pushLoop 进入循环作用域
// 参数 start 为循环开始位置（用于 continue 跳转目标）
func (c *Compiler) pushLoop(start int) {
	c.loopStarts = append(c.loopStarts, start)
	c.loopBreaks = append(c.loopBreaks, []int{})
	c.loopContinues = append(c.loopContinues, []int{})
}

// popLoop 退出循环作用域
// 在循环编译完成后调用，清理循环状态
func (c *Compiler) popLoop() {
	c.loopStarts = c.loopStarts[:len(c.loopStarts)-1]
	c.loopBreaks = c.loopBreaks[:len(c.loopBreaks)-1]
	c.loopContinues = c.loopContinues[:len(c.loopContinues)-1]
}

// addBreak 添加 break 跳转位置
// 参数 pos 为跳转指令的位置，目标稍后通过 patchBreaks 回填
func (c *Compiler) addBreak(pos int) {
	if len(c.loopBreaks) > 0 {
		c.loopBreaks[len(c.loopBreaks)-1] = append(c.loopBreaks[len(c.loopBreaks)-1], pos)
	}
}

// patchBreaks 回填所有 break 跳转目标
// 参数 target 为循环结束位置，所有 break 语句跳转到此处
func (c *Compiler) patchBreaks(target int) {
	if len(c.loopBreaks) > 0 {
		for _, pos := range c.loopBreaks[len(c.loopBreaks)-1] {
			c.instructions[pos].Args[0] = target
		}
	}
}

// addContinue 添加 continue 跳转位置
// 参数 pos 为跳转指令的位置，目标稍后通过 patchContinues 回填
func (c *Compiler) addContinue(pos int) {
	if len(c.loopContinues) > 0 {
		c.loopContinues[len(c.loopContinues)-1] = append(c.loopContinues[len(c.loopContinues)-1], pos)
	}
}

// patchContinues 回填所有 continue 跳转目标
// 参数 target 为循环继续位置（通常是循环条件检查处）
func (c *Compiler) patchContinues(target int) {
	if len(c.loopContinues) > 0 {
		for _, pos := range c.loopContinues[len(c.loopContinues)-1] {
			c.instructions[pos].Args[0] = target
		}
	}
}

// inLoop 检查当前是否在循环内部
// 用于编译时验证 break/continue 的合法性
func (c *Compiler) inLoop() bool {
	return len(c.loopBreaks) > 0
}
