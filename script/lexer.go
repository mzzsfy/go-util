package script

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ========== Token类型定义 ==========

// TokenType 表示词法记号的类型
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError
	TokenComment
	TokenInt
	TokenFloat
	TokenString
	TokenBool
	TokenNil
	TokenIdent
	TokenFn
	TokenIf
	TokenElse
	TokenFor
	TokenRange
	TokenReturn
	TokenBreak
	TokenContinue
	TokenThrow
	TokenDef
	TokenPlus
	TokenMinus
	TokenStar
	TokenSlash
	TokenPercent
	TokenAnd
	TokenOr
	TokenNot
	TokenBitAnd
	TokenBitOr
	TokenBitXor
	TokenLShift
	TokenRShift
	TokenEq
	TokenNeq
	TokenLt
	TokenGt
	TokenLe
	TokenGe
	TokenAssign
	TokenPlusAssign
	TokenMinusAssign
	TokenStarAssign
	TokenSlashAssign
	TokenArrow
	TokenTypeAnnot
	TokenLParen
	TokenRParen
	TokenLBrace
	TokenRBrace
	TokenLBracket
	TokenRBracket
	TokenComma
	TokenColon
	TokenSemicolon   // 新增：分号
	TokenTypedAssign // :=> 类型安全声明
)

// tokenNames Token类型到可读名称的映射
var tokenNames = map[TokenType]string{
	TokenEOF:         "EOF（文件结束）",
	TokenError:       "错误",
	TokenComment:     "注释",
	TokenInt:         "整数",
	TokenFloat:       "浮点数",
	TokenString:      "字符串",
	TokenBool:        "布尔值",
	TokenNil:         "nil",
	TokenIdent:       "标识符",
	TokenFn:          "fn",
	TokenIf:          "if",
	TokenElse:        "else",
	TokenFor:         "for",
	TokenRange:       "range",
	TokenReturn:      "return",
	TokenBreak:       "break",
	TokenContinue:    "continue",
	TokenThrow:       "throw",
	TokenDef:         "#fn",
	TokenPlus:        "+",
	TokenMinus:       "-",
	TokenStar:        "*",
	TokenSlash:       "/",
	TokenPercent:     "%",
	TokenAnd:         "&&",
	TokenOr:          "||",
	TokenNot:         "!",
	TokenBitAnd:      "&",
	TokenBitOr:       "|",
	TokenBitXor:      "^",
	TokenLShift:      "<<",
	TokenRShift:      ">>",
	TokenEq:          "==",
	TokenNeq:         "!=",
	TokenLt:          "<",
	TokenGt:          ">",
	TokenLe:          "<=",
	TokenGe:          ">=",
	TokenAssign:      ":=",
	TokenPlusAssign:  "+=",
	TokenMinusAssign: "-=",
	TokenStarAssign:  "*=",
	TokenSlashAssign: "/=",
	TokenArrow:       "=>",
	TokenTypeAnnot:   ":",
	TokenLParen:      "(",
	TokenRParen:      ")",
	TokenLBrace:      "{",
	TokenRBrace:      "}",
	TokenLBracket:    "[",
	TokenRBracket:    "]",
	TokenComma:       ",",
	TokenColon:       ":",
	TokenSemicolon:   ";",
	TokenTypedAssign: ":=>",
}

// String 返回TokenType的可读字符串表示
// 用于错误消息和调试输出
func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("Token(%d)", int(t))
}

// Token 表示一个词法记号
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

// Position 表示源码中的位置
type Position struct {
	Line   int
	Column int
}

// ========== Lexer词法分析器 ==========

// Lexer 将源代码转换为Token序列
type Lexer struct {
	// input 源代码输入
	input string
	// pos 当前读取位置（字节偏移）
	pos int
	// line 当前行号（从1开始）
	line int
	// column 当前列号
	column int
	// ch 当前字符
	ch rune
	// peekCh 下一个字符（预读）
	peekCh rune
}

// ========== 操作符辅助函数 ==========

// handleColon 处理冒号

// makeAssignOp 创建复合赋值操作符Token的辅助函数
func makeAssignOp(l *Lexer, baseType TokenType, baseVal string, assignType TokenType, assignVal string) Token {
	if l.peekCh == '=' {
		l.readChar()
		return Token{Type: assignType, Value: assignVal, Line: l.line, Column: l.column - 1}
	}
	return Token{Type: baseType, Value: baseVal, Line: l.line, Column: l.column}
}

