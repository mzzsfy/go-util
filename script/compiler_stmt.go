package script

import "fmt"

// ========== 语句编译方法 ==========
// 包含所有语句类型的编译实现

// compileVarDecl 编译变量声明语句
// 将初始值压栈并存储到局部变量表
// :=> 声明在编译期校验注解类型名, 并对非 any 注解生成运行时类型校验
func (c *Compiler) compileVarDecl(stmt *VarDeclStmt) error {
	annotName := ""
	if stmt.TypeAnnot != nil {
		name := typeAnnotName(stmt.TypeAnnot)
		runtimeType, ok := typedAnnotTypes[name]
		if !ok {
			return NewCompileErrorFromPos(stmt,
				"类型错误：未知的类型注解 '%s'。\n"+
					"→ 问题：:=> 声明只支持以下类型注解。\n"+
					"→ 支持的类型：int(i)、float(f)、string(s/str)、bool(b)、any、arr(array)、map\n"+
					"→ 建议：使用支持的类型注解，或使用 any 跳过类型校验",
				name)
		}
		c.inTypedDeclaration = true
		c.typedAnnotName = name
		annotName = runtimeType
	}

	// 编译初始值
	err := c.compileExpr(stmt.Value)
	// 无论成功失败都重置声明标记, 防止泄漏到后续语句
	c.inTypedDeclaration = false
	c.typedAnnotName = ""
	if err != nil {
		return err
	}

	// 非 any 注解生成运行时类型校验序列
	if annotName != "" {
		c.emitTypeCheck(annotName)
	}

	// 存储变量
	if _, ok := c.localVars[stmt.Name]; !ok {
		c.localVars[stmt.Name] = len(c.localVars)
	}

	// 生成存储指令
	c.emit1(OpStoreLocal, c.localVars[stmt.Name])

	return nil
}

// typeAnnotName 提取类型注解的基础类型名
// 复合类型注解(数组/map等)返回其字符串形式, 不在支持集合内由调用方报错
func typeAnnotName(t TypeExpr) string {
	if base, ok := t.(*BaseTypeExpr); ok {
		return base.Name
	}
	return t.String()
}

// emitTypeCheck 生成运行时类型校验指令序列
// 栈顶为待校验值, 匹配时保持该值在栈顶, 不匹配时抛出运行时错误
func (c *Compiler) emitTypeCheck(annotName string) {
	c.emit(OpDup)
	c.emit(OpTypeOf)
	c.emit1(OpConst, c.addConstant(NewValue(annotName)))
	c.emit(OpEqual)
	jumpPos := c.emit1(OpJumpIfTrue, 0)
	c.emit(OpPop)
	c.emit1(OpConst, c.addConstant(NewValue(
		"类型错误：getBindValue 返回值类型与注解不匹配。\n"+
			"→ 要求：:=>"+annotName+" 声明的值必须是 "+annotName+" 类型\n"+
			"→ 建议：检查绑定值类型，或使用 :=>any 跳过校验")))
	c.emit(OpThrow)
	c.patchJump(jumpPos)
}

// compileExprStmt 编译表达式语句
// 计算表达式并弹出结果（表达式结果不作为返回值）
func (c *Compiler) compileExprStmt(stmt *ExprStmt) error {
	if err := c.compileExpr(stmt.Expr); err != nil {
		return err
	}
	// 弹出栈顶值（如果不是最后一个语句）
	c.emit(OpPop)
	return nil
}

// compileBlockStmt 编译块语句
// 依次编译块内的所有语句
func (c *Compiler) compileBlockStmt(stmt *BlockStmt) error {
	for _, s := range stmt.Statements {
		if err := c.compileStmt(s); err != nil {
			return err
		}
	}
	return nil
}

// compileIfStmt 编译if条件语句(语句上下文)
// 不保留then/else块的值在栈上
func (c *Compiler) compileIfStmt(stmt *IfStmt) error {
	// 编译条件
	if err := c.compileExpr(stmt.Condition); err != nil {
		return err
	}

	// 条件跳转（假跳转）
	jumpIfFalsePos := c.emit1(OpJumpIfFalse, 0)

	// then分支 - 语句上下文不留值
	if err := c.compileBlockStmt(stmt.Then); err != nil {
		return err
	}

	// 处理else分支(语句上下文)
	return c.compileIfElseBranch(stmt.Else, jumpIfFalsePos)
}

