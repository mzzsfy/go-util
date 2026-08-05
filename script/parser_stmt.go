package script

import (
	"fmt"
	"strings"
)

// ========== 语句解析 ==========
// 本文件包含所有语句类型的解析逻辑
// 包括：变量声明、函数定义、控制流语句、类型解析等

// parseVarDecl 解析变量声明语句
// 语法格式: name := value 或 name: type = value
// 示例: x := 1, name: string = "hello"
// 特殊情况: for循环中的 name := range arr 不会解析值
func (p *Parser) parseVarDecl() *VarDeclStmt {
	stmt := &VarDeclStmt{
		Position: p.currentPosition(),
		Name:     p.curToken.Value,
	}
	p.nextToken()
	stmt.TypeAnnot = p.tryParseTypeAnnotation()
	if p.curTokenIs(TokenAssign) {
		p.nextToken()
	}
	// 特殊处理：如果下一个token是range，不解析值（用于for循环）
	if !p.curTokenIs(TokenRange) {
		stmt.Value = p.parseExpr(LOWEST)
	}
	return stmt
}

// parseTypedVarDecl 解析类型安全变量声明语句
// 语法格式: name :=>type value
// 示例: x :=>int 10, name :=>string "hello"
// 类型是必须的，不是可选的
func (p *Parser) parseTypedVarDecl() *VarDeclStmt {
	stmt := &VarDeclStmt{
		Position: p.currentPosition(),
		Name:     p.curToken.Value,
	}
	// 跳过 :=>
	p.nextToken()
	// 跳过 :=> 到达类型
	p.nextToken()
	// 解析类型表达式
	stmt.TypeAnnot = p.parseTypeExpr()
	if stmt.TypeAnnot == nil {
		p.addError(fmt.Sprintf(
			"类型错误：':=>' 后必须指定有效类型（第%d行第%d列）。\n"+
				"→ 问题：类型安全声明需要明确的类型。\n"+
				"→ 正确格式：name :=>type value\n"+
				"→ 示例：x :=>int 10, name :=>string \"hello\"\n"+
				"→ 支持的类型：int, float, string, bool, any, array, function\n"+
				"→ 类型别名：i, f, s, str, b, arr, fn",
			p.curToken.Line, p.curToken.Column))
		return nil
	}
	// 前进到值表达式
	p.nextToken()
	// 解析值表达式
	stmt.Value = p.parseExpr(LOWEST)
	return stmt
}

// parseExprStmt 解析表达式语句
// 将表达式包装为语句，用于独立表达式
// 示例: funcCall(), x + y
func (p *Parser) parseExprStmt() *ExprStmt {
	return &ExprStmt{
		Position: p.currentPosition(),
		Expr:     p.parseExpr(LOWEST),
	}
}

// parseFuncDecl 解析函数声明
// 语法格式: fn name(params) -> returnType { body }
// 示例: fn add(a, b) -> int { return a + b }
func (p *Parser) parseFuncDecl() *FuncDeclStmt {
	stmt := &FuncDeclStmt{Position: p.currentPosition()}
	if !p.expectPeek(TokenIdent) {
		return nil
	}
	stmt.Name = p.curToken.Value
	if !p.expectPeek(TokenLParen) {
		return nil
	}
	stmt.Params = p.parseParams()
	if !p.curTokenIs(TokenRParen) {
		return nil
	}
	if p.peekTokenIs(TokenArrow) {
		p.nextToken()
		p.nextToken()
		stmt.Return = p.parseTypeExpr()
	}
	stmt.Body = p.expectAndParseBlock()
	if stmt.Body == nil {
		return nil
	}
	return stmt
}

// parseParams 解析函数参数列表
// 语法格式: (param1, param2, ...)
// 支持空参数列表和带类型注解的参数
func (p *Parser) parseParams() []Param {
	if p.peekTokenIs(TokenRParen) {
		p.nextToken()
		return nil
	}
	p.nextToken()
	params := p.parseParamList(nil)
	if !p.expectPeek(TokenRParen) {
		return nil
	}
	return params
}

// parseParamList 解析参数列表中的多个参数
// 处理逗号分隔的参数列表
func (p *Parser) parseParamList(params []Param) []Param {
	params = append(params, p.parseSingleParam())
	for p.peekTokenIs(TokenComma) {
		p.nextToken()
		p.nextToken()
		params = append(params, p.parseSingleParam())
	}
	return params
}

// parseSingleParam 解析单个参数
// 支持可选的类型注解: name -> type
func (p *Parser) parseSingleParam() Param {
	param := Param{Name: p.curToken.Value}
	if p.peekTokenIs(TokenArrow) {
		p.nextToken()
		p.nextToken()
		param.TypeAnnot = p.parseTypeExpr()
	}
	return param
}

