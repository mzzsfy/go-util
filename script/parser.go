package script

import (
	"fmt"
	"sync"
)

// ========== 解析器核心结构 ==========

// Parser 将Token序列转换为AST
type Parser struct {
	lexer *Lexer
	// curToken 当前Token
	curToken Token
	// peekToken 下一个Token（预读）
	peekToken Token
	// errors 编译错误列表
	errors []*CompileError
}

// parserPool 解析器实例池，减少内存分配
var parserPool = sync.Pool{
	New: func() interface{} {
		return &Parser{}
	},
}

// NewParser 创建语法分析器
func NewParser() *Parser {
	return parserPool.Get().(*Parser)
}

// returnParserToPool 将Parser归还到池中
func returnParserToPool(p *Parser) {
	// 清空Parser状态以便复用
	p.lexer = nil
	p.curToken = Token{}
	p.peekToken = Token{}
	p.errors = nil
	parserPool.Put(p)
}

// initLexer 初始化词法分析器并预读token
func (p *Parser) initLexer(source string) {
	p.lexer = NewLexer(source)
	p.errors = nil
	p.nextToken()
	p.nextToken()
}

// Compile 编译源码为可执行的字节码
// 参数source为脚本源代码字符串
// 返回编译后的脚本和可能的错误
func (p *Parser) Compile(source string) (*CompiledScript, error) {
	p.initLexer(source)

	// 解析整个程序
	program := p.parseProgram()

	if len(p.errors) > 0 {
		return nil, p.errors[0]
	}

	// 编译AST到字节码
	compiler := NewCompiler()
	return compiler.Compile(program)
}

// Validate 验证源码语法是否正确
// 只进行语法检查，不生成代码
func (p *Parser) Validate(source string) error {
	p.initLexer(source)

	p.parseProgram()

	if len(p.errors) > 0 {
		return p.errors[0]
	}

	return nil
}

// ========== Token操作方法 ==========

// nextToken 前进到下一个token
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

// curTokenIs 检查当前token类型
func (p *Parser) curTokenIs(t TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs 检查下一个token类型
func (p *Parser) peekTokenIs(t TokenType) bool {
	return p.peekToken.Type == t
}

// advanceIfPeekIs 如果下一个token匹配则前进
// 返回是否匹配并前进
func (p *Parser) advanceIfPeekIs(t TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	return false
}

// expectPeek 期望下一个token类型
func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// peekError 报告token不匹配错误
func (p *Parser) peekError(t TokenType) {
	context := ""
	suggestion := ""

	if p.curToken.Type == TokenIdent {
		context = fmt.Sprintf("（在标识符 '%s' 之后）", p.curToken.Value)
	} else if p.curToken.Type != TokenEOF && p.curToken.Value != "" {
		context = fmt.Sprintf("（在 '%s' 之后）", p.curToken.Value)
	}

	switch t {
	case TokenRParen:
		suggestion = "\n→ 建议：检查括号是否匹配，确保每个 ( 都有对应的 )"
	case TokenRBrace:
		suggestion = "\n→ 建议：检查花括号是否匹配，确保每个 { 都有对应的 }"
	case TokenRBracket:
		suggestion = "\n→ 建议：检查方括号是否匹配，确保每个 [ 都有对应的 ]"
	case TokenIdent:
		suggestion = "\n→ 建议：期望一个标识符（变量名或函数名），如 myVar 或 myFunc"
	case TokenLBrace:
		suggestion = "\n→ 建议：检查是否缺少代码块开始标记 {"
	case TokenAssign:
		suggestion = "\n→ 建议：检查是否缺少赋值运算符 :="
	}

	p.addError(fmt.Sprintf("语法错误：期望 %v，但得到 %v%s%s", t, p.peekToken.Type, context, suggestion))
}

// addError 添加编译错误
// 在当前位置记录错误信息
func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, &CompileError{
		Line:    p.curToken.Line,
		Column:  p.curToken.Column,
		Message: msg,
	})
}

// currentPosition 获取当前位置信息
// 这是一个辅助方法，减少重复的Position创建代码
func (p *Parser) currentPosition() Position {
	return Position{Line: p.curToken.Line, Column: p.curToken.Column}
}

// ========== 程序解析 ==========

// parseProgram 解析整个程序
// 程序由一系列语句组成，直到遇到EOF
func (p *Parser) parseProgram() *Program {
	var statements []Stmt

	for !p.curTokenIs(TokenEOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			statements = append(statements, stmt)
		}
		p.nextToken()
	}

	return &Program{Statements: statements}
}

// parseStatement 解析语句
func (p *Parser) parseStatement() Stmt {
	return p.parseStatementWithDispatch()
}

// parseProgram 解析整个程序
func (p *Parser) parseStatementWithDispatch() Stmt {
	// 处理标识符特殊情况（需要查看 peek token 判断是变量声明还是表达式）
	if p.curTokenIs(TokenIdent) && p.peekTokenIs(TokenAssign) {
		return p.parseVarDecl()
	}

	// 处理类型安全声明语法：ident :=>type value
	if p.curTokenIs(TokenIdent) && p.peekTokenIs(TokenTypedAssign) {
		return p.parseTypedVarDecl()
	}

	switch p.curToken.Type {
	case TokenFn:
		return p.parseFuncDecl()
	case TokenIf:
		return p.parseIfStmt()
	case TokenFor:
		return p.parseForStmt()
	case TokenReturn:
		return p.parseReturnStmt()
	case TokenBreak:
		return p.parseBreakStmt()
	case TokenContinue:
		return p.parseContinueStmt()
	case TokenThrow:
		return p.parseThrowStmt()
	case TokenDef:
		return p.parseDefDirective()
	default:
		return p.parseExprStmt()
	}
}

// ========== 表达式优先级 ==========

// parseBinaryExpr 解析二元运算表达式
// 统一处理所有二元运算符（算术、比较、逻辑、位运算）
func (p *Parser) parseBinaryExpr(left Expr) Expr {
	expr := &BinaryExpr{
		Position: p.currentPosition(),
		Left:     left,
		Operator: p.curToken.Value,
	}
	prec := p.curPrecedence()
	p.nextToken()
	expr.Right = p.parseExpr(prec)
	return expr
}
