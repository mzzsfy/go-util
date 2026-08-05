package script

// ========== 表达式解析核心 ==========
// 使用Pratt解析器处理运算符优先级

import (
	"fmt"
	"strconv"
)

// ========== 运算符优先级 ==========
// 数值越大优先级越高
const (
	_ int = iota
	// LOWEST 最低优先级，用于初始表达式解析
	LOWEST
	// ASSIGN 赋值运算符优先级: = += -= *= /=
	ASSIGN
	// OR 逻辑或优先级: ||
	OR
	// AND 逻辑与优先级: &&
	AND
	// EQUALS 相等比较优先级: == !=
	EQUALS
	// LESSGREATER 大小比较优先级: < > <= >=
	LESSGREATER
	// BITOP 位运算优先级: | ^ & << >>
	BITOP
	// SUM 加减运算优先级: + -
	SUM
	// PRODUCT 乘除运算优先级: * / %
	PRODUCT
	// PREFIX 前缀运算优先级: - ! ^
	PREFIX
	// CALL 函数调用优先级: ( [
	CALL
	// INDEX 索引访问优先级: [
	INDEX
)

// tokenPrecedence 运算符优先级查找
func tokenPrecedence(t TokenType) int {
	switch t {
	case TokenAssign, TokenPlusAssign, TokenMinusAssign, TokenStarAssign, TokenSlashAssign:
		return ASSIGN
	case TokenOr:
		return OR
	case TokenAnd:
		return AND
	case TokenEq, TokenNeq:
		return EQUALS
	case TokenLt, TokenGt, TokenLe, TokenGe:
		return LESSGREATER
	case TokenBitOr, TokenBitXor, TokenBitAnd, TokenLShift, TokenRShift:
		return BITOP
	case TokenPlus, TokenMinus:
		return SUM
	case TokenStar, TokenSlash, TokenPercent:
		return PRODUCT
	case TokenLParen:
		return CALL
	case TokenLBracket:
		return INDEX
	}
	return LOWEST
}

// parseExpr 解析表达式
func (p *Parser) parseExpr(precedence int) Expr {
	left := p.parsePrefix()
	if left == nil {
		return nil
	}
	return p.parseInfixExpr(left, precedence)
}

// parsePrefix 根据当前token类型分发到对应的前缀解析器
func (p *Parser) parsePrefix() Expr {
	switch p.curToken.Type {
	case TokenIdent:
		return p.parseIdentExpr()
	case TokenInt:
		return p.parseIntLiteral()
	case TokenFloat:
		return p.parseFloatLiteral()
	case TokenString:
		return p.parseStringLiteral()
	case TokenBool:
		return p.parseBoolLiteral()
	case TokenNil:
		return p.parseNilLiteral()
	case TokenLBracket:
		return p.parseArrayLiteral()
	case TokenLBrace:
		return p.parseBraceLiteral()
	case TokenIf:
		return p.parseIfExpr()
	case TokenLParen:
		return p.parseGroupedExpr()
	case TokenMinus, TokenNot, TokenBitXor:
		return p.parsePrefixExpr()
	default:
		p.addError(fmt.Sprintf(
			"语法错误：无法识别 '%s'（第%d行第%d列）。\n"+
				"→ 问题：编译器在此位置期望一个表达式的开始。\n"+
				"→ 期望：标识符（如 x、myVar）、字面量（如 123、3.14、\"text\"、true）、"+
				"运算符（如 -、!）或括号 (。\n"+
				"→ 建议：检查是否有拼写错误、缺少操作数或使用了不支持的操作符",
			p.curToken.Value, p.curToken.Line, p.curToken.Column))
		return nil
	}
}

// parseInfixExpr 解析中缀表达式
func (p *Parser) parseInfixExpr(left Expr, precedence int) Expr {
	for !p.peekTokenIs(TokenEOF) && precedence < p.peekPrecedence() {
		if !p.hasInfixHandler(p.peekToken.Type) {
			return left
		}
		p.nextToken()
		left = p.applyInfix(left)
	}
	return left
}

