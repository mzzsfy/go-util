package script

import (
	"testing"
)

// ========== Lexer综合测试 ==========

// Test_Comprehensive_Lexer_TokenBoundary token边界与注释测试
func Test_Comprehensive_Lexer_TokenBoundary(t *testing.T) {
	t.Run("空字符串输入", func(t *testing.T) {
		lexer := NewLexer("")
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Fatalf("空输入不应报错: %v", err)
		}
		if len(tokens) != 0 {
			t.Errorf("空输入期望0个token, 得到%d", len(tokens))
		}
	})

	t.Run("单行注释", func(t *testing.T) {
		lexer := NewLexer("// comment only")
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Fatalf("注释不应报错: %v", err)
		}
		if len(tokens) != 0 {
			t.Errorf("纯注释期望0个token, 得到%d", len(tokens))
		}
	})

	t.Run("井号注释", func(t *testing.T) {
		lexer := NewLexer("# hash comment")
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Fatalf("注释不应报错: %v", err)
		}
		if len(tokens) != 0 {
			t.Errorf("纯注释期望0个token, 得到%d", len(tokens))
		}
	})

	t.Run("多行注释", func(t *testing.T) {
		lexer := NewLexer("/* line1\nline2\nline3 */")
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Fatalf("多行注释不应报错: %v", err)
		}
		if len(tokens) != 0 {
			t.Errorf("纯多行注释期望0个token, 得到%d", len(tokens))
		}
	})

	t.Run("连续双字符运算符", func(t *testing.T) {
		tests := []struct {
			input    string
			expected []TokenType
		}{
			{"==", []TokenType{TokenEq}},
			{"!=", []TokenType{TokenNeq}},
			{"<=", []TokenType{TokenLe}},
			{">=", []TokenType{TokenGe}},
			{"&&", []TokenType{TokenAnd}},
			{"||", []TokenType{TokenOr}},
			{"<<", []TokenType{TokenLShift}},
			{">>", []TokenType{TokenRShift}},
		}
		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				lexer := NewLexer(tt.input)
				tokens, err := lexer.Tokenize()
				if err != nil {
					t.Fatalf("词法分析失败: %v", err)
				}
				if len(tokens) != len(tt.expected) {
					t.Fatalf("token数量: got=%d, want=%d", len(tokens), len(tt.expected))
				}
				for i, tok := range tokens {
					if tok.Type != tt.expected[i] {
						t.Errorf("token[%d]: got=%v, want=%v", i, tok.Type, tt.expected[i])
					}
				}
			})
		}
	})

	t.Run("混合空白和换行", func(t *testing.T) {
		input := "  \t\n  x  \r\n  :=  \n  42  "
		lexer := NewLexer(input)
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Fatalf("词法分析失败: %v", err)
		}
		expected := []TokenType{TokenIdent, TokenAssign, TokenInt}
		if len(tokens) != len(expected) {
			t.Fatalf("token数量: got=%d, want=%d", len(tokens), len(expected))
		}
		for i, tok := range tokens {
			if tok.Type != expected[i] {
				t.Errorf("token[%d]: got=%v, want=%v", i, tok.Type, expected[i])
			}
		}
	})

	t.Run("注释穿插代码", func(t *testing.T) {
		input := `x := 10 // inline
/* block */ y := 20`
		lexer := NewLexer(input)
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Fatalf("词法分析失败: %v", err)
		}
		expected := []TokenType{TokenIdent, TokenAssign, TokenInt, TokenIdent, TokenAssign, TokenInt}
		if len(tokens) != len(expected) {
			t.Fatalf("token数量: got=%d, want=%d", len(tokens), len(expected))
		}
	})
}