// compileIfStmtValue 编译if语句(值上下文)
// 保留then/else块最后一个表达式的值在栈上, 无else时压入nil
func (c *Compiler) compileIfStmtValue(stmt *IfStmt) error {
	if err := c.compileExpr(stmt.Condition); err != nil {
		return err
	}
	jumpIfFalsePos := c.emit1(OpJumpIfFalse, 0)
	if err := c.compileBlockExpr(stmt.Then); err != nil {
		return err
	}
	return c.compileIfElseBranchExpr(stmt.Else, jumpIfFalsePos)
}

// compileIfElseBranchExpr 编译if语句的else分支(值上下文,保留值)
func (c *Compiler) compileIfElseBranchExpr(elseStmt Stmt, jumpIfFalsePos int) error {
	// 跳过else分支(then块已有值在栈上)
	jumpPos := c.emit1(OpJump, 0)
	c.patchJump(jumpIfFalsePos)

	if elseStmt == nil {
		// 无else, 压入nil作为值
		c.emit(OpNil)
	} else {
		// 处理 else 分支
		switch e := elseStmt.(type) {
		case *BlockStmt:
			if err := c.compileBlockExpr(e); err != nil {
				return err
			}
		case *IfStmt:
			// else if - 递归处理(值上下文)
			if err := c.compileIfStmtValue(e); err != nil {
				return err
			}
		default:
			if err := c.compileStmt(elseStmt); err != nil {
				return err
			}
			c.emit(OpNil)
		}
	}

	c.patchJump(jumpPos)
	return nil
}

// compileIfElseBranch 编译if语句的else分支(语句上下文,不留值)
func (c *Compiler) compileIfElseBranch(elseStmt Stmt, jumpIfFalsePos int) error {
	if elseStmt == nil {
		c.patchJump(jumpIfFalsePos)
		return nil
	}

	jumpPos := c.emit1(OpJump, 0)
	c.patchJump(jumpIfFalsePos)

	// 处理 else 分支
	switch e := elseStmt.(type) {
	case *BlockStmt:
		// else { block } - 语句上下文不留值
		if err := c.compileBlockStmt(e); err != nil {
			return err
		}
	case *IfStmt:
		// else if - 递归处理(语句上下文)
		if err := c.compileIfStmt(e); err != nil {
			return err
		}
	default:
		// 其他情况
		if err := c.compileStmt(elseStmt); err != nil {
			return err
		}
	}

	c.patchJump(jumpPos)
	return nil
}

// compileForStmt 编译for语句
func (c *Compiler) compileForStmt(stmt *ForStmt) error {
	switch stmt.Mode {
	case ForWhile:
		return c.compileWhileFor(stmt)
	case ForCount:
		return c.compileCountFor(stmt)
	case ForStandard:
		return c.compileStandardFor(stmt)
	case ForRange:
		return c.compileRangeFor(stmt)
	default:
		return NewCompileErrorFromPos(stmt, "未知的for循环模式: %v", stmt.Mode)
	}
}

// compileWhileFor 编译条件循环 (while)
// 语法: for condition { body }
// 实现: 条件判断 -> 循环体 -> 跳回条件判断
func (c *Compiler) compileWhileFor(stmt *ForStmt) error {
	// 记录循环开始位置
	loopStart := len(c.instructions)
	// 进入循环作用域
	c.pushLoop(loopStart)

	// 编译条件表达式
	if err := c.compileExpr(stmt.Cond); err != nil {
		return err
	}

	// 条件为假时跳出循环（占位，稍后回填）
	jumpIfFalseIdx := len(c.instructions)
	c.emit1(OpJumpIfFalse, 0)

	// 编译循环体
	if err := c.compileBlockStmt(stmt.Body); err != nil {
		return err
	}

	// 跳回循环开始（使用绝对位置）
	c.emit1(OpJump, loopStart)

	// 回填：条件为假时跳到这里（使用绝对位置）
	loopEnd := len(c.instructions)
	c.instructions[jumpIfFalseIdx].Args[0] = loopEnd

	// 回填所有 break 和 continue 跳转目标
	c.patchBreaks(loopEnd)
	c.patchContinues(loopStart)
	// 退出循环作用域
	c.popLoop()

	return nil
}