// hasInfixHandler 判断指定token类型是否有中缀处理器
func (p *Parser) hasInfixHandler(t TokenType) bool {
	switch t {
	case TokenLParen, TokenLBracket,
		TokenAssign, TokenPlusAssign, TokenMinusAssign, TokenStarAssign, TokenSlashAssign,
		TokenPlus, TokenMinus, TokenStar, TokenSlash, TokenPercent,
		TokenEq, TokenNeq, TokenLt, TokenGt, TokenLe, TokenGe,
		TokenAnd, TokenOr,
		TokenBitAnd, TokenBitOr, TokenBitXor, TokenLShift, TokenRShift:
		return true
	}
	return false
}

// applyInfix 根据当前token类型分发到对应的中缀处理器
func (p *Parser) applyInfix(left Expr) Expr {
	switch p.curToken.Type {
	case TokenLParen:
		return p.parseCallExpr(left)
	case TokenLBracket:
		return p.parseIndexExpr(left)
	case TokenAssign, TokenPlusAssign, TokenMinusAssign, TokenStarAssign, TokenSlashAssign:
		return p.parseAssignExpr(left)
	default:
		return p.parseBinaryExpr(left)
	}
}

// getPrecedence 获取指定token类型的优先级
func (p *Parser) getPrecedence(t TokenType) int {
	return tokenPrecedence(t)
}

// peekPrecedence 获取下一个token的优先级
func (p *Parser) peekPrecedence() int {
	return p.getPrecedence(p.peekToken.Type)
}

// curPrecedence 获取当前token的优先级
func (p *Parser) curPrecedence() int {
	return p.getPrecedence(p.curToken.Type)
}

// ========== 字面量解析 ==========

// parseIntLiteral 解析整数字面量
func (p *Parser) parseIntLiteral() *LiteralExpr {
	value, err := strconv.ParseInt(p.curToken.Value, 0, 64)
	if err != nil {
		p.addError(fmt.Sprintf(
			"整数格式错误：'%s' 不是有效的整数字面量（第%d行第%d列）。\n"+
				"→ 问题：该值无法解析为整数。\n"+
				"→ 支持的格式：十进制（如 123, -456）、十六进制（如 0xFF, 0xAB）。\n"+
				"→ 建议：检查数字格式是否正确，避免使用前导零（如 0123），使用下划线分隔提高可读性（如 1_000_000）",
			p.curToken.Value, p.curToken.Line, p.curToken.Column))
		return nil
	}
	return p.newLiteralExpr(LiteralInt, int(value))
}

// parseFloatLiteral 解析浮点数字面量
func (p *Parser) parseFloatLiteral() *LiteralExpr {
	value, err := strconv.ParseFloat(p.curToken.Value, 64)
	if err != nil {
		p.addError(fmt.Sprintf(
			"浮点数格式错误：'%s' 不是有效的浮点数字面量（第%d行第%d列）。\n"+
				"→ 问题：该值无法解析为浮点数。\n"+
				"→ 支持的格式：标准浮点数（如 3.14, -2.5）、科学计数法（如 1.5e10）。\n"+
				"→ 建议：检查小数点位置和格式是否正确，避免使用多个小数点（如 1.2.3）",
			p.curToken.Value, p.curToken.Line, p.curToken.Column))
		return nil
	}
	return p.newLiteralExpr(LiteralFloat, value)
}

// parseStringLiteral 解析字符串字面量
func (p *Parser) parseStringLiteral() *LiteralExpr {
	return p.newLiteralExpr(LiteralString, p.curToken.Value)
}

// parseBoolLiteral 解析布尔字面量
func (p *Parser) parseBoolLiteral() *LiteralExpr {
	return p.newLiteralExpr(LiteralBool, p.curToken.Value == "true")
}

// parseNilLiteral 解析nil字面量
func (p *Parser) parseNilLiteral() *LiteralExpr {
	return p.newLiteralExpr(LiteralNil, nil)
}

// newLiteralExpr 创建字面量表达式
func (p *Parser) newLiteralExpr(litType LiteralType, value any) *LiteralExpr {
	return &LiteralExpr{
		Position: p.currentPosition(),
		Type:     litType,
		Value:    value,
	}
}

// parseIdentExpr 解析标识符表达式
func (p *Parser) parseIdentExpr() *IdentExpr {
	return &IdentExpr{
		Position: p.currentPosition(),
		Name:     p.curToken.Value,
	}
}

// ========== 访问操作解析方法 ==========