// makeDoubleOp 创建双相同字符操作符Token的辅助函数
func makeDoubleOp(l *Lexer, ch rune, singleType TokenType, singleVal string, doubleType TokenType, doubleVal string) Token {
	if l.peekCh == ch {
		l.readChar()
		return Token{Type: doubleType, Value: doubleVal, Line: l.line, Column: l.column - 1}
	}
	return Token{Type: singleType, Value: singleVal, Line: l.line, Column: l.column}
}

// handleColon 处理冒号
// 支持: :（冒号）、:=（赋值）、:=>（类型安全声明）
func handleColon(l *Lexer) Token {
	// 检查 :=
	if l.peekCh == '=' {
		l.readChar()
		// 检查 :=>
		if l.peekCh == '>' {
			l.readChar()
			return Token{Type: TokenTypedAssign, Value: ":=>", Line: l.line, Column: l.column - 2}
		}
		return Token{Type: TokenAssign, Value: ":=", Line: l.line, Column: l.column - 1}
	}
	return Token{Type: TokenColon, Value: ":", Line: l.line, Column: l.column}
}

// NewLexer 创建词法分析器实例
// 参数input为要分析的源代码字符串
func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

// newToken 创建Token
// col参数为0时使用当前列号，否则使用指定列号
func (l *Lexer) newToken(tokType TokenType, value string, col int) Token {
	if col == 0 {
		col = l.column
	}
	return Token{Type: tokType, Value: value, Line: l.line, Column: col}
}

// readChar 读取下一个字符
// ASCII快速路径避免UTF-8解码开销, 多字节字符按实际宽度前进
func (l *Lexer) readChar() {
	if l.pos >= len(l.input) {
		l.ch = 0
		l.peekCh = 0
		l.pos++
		l.column++
		return
	}
	c := l.input[l.pos]
	if c < utf8.RuneSelf {
		l.ch = rune(c)
		l.pos++
	} else {
		var size int
		l.ch, size = utf8.DecodeRuneInString(l.input[l.pos:])
		l.pos += size
	}
	// peekCh 预读下一个字符
	if l.pos < len(l.input) {
		c2 := l.input[l.pos]
		if c2 < utf8.RuneSelf {
			l.peekCh = rune(c2)
		} else {
			l.peekCh, _ = utf8.DecodeRuneInString(l.input[l.pos:])
		}
	} else {
		l.peekCh = 0
	}
	l.column++
}

// NextToken 获取下一个Token
// 自动跳过注释Token, 确保调用者无需处理TokenComment
func (l *Lexer) NextToken() Token {
	for {
		tok := l.readNextToken()
		if tok.Type != TokenComment {
			return tok
		}
	}
}