// compileCountFor 编译计数循环
// 语法: for i := n { body }
// 实现: 初始化i=1 -> 条件判断i<=n -> 循环体 -> i++ -> 跳回条件判断
// 注意：如果n是数组或Map，则自动转为range循环
// continue 跳转到 i++ 执行，break 跳转到循环结束
func (c *Compiler) compileCountFor(stmt *ForStmt) error {
	// 检查init值的类型
	initValue := stmt.Init.(*VarDeclStmt).Value

	// 如果是数组/Map字面量，直接编译为range循环
	if c.isLikelyRangeLoop(initValue) {
		return c.compileRangeFor(stmt)
	}

	// 如果是标识符，需要运行时判断类型
	if _, ok := initValue.(*IdentExpr); ok {
		return c.compileDynamicFor(stmt)
	}

	// 编译循环次数（存储到常量池）
	if err := c.compileExpr(initValue); err != nil {
		return err
	}

	// 将循环次数存储到临时变量
	countVarIdx := c.allocLoopTemp("count")
	c.emit1(OpStoreLocal, countVarIdx)

	// 创建计数器变量i=1（从1开始）
	// 用户变量名指向计数器，这样在循环体中 i 就是计数器的值
	// 同名变量每层循环独立槽位, 避免嵌套时槽越界或计数器互踩
	counterName := stmt.Init.(*VarDeclStmt).Name
	iVarIdx, counterShadow := c.shadowLoopVar(counterName)
	c.emit1(OpConst, c.addConstant(NewValue(1)))
	c.emit1(OpStoreLocal, iVarIdx)

	// 记录循环开始位置
	loopStart := len(c.instructions)
	// 进入循环作用域
	c.pushLoop(loopStart)

	// 编译条件：i <= count
	c.emit1(OpLoadLocal, iVarIdx)
	c.emit1(OpLoadLocal, countVarIdx)
	c.emit(OpLessEq)

	// 条件为假时跳出循环（占位）
	jumpIfFalseIdx := len(c.instructions)
	c.emit1(OpJumpIfFalse, 0)

	// 编译循环体
	if err := c.compileBlockStmt(stmt.Body); err != nil {
		return err
	}

	// 记录 i++ 位置（continue 跳转目标）
	postStart := len(c.instructions)

	// i++
	c.emit1(OpLoadLocal, iVarIdx)
	c.emit1(OpConst, c.addConstant(NewValue(1)))
	c.emit(OpAdd)
	c.emit1(OpStoreLocal, iVarIdx)

	// 跳回循环开始（使用绝对位置）
	c.emit1(OpJump, loopStart)

	// 回填：条件为假时跳到这里（使用绝对位置）
	loopEnd := len(c.instructions)
	c.instructions[jumpIfFalseIdx].Args[0] = loopEnd

	// 回填所有 break 和 continue 跳转目标
	// break 跳到循环结束，continue 跳到 i++ 处
	c.patchBreaks(loopEnd)
	c.patchContinues(postStart)
	// 退出循环作用域
	c.popLoop()
	// 恢复被遮蔽的外层同名变量绑定
	c.unshadowLoopVar(counterShadow)

	return nil
}