// parseBlockStmt 解析块语句
// 语法格式: { statement1 statement2 ... }
// 块语句包含零个或多个语句
func (p *Parser) parseBlockStmt() *BlockStmt {
	var statements []Stmt
	for !p.curTokenIs(TokenRBrace) && !p.curTokenIs(TokenEOF) {
		if stmt := p.parseStatement(); stmt != nil {
			statements = append(statements, stmt)
		}
		p.nextToken()
	}
	return &BlockStmt{
		Position:   p.currentPosition(),
		Statements: statements,
	}
}

// expectAndParseBlock 期望左花括号并解析块语句
func (p *Parser) expectAndParseBlock() *BlockStmt {
	if !p.expectPeek(TokenLBrace) {
		return nil
	}
	p.nextToken()
	return p.parseBlockStmt()
}

// ========== 控制流语句解析 ==========
// 包含if条件语句、for循环语句、return/break/continue等控制流解析

// parseIfStmt 解析if条件语句
// 语法格式: if condition { then } else { else }
// parseExpr自动处理条件中的括号, 无需手动消费
// 示例: if x > 0 { return x }, if (1+2)*3 > 5 { 1 }
func (p *Parser) parseIfStmt() *IfStmt {
	stmt := &IfStmt{Position: p.currentPosition()}
	p.nextToken()

	stmt.Condition = p.parseExpr(LOWEST)

	stmt.Then = p.expectAndParseBlock()
	if stmt.Then == nil {
		return nil
	}
	return p.finalizeIfStmt(stmt)
}

// finalizeIfStmt 完成if语句的else分支解析
// 处理可选的else分支，支持else if链
func (p *Parser) finalizeIfStmt(stmt *IfStmt) *IfStmt {
	if !p.peekTokenIs(TokenElse) {
		return stmt
	}
	p.nextToken()
	elseStmt, ok := p.parseIfElseBranch()
	if !ok {
		return nil
	}
	stmt.Else = elseStmt
	return stmt
}

// parseIfElseBranch 解析if语句的else分支
// 支持两种形式: else if 或 else { }
func (p *Parser) parseIfElseBranch() (Stmt, bool) {
	if p.advanceIfPeekIs(TokenIf) {
		return p.parseIfStmt(), true
	}
	block := p.expectAndParseBlock()
	if block == nil {
		return nil, false
	}
	return block, true
}

// parseForStmt 解析for循环语句
// 支持多种格式:
//   - 无限循环: for { body } 或 for ( ) { body }
//   - 条件循环: for condition { body } 或 for (condition) { body }
//   - 计数循环: for i := n { body } 或 for (i := n) { body }
//   - 标准循环: for i := 0; i < n; i++ { body } 或 for (i := 0; i < n; i++) { body }
//   - Range循环: for k := range arr { body } 或 for (k := range arr) { body }
//   - 省略形式: for i := 0; ; i++ { } 或 for ; ; { } 等
func (p *Parser) parseForStmt() *ForStmt {
	stmt := &ForStmt{Position: p.currentPosition()}
	p.nextToken()

	// 跳过可选的左括号
	hasParens := false
	if p.curTokenIs(TokenLParen) {
		hasParens = true
		p.nextToken()
	}

	switch {
	case p.curTokenIs(TokenLBrace) || (hasParens && p.curTokenIs(TokenRParen)):
		// for { } 或 for ( )
		if hasParens && p.curTokenIs(TokenRParen) {
			p.nextToken() // 跳过右括号
		}
		return p.parseInfiniteForLoop(stmt)
	case p.curTokenIs(TokenSemicolon):
		// for ; ... 省略init的标准循环
		return p.parseStandardForLoopWithoutInit(stmt, hasParens)
	case p.curTokenIs(TokenIdent) && p.peekTokenIs(TokenAssign):
		return p.parseInitForLoop(stmt, hasParens)
	case p.curTokenIs(TokenIdent) && p.peekTokenIs(TokenComma):
		return p.parseTwoVarRangeForLoop(stmt, hasParens)
	default:
		return p.parseCondOrRangeForLoop(stmt, hasParens)
	}
}

// parseInfiniteForLoop 解析无限循环
// 语法格式: for { body }
// 等价于 for true { body }
func (p *Parser) parseInfiniteForLoop(stmt *ForStmt) *ForStmt {
	stmt.Mode = ForWhile
	stmt.Cond = &LiteralExpr{Type: LiteralBool, Value: true}
	p.nextToken() // 前进到块内第一个语句
	stmt.Body = p.parseBlockStmt()
	return stmt
}