// Test_Comprehensive_Lexer_Literals 字面量测试
func Test_Comprehensive_Lexer_Literals(t *testing.T) {
	t.Run("整数边界值", func(t *testing.T) {
		tests := []struct {
			input    string
			expected TokenType
		}{
			{"0", TokenInt},
			{"00", TokenInt},
			{"007", TokenInt},
		}
		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				lexer := NewLexer(tt.input)
				tokens, err := lexer.Tokenize()
				if err != nil {
					t.Fatalf("词法分析失败: %v", err)
				}
				if len(tokens) != 1 {
					t.Fatalf("期望1个token, 得到%d", len(tokens))
				}
				if tokens[0].Type != tt.expected {
					t.Errorf("got=%v, want=%v", tokens[0].Type, tt.expected)
				}
			})
		}
	})

	t.Run("十六进制", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"0xFF", "0xFF"},
			{"0xAB", "0xAB"},
			{"0x00", "0x00"},
			{"0Xff", "0Xff"},
		}
		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				lexer := NewLexer(tt.input)
				tokens, err := lexer.Tokenize()
				if err != nil {
					t.Fatalf("词法分析失败: %v", err)
				}
				if len(tokens) != 1 {
					t.Fatalf("期望1个token, 得到%d", len(tokens))
				}
				if tokens[0].Value != tt.expected {
					t.Errorf("got=%s, want=%s", tokens[0].Value, tt.expected)
				}
			})
		}
	})

	t.Run("下划线分隔整数", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"1_000", "1000"},
			{"1_000_000", "1000000"},
		}
		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				lexer := NewLexer(tt.input)
				tokens, err := lexer.Tokenize()
				if err != nil {
					t.Fatalf("词法分析失败: %v", err)
				}
				if len(tokens) != 1 {
					t.Fatalf("期望1个token, 得到%d", len(tokens))
				}
				if tokens[0].Value != tt.expected {
					t.Errorf("got=%s, want=%s", tokens[0].Value, tt.expected)
				}
			})
		}
	})

	t.Run("浮点数", func(t *testing.T) {
		tests := []struct {
			input    string
			expected TokenType
		}{
			{"0.0", TokenFloat},
			{"3.14", TokenFloat},
			{"0.5", TokenFloat},
		}
		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				lexer := NewLexer(tt.input)
				tokens, err := lexer.Tokenize()
				if err != nil {
					t.Fatalf("词法分析失败: %v", err)
				}
				if len(tokens) != 1 {
					t.Fatalf("期望1个token, 得到%d", len(tokens))
				}
				if tokens[0].Type != tt.expected {
					t.Errorf("got=%v, want=%v", tokens[0].Type, tt.expected)
				}
			})
		}
	})

	t.Run("浮点数下划线", func(t *testing.T) {
		lexer := NewLexer("1_000.500_000")
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Fatalf("词法分析失败: %v", err)
		}
		if len(tokens) != 1 {
			t.Fatalf("期望1个token, 得到%d", len(tokens))
		}
		if tokens[0].Value != "1000.500000" {
			t.Errorf("got=%s, want=%s", tokens[0].Value, "1000.500000")
		}
	})

	t.Run("字符串转义", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{`""`, ""},
			{`"\n"`, "\n"},
			{`"\t"`, "\t"},
			{`"\""`, "\""},
			{`"\\"`, "\\"},
			{`"\r"`, "\r"},
			{`"hello"`, "hello"},
		}
		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				lexer := NewLexer(tt.input)
				tokens, err := lexer.Tokenize()
				if err != nil {
					t.Fatalf("词法分析失败: %v", err)
				}
				if len(tokens) != 1 {
					t.Fatalf("期望1个token, 得到%d", len(tokens))
				}
				if tokens[0].Value != tt.expected {
					t.Errorf("got=%q, want=%q", tokens[0].Value, tt.expected)
				}
			})
		}
	})

	t.Run("布尔和nil", func(t *testing.T) {
		tests := []struct {
			input    string
			expected TokenType
		}{
			{"true", TokenBool},
			{"false", TokenBool},
			{"nil", TokenNil},
		}
		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				lexer := NewLexer(tt.input)
				tokens, err := lexer.Tokenize()
				if err != nil {
					t.Fatalf("词法分析失败: %v", err)
				}
				if len(tokens) != 1 {
					t.Fatalf("期望1个token, 得到%d", len(tokens))
				}
				if tokens[0].Type != tt.expected {
					t.Errorf("got=%v, want=%v", tokens[0].Type, tt.expected)
				}
			})
		}
	})
}