// readNextToken 读取下一个原始Token(含注释)
func (l *Lexer) readNextToken() Token {
	l.skipWhitespace()

	ch := l.ch
	if ch == 0 {
		return l.newToken(TokenEOF, "", 0)
	}

	col := l.column

	switch ch {
	// 单字符token
	case '(':
		l.readChar()
		return Token{Type: TokenLParen, Value: "(", Line: l.line, Column: col}
	case ')':
		l.readChar()
		return Token{Type: TokenRParen, Value: ")", Line: l.line, Column: col}
	case '{':
		l.readChar()
		return Token{Type: TokenLBrace, Value: "{", Line: l.line, Column: col}
	case '}':
		l.readChar()
		return Token{Type: TokenRBrace, Value: "}", Line: l.line, Column: col}
	case '[':
		l.readChar()
		return Token{Type: TokenLBracket, Value: "[", Line: l.line, Column: col}
	case ']':
		l.readChar()
		return Token{Type: TokenRBracket, Value: "]", Line: l.line, Column: col}
	case ',':
		l.readChar()
		return Token{Type: TokenComma, Value: ",", Line: l.line, Column: col}
	case '%':
		l.readChar()
		return Token{Type: TokenPercent, Value: "%", Line: l.line, Column: col}
	case '^':
		l.readChar()
		return Token{Type: TokenBitXor, Value: "^", Line: l.line, Column: col}
	case ';':
		l.readChar()
		return Token{Type: TokenSemicolon, Value: ";", Line: l.line, Column: col}

	// 多字符操作符
	case '+':
		tok := makeAssignOp(l, TokenPlus, "+", TokenPlusAssign, "+=")
		l.readChar()
		return tok
	case '-':
		tok := makeAssignOp(l, TokenMinus, "-", TokenMinusAssign, "-=")
		l.readChar()
		return tok
	case '*':
		tok := makeAssignOp(l, TokenStar, "*", TokenStarAssign, "*=")
		l.readChar()
		return tok
	case '/':
		if l.peekCh == '/' {
			tok := l.readComment()
			l.readChar()
			return tok
		}
		if l.peekCh == '*' {
			tok := l.readMultilineComment()
			l.readChar()
			return tok
		}
		tok := makeAssignOp(l, TokenSlash, "/", TokenSlashAssign, "/=")
		l.readChar()
		return tok
	case '&':
		tok := makeDoubleOp(l, '&', TokenBitAnd, "&", TokenAnd, "&&")
		l.readChar()
		return tok
	case '|':
		tok := makeDoubleOp(l, '|', TokenBitOr, "|", TokenOr, "||")
		l.readChar()
		return tok
	case '!':
		tok := makeAssignOp(l, TokenNot, "!", TokenNeq, "!=")
		l.readChar()
		return tok
	case '<':
		if l.peekCh == '<' {
			l.readChar()
			tok := Token{Type: TokenLShift, Value: "<<", Line: l.line, Column: l.column - 1}
			l.readChar()
			return tok
		}
		tok := makeAssignOp(l, TokenLt, "<", TokenLe, "<=")
		l.readChar()
		return tok
	case '>':
		if l.peekCh == '>' {
			l.readChar()
			tok := Token{Type: TokenRShift, Value: ">>", Line: l.line, Column: l.column - 1}
			l.readChar()
			return tok
		}
		tok := makeAssignOp(l, TokenGt, ">", TokenGe, ">=")
		l.readChar()
		return tok
	case '=':
		if l.peekCh == '>' {
			l.readChar()
			tok := l.newToken(TokenArrow, "=>", 0)
			l.readChar()
			return tok
		}
		tok := makeAssignOp(l, TokenAssign, "=", TokenEq, "==")
		l.readChar()
		return tok
	case ':':
		tok := handleColon(l)
		l.readChar()
		return tok

	// 特殊字符
	case '"':
		return l.readString()
	case '#':
		return l.handleHashChar()

	default:
		if ch >= '0' && ch <= '9' {
			return l.readNumber()
		}
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' {
			return l.readIdentifier()
		}
		if unicode.IsLetter(ch) {
			return l.readIdentifier()
		}
		tok := l.newToken(TokenError, string(ch), 0)
		l.readChar()
		return tok
	}
}

// handleHashChar 处理#字符（可能是注释或指令）
// 仅 #fn 与 #def 前缀(后跟空格)识别为指令, 其余一律按注释处理
func (l *Lexer) handleHashChar() Token {
	rest := l.input[l.pos:]
	if strings.HasPrefix(rest, "fn ") || strings.HasPrefix(rest, "def ") {
		return l.readDirective()
	}
	return l.readComment()
}

// readComment 读取单行注释
// 注释以#或//开头，持续到行尾
func (l *Lexer) readComment() Token {
	col := l.column
	l.readUntil('\n')
	return l.newToken(TokenComment, "", col)
}

// readDirective 读取编译指令
// 指令以#fn开头，持续到行尾
func (l *Lexer) readDirective() Token {
	col := l.column
	start := l.pos - 1
	l.readUntil('\n')
	return l.newToken(TokenDef, l.input[start:l.pos-1], col)
}

// readIdentifier 读取标识符
// 标识符以字母或下划线开头，后跟字母、数字或下划线
// 同时处理关键字识别
func (l *Lexer) readIdentifier() Token {
	start := l.pos - 1
	col := l.column

	l.readIdentifierChars()

	value := l.input[start : l.pos-1]
	tokenType := l.lookupKeyword(value)

	return l.newToken(tokenType, value, col)
}

// readIdentifierChars 读取标识符字符
func (l *Lexer) readIdentifierChars() {
	for l.isIdentifierChar(l.ch) {
		l.readChar()
	}
}