// parseInitForLoop 解析带初始化的循环
// 根据后续token判断是计数循环、range循环还是标准循环
func (p *Parser) parseInitForLoop(stmt *ForStmt, hasParens bool) *ForStmt {
	initStmt := p.parseVarDecl()

	// 检查是否是range循环: for k := range arr
	// 注意：parseVarDecl在遇到range时会停止，所以当前token应该在range
	if p.curTokenIs(TokenRange) {
		p.nextToken() // 跳过range关键字
		return p.parseExplicitRangeForLoop(stmt, initStmt, hasParens)
	}

	// 前进到下一个token以便检查
	p.nextToken()

	// 检查是否是计数循环：for i := n { } 或 for (i := n)
	if p.curTokenIs(TokenLBrace) || (hasParens && p.curTokenIs(TokenRParen)) {
		if hasParens && p.curTokenIs(TokenRParen) {
			p.nextToken() // 跳过 )
		}
		return p.parseCountForLoop(stmt, initStmt)
	}

	// 否则是标准for循环
	return p.parseStandardForLoopWithInit(stmt, initStmt, hasParens)
}

// parseCountForLoop 解析次数循环
// 语法格式: for i := n { body }
// 循环n次，i从1到n
func (p *Parser) parseCountForLoop(stmt *ForStmt, initStmt *VarDeclStmt) *ForStmt {
	stmt.Mode = ForCount
	stmt.Init = initStmt
	p.nextToken() // 前进到块内第一个语句
	stmt.Body = p.parseBlockStmt()
	return stmt
}

// parseStandardForLoopWithInit 解析带初始化的标准for循环
// 语法格式: for init; cond; post { body } 或 for (init; cond; post) { body }
func (p *Parser) parseStandardForLoopWithInit(stmt *ForStmt, initStmt *VarDeclStmt, hasParens bool) *ForStmt {
	stmt.Mode = ForStandard
	stmt.Init = initStmt

	// 当前token应该是分号
	if !p.curTokenIs(TokenSemicolon) {
		p.addError(fmt.Sprintf(
			"语法错误：标准for循环需要分号分隔各部分（第%d行）。\n"+
				"→ 正确格式：for init; cond; post { body }\n"+
				"→ 示例：for i := 0; i < 10; i = i + 1 { print(i) }\n"+
				"→ 注意：各部分之间必须用分号 ; 分隔",
			p.curToken.Line))
		return nil
	}

	// 解析条件表达式（可能为空）
	p.nextToken()
	if !p.curTokenIs(TokenSemicolon) {
		stmt.Cond = p.parseExpr(LOWEST)
		// 前进到下一个token（应该是分号）
		p.nextToken()
	}

	// 期望第二个分号
	if !p.curTokenIs(TokenSemicolon) {
		p.addError(fmt.Sprintf(
			"语法错误：标准for循环需要分号分隔各部分（第%d行）。\n"+
				"→ 正确格式：for init; cond; post { body }\n"+
				"→ 示例：for i := 0; i < 10; i = i + 1 { print(i) }\n"+
				"→ 注意：各部分之间必须用分号 ; 分隔",
			p.curToken.Line))
		return nil
	}

	// 解析post语句（可能为空）
	p.nextToken()
	if hasParens && p.curTokenIs(TokenRParen) {
		// for (init; cond;) { } - post为空且有右括号
		// 当前在 )，不要前进，让后面的逻辑处理
	} else if !p.curTokenIs(TokenRParen) && !p.curTokenIs(TokenLBrace) {
		// 有post语句
		stmt.Post = p.parseSimpleStmt()
		// 前进到下一个token（可能是 ) 或 {）
		p.nextToken()
	}

	// 如果有右括号，跳过它
	if hasParens && p.curTokenIs(TokenRParen) {
		p.nextToken()
	}

	// 解析循环体
	// 当前应该在 { ，需要检查并进入块
	if !p.curTokenIs(TokenLBrace) {
		// 尝试使用 expectAndParseBlock（它会 peek 检查）
		stmt.Body = p.expectAndParseBlock()
	} else {
		// 当前已经在 {，直接解析块
		p.nextToken() // 前进到块内第一个语句
		stmt.Body = p.parseBlockStmt()
	}
	if stmt.Body == nil {
		return nil
	}
	return stmt
}

