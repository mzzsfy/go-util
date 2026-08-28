package script

import (
	"testing"
)

func Test_lexer_BasicTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected []TokenType
	}{
		{
			"x := 10",
			[]TokenType{TokenIdent, TokenAssign, TokenInt},
		},
		{
			"x + y",
			[]TokenType{TokenIdent, TokenPlus, TokenIdent},
		},
		{
			"if x > 10",
			[]TokenType{TokenIf, TokenIdent, TokenGt, TokenInt},
		},
		{
			"fn add()",
			[]TokenType{TokenFn, TokenIdent, TokenLParen, TokenRParen},
		},
		{
			"true false nil",
			[]TokenType{TokenBool, TokenBool, TokenNil},
		},
		{
			`"hello"`,
			[]TokenType{TokenString},
		},
		{
			"3.14",
			[]TokenType{TokenFloat},
		},
	}

	for _, tt := range tests {
		lexer := NewLexer(tt.input)
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Errorf("词法分析失败: %v", err)
			continue
		}

		if len(tokens) != len(tt.expected) {
			t.Errorf("Token数量不匹配: got=%d, want=%d", len(tokens), len(tt.expected))
			continue
		}

		for i, tok := range tokens {
			if tok.Type != tt.expected[i] {
				t.Errorf("Token类型不匹配 [%d]: got=%v, want=%v", i, tok.Type, tt.expected[i])
			}
		}
	}
}

func Test_lexer_Comments(t *testing.T) {
	input := `
// 这是注释
x := 10
/* 多行
注释 */
y := 20
`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("词法分析失败: %v", err)
	}

	// 注释应该被过滤掉
	expected := []TokenType{TokenIdent, TokenAssign, TokenInt, TokenIdent, TokenAssign, TokenInt}
	if len(tokens) != len(expected) {
		t.Errorf("Token数量不匹配: got=%d, want=%d", len(tokens), len(expected))
	}

	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("Token类型不匹配 [%d]: got=%v, want=%v", i, tok.Type, expected[i])
		}
	}
}

func Test_lexer_Numbers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// 基础数字
		{"123", "123"},
		{"3.14", "3.14"},
		// 下划线分隔
		{"1_000", "1000"},
		{"1_000_000", "1000000"},
		// 十六进制
		{"0xFF", "0xFF"},
		{"0x10", "0x10"},
		{"0xAB", "0xAB"},
		{"0x00", "0x00"},
		{"0XFF", "0XFF"},
		{"0xFF_FF", "0xFFFF"},
	}

	for _, tt := range tests {
		lexer := NewLexer(tt.input)
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Errorf("词法分析失败: %v", err)
			continue
		}

		if len(tokens) != 1 {
			t.Errorf("期望1个token，got=%d", len(tokens))
			continue
		}

		if tokens[0].Value != tt.expected {
			t.Errorf("值不匹配: got=%s, want=%s", tokens[0].Value, tt.expected)
		}
	}
}

func Test_lexer_EscapeSequences(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello\nworld"`, "hello\nworld"},
		{`"tab\there"`, "tab\there"},
		{`"quote\"test"`, "quote\"test"},
		{`"backslash\\"`, "backslash\\"},
		{`"carriage\rreturn"`, "carriage\rreturn"},
	}

	for _, tt := range tests {
		lexer := NewLexer(tt.input)
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Errorf("词法分析失败: %v", err)
			continue
		}

		if len(tokens) != 1 {
			t.Errorf("期望1个token，got=%d", len(tokens))
			continue
		}

		if tokens[0].Value != tt.expected {
			t.Errorf("值不匹配: got=%q, want=%q", tokens[0].Value, tt.expected)
		}
	}
}

// ========== Lexer边缘场景测试（从lexer_edge_test.go合并） ==========