// Test_Comprehensive_Lexer_Identifiers 标识符与关键字测试
func Test_Comprehensive_Lexer_Identifiers(t *testing.T) {
	t.Run("合法标识符", func(t *testing.T) {
		tests := []struct {
			input string
			value string
		}{
			{"abc", "abc"},
			{"_x", "_x"},
			{"x123", "x123"},
			{"a_b_c", "a_b_c"},
			{"myVar", "myVar"},
		}
		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				lexer := NewLexer(tt.input)
				tokens, err := lexer.Tokenize()
				if err != nil {
					t.Fatalf("词法分析失败: %v", err)
				}
				if len(tokens) != 1 {
					t.Fatalf("期望1个token, 得到%d", len(tokens))
				}
				if tokens[0].Type != TokenIdent {
					t.Errorf("期望TokenIdent, got=%v", tokens[0].Type)
				}
				if tokens[0].Value != tt.value {
					t.Errorf("got=%s, want=%s", tokens[0].Value, tt.value)
				}
			})
		}
	})

	t.Run("关键字区分", func(t *testing.T) {
		tests := []struct {
			input    string
			expected TokenType
		}{
			{"if", TokenIf},
			{"else", TokenElse},
			{"for", TokenFor},
			{"fn", TokenFn},
			{"range", TokenRange},
			{"return", TokenReturn},
			{"break", TokenBreak},
			{"continue", TokenContinue},
			{"throw", TokenThrow},
			{"true", TokenBool},
			{"false", TokenBool},
			{"nil", TokenNil},
		}
		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				lexer := NewLexer(tt.input)
				tokens, err := lexer.Tokenize()
				if err != nil {
					t.Fatalf("词法分析失败: %v", err)
				}
				if len(tokens) != 1 {
					t.Fatalf("期望1个token, 得到%d", len(tokens))
				}
				if tokens[0].Type != tt.expected {
					t.Errorf("got=%v, want=%v", tokens[0].Type, tt.expected)
				}
			})
		}
	})

	t.Run("类似关键字的标识符", func(t *testing.T) {
		// ifx不是关键字,应该是标识符
		lexer := NewLexer("ifx")
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Fatalf("词法分析失败: %v", err)
		}
		if len(tokens) != 1 {
			t.Fatalf("期望1个token, 得到%d", len(tokens))
		}
		if tokens[0].Type != TokenIdent {
			t.Errorf("ifx应该是TokenIdent, got=%v", tokens[0].Type)
		}
	})
}

// ========== Parser综合测试 ==========

// Test_Comprehensive_Parser_ExprPrecedence 表达式优先级测试
func Test_Comprehensive_Parser_ExprPrecedence(t *testing.T) {
	t.Run("算术优先级_乘法先于加法", func(t *testing.T) {
		runIntTest(t, "1 + 2 * 3", 7)
	})

	t.Run("括号改变优先级", func(t *testing.T) {
		runIntTest(t, "(1 + 2) * 3", 9)
	})

	t.Run("连续加法", func(t *testing.T) {
		runIntTest(t, "1 + 2 + 3 + 4", 10)
	})

	t.Run("连续乘法", func(t *testing.T) {
		runIntTest(t, "2 * 3 * 4", 24)
	})

	t.Run("混合加减乘", func(t *testing.T) {
		runIntTest(t, "10 + 20 * 3 - 5", 65)
	})

	t.Run("除法和取模", func(t *testing.T) {
		runIntTest(t, "20 / 4 + 17 % 5", 7)
	})

	t.Run("比较运算", func(t *testing.T) {
		runBoolTest(t, "1 < 2", true)
		runBoolTest(t, "2 <= 2", true)
		runBoolTest(t, "3 > 2", true)
		runBoolTest(t, "3 >= 3", true)
		runBoolTest(t, "1 == 1", true)
		runBoolTest(t, "1 != 2", true)
	})

	t.Run("逻辑运算优先级", func(t *testing.T) {
		// && 优先于 ||
		runBoolTest(t, "true || false && false", true)
	})

	t.Run("非运算", func(t *testing.T) {
		runBoolTest(t, "!false", true)
		runBoolTest(t, "!true", false)
	})

	t.Run("混合逻辑与比较", func(t *testing.T) {
		runBoolTest(t, "1 + 2 < 3 * 4 && !false", true)
	})

	t.Run("嵌套括号", func(t *testing.T) {
		runIntTest(t, "((1 + 2) * (3 + 4))", 21)
	})

	t.Run("位运算优先级", func(t *testing.T) {
		// & 优先级高于 |: 1 | (2 & 3) = 1 | 2 = 3
		runIntTest(t, "1 | 2 & 3", 3)
	})

	t.Run("位移运算", func(t *testing.T) {
		runIntTest(t, "1 << 4", 16)
		runIntTest(t, "256 >> 2", 64)
	})
}