// parseStandardForLoopWithoutInit 解析省略init的标准for循环
// 语法格式: for ; cond; post { body } 或 for (; cond; post) { body }
func (p *Parser) parseStandardForLoopWithoutInit(stmt *ForStmt, hasParens bool) *ForStmt {
	stmt.Mode = ForStandard

	// 当前token是分号，解析条件表达式（可能为空）
	p.nextToken()
	if !p.curTokenIs(TokenSemicolon) {
		stmt.Cond = p.parseExpr(LOWEST)
		// 前进到下一个token（应该是分号）
		p.nextToken()
	}

	// 期望第二个分号
	if !p.curTokenIs(TokenSemicolon) {
		p.addError(fmt.Sprintf(
			"语法错误：标准for循环需要分号分隔各部分（第%d行）。\n"+
				"→ 正确格式：for ; cond; post { body }\n"+
				"→ 示例：for ; i < 10; i = i + 1 { print(i) }\n"+
				"→ 注意：各部分之间必须用分号 ; 分隔",
			p.curToken.Line))
		return nil
	}

	// 解析post语句（可能为空）
	p.nextToken()
	if hasParens && p.curTokenIs(TokenRParen) {
		// for (; cond;) { } - post为空且有右括号
		// 当前在 )，不要前进，让后面的逻辑处理
	} else if !p.curTokenIs(TokenRParen) && !p.curTokenIs(TokenLBrace) {
		// 有post语句
		stmt.Post = p.parseSimpleStmt()
		// 前进到下一个token（可能是 ) 或 {）
		p.nextToken()
	}

	// 如果有右括号，跳过它
	if hasParens && p.curTokenIs(TokenRParen) {
		p.nextToken()
	}

	// 解析循环体
	// 当前应该在 { ，需要检查并进入块
	if !p.curTokenIs(TokenLBrace) {
		// 尝试使用 expectAndParseBlock（它会 peek 检查）
		stmt.Body = p.expectAndParseBlock()
	} else {
		// 当前已经在 {，直接解析块
		p.nextToken() // 前进到块内第一个语句
		stmt.Body = p.parseBlockStmt()
	}
	if stmt.Body == nil {
		return nil
	}
	return stmt
}

// parseExplicitRangeForLoop 解析显式range关键字的遍历循环
// 语法格式: for k := range arr { body } 或 for (k := range arr) { body }
func (p *Parser) parseExplicitRangeForLoop(stmt *ForStmt, initStmt *VarDeclStmt, hasParens bool) *ForStmt {
	// 跳过range关键字已经在调用者完成
	// 解析要遍历的集合
	stmt.Mode = ForRange
	stmt.Init = &VarDeclStmt{
		Position: initStmt.Position,
		Name:     initStmt.Name,
		Value:    p.parseExpr(LOWEST),
	}

	// 前进到下一个token（可能是 ) 或 {）
	p.nextToken()

	// 处理可选的右括号
	if hasParens && p.curTokenIs(TokenRParen) {
		p.nextToken()
	}

	// 解析循环体
	// 当前应该在 { ，需要检查并进入块
	if !p.curTokenIs(TokenLBrace) {
		// 尝试使用 expectAndParseBlock（它会 peek 检查）
		stmt.Body = p.expectAndParseBlock()
	} else {
		// 当前已经在 {，直接解析块
		p.nextToken() // 前进到块内第一个语句
		stmt.Body = p.parseBlockStmt()
	}
	if stmt.Body == nil {
		return nil
	}
	return stmt
}

// parseCondOrRangeForLoop 解析条件循环或遍历循环
// 根据后续token判断是while循环还是range循环
func (p *Parser) parseCondOrRangeForLoop(stmt *ForStmt, hasParens bool) *ForStmt {
	expr := p.parseExpr(LOWEST)

	// 前进到下一个token以检查后续内容
	p.nextToken()

	// 带括号时, ) 后可能还有表达式运算符 (如 for (i+1)*2 > 0 { })
	if hasParens && p.curTokenIs(TokenRParen) {
		expr = p.parseInfixExpr(expr, LOWEST)
		p.nextToken()
	}

	// 检查是否是条件循环：后面直接跟 { 或 )
	// 如果是 )，说明有括号，需要跳过然后检查 {
	if p.curTokenIs(TokenRParen) {
		// for (cond) { body }
		p.nextToken() // 跳过 )
		// 现在应该是 {
		if !p.curTokenIs(TokenLBrace) {
			p.addError(fmt.Sprintf(
				"语法错误：期望 '{'，但得到 %v（第%d行）。\n"+
					"→ 问题：for循环的循环体必须用花括号包围。\n"+
					"→ 正确格式：for (condition) { body }\n"+
					"→ 示例：for (i < 10) { print(i) }",
				p.curToken.Type, p.curToken.Line))
			return nil
		}
		return p.parseWhileForLoop(stmt, expr)
	}

	if p.curTokenIs(TokenLBrace) {
		// for cond { body }
		return p.parseWhileForLoop(stmt, expr)
	}

	// 否则是range循环: for k := arr 或 for k := range arr
	return p.parseRangeForLoop(stmt, expr)
}