// parseIndexExpr 解析索引表达式
// 支持两种模式:
//   - 普通索引: arr[i]
//   - 切片: arr[start:end] 或 arr[:end] 或 arr[start:]
func (p *Parser) parseIndexExpr(left Expr) Expr {
	p.nextToken()

	// 检查是否是省略起始的切片 [:end]
	if p.curTokenIs(TokenColon) {
		return p.parseSlice(left, nil)
	}

	index := p.parseExpr(LOWEST)

	// 检查是否是切片 [start:end] 或 [start:]
	if p.peekTokenIs(TokenColon) {
		return p.parseSlice(left, index)
	}

	// 普通索引访问
	if !p.expectPeek(TokenRBracket) {
		return nil
	}

	return &IndexExpr{
		Position: p.currentPosition(),
		Object:   left,
		Index:    index,
	}
}

// parseSlice 解析切片表达式
// 支持三种模式: [:end], [start:end], [start:]
// start为nil时表示省略起始索引
func (p *Parser) parseSlice(left Expr, start Expr) *SliceExpr {
	slice := &SliceExpr{
		Position: p.currentPosition(),
		Object:   left,
		Start:    start,
	}

	// 如果有起始索引，需要移动到冒号
	if start != nil {
		p.nextToken()
	}

	// 解析结束索引（如果存在）
	if !p.peekTokenIs(TokenRBracket) {
		p.nextToken()
		slice.End = p.parseExpr(LOWEST)
	}

	if !p.expectPeek(TokenRBracket) {
		return nil
	}

	return slice
}

// parseCallExpr 解析函数调用表达式
func (p *Parser) parseCallExpr(fn Expr) *CallExpr {
	call := &CallExpr{
		Position: p.currentPosition(),
		Func:     fn,
	}

	if p.peekTokenIs(TokenRParen) {
		p.nextToken()
		return call
	}

	call.Args = p.parseCommaSeparatedExprs(TokenRParen)
	if call.Args == nil {
		return nil
	}

	return call
}

// parsePrefixExpr 解析前缀表达式
func (p *Parser) parsePrefixExpr() *UnaryExpr {
	expr := &UnaryExpr{
		Position: p.currentPosition(),
		Operator: p.curToken.Value,
	}

	p.nextToken()
	expr.Operand = p.parseExpr(PREFIX)

	return expr
}

// parseAssignExpr 解析赋值表达式
func (p *Parser) parseAssignExpr(left Expr) *BinaryExpr {
	expr := &BinaryExpr{
		Position: p.currentPosition(),
		Left:     left,
		Operator: p.curToken.Value,
	}
	p.nextToken()
	expr.Right = p.parseExpr(LOWEST)
	return expr
}

// ========== 复合类型解析 ==========

// isEmptyCollection 检查是否为空集合并前进token
func (p *Parser) isEmptyCollection(closeToken TokenType) bool {
	if p.peekTokenIs(closeToken) {
		p.nextToken()
		return true
	}
	return false
}

// parseArrayLiteral 解析数组字面量
func (p *Parser) parseArrayLiteral() *ArrayExpr {
	arr := &ArrayExpr{Position: p.currentPosition()}
	if p.isEmptyCollection(TokenRBracket) {
		return arr
	}
	arr.Elements = p.parseCommaSeparatedExprs(TokenRBracket)
	if arr.Elements == nil {
		return nil
	}
	return arr
}

// parseCommaSeparatedExprs 解析逗号分隔的表达式列表
// 支持尾随逗号: [1, 2, 3,]
func (p *Parser) parseCommaSeparatedExprs(endToken TokenType) []Expr {
	var exprs []Expr
	p.nextToken()
	exprs = append(exprs, p.parseExpr(LOWEST))
	for p.peekTokenIs(TokenComma) {
		p.nextToken()
		if p.peekTokenIs(endToken) {
			break
		}
		p.nextToken()
		exprs = append(exprs, p.parseExpr(LOWEST))
	}
	if !p.expectPeek(endToken) {
		return nil
	}
	return exprs
}