// isIdentifierChar 判断是否为标识符字符
func (l *Lexer) isIdentifierChar(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_'
}

// lookupKeyword 查找关键字类型
func (l *Lexer) lookupKeyword(value string) TokenType {
	switch value {
	case "fn":
		return TokenFn
	case "if":
		return TokenIf
	case "else":
		return TokenElse
	case "for":
		return TokenFor
	case "range":
		return TokenRange
	case "return":
		return TokenReturn
	case "break":
		return TokenBreak
	case "continue":
		return TokenContinue
	case "throw":
		return TokenThrow
	case "true", "false":
		return TokenBool
	case "nil":
		return TokenNil
	}
	return TokenIdent
}

// skipWhitespace 跳过空白字符
func (l *Lexer) skipWhitespace() {
	for {
		switch l.ch {
		case ' ', '\t', '\r':
			l.readChar()
		case '\n':
			l.handleNewline()
			l.readChar()
		default:
			return
		}
	}
}

// readUntil 读取直到遇到指定字符或EOF
func (l *Lexer) readUntil(stop rune) {
	for l.ch != stop && l.ch != 0 {
		l.readChar()
	}
}

// handleNewline 处理换行
func (l *Lexer) handleNewline() {
	l.line++
	l.column = 0
}

// readMultilineComment 读取多行注释
func (l *Lexer) readMultilineComment() Token {
	col := l.column
	l.readChar()
	l.readChar()

	for !(l.ch == '*' && l.peekCh == '/') && l.ch != 0 {
		if l.ch == '\n' {
			l.handleNewline()
		}
		l.readChar()
	}

	if l.ch == 0 {
		return l.newToken(TokenError, fmt.Sprintf(
			"多行注释未闭合：从第%d行第%d列开始的注释缺少结束标记 */。\n"+
				"→ 问题：多行注释必须用 /* 开始，用 */ 结束。\n"+
				"→ 建议：在注释末尾添加 */，或检查是否误删了结束标记。\n"+
				"→ 示例：/* 这是注释内容 */",
			l.line, col), col)
	}
	l.readChar()
	l.readChar()
	return l.newToken(TokenComment, "", col)
}

// Tokenize 执行完整词法分析
// 返回Token切片和可能的错误
// 注释Token会被自动过滤
// 遇到无效字符时立即返回错误
func (l *Lexer) Tokenize() ([]Token, error) {
	// 预分配: 估算token数量约为输入长度的1/4
	estimated := len(l.input) / 4
	if estimated < 8 {
		estimated = 8
	}
	tokens := make([]Token, 0, estimated)

	for {
		tok := l.NextToken()

		if tok.Type == TokenEOF {
			break
		}

		if tok.Type == TokenError {
			return nil, &CompileError{
				Line:   tok.Line,
				Column: tok.Column,
				Message: fmt.Sprintf("无效字符 '%s'（第%d行第%d列）：编译器无法识别此字符。\n"+
					"→ 期望：标识符（如 myVar）、数字（如 123）、字符串（如 \"hello\"）、运算符（如 +、-）或分隔符（如 (、)）。\n"+
					"→ 建议：检查是否有特殊字符或非ASCII字符，确保使用英文标点符号", tok.Value, tok.Line, tok.Column),
			}
		}

		// NextToken已自动跳过注释, 无需额外过滤
		tokens = append(tokens, tok)
	}

	return tokens, nil
}

// ========== 字符串处理 ==========

// readString 读取双引号字符串字面量
// 支持转义字符: \n, \t, \r, \\, \"
// 未识别的转义序列原样保留
// 字符串未闭合时返回错误Token
func (l *Lexer) readString() Token {
	col := l.column
	// 跳过开始引号
	l.readChar()

	// 预分配容量：估计字符串长度为剩余输入的1/4（经验值）
	// 减少strings.Builder的扩容次数
	estimatedLen := (len(l.input) - l.pos) / 4
	if estimatedLen > 64 {
		estimatedLen = 64 // 限制预分配大小，避免浪费
	}

	var sb strings.Builder
	sb.Grow(estimatedLen)

	if !l.readStringContent(&sb) {
		return l.newToken(TokenError, fmt.Sprintf(
			"字符串未闭合：从第%d行第%d列开始的字符串缺少结束引号 \"。\n"+
				"→ 问题：字符串字面量必须用双引号包围，如 \"hello\"。\n"+
				"→ 建议：检查字符串是否在行尾或文件末尾正确闭合，或是否意外换行。\n"+
				"→ 提示：多行字符串可以使用字符串拼接：\"line1\" + \"line2\"",
			l.line, col), col)
	}

	// 跳过结束引号
	l.readChar()
	return l.newToken(TokenString, sb.String(), col)
}