// parseWhileForLoop 解析条件循环
// 语法格式: for condition { body }
// 当condition为真时执行循环体
func (p *Parser) parseWhileForLoop(stmt *ForStmt, cond Expr) *ForStmt {
	stmt.Mode = ForWhile
	stmt.Cond = cond
	p.nextToken() // 前进到块内第一个语句
	stmt.Body = p.parseBlockStmt()
	return stmt
}

// parseRangeForLoop 解析遍历循环
// 语法格式: for k := arr { body }
// 遍历数组或Map的每个元素
func (p *Parser) parseRangeForLoop(stmt *ForStmt, keyExpr Expr) *ForStmt {
	identExpr, ok := keyExpr.(*IdentExpr)
	if !ok {
		p.addError(fmt.Sprintf(
			"语法错误：range循环的键必须是标识符（第%d行）。\n"+
				"→ 问题：for...range循环要求键变量必须是简单的标识符。\n"+
				"→ 正确示例：for k := arr { print(k) }\n"+
				"→ 错误示例：for arr[0] := arr { ... }  // 不能使用表达式作为键\n"+
				"→ 建议：使用简单的变量名作为键，如 k、key、item",
			p.curToken.Line))
		return nil
	}
	if !p.curTokenIs(TokenAssign) {
		p.addError(fmt.Sprintf(
			"语法错误：期望 := 运算符，但得到 %v（第%d行）。\n"+
				"→ 问题：range循环必须使用 := 进行变量声明。\n"+
				"→ 正确格式：for k := arr { ... }\n"+
				"→ 错误格式：for k = arr { ... }  // 应使用 := 而非 =\n"+
				"→ 建议：将 = 改为 :=",
			p.curToken.Type, p.curToken.Line))
		return nil
	}
	p.nextToken()
	stmt.Mode = ForRange
	stmt.Init = &VarDeclStmt{
		Position: p.currentPosition(),
		Name:     identExpr.Name,
		Value:    p.parseExpr(LOWEST),
	}
	stmt.Body = p.expectAndParseBlock()
	if stmt.Body == nil {
		return nil
	}
	return stmt
}

// parseTwoVarRangeForLoop 解析双变量range循环
// 语法格式: for i, v := range arr { body } 或 for (i, v := range arr) { body }
// 也支持隐式形式: for i, v := arr { body }
func (p *Parser) parseTwoVarRangeForLoop(stmt *ForStmt, hasParens bool) *ForStmt {
	// 读取第一个变量名（key/index变量）
	keyName := p.curToken.Value
	p.nextToken() // 跳过key标识符

	// 期望逗号
	if !p.curTokenIs(TokenComma) {
		p.addError(fmt.Sprintf(
			"语法错误：期望 ',' 分隔双变量（第%d行）。\n"+
				"→ 格式：for i, v := range arr { ... }\n"+
				"→ 示例：for index, value := range [1,2,3] { print(index) }",
			p.curToken.Line))
		return nil
	}
	p.nextToken() // 跳过逗号

	// 读取第二个变量名（value变量）
	if !p.curTokenIs(TokenIdent) {
		p.addError(fmt.Sprintf(
			"语法错误：期望value变量名，但得到 %v（第%d行）。\n"+
				"→ 格式：for i, v := range arr { ... }\n"+
				"→ 示例：for index, value := range [1,2,3] { print(value) }",
			p.curToken.Type, p.curToken.Line))
		return nil
	}
	valueName := p.curToken.Value
	p.nextToken() // 跳过value标识符

	// 期望 :=
	if !p.curTokenIs(TokenAssign) {
		p.addError(fmt.Sprintf(
			"语法错误：期望 ':=' 但得到 %v（第%d行）。\n"+
				"→ 格式：for i, v := range arr { ... }",
			p.curToken.Type, p.curToken.Line))
		return nil
	}
	p.nextToken() // 跳过 :=

	// 可选range关键字
	if p.curTokenIs(TokenRange) {
		p.nextToken() // 跳过range
	}

	stmt.Mode = ForRange
	stmt.RangeValueVar = valueName
	stmt.Init = &VarDeclStmt{
		Position: p.currentPosition(),
		Name:     keyName,
		Value:    p.parseExpr(LOWEST),
	}

	// 处理可选右括号
	if hasParens {
		p.advanceIfPeekIs(TokenRParen)
	}

	stmt.Body = p.expectAndParseBlock()
	if stmt.Body == nil {
		return nil
	}
	return stmt
}