// compileDynamicFor 编译动态for循环
// 在运行时判断是计数循环还是range循环
func (c *Compiler) compileDynamicFor(stmt *ForStmt) error {
	initValue := stmt.Init.(*VarDeclStmt).Value
	varName := stmt.Init.(*VarDeclStmt).Name

	// 编译获取循环变量的值
	if err := c.compileExpr(initValue); err != nil {
		return err
	}

	// 复制一份用于类型检查
	c.emit(OpDup)

	// 检查是否是数组类型
	// 使用 OpTypeOf 获取类型
	c.emit(OpTypeOf)

	// 加载 "array" 字符串用于比较
	c.emit1(OpConst, c.addConstant(NewValue("array")))

	// 比较是否相等
	c.emit(OpEqual)

	// 如果是数组，跳转到 range 循环
	isArrayJumpIdx := len(c.instructions)
	c.emit1(OpJumpIfTrue, 0)

	// 检查是否是 map 类型
	// 再次获取类型（因为之前被比较消耗了）
	c.emit(OpDup)
	c.emit(OpTypeOf)
	c.emit1(OpConst, c.addConstant(NewValue("map")))
	c.emit(OpEqual)

	// 如果是 map，跳转到 range 循环
	isMapJumpIdx := len(c.instructions)
	c.emit1(OpJumpIfTrue, 0)

	// === 计数循环路径 ===
	// 栈顶是循环次数（数字）

	// 将循环次数存储到临时变量
	countVarIdx := c.allocLoopTemp("count")
	c.emit1(OpStoreLocal, countVarIdx)

	// 创建计数器变量i=1（从1开始）
	iVarIdx, counterShadow := c.shadowLoopVar(varName)
	c.emit1(OpConst, c.addConstant(NewValue(1)))
	c.emit1(OpStoreLocal, iVarIdx)

	// 记录计数循环开始位置
	countLoopStart := len(c.instructions)
	c.pushLoop(countLoopStart)

	// 编译条件：i <= count
	c.emit1(OpLoadLocal, iVarIdx)
	c.emit1(OpLoadLocal, countVarIdx)
	c.emit(OpLessEq)

	// 条件为假时跳出循环
	countJumpIfFalseIdx := len(c.instructions)
	c.emit1(OpJumpIfFalse, 0)

	// 编译循环体
	if err := c.compileBlockStmt(stmt.Body); err != nil {
		return err
	}

	// 记录 i++ 位置
	countPostStart := len(c.instructions)

	// i++
	c.emit1(OpLoadLocal, iVarIdx)
	c.emit1(OpConst, c.addConstant(NewValue(1)))
	c.emit(OpAdd)
	c.emit1(OpStoreLocal, iVarIdx)

	// 跳回循环开始
	c.emit1(OpJump, countLoopStart)

	// 计数循环结束
	countLoopEnd := len(c.instructions)
	c.instructions[countJumpIfFalseIdx].Args[0] = countLoopEnd
	c.patchBreaks(countLoopEnd)
	c.patchContinues(countPostStart)
	c.popLoop()
	c.unshadowLoopVar(counterShadow)

	// 跳过 range 循环
	skipRangeJumpIdx := len(c.instructions)
	c.emit1(OpJump, 0)

	// === Range 循环路径 ===
	rangeLoopStart := len(c.instructions)

	// 回填：如果是数组/map，跳转到这里
	c.instructions[isArrayJumpIdx].Args[0] = rangeLoopStart
	c.instructions[isMapJumpIdx].Args[0] = rangeLoopStart

	// 栈顶是集合（数组或map），存储到临时变量
	collectionVarIdx := c.allocLoopTemp("collection")
	c.emit1(OpStoreLocal, collectionVarIdx)

	// Map处理: 运行时检测并提取keys
	mapVarIdx := c.allocLoopTemp("map")
	c.emit(OpNil)
	c.emit1(OpStoreLocal, mapVarIdx)
	c.emit1(OpLoadLocal, collectionVarIdx)
	c.emit(OpTypeOf)
	c.emit1(OpConst, c.addConstant(NewValue("map")))
	c.emit(OpEqual)
	notMapJump := c.emit1(OpJumpIfFalse, 0)
	// 是Map: 保存原始map, 提取keys
	c.emit1(OpLoadLocal, collectionVarIdx)
	c.emit1(OpStoreLocal, mapVarIdx)
	c.emit1(OpLoadLocal, mapVarIdx)
	c.emit(OpMapKeys)
	c.emit1(OpStoreLocal, collectionVarIdx)
	c.instructions[notMapJump].Args[0] = len(c.instructions)

	// 创建索引变量idx=0
	idxVarIdx := c.allocLoopTemp("idx")
	c.emit1(OpConst, c.addConstant(NewValue(0)))
	c.emit1(OpStoreLocal, idxVarIdx)

	// 获取集合长度
	lenVarIdx := c.allocLoopTemp("len")
	c.emit1(OpLoadLocal, collectionVarIdx)
	c.emit1(OpLen, 0)
	c.emit1(OpStoreLocal, lenVarIdx)

	// 创建键变量（用户定义的变量名）
	keyVarIdx, dynamicKeyShadow := c.shadowLoopVar(varName)

	// 记录循环条件检查位置（跳回目标）
	rangeCondStart := len(c.instructions)
	// 进入循环作用域
	c.pushLoop(rangeCondStart)

	// 编译条件：idx < len
	c.emit1(OpLoadLocal, idxVarIdx)
	c.emit1(OpLoadLocal, lenVarIdx)
	c.emit(OpLess)

	// 条件为假时跳出循环
	rangeJumpIfFalseIdx := len(c.instructions)
	c.emit1(OpJumpIfFalse, 0)

	// 获取当前元素
	// 检查是否是Map(通过mapVarIdx是否为nil)
	c.emit1(OpLoadLocal, mapVarIdx)
	c.emit(OpNil)
	c.emit(OpEqual)
	isNotMapJump := c.emit1(OpJumpIfFalse, 0)
	// 是数组: 变量 = collection[idx] (元素值)
	c.emit1(OpLoadLocal, collectionVarIdx)
	c.emit1(OpLoadLocal, idxVarIdx)
	c.emit(OpIndex)
	c.emit1(OpStoreLocal, keyVarIdx)
	skipMapBinding := c.emit1(OpJump, 0)
	// 是Map: 变量 = keys[idx] (string key)
	c.instructions[isNotMapJump].Args[0] = len(c.instructions)
	c.emit1(OpLoadLocal, collectionVarIdx)
	c.emit1(OpLoadLocal, idxVarIdx)
	c.emit(OpIndex)
	c.emit1(OpStoreLocal, keyVarIdx)
	c.instructions[skipMapBinding].Args[0] = len(c.instructions)

	// 编译循环体
	if err := c.compileBlockStmt(stmt.Body); err != nil {
		return err
	}

	// 记录 idx++ 位置
	rangePostStart := len(c.instructions)

	// idx++
	c.emit1(OpLoadLocal, idxVarIdx)
	c.emit1(OpConst, c.addConstant(NewValue(1)))
	c.emit(OpAdd)
	c.emit1(OpStoreLocal, idxVarIdx)

	// 跳回条件检查
	c.emit1(OpJump, rangeCondStart)

	// Range 循环结束
	rangeLoopEnd := len(c.instructions)
	c.instructions[rangeJumpIfFalseIdx].Args[0] = rangeLoopEnd
	c.patchBreaks(rangeLoopEnd)
	c.patchContinues(rangePostStart)
	c.popLoop()
	c.unshadowLoopVar(dynamicKeyShadow)

	// 回填：跳过 range 循环的跳转
	c.instructions[skipRangeJumpIdx].Args[0] = rangeLoopEnd

	return nil
}