// TestLexer_EmptyInput 测试空输入
func Test_lexer_EmptyInput(t *testing.T) {
	parser := NewParser()
	script, err := parser.Compile("")
	if err != nil {
		t.Fatalf("空输入编译失败: %v", err)
	}

	ctx := NewContext()
	engine := NewEngine()
	_, err = engine.Run(ctx, script)
	if err != nil {
		t.Fatalf("空输入执行失败: %v", err)
	}
}

// TestLexer_Whitespace 测试空白字符
func Test_lexer_Whitespace(t *testing.T) {
	input := "  \t  42  \n  "
	parser := NewParser()
	script, err := parser.Compile(input)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	engine := NewEngine()
	result, err := engine.Run(ctx, script)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	if result.Int() != 42 {
		t.Errorf("期望 42, 得到 %d", result.Int())
	}
}

// TestLexer_FloatUnderscore 测试浮点数中的下划线
func Test_lexer_FloatUnderscore(t *testing.T) {
	input := "1_000.500_000"
	parser := NewParser()
	script, err := parser.Compile(input)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	engine := NewEngine()
	result, err := engine.Run(ctx, script)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	expected := 1000.5
	if result.Float() != expected {
		t.Errorf("期望 %f, 得到 %f", expected, result.Float())
	}
}

// TestLexer_LongString 测试长字符串
func Test_lexer_LongString(t *testing.T) {
	longStr := ""
	for i := 0; i < 100; i++ {
		longStr += "a"
	}

	input := `"` + longStr + `"`
	parser := NewParser()
	script, err := parser.Compile(input)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	engine := NewEngine()
	result, err := engine.Run(ctx, script)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	if len(result.String()) != 100 {
		t.Errorf("期望长度100, 得到 %d", len(result.String()))
	}
}

// TestLexer_Operators 测试操作符解析
func Test_lexer_Operators(t *testing.T) {
	tests := []struct {
		input    string
		expected []TokenType
	}{
		{"x == y", []TokenType{TokenIdent, TokenEq, TokenIdent}},
		{"x != y", []TokenType{TokenIdent, TokenNeq, TokenIdent}},
		{"x >= y", []TokenType{TokenIdent, TokenGe, TokenIdent}},
		{"x <= y", []TokenType{TokenIdent, TokenLe, TokenIdent}},
		{"x && y", []TokenType{TokenIdent, TokenAnd, TokenIdent}},
		{"x || y", []TokenType{TokenIdent, TokenOr, TokenIdent}},
		{"x >> 1", []TokenType{TokenIdent, TokenRShift, TokenInt}},
		{"x << 1", []TokenType{TokenIdent, TokenLShift, TokenInt}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("词法分析失败: %v", err)
			}

			if len(tokens) != len(tt.expected) {
				t.Fatalf("Token数量不匹配: got=%d, want=%d", len(tokens), len(tt.expected))
			}

			for i, tok := range tokens {
				if tok.Type != tt.expected[i] {
					t.Errorf("Token类型不匹配 [%d]: got=%v, want=%v", i, tok.Type, tt.expected[i])
				}
			}
		})
	}
}

// TestLexer_ArrowOperator 测试箭头操作符
func Test_lexer_ArrowOperator(t *testing.T) {
	input := "x => int"
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("词法分析失败: %v", err)
	}

	expected := []TokenType{TokenIdent, TokenArrow, TokenIdent}
	if len(tokens) != len(expected) {
		t.Fatalf("Token数量不匹配: got=%d, want=%d", len(tokens), len(expected))
	}

	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("Token类型不匹配 [%d]: got=%v, want=%v", i, tok.Type, expected[i])
		}
	}
}