// parseMapPairs 解析Map的键值对列表
// 支持尾随逗号: {"a": 1, "b": 2,}
func (p *Parser) parseMapPairs(m *MapExpr) bool {
	for {
		pair, ok := p.parseMapPair()
		if !ok {
			return false
		}
		m.Pairs = append(m.Pairs, pair)
		if !p.peekTokenIs(TokenComma) {
			return true
		}
		p.nextToken()
		if p.peekTokenIs(TokenRBrace) {
			return true
		}
		p.nextToken()
	}
}

// parseMapPair 解析Map的键值对
func (p *Parser) parseMapPair() (MapPair, bool) {
	key := p.parseExpr(LOWEST)
	if !p.expectPeek(TokenColon) {
		return MapPair{}, false
	}
	p.nextToken()
	value := p.parseExpr(LOWEST)
	return MapPair{Key: key, Value: value}, true
}

// parseIfExpr 解析if表达式
func (p *Parser) parseIfExpr() *IfExpr {
	expr := &IfExpr{Position: p.currentPosition()}
	p.nextToken()
	expr.Condition = p.parseExpr(LOWEST)
	if !p.expectPeek(TokenLBrace) {
		return nil
	}
	p.nextToken()
	thenExpr, ok := p.parseSingleExprBranch("then")
	if !ok {
		return nil
	}
	expr.Then = thenExpr
	// else 分支是可选的，没有 else 时 Else 为 nil
	// 编译时如果没有 else 且条件为假，会返回 nil
	if p.peekTokenIs(TokenElse) {
		p.nextToken()
		p.nextToken()
		expr.Else = p.parseIfExprElse()
	}
	return expr
}

// parseSingleExprBranch 解析if表达式的单个表达式分支
func (p *Parser) parseSingleExprBranch(branchName string) (Expr, bool) {
	block := p.parseBlockStmt()
	if len(block.Statements) != 1 {
		p.addError(fmt.Sprintf(
			"if表达式语法错误：%s分支只能包含单个表达式（第%d行）。\n"+
				"→ 问题：if表达式要求每个分支必须是单个表达式，不是语句块。\n"+
				"→ 示例：if cond { 1 } else { 2 }  // 正确：返回数值\n"+
				"→ 错误：if cond { x := 1; x }    // 错误：包含多个语句\n"+
				"→ 建议：如果需要多个语句，请使用if语句而非if表达式",
			branchName, p.curToken.Line))
		return nil, false
	}
	exprStmt, ok := block.Statements[0].(*ExprStmt)
	if !ok {
		p.addError(fmt.Sprintf(
			"if表达式语法错误：%s分支必须是表达式，不能是语句（第%d行）。\n"+
				"→ 问题：if表达式的分支必须产生一个值，但语句不产生值。\n"+
				"→ 示例：if cond { x+1 } else { x*2 }  // 正确：表达式\n"+
				"→ 错误：if cond { return x } else { y := 1 }  // 错误：语句\n"+
				"→ 建议：将语句改为表达式，或使用if语句",
			branchName, p.curToken.Line))
		return nil, false
	}
	return exprStmt.Expr, true
}

// parseIfExprElse 解析if表达式的else分支
func (p *Parser) parseIfExprElse() Expr {
	if p.curTokenIs(TokenIf) {
		return p.parseIfExpr()
	}
	if !p.curTokenIs(TokenLBrace) {
		p.peekError(TokenLBrace)
		return nil
	}
	p.nextToken()
	elseExpr, ok := p.parseSingleExprBranch("else")
	if !ok {
		return nil
	}
	return elseExpr
}

// parseGroupedExpr 解析分组表达式
func (p *Parser) parseGroupedExpr() Expr {
	p.nextToken()
	expr := p.parseExpr(LOWEST)
	if !p.expectPeek(TokenRParen) {
		return nil
	}
	return expr
}

// parseBraceLiteral 解析 {} 开头的表达式
// 在表达式上下文中，{} 被解析为 map 字面量
// 支持：
//   - {}        -> 空 map
//   - {"k": v}  -> 带键值对的 map
func (p *Parser) parseBraceLiteral() Expr {
	m := &MapExpr{Position: p.currentPosition()}
	// 检查空 map: {}
	if p.isEmptyCollection(TokenRBrace) {
		return m
	}
	p.nextToken()
	if !p.parseMapPairs(m) {
		return nil
	}
	if !p.expectPeek(TokenRBrace) {
		return nil
	}
	return m
}