// isLikelyRangeLoop 判断是否可能是range循环
// 只有数组/Map字面量才确定为range循环
// 标识符需要运行时判断类型
func (c *Compiler) isLikelyRangeLoop(expr Expr) bool {
	// 只有数组或Map字面量才确定为range循环
	switch expr.(type) {
	case *ArrayExpr:
		return true
	case *MapExpr:
		return true
	}

	// 标识符需要运行时判断类型，由 compileDynamicFor 处理
	return false
}

// compileStandardFor 编译标准for循环
// 语法: for init; cond; post { body }
// 类似C语言的for循环，各部分可以省略
// continue 跳转到 post 语句执行，break 跳转到循环结束
func (c *Compiler) compileStandardFor(stmt *ForStmt) error {
	// 编译初始化语句（可能为空）
	if stmt.Init != nil {
		if err := c.compileStmt(stmt.Init); err != nil {
			return err
		}
	}

	// 记录循环开始位置（条件判断）
	loopStart := len(c.instructions)
	// 进入循环作用域（continue 目标先设为 loopStart，有 post 时会更新）
	c.pushLoop(loopStart)

	// 编译条件表达式（可能为空，空条件等价于true）
	if stmt.Cond != nil {
		if err := c.compileExpr(stmt.Cond); err != nil {
			return err
		}
	} else {
		// 空条件，相当于 true，总是继续循环
		c.emit1(OpConst, c.addConstant(NewValue(true)))
	}

	// 条件为假时跳出循环（占位）
	jumpIfFalseIdx := len(c.instructions)
	c.emit1(OpJumpIfFalse, 0)

	// 编译循环体
	if err := c.compileBlockStmt(stmt.Body); err != nil {
		return err
	}

	// 记录 post 语句位置（continue 跳转目标）
	postStart := len(c.instructions)

	// 编译post语句（可能为空）
	if stmt.Post != nil {
		if err := c.compileStmt(stmt.Post); err != nil {
			return err
		}
	}

	// 跳回循环开始（条件判断）
	c.emit1(OpJump, loopStart)

	// 回填：条件为假时跳到这里
	loopEnd := len(c.instructions)
	c.instructions[jumpIfFalseIdx].Args[0] = loopEnd

	// 回填所有 break 和 continue 跳转目标
	// break 跳到循环结束，continue 跳到 post 语句
	c.patchBreaks(loopEnd)
	c.patchContinues(postStart)
	// 退出循环作用域
	c.popLoop()

	return nil
}

// getOrAllocLocalVar 获取或分配局部变量索引
// 如果变量已存在，返回现有索引；否则分配新索引
func (c *Compiler) getOrAllocLocalVar(name string) int {
	if idx, ok := c.localVars[name]; ok {
		return idx
	}
	idx := len(c.localVars)
	c.localVars[name] = idx
	return idx
}

// allocLoopTemp 分配循环专用临时变量
// 根据当前循环嵌套深度生成唯一名称, 避免嵌套循环临时变量冲突
func (c *Compiler) allocLoopTemp(base string) int {
	depth := len(c.loopBreaks)
	return c.getOrAllocLocalVar(fmt.Sprintf("__for_%s_%d", base, depth))
}