// parseSimpleStmt 解析简单语句
// 用于for循环的init和post部分
// 可能是变量声明或表达式语句
func (p *Parser) parseSimpleStmt() Stmt {
	if p.curTokenIs(TokenIdent) && p.peekTokenIs(TokenAssign) {
		return p.parseVarDecl()
	}
	return p.parseExprStmt()
}

// parseReturnStmt 解析return返回语句
// 语法格式: return 或 return value
// 可选返回值，默认返回nil
func (p *Parser) parseReturnStmt() *ReturnStmt {
	stmt := &ReturnStmt{Position: p.currentPosition()}
	if p.peekTokenIs(TokenRBrace) || p.peekTokenIs(TokenEOF) {
		return stmt
	}
	p.nextToken()
	stmt.Value = p.parseExpr(LOWEST)
	return stmt
}

// parseBreakStmt 解析break中断语句
// 用于跳出最近的循环
func (p *Parser) parseBreakStmt() *BreakStmt {
	return &BreakStmt{Position: p.currentPosition()}
}

// parseContinueStmt 解析continue继续语句
// 用于跳过当前循环迭代，继续下一次迭代
func (p *Parser) parseContinueStmt() *ContinueStmt {
	return &ContinueStmt{Position: p.currentPosition()}
}

// parseThrowStmt 解析throw抛出异常语句
// 语法格式: throw value
// 抛出一个异常值
func (p *Parser) parseThrowStmt() *ThrowStmt {
	stmt := &ThrowStmt{Position: p.currentPosition()}
	p.nextToken()
	stmt.Value = p.parseExpr(LOWEST)
	return stmt
}

// parseDefDirective 解析#def或#fn指令
// 用于声明外部函数
// #def 语法格式: #def name(params) -> returnType
// #def 示例: #def externalFunc(a int, b int) -> int
// #fn 语法格式: #fn name(type1, type2)=>returnType
// #fn 示例: #fn externalFunc(int, int)=>int
func (p *Parser) parseDefDirective() *DefDirectiveStmt {
	stmt := &DefDirectiveStmt{Position: p.currentPosition()}
	comps := p.parseDirectiveComponents(p.curToken.Value)
	stmt.Name = comps.name
	if comps.params != "" {
		stmt.Params = p.parseDefParams(comps.params)
	}
	stmt.Return = p.parseDefReturnType(comps.ret)
	return stmt
}

// directiveComponents #def指令解析结果
// 包含函数名、参数字符串和返回类型字符串
type directiveComponents struct {
	name   string
	params string
	ret    string
}

// findParenIndices 查找括号的索引位置
func findParenIndices(s string) (openIdx, closeIdx int, valid bool) {
	openIdx = strings.Index(s, "(")
	if openIdx == -1 {
		return -1, -1, false
	}
	closeIdx = strings.Index(s, ")")
	if closeIdx == -1 || closeIdx <= openIdx {
		return openIdx, -1, false
	}
	return openIdx, closeIdx, true
}

// parseDirectiveComponents 解析#def或#fn指令字符串的各个组成部分
// #def 格式: name(param1 type1, param2 type2) -> returnType
// #fn 格式: name(type1, type2)=>returnType
func (p *Parser) parseDirectiveComponents(directive string) directiveComponents {
	// 检测是否是 #fn 格式
	isFnFormat := strings.HasPrefix(directive, "#fn ") || strings.HasPrefix(directive, "fn ")

	// 移除前缀
	directive, _ = stripAnyPrefix(directive, defPrefixes)

	// 查找括号位置
	openIdx, closeIdx, valid := findParenIndices(directive)
	if !valid {
		name := directive
		if openIdx != -1 {
			name = directive[:openIdx]
		}
		return directiveComponents{name: strings.TrimSpace(name)}
	}

	name := strings.TrimSpace(directive[:openIdx])
	params := directive[openIdx+1 : closeIdx]
	ret := directive[closeIdx+1:]

	// 如果是 #fn 格式，使用特殊的参数解析方式
	if isFnFormat {
		return directiveComponents{
			name:   name,
			params: params,
			ret:    p.parseFnReturnType(ret),
		}
	}

	return directiveComponents{
		name:   name,
		params: params,
		ret:    ret,
	}
}

// parseFnReturnType 解析 #fn 格式的返回类型
// 格式: =>returnType 或 => returnType
func (p *Parser) parseFnReturnType(ret string) string {
	ret = strings.TrimSpace(ret)
	// 移除 => 前缀
	if stripped, ok := stripAnyPrefix(ret, fnArrowPrefixes); ok {
		ret = stripped
	}
	return strings.TrimSpace(ret)
}