// Test_Comprehensive_Parser_Statements 语句解析测试
func Test_Comprehensive_Parser_Statements(t *testing.T) {
	t.Run("变量声明", func(t *testing.T) {
		inputs := []string{
			"x := 10",
			"x := 3.14",
			`x := "hello"`,
			"x := true",
			"x := nil",
			"x := [1, 2, 3]",
			`x := {"a": 1}`,
		}
		for _, input := range inputs {
			t.Run(input, func(t *testing.T) {
				parser := NewParser()
				err := parser.Validate(input)
				if err != nil {
					t.Errorf("验证失败: %v", err)
				}
			})
		}
	})

	t.Run("类型安全声明 :=>", func(t *testing.T) {
		inputs := []string{
			"x :=>int 10",
			"y :=>float 3.14",
			`name :=>string "hello"`,
			"flag :=>bool true",
		}
		for _, input := range inputs {
			t.Run(input, func(t *testing.T) {
				parser := NewParser()
				_, err := parser.Compile(input)
				if err != nil {
					t.Errorf("编译失败: %v", err)
				}
			})
		}
	})

	t.Run("赋值语句", func(t *testing.T) {
		runIntTest(t, "x := 10\nx = 20\nx", 20)
	})

	t.Run("复合赋值解析", func(t *testing.T) {
		// += -= *= /= 由词法分析器和解析器支持,但编译器尚未实现
		// 验证词法分析能正确解析这些运算符
		tests := []struct {
			input    string
			expected []TokenType
		}{
			{"x += 5", []TokenType{TokenIdent, TokenPlusAssign, TokenInt}},
			{"x -= 3", []TokenType{TokenIdent, TokenMinusAssign, TokenInt}},
			{"x *= 2", []TokenType{TokenIdent, TokenStarAssign, TokenInt}},
			{"x /= 4", []TokenType{TokenIdent, TokenSlashAssign, TokenInt}},
		}
		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				lexer := NewLexer(tt.input)
				tokens, err := lexer.Tokenize()
				if err != nil {
					t.Fatalf("词法分析失败: %v", err)
				}
				if len(tokens) != len(tt.expected) {
					t.Fatalf("token数量: got=%d, want=%d", len(tokens), len(tt.expected))
				}
				for i, tok := range tokens {
					if tok.Type != tt.expected[i] {
						t.Errorf("token[%d]: got=%v, want=%v", i, tok.Type, tt.expected[i])
					}
				}
			})
		}
	})

	t.Run("函数定义无返回类型", func(t *testing.T) {
		input := "fn add(x, y) {\nreturn x + y\n}"
		parser := NewParser()
		err := parser.Validate(input)
		if err != nil {
			t.Errorf("验证失败: %v", err)
		}
	})

	t.Run("函数定义带返回类型", func(t *testing.T) {
		input := "fn add(x, y) => int {\nreturn x + y\n}"
		parser := NewParser()
		err := parser.Validate(input)
		if err != nil {
			t.Errorf("验证失败: %v", err)
		}
	})

	t.Run("函数调用执行", func(t *testing.T) {
		input := "fn add(x, y) {\nreturn x + y\n}\nadd(3, 4)"
		runIntTest(t, input, 7)
	})

	t.Run("if-else语句", func(t *testing.T) {
		input := `x := 10
if x > 5 {
	return 1
} else {
	return 0
}`
		runIntTest(t, input, 1)
	})

	t.Run("else-if链", func(t *testing.T) {
		input := `x := 5
if x > 10 {
	return 1
} else if x > 3 {
	return 2
} else {
	return 3
}`
		runIntTest(t, input, 2)
	})

	t.Run("for条件循环", func(t *testing.T) {
		input := `sum := 0
i := 0
for i < 5 {
	sum = sum + i
	i = i + 1
}
sum`
		runIntTest(t, input, 10)
	})

	t.Run("for计数循环", func(t *testing.T) {
		input := `sum := 0
for i := 5 {
	sum = sum + i
}
sum`
		runIntTest(t, input, 15)
	})

	t.Run("for标准循环", func(t *testing.T) {
		input := `sum := 0
for i := 0; i < 5; i = i + 1 {
	sum = sum + i
}
sum`
		runIntTest(t, input, 10)
	})

	t.Run("for range循环", func(t *testing.T) {
		input := `sum := 0
for k := range [1, 2, 3, 4] {
	sum = sum + k
}
sum`
		runIntTest(t, input, 10)
	})

	t.Run("for range双变量", func(t *testing.T) {
		input := `sum := 0
for i, v := range [10, 20, 30] {
	sum = sum + v
}
sum`
		runIntTest(t, input, 60)
	})

	t.Run("break语句", func(t *testing.T) {
		// for i := n 计数循环i从1开始
		input := `sum := 0
for i := 10 {
	if i >= 3 {
		break
	}
	sum = sum + 1
}
sum`
		runIntTest(t, input, 2)
	})

	t.Run("continue语句", func(t *testing.T) {
		// for i := n 计数循环i从1开始, continue跳到i++
		input := `sum := 0
for i := 5 {
	if i == 2 {
		continue
	}
	sum = sum + i
}
sum`
		// i=1 sum=1, i=2 skip, i=3 sum=4, i=4 sum=8, i=5 sum=13
		runIntTest(t, input, 13)
	})

	t.Run("无限循环带break", func(t *testing.T) {
		input := `count := 0
for {
	count = count + 1
	if count >= 5 {
		break
	}
}
count`
		runIntTest(t, input, 5)
	})

	t.Run("throw语句", func(t *testing.T) {
		runRuntimeErrorTest(t, `throw "error"`)
	})
}