// readStringContent 读取字符串内容到StringBuilder
// 返回是否正常找到结束引号
func (l *Lexer) readStringContent(sb *strings.Builder) bool {
	for !l.isStringEnd() {
		l.processStringChar(sb)
	}
	return l.ch == '"'
}

// isStringEnd 判断是否到字符串结束
func (l *Lexer) isStringEnd() bool {
	return l.ch == '"' || l.ch == 0
}

// processStringChar 处理字符串字符
func (l *Lexer) processStringChar(sb *strings.Builder) {
	if l.ch == '\\' {
		l.readChar()
		l.handleEscapeSequence(sb)
		return
	}
	// 跨行字符串内的裸换行同样更新行号
	if l.ch == '\n' {
		sb.WriteRune(l.ch)
		l.handleNewline()
		l.readChar()
		return
	}
	sb.WriteRune(l.ch)
	l.readChar()
}

// handleEscapeSequence 处理转义序列
// 对于无法识别的转义字符，保留反斜杠和原字符
func (l *Lexer) handleEscapeSequence(sb *strings.Builder) {
	switch l.ch {
	case 'n':
		sb.WriteByte('\n')
	case 't':
		sb.WriteByte('\t')
	case 'r':
		sb.WriteByte('\r')
	case '\\':
		sb.WriteByte('\\')
	case '"':
		sb.WriteByte('"')
	default:
		// 保留无法识别的转义序列
		sb.WriteByte('\\')
		sb.WriteRune(l.ch)
		l.readChar()
		return
	}
	l.readChar()
}

// ========== 数字处理 ==========

// readNumber 读取数字
// 支持格式:
//   - 十进制整数: 123, 1_000_000
//   - 十六进制整数: 0xFF, 0XAB
//   - 浮点数: 3.14, 1.5_5
//
// 下划线作为数字分隔符会被自动移除
func (l *Lexer) readNumber() Token {
	start := l.pos - 1
	col := l.column

	// 检查十六进制前缀 0x 或 0X
	if l.ch == '0' && (l.peekCh == 'x' || l.peekCh == 'X') {
		return l.readHexNumber(start, col)
	}

	// 十进制数字
	return l.readDecimalNumber(start, col)
}

// readHexNumber 读取十六进制数字
func (l *Lexer) readHexNumber(start int, col int) Token {
	// 读取 0
	l.readChar()
	// 读取 x
	l.readChar()
	for isHexDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	value := l.cleanNumberString(start)
	return l.newToken(TokenInt, value, col)
}

// readDecimalNumber 读取十进制数字
func (l *Lexer) readDecimalNumber(start int, col int) Token {
	l.readIntegerPart()

	if l.isFloatStart() {
		l.readFractionalPart()
		value := l.cleanNumberString(start)
		return l.newToken(TokenFloat, value, col)
	}

	value := l.cleanNumberString(start)
	return l.newToken(TokenInt, value, col)
}

// readIntegerPart 读取整数部分
func (l *Lexer) readIntegerPart() {
	for unicode.IsDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
}

// isFloatStart 判断是否为浮点数开始
func (l *Lexer) isFloatStart() bool {
	return l.ch == '.' && unicode.IsDigit(l.peekCh)
}

// readFractionalPart 读取小数部分
func (l *Lexer) readFractionalPart() {
	// 读取小数点
	l.readChar()
	for unicode.IsDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
}

// cleanNumberString 清理数字字符串（移除下划线）
// 直接使用ReplaceAll，在没有下划线时开销很小
func (l *Lexer) cleanNumberString(start int) string {
	return strings.ReplaceAll(l.input[start:l.pos-1], "_", "")
}

// isHexDigit 判断是否为十六进制数字
func isHexDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}