// ========== 类型解析 ==========
// 包含类型表达式的解析逻辑
// 支持：基本类型、数组类型、Map类型

// hasTypeAnnotation 检查是否有类型注解
// 类型注解可以是冒号(:)或大于号(>)
func (p *Parser) hasTypeAnnotation() bool {
	if p.curTokenIs(TokenTypeAnnot) {
		return true
	}
	return p.curTokenIs(TokenGt) && !p.peekTokenIs(TokenAssign)
}

// tryParseTypeAnnotation 尝试解析类型注解
// 如果存在类型注解则解析，否则返回nil
func (p *Parser) tryParseTypeAnnotation() TypeExpr {
	if !p.hasTypeAnnotation() {
		return nil
	}
	p.nextToken()
	typeExpr := p.parseTypeExpr()
	p.nextToken()
	return typeExpr
}

// parseTypeExpr 解析类型表达式
// 根据当前token类型分派到对应的类型解析器
// 支持: 标识符(基本类型)、[数组类型)、map{Map类型)
func (p *Parser) parseTypeExpr() TypeExpr {
	if parser, ok := typeParsers[p.curToken.Type]; ok {
		return parser(p)
	}
	p.addError(fmt.Sprintf(
		"类型错误：'%s' 不是有效的类型（第%d行第%d列）。\n"+
			"→ 问题：该位置需要类型声明，但编译器无法识别此类型。\n"+
			"→ 支持的类型：\n"+
			"  - 基本类型：int, float, string, bool\n"+
			"  - 数组类型：[]T （如 []int, []string）\n"+
			"  - Map类型：map{K:V} （如 map{string:int}）\n"+
			"→ 建议：检查类型名称是否正确，确保没有拼写错误",
		p.curToken.Value, p.curToken.Line, p.curToken.Column))
	return nil
}

// typeParser 类型解析函数类型
// 定义类型解析器的函数签名
type typeParser func(*Parser) TypeExpr

// typeParsers 类型解析器映射表
// 将TokenType映射到对应的类型解析函数
var typeParsers map[TokenType]typeParser

func init() {
	typeParsers = map[TokenType]typeParser{
		TokenIdent:    (*Parser).parseBaseType,
		TokenLBracket: (*Parser).parseArrayType,
	}
}

// parseBaseType 解析基本类型
// 支持的基本类型: int, float, string, bool, any, array, function
// 同时支持类型别名: i, f, s, str, b, arr, fn
func (p *Parser) parseBaseType() TypeExpr {
	name := p.curToken.Value
	// 解析类型别名
	resolvedName := resolveTypeAlias(name)
	if !isValidBaseTypes(name) {
		p.addError(fmt.Sprintf(
			"类型错误：'%s' 不是有效的基本类型（第%d行第%d列）。\n"+
				"→ 问题：该类型名不被识别为基本类型。\n"+
				"→ 支持的基本类型：\n"+
				"  - int（或别名 i）：整数类型（如 1, -2, 100）\n"+
				"  - float（或别名 f）：浮点数类型（如 3.14, -0.5）\n"+
				"  - string（或别名 s, str）：字符串类型（如 \"hello\"）\n"+
				"  - bool（或别名 b）：布尔类型（true 或 false）\n"+
				"  - any：任意类型\n"+
				"  - array（或别名 arr）：数组类型\n"+
				"  - function（或别名 fn）：函数类型\n"+
				"→ 建议：检查类型名称是否有拼写错误",
			name, p.curToken.Line, p.curToken.Column))
		return nil
	}
	// 使用解析后的类型名创建类型表达式
	return &BaseTypeExpr{
		Position: p.currentPosition(),
		Name:     resolvedName,
	}
}

// parseArrayType 解析数组类型
// 语法格式: []ElementType
// 示例: []int, []string
func (p *Parser) parseArrayType() TypeExpr {
	p.nextToken()
	elemType := p.parseTypeExpr()
	return &ArrayTypeExpr{
		Position: p.currentPosition(),
		ElemType: elemType,
	}
}

// validBaseTypes 有效基本类型集合
// 包含脚本支持的所有基本类型
var validBaseTypes = map[string]bool{
	"int":      true,
	"float":    true,
	"string":   true,
	"bool":     true,
	"any":      true,
	"array":    true,
	"function": true,
	"map":      true,
}

// typeAliases 类型别名映射
// 提供简短的类型名称以提高开发效率
var typeAliases = map[string]string{
	"i":   "int",
	"f":   "float",
	"s":   "string",
	"str": "string",
	"b":   "bool",
	"arr": "array",
	"fn":  "function",
	"any": "any",
}