// compileRangeFor 编译range循环
// 语法: for k := arr { body }
// 遍历数组或Map的每个元素
// continue 跳转到 idx++ 执行，break 跳转到循环结束
func (c *Compiler) compileRangeFor(stmt *ForStmt) error {
	initValue := stmt.Init.(*VarDeclStmt).Value

	// 编译要遍历的集合（数组或Map）
	if err := c.compileExpr(initValue); err != nil {
		return err
	}

	// 存储集合到局部变量
	collectionVarIdx := c.allocLoopTemp("collection")
	c.emit1(OpStoreLocal, collectionVarIdx)

	// Map处理: MapLiteral在编译期可知, 标识符需运行时检测
	isMap := false
	isDynamic := false
	var mapVarIdx int
	if _, ok := initValue.(*MapExpr); ok {
		isMap = true
		mapVarIdx = c.allocLoopTemp("map")
		c.emit1(OpLoadLocal, collectionVarIdx)
		c.emit1(OpStoreLocal, mapVarIdx)
		c.emit1(OpLoadLocal, mapVarIdx)
		c.emit(OpMapKeys)
		c.emit1(OpStoreLocal, collectionVarIdx)
	} else if _, ok := initValue.(*IdentExpr); ok {
		isDynamic = true
		mapVarIdx = c.allocLoopTemp("map")
		// 初始化为nil, 运行时若为map则覆盖为原始map
		c.emit(OpNil)
		c.emit1(OpStoreLocal, mapVarIdx)
		// 运行时检测: 若是Map则提取keys替换collection
		c.emit1(OpLoadLocal, collectionVarIdx)
		c.emit(OpTypeOf)
		c.emit1(OpConst, c.addConstant(NewValue("map")))
		c.emit(OpEqual)
		notMapJump := c.emit1(OpJumpIfFalse, 0)
		// 是Map: 保存原始map, 提取keys
		c.emit1(OpLoadLocal, collectionVarIdx)
		c.emit1(OpStoreLocal, mapVarIdx)
		c.emit1(OpLoadLocal, mapVarIdx)
		c.emit(OpMapKeys)
		c.emit1(OpStoreLocal, collectionVarIdx)
		notMapIdx := len(c.instructions)
		c.instructions[notMapJump].Args[0] = notMapIdx
	}

	// 创建索引变量idx=0
	idxVarIdx := c.allocLoopTemp("idx")
	c.emit1(OpConst, c.addConstant(NewValue(0)))
	c.emit1(OpStoreLocal, idxVarIdx)

	// 获取集合长度
	lenVarIdx := c.allocLoopTemp("len")
	c.emit1(OpLoadLocal, collectionVarIdx)
	c.emit1(OpLen, 0)
	c.emit1(OpStoreLocal, lenVarIdx)

	// 创建键变量k（用户定义的变量名）
	keyVarName := stmt.Init.(*VarDeclStmt).Name
	keyVarIdx, keyShadow := c.shadowLoopVar(keyVarName)

	// 双变量模式下预分配value变量
	valueVarIdx := -1
	var valueShadow loopVarShadow
	if stmt.RangeValueVar != "" {
		valueVarIdx, valueShadow = c.shadowLoopVar(stmt.RangeValueVar)
	}

	// 记录循环开始位置
	loopStart := len(c.instructions)
	// 进入循环作用域
	c.pushLoop(loopStart)

	// 编译条件：idx < len
	c.emit1(OpLoadLocal, idxVarIdx)
	c.emit1(OpLoadLocal, lenVarIdx)
	c.emit(OpLess)

	// 条件为假时跳出循环（占位）
	jumpIfFalseIdx := len(c.instructions)
	c.emit1(OpJumpIfFalse, 0)

	// 绑定循环变量
	if valueVarIdx >= 0 {
		if isMap {
			// Map字面量双变量: key = keys[idx], value = map[key]
			c.emit1(OpLoadLocal, collectionVarIdx)
			c.emit1(OpLoadLocal, idxVarIdx)
			c.emit(OpIndex)
			c.emit1(OpStoreLocal, keyVarIdx)
			c.emit1(OpLoadLocal, mapVarIdx)
			c.emit1(OpLoadLocal, keyVarIdx)
			c.emit(OpIndex)
			c.emit1(OpStoreLocal, valueVarIdx)
		} else if isDynamic {
			// 运行时检测: 根据mapVarIdx是否为nil决定绑定方式
			// 检查mapVarIdx是否为nil
			c.emit1(OpLoadLocal, mapVarIdx)
			c.emit(OpNil)
			c.emit(OpEqual)
			isNotMapJump := c.emit1(OpJumpIfFalse, 0)
			// 是数组: key = idx, value = collection[idx]
			c.emit1(OpLoadLocal, idxVarIdx)
			c.emit1(OpStoreLocal, keyVarIdx)
			c.emit1(OpLoadLocal, collectionVarIdx)
			c.emit1(OpLoadLocal, idxVarIdx)
			c.emit(OpIndex)
			c.emit1(OpStoreLocal, valueVarIdx)
			skipMapBinding := c.emit1(OpJump, 0)
			// 是Map: key = keys[idx], value = map[key]
			c.instructions[isNotMapJump].Args[0] = len(c.instructions)
			c.emit1(OpLoadLocal, collectionVarIdx)
			c.emit1(OpLoadLocal, idxVarIdx)
			c.emit(OpIndex)
			c.emit1(OpStoreLocal, keyVarIdx)
			c.emit1(OpLoadLocal, mapVarIdx)
			c.emit1(OpLoadLocal, keyVarIdx)
			c.emit(OpIndex)
			c.emit1(OpStoreLocal, valueVarIdx)
			c.instructions[skipMapBinding].Args[0] = len(c.instructions)
		} else {
			// 数组双变量: key = idx, value = collection[idx]
			c.emit1(OpLoadLocal, idxVarIdx)
			c.emit1(OpStoreLocal, keyVarIdx)
			c.emit1(OpLoadLocal, collectionVarIdx)
			c.emit1(OpLoadLocal, idxVarIdx)
			c.emit(OpIndex)
			c.emit1(OpStoreLocal, valueVarIdx)
		}
	} else {
		// 单变量: key = collection[idx]
		// 数组: collection是原数组, 得到元素值
		// Map: collection是keys数组, 得到string key
		c.emit1(OpLoadLocal, collectionVarIdx)
		c.emit1(OpLoadLocal, idxVarIdx)
		c.emit(OpIndex)
		c.emit1(OpStoreLocal, keyVarIdx)
	}

	// 编译循环体
	if err := c.compileBlockStmt(stmt.Body); err != nil {
		return err
	}

	// 记录 idx++ 位置（continue 跳转目标）
	postStart := len(c.instructions)

	// idx++
	c.emit1(OpLoadLocal, idxVarIdx)
	c.emit1(OpConst, c.addConstant(NewValue(1)))
	c.emit(OpAdd)
	c.emit1(OpStoreLocal, idxVarIdx)

	// 跳回循环开始
	c.emit1(OpJump, loopStart)

	// 回填：条件为假时跳到这里
	loopEnd := len(c.instructions)
	c.instructions[jumpIfFalseIdx].Args[0] = loopEnd

	// 回填所有 break 和 continue 跳转目标
	// break 跳到循环结束，continue 跳到 idx++ 处
	c.patchBreaks(loopEnd)
	c.patchContinues(postStart)
	// 退出循环作用域
	c.popLoop()
	// 恢复被遮蔽的外层同名变量绑定
	c.unshadowLoopVar(keyShadow)
	if valueVarIdx >= 0 {
		c.unshadowLoopVar(valueShadow)
	}

	return nil
}