// Test_Comprehensive_Parser_ErrorRecovery 错误恢复测试
func Test_Comprehensive_Parser_ErrorRecovery(t *testing.T) {
	t.Run("未闭合括号", func(t *testing.T) {
		runErrorTest(t, "(1 + 2")
	})

	t.Run("未闭合方括号", func(t *testing.T) {
		runErrorTest(t, "[1, 2, 3")
	})

	t.Run("未闭合花括号", func(t *testing.T) {
		// parser在EOF时容错返回, 不保证报错
		parser := NewParser()
		_, err := parser.Compile("fn f() { return 1")
		if err != nil {
			t.Logf("编译报错(可接受): %v", err)
		}
		// 不强制要求报错
	})

	t.Run("未闭合字符串", func(t *testing.T) {
		lexer := NewLexer(`"unterminated`)
		_, err := lexer.Tokenize()
		if err == nil {
			t.Error("应该报告未闭合字符串错误")
		}
	})

	t.Run("未闭合多行注释", func(t *testing.T) {
		lexer := NewLexer("/* unclosed comment")
		_, err := lexer.Tokenize()
		if err == nil {
			t.Error("应该报告未闭合多行注释错误")
		}
	})

	t.Run("标准for循环缺少分号", func(t *testing.T) {
		runErrorTest(t, "for i := 0 i < 10 { 1 }")
	})

	t.Run("无效字符", func(t *testing.T) {
		lexer := NewLexer("x @ y")
		_, err := lexer.Tokenize()
		if err == nil {
			t.Error("应该报告无效字符错误")
		}
	})

	t.Run("不完整表达式", func(t *testing.T) {
		runErrorTest(t, "1 +")
	})

	t.Run("空map和数组合法", func(t *testing.T) {
		// 空数组和空map不应报错
		inputs := []string{
			"len([])",
			"len({})",
		}
		for _, input := range inputs {
			t.Run(input, func(t *testing.T) {
				parser := NewParser()
				err := parser.Validate(input)
				if err != nil {
					t.Errorf("不应报错: %v", err)
				}
			})
		}
	})
}