// resolveTypeAlias 解析类型别名
// 如果给定的名称是别名，返回对应的完整类型名
// 否则返回原始名称
func resolveTypeAlias(name string) string {
	if resolved, ok := typeAliases[name]; ok {
		return resolved
	}
	return name
}

// isValidBaseTypes 检查是否是有效的基本类型名
// 会先解析类型别名，再检查是否在有效基本类型集合中
func isValidBaseTypes(name string) bool {
	// 先解析类型别名
	resolved := resolveTypeAlias(name)
	return validBaseTypes[resolved]
}

// stripAnyPrefix 尝试移除任意一个前缀
// 如果匹配则返回去除前缀的字符串和true
func stripAnyPrefix(s string, prefixes []string) (string, bool) {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return s[len(prefix):], true
		}
	}
	return s, false
}

// defPrefixes def指令前缀列表
// 支持 "#def ", "def ", "#fn ", "fn " 四种格式
var defPrefixes = []string{"#def ", "def ", "#fn ", "fn "}

// fnArrowPrefixes fn格式箭头前缀列表
// 支持 "=>" 和 "=> " 两种格式
var fnArrowPrefixes = []string{"=>", "=> "}

// arrowPrefixes 箭头前缀列表
// 支持 "-> " 和 "->" 两种格式
var arrowPrefixes = []string{"-> ", "->"}

// parseDefReturnType 解析返回类型部分
// 处理可选的箭头前缀并规范化类型名
func (p *Parser) parseDefReturnType(returnPart string) TypeExpr {
	normalized := p.normalizeReturnType(returnPart)
	if normalized == "" {
		return nil
	}
	return &BaseTypeExpr{Name: normalized}
}

// normalizeReturnType 规范化返回类型字符串
// 去除空白和可选的箭头前缀
func (p *Parser) normalizeReturnType(s string) string {
	s = strings.TrimSpace(s)
	if stripped, ok := stripAnyPrefix(s, arrowPrefixes); ok {
		s = stripped
	}
	return strings.TrimSpace(s)
}

// parseDefParams 解析 #def 或 #fn 指令中的参数列表
// #def 格式: name1 type1, name2 type2, ... (参数名 + 类型)
// #fn 格式: type1, type2, ... (只有类型，没有参数名)
func (p *Parser) parseDefParams(paramsStr string) []Param {
	paramsStr = strings.TrimSpace(paramsStr)
	if paramsStr == "" {
		return nil
	}
	parts := strings.Split(paramsStr, ",")
	params := make([]Param, 0, len(parts))

	// 判断格式：检查第一个参数是否只有一个单词
	// 如果只有一个单词，则认为是 #fn 格式（只有类型）
	// 如果有多个单词，则认为是 #def 格式（参数名 + 类型）
	isFnFormat := false
	if len(parts) > 0 {
		firstPart := strings.TrimSpace(parts[0])
		fields := strings.Fields(firstPart)
		// 如果只有一个字段，且是有效类型（或类型别名），则认为是 #fn 格式
		if len(fields) == 1 && isValidTypeOrAlias(fields[0]) {
			isFnFormat = true
		}
	}

	for i, part := range parts {
		var param Param
		if isFnFormat {
			param = p.parseFnSingleParam(part, i)
		} else {
			param = p.parseDefSingleParam(part)
		}
		if param.Name != "" || param.TypeAnnot != nil {
			params = append(params, param)
		}
	}
	return params
}

// isValidTypeOrAlias 检查是否是有效的类型名或类型别名
func isValidTypeOrAlias(name string) bool {
	// 先检查是否是类型别名
	if resolveTypeAlias(name) != name {
		return true
	}
	// 再检查是否是有效的基本类型
	return validBaseTypes[name]
}

// parseFnSingleParam 解析 #fn 指令中的单个参数
// 格式: 只有类型，没有参数名
// 自动生成参数名: _0, _1, _2, ...
func (p *Parser) parseFnSingleParam(part string, index int) Param {
	part = strings.TrimSpace(part)
	if part == "" {
		return Param{}
	}
	// 解析类型别名
	resolvedType := resolveTypeAlias(part)
	return Param{
		Name:      fmt.Sprintf("_%d", index),
		TypeAnnot: &BaseTypeExpr{Name: resolvedType},
	}
}

// parseDefSingleParam 解析 #def 指令中的单个参数
// 格式: name 或 name type
func (p *Parser) parseDefSingleParam(part string) Param {
	part = strings.TrimSpace(part)
	if part == "" {
		return Param{}
	}
	fields := strings.Fields(part)
	param := Param{Name: fields[0]}
	if len(fields) >= 2 {
		param.TypeAnnot = &BaseTypeExpr{Name: fields[1]}
	}
	return param
}