// TestLexer_EscapeSequences_EdgeCases 测试转义序列边缘情况
func Test_lexer_EscapeSequences_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// 无法识别的转义序列保留原样
		{"未知转义", `"hello\xworld"`, "hello\\xworld"},
		{"数字转义", `"test\1escape"`, "test\\1escape"},
		{"字母转义", `"test\aescope"`, "test\\aescope"},
		// 空字符串
		{"空字符串", `""`, ""},
		// 单字符字符串
		{"单字符", `"a"`, "a"},
		// 转义序列在结尾
		{"结尾换行", `"test\n"`, "test\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Errorf("词法分析失败: %v", err)
				return
			}

			if len(tokens) != 1 {
				t.Errorf("期望1个token，got=%d", len(tokens))
				return
			}

			if tokens[0].Value != tt.expected {
				t.Errorf("值不匹配: got=%q, want=%q", tokens[0].Value, tt.expected)
			}
		})
	}
}

// TestLexer_UnterminatedString2 测试未闭合字符串
func Test_lexer_UnterminatedString2(t *testing.T) {
	input := `"hello world`
	lexer := NewLexer(input)
	_, err := lexer.Tokenize()
	if err == nil {
		t.Error("应该报告未闭合字符串错误")
	}
}

// TestLexer_UnterminatedMultilineComment 测试未闭合多行注释
func Test_lexer_UnterminatedMultilineComment(t *testing.T) {
	input := `/* 这是一个
未闭合的
注释`
	lexer := NewLexer(input)
	_, err := lexer.Tokenize()
	if err == nil {
		t.Error("应该报告未闭合多行注释错误")
	}
}

// TestLexer_DefDirective 测试#fn指令
func Test_lexer_DefDirective(t *testing.T) {
	tests := []struct {
		input string
		value string
	}{
		{"#fn myFunc()", "#fn myFunc()"},
		{"#fn myFunc(x, y)", "#fn myFunc(x, y)"},
		{"#fn myFunc(x)=>int", "#fn myFunc(x)=>int"},
	}

	for _, tt := range tests {
		lexer := NewLexer(tt.input)
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Errorf("词法分析失败: %v", err)
			continue
		}

		if len(tokens) != 1 {
			t.Errorf("期望1个token，got=%d", len(tokens))
			continue
		}

		if tokens[0].Type != TokenDef {
			t.Errorf("期望TokenDef，got=%v", tokens[0].Type)
		}

		if tokens[0].Value != tt.value {
			t.Errorf("值不匹配: got=%q, want=%q", tokens[0].Value, tt.value)
		}
	}
}

// TestLexer_InvalidCharacter 测试无效字符
func Test_lexer_InvalidCharacter(t *testing.T) {
	input := "x @ y"
	lexer := NewLexer(input)
	_, err := lexer.Tokenize()
	if err == nil {
		t.Error("应该报告无效字符错误")
	}
}

// TestLexer_MultilineComment2 测试多行注释
func Test_lexer_MultilineComment2(t *testing.T) {
	input := `/* 注释第一行
注释第二行
注释第三行 */
42`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("词法分析失败: %v", err)
	}

	// 应该只有一个整数token（注释被过滤）
	if len(tokens) != 1 {
		t.Errorf("期望1个token，got=%d", len(tokens))
	}

	if tokens[0].Type != TokenInt {
		t.Errorf("期望TokenInt，got=%v", tokens[0].Type)
	}
}

// TestLexer_TypeAnnotation 测试类型注解
func Test_lexer_TypeAnnotation(t *testing.T) {
	tests := []struct {
		input    string
		expected []TokenType
	}{
		// >int 被正确解析为TokenGt + TokenIdent，TokenTypeAnnot需要紧凑语法如>int
		{"x <string", []TokenType{TokenIdent, TokenLt, TokenIdent}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Fatalf("词法分析失败: %v", err)
			}

			if len(tokens) != len(tt.expected) {
				t.Fatalf("Token数量不匹配: got=%d, want=%d", len(tokens), len(tt.expected))
			}

			for i, tok := range tokens {
				if tok.Type != tt.expected[i] {
					t.Errorf("Token类型不匹配 [%d]: got=%v, want=%v", i, tok.Type, tt.expected[i])
				}
			}
		})
	}
}