// compileBreakStmt 编译break语句
// break 语句用于跳出当前循环，跳转目标在循环结束时回填
func (c *Compiler) compileBreakStmt(stmt *BreakStmt) error {
	if !c.inLoop() {
		return NewCompileErrorFromPos(stmt,
			"语法错误：break 语句必须在循环内部使用。\n"+
				"→ 问题：break 只能用于 for 循环中。\n"+
				"→ 建议：检查 break 是否误写在循环外部")
	}
	// 生成跳转到循环结束的指令（目标稍后回填）
	pos := c.emit1(OpJump, 0)
	c.addBreak(pos)
	return nil
}

// compileContinueStmt 编译continue语句
// continue 语句用于跳过当前迭代，跳转到循环条件检查处
func (c *Compiler) compileContinueStmt(stmt *ContinueStmt) error {
	if !c.inLoop() {
		return NewCompileErrorFromPos(stmt,
			"语法错误：continue 语句必须在循环内部使用。\n"+
				"→ 问题：continue 只能用于 for 循环中。\n"+
				"→ 建议：检查 continue 是否误写在循环外部")
	}
	// 生成跳转到循环开始的指令（目标稍后回填）
	pos := c.emit1(OpJump, 0)
	c.addContinue(pos)
	return nil
}

// compilerState 保存编译器状态
type compilerState struct {
	instructions        []Instruction
	constants           []Value
	localVars           map[string]int
	constantCache       map[cacheKey]int
	pendingFunctionRefs map[int]int
	sliceStartIdx       int
	sliceEndIdx         int
}

// saveState 保存当前编译器状态
// 覆盖 resetForFunction 重置的全部字段, 保证嵌套函数编译后状态完整恢复
func (c *Compiler) saveState() compilerState {
	return compilerState{
		instructions:        c.instructions,
		constants:           c.constants,
		localVars:           c.localVars,
		constantCache:       c.constantCache,
		pendingFunctionRefs: c.pendingFunctionRefs,
		sliceStartIdx:       c.sliceStartIdx,
		sliceEndIdx:         c.sliceEndIdx,
	}
}

// restoreState 恢复编译器状态
func (c *Compiler) restoreState(state compilerState) {
	c.instructions = state.instructions
	c.constants = state.constants
	c.localVars = state.localVars
	c.constantCache = state.constantCache
	c.pendingFunctionRefs = state.pendingFunctionRefs
	c.sliceStartIdx = state.sliceStartIdx
	c.sliceEndIdx = state.sliceEndIdx
}

// resetForFunction 重置编译器为函数编译环境
func (c *Compiler) resetForFunction() {
	c.instructions = nil
	c.constants = nil
	c.localVars = make(map[string]int)
	c.constantCache = make(map[cacheKey]int)
	// 初始化待修复函数引用表
	c.pendingFunctionRefs = make(map[int]int)
	// 重新初始化切片常量
	c.sliceStartIdx = c.addConstant(NewValue(0))
	c.sliceEndIdx = c.addConstant(NewValue(SliceEndDefault))
}

// compileFuncDecl 编译函数声明
// 支持递归调用:函数名在函数体内可见
func (c *Compiler) compileFuncDecl(stmt *FuncDeclStmt) error {
	// 在外部常量池中预先创建一个占位符,用于递归调用
	placeholderIdx := c.addConstant(NewValue(nil))

	// 预先注册函数常量索引,支持递归调用
	if c.functionConstIndices == nil {
		c.functionConstIndices = make(map[string]int)
	}
	c.functionConstIndices[stmt.Name] = placeholderIdx

	// 保存编译状态
	oldState := c.saveState()
	c.resetForFunction()

	// 参数从槽位0开始
	for _, param := range stmt.Params {
		c.localVars[param.Name] = len(c.localVars)
	}

	// 编译函数体,使用compileBlockExpr保留最后一个表达式的值作为返回值
	if err := c.compileBlockExpr(stmt.Body); err != nil {
		return err
	}

	// 添加隐式return
	c.emit(OpReturn)

	// 创建函数
	fn := &CompiledFunction{
		Name:         stmt.Name,
		Instructions: c.instructions,
		Constants:    c.constants,
		NumLocals:    len(c.localVars),
		NumParams:    len(stmt.Params),
	}
	c.functions = append(c.functions, fn)

	// 捕获本函数的待修复引用, 避免恢复外层状态后归属错位
	pendingRefs := c.pendingFunctionRefs

	// 恢复编译状态
	c.restoreState(oldState)

	// 创建函数值并更新占位符
	fnValue := NewValue(&FunctionValue{
		Compiled: fn,
	})
	c.constants[placeholderIdx] = fnValue

	// 引用修复延后到编译收尾统一执行, 此时主池占位才全部填充完毕
	c.funcRefFixes = append(c.funcRefFixes, funcRefFix{fn: fn, refs: pendingRefs})

	// 注册函数名到局部变量表
	if _, exists := c.localVars[stmt.Name]; !exists {
		c.localVars[stmt.Name] = len(c.localVars)
	}

	// 生成代码：加载函数常量并存储到局部变量
	c.emit1(OpConst, placeholderIdx)

	c.emit1(OpStoreLocal, c.localVars[stmt.Name])

	return nil
}

// compileReturnStmt 编译return语句
func (c *Compiler) compileReturnStmt(stmt *ReturnStmt) error {
	if err := c.compileExprOrNil(stmt.Value); err != nil {
		return err
	}
	c.emit(OpReturn)
	return nil
}

// compileThrowStmt 编译throw语句
func (c *Compiler) compileThrowStmt(stmt *ThrowStmt) error {
	if err := c.compileExpr(stmt.Value); err != nil {
		return err
	}
	c.emit(OpThrow)
	return nil
}

// compileDefDirective 编译#def指令
func (c *Compiler) compileDefDirective(stmt *DefDirectiveStmt) error {
	idx := len(c.externals)
	c.externals = append(c.externals, ExternalFunc{
		Name:   stmt.Name,
		Params: stmt.Params,
		Return: stmt.Return,
	})
	// 缓存外部函数索引
	c.externalCache[stmt.Name] = idx
	return nil
}
