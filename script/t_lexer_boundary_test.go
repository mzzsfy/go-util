package script

import (
	"fmt"
	"testing"
)

// ========== 辅助函数 ==========

// lexAll 使用NextToken收集所有非EOF的Token(含错误Token)
func lexAll(input string) []Token {
	l := NewLexer(input)
	var toks []Token
	for {
		tok := l.NextToken()
		if tok.Type == TokenEOF {
			break
		}
		toks = append(toks, tok)
	}
	return toks
}

// assertTypes 检查Token序列的类型是否匹配
func assertTypes(t *testing.T, toks []Token, want []TokenType) {
	t.Helper()
	if len(toks) != len(want) {
		t.Fatalf("token数量: got %d, want %d", len(toks), len(want))
	}
	for i, w := range want {
		if toks[i].Type != w {
			t.Errorf("token[%d].Type = %v, want %v", i, toks[i].Type, w)
		}
	}
}

// ========== 数字Token测试 ==========

// Test_LexerBoundary_DecimalIntegers 十进制整数
func Test_LexerBoundary_DecimalIntegers(t *testing.T) {
	tests := []struct {
		input string
		value string
	}{
		{"0", "0"},
		{"1", "1"},
		{"42", "42"},
		{"999999", "999999"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			toks := lexAll(tt.input)
			if len(toks) != 1 {
				t.Fatalf("token数量: got %d, want 1", len(toks))
			}
			if toks[0].Type != TokenInt {
				t.Errorf("Type = %v, want TokenInt", toks[0].Type)
			}
			if toks[0].Value != tt.value {
				t.Errorf("Value = %q, want %q", toks[0].Value, tt.value)
			}
		})
	}
}

// Test_LexerBoundary_NegativeNumberAsTokens 负数在词法层产生TokenMinus+TokenInt
func Test_LexerBoundary_NegativeNumberAsTokens(t *testing.T) {
	t.Run("-1", func(t *testing.T) {
		toks := lexAll("-1")
		assertTypes(t, toks, []TokenType{TokenMinus, TokenInt})
		if toks[1].Value != "1" {
			t.Errorf("int value = %q, want \"1\"", toks[1].Value)
		}
	})
	t.Run("-42", func(t *testing.T) {
		toks := lexAll("-42")
		assertTypes(t, toks, []TokenType{TokenMinus, TokenInt})
	})
}

// Test_LexerBoundary_HexNumbers 十六进制整数
func Test_LexerBoundary_HexNumbers(t *testing.T) {
	tests := []struct {
		input string
		value string
	}{
		{"0xFF", "0xFF"},
		{"0XAB", "0XAB"},
		{"0x0", "0x0"},
		{"0xDEADBEEF", "0xDEADBEEF"},
		{"0xabcdef", "0xabcdef"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			toks := lexAll(tt.input)
			if len(toks) != 1 {
				t.Fatalf("token数量: got %d, want 1", len(toks))
			}
			if toks[0].Type != TokenInt {
				t.Errorf("Type = %v, want TokenInt", toks[0].Type)
			}
			if toks[0].Value != tt.value {
				t.Errorf("Value = %q, want %q", toks[0].Value, tt.value)
			}
		})
	}
}

// Test_LexerBoundary_HexUnderscore 十六进制数字分隔符
func Test_LexerBoundary_HexUnderscore(t *testing.T) {
	t.Run("0xFF_FF", func(t *testing.T) {
		toks := lexAll("0xFF_FF")
		if len(toks) != 1 || toks[0].Type != TokenInt {
			t.Fatalf("应为单个TokenInt")
		}
		if toks[0].Value != "0xFFFF" {
			t.Errorf("Value = %q, want \"0xFFFF\"", toks[0].Value)
		}
	})
}

// Test_LexerBoundary_NumberUnderscoreSeparator 数字分隔符下划线被移除
func Test_LexerBoundary_NumberUnderscoreSeparator(t *testing.T) {
	tests := []struct {
		input string
		value string
	}{
		{"1_000_000", "1000000"},
		{"1_2_3", "123"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			toks := lexAll(tt.input)
			if len(toks) != 1 || toks[0].Type != TokenInt {
				t.Fatalf("应为单个TokenInt")
			}
			if toks[0].Value != tt.value {
				t.Errorf("Value = %q, want %q", toks[0].Value, tt.value)
			}
		})
	}
}

// Test_LexerBoundary_FloatNumbers 浮点数
func Test_LexerBoundary_FloatNumbers(t *testing.T) {
	tests := []struct {
		input string
		value string
	}{
		{"0.0", "0.0"},
		{"3.14", "3.14"},
		{"0.5", "0.5"},
		{"100.001", "100.001"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			toks := lexAll(tt.input)
			if len(toks) != 1 {
				t.Fatalf("token数量: got %d, want 1", len(toks))
			}
			if toks[0].Type != TokenFloat {
				t.Errorf("Type = %v, want TokenFloat", toks[0].Type)
			}
			if toks[0].Value != tt.value {
				t.Errorf("Value = %q, want %q", toks[0].Value, tt.value)
			}
		})
	}
}

// Test_LexerBoundary_DotPrefixedNumber .5在词法层产生错误Token(引擎行为)
func Test_LexerBoundary_DotPrefixedNumber(t *testing.T) {
	t.Run(".5", func(t *testing.T) {
		toks := lexAll(".5")
		// 点号不是有效起始字符, 产生TokenError
		if len(toks) < 1 {
			t.Fatal("应至少产生1个token")
		}
		if toks[0].Type != TokenError {
			t.Errorf("首个token应为TokenError, got %v", toks[0].Type)
		}
	})
}

// Test_LexerBoundary_NumberTrailingDot 5.在词法层产生TokenInt后跟TokenError(引擎行为)
func Test_LexerBoundary_NumberTrailingDot(t *testing.T) {
	t.Run("5.", func(t *testing.T) {
		toks := lexAll("5.")
		if len(toks) < 2 {
			t.Fatal("应至少产生2个token")
		}
		if toks[0].Type != TokenInt {
			t.Errorf("首个token应为TokenInt, got %v", toks[0].Type)
		}
		if toks[0].Value != "5" {
			t.Errorf("Value = %q, want \"5\"", toks[0].Value)
		}
		if toks[1].Type != TokenError {
			t.Errorf("第二个token应为TokenError(点号), got %v", toks[1].Type)
		}
	})
}

// Test_LexerBoundary_NumberFollowedByIdentifier 数字后跟标识符(引擎行为)
func Test_LexerBoundary_NumberFollowedByIdentifier(t *testing.T) {
	t.Run("123abc", func(t *testing.T) {
		toks := lexAll("123abc")
		assertTypes(t, toks, []TokenType{TokenInt, TokenIdent})
		if toks[0].Value != "123" {
			t.Errorf("int value = %q, want \"123\"", toks[0].Value)
		}
		if toks[1].Value != "abc" {
			t.Errorf("ident value = %q, want \"abc\"", toks[1].Value)
		}
	})
}

// ========== 字符串Token测试 ==========

// Test_LexerBoundary_EmptyString 空字符串
func Test_LexerBoundary_EmptyString(t *testing.T) {
	t.Run(`""`, func(t *testing.T) {
		toks := lexAll(`""`)
		if len(toks) != 1 {
			t.Fatalf("token数量: got %d, want 1", len(toks))
		}
		if toks[0].Type != TokenString {
			t.Errorf("Type = %v, want TokenString", toks[0].Type)
		}
		if toks[0].Value != "" {
			t.Errorf("Value = %q, want empty", toks[0].Value)
		}
	})
}

// Test_LexerBoundary_SimpleString 普通字符串
func Test_LexerBoundary_SimpleString(t *testing.T) {
	t.Run(`"hello"`, func(t *testing.T) {
		toks := lexAll(`"hello"`)
		if len(toks) != 1 {
			t.Fatalf("token数量: got %d, want 1", len(toks))
		}
		if toks[0].Type != TokenString {
			t.Errorf("Type = %v, want TokenString", toks[0].Type)
		}
		if toks[0].Value != "hello" {
			t.Errorf("Value = %q, want \"hello\"", toks[0].Value)
		}
	})
}

// Test_LexerBoundary_StringBackslashEscape 反斜杠转义序列
func Test_LexerBoundary_StringBackslashEscape(t *testing.T) {
	t.Run("backslash_escape", func(t *testing.T) {
		// 输入字符串: "a\\b" (字面双反斜杠, 词法器解释为单个反斜杠)
		input := "\"a" + "\\\\" + "b\""
		toks := lexAll(input)
		if len(toks) != 1 {
			t.Fatalf("token数量: got %d, want 1", len(toks))
		}
		if toks[0].Type != TokenString {
			t.Errorf("Type = %v, want TokenString", toks[0].Type)
		}
		want := "a" + "\\" + "b"
		if toks[0].Value != want {
			t.Errorf("Value = %q, want %q", toks[0].Value, want)
		}
	})
}

// Test_LexerBoundary_StringBasicEscapes 字符串转义序列
func Test_LexerBoundary_StringBasicEscapes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{`a\nb`, `"a\nb"`, "a\nb"},
		{`a\tb`, `"a\tb"`, "a\tb"},
		{`a\rb`, `"a\rb"`, "a\rb"},
		{`a\"b`, `"a\"b"`, "a\"b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toks := lexAll(tt.input)
			if len(toks) != 1 {
				t.Fatalf("token数量: got %d, want 1", len(toks))
			}
			if toks[0].Type != TokenString {
				t.Errorf("Type = %v, want TokenString", toks[0].Type)
			}
			if toks[0].Value != tt.want {
				t.Errorf("Value = %q, want %q", toks[0].Value, tt.want)
			}
		})
	}
}

// Test_LexerBoundary_StringUnknownEscape 未识别转义保留原样
func Test_LexerBoundary_StringUnknownEscape(t *testing.T) {
	t.Run(`a\xb`, func(t *testing.T) {
		toks := lexAll(`"a\xb"`)
		if len(toks) != 1 {
			t.Fatalf("token数量: got %d, want 1", len(toks))
		}
		if toks[0].Type != TokenString {
			t.Errorf("Type = %v, want TokenString", toks[0].Type)
		}
		// 未识别转义\x保留为字面量
		if toks[0].Value != `a\xb` {
			t.Errorf("Value = %q, want %q", toks[0].Value, `a\xb`)
		}
	})
}

// Test_LexerBoundary_StringContainsNewline 字符串中包含转义换行
func Test_LexerBoundary_StringContainsNewline(t *testing.T) {
	t.Run(`hello\nworld`, func(t *testing.T) {
		toks := lexAll(`"hello\nworld"`)
		if len(toks) != 1 {
			t.Fatalf("token数量: got %d, want 1", len(toks))
		}
		if toks[0].Value != "hello\nworld" {
			t.Errorf("Value = %q, want \"hello\nworld\"", toks[0].Value)
		}
	})
}

// Test_LexerBoundary_UnicodeString Unicode字符串
func Test_LexerBoundary_UnicodeString(t *testing.T) {
	t.Run("chinese_chars", func(t *testing.T) {
		toks := lexAll(`"中文"`)
		if len(toks) != 1 {
			t.Fatalf("token数量: got %d, want 1", len(toks))
		}
		// 验证Unicode字符串被识别为TokenString
		if toks[0].Type != TokenString {
			t.Errorf("Type = %v, want TokenString", toks[0].Type)
		}
	})
}

// Test_LexerBoundary_UnclosedStringError 未闭合字符串产生错误Token
func Test_LexerBoundary_UnclosedStringError(t *testing.T) {
	t.Run(`"hello`, func(t *testing.T) {
		toks := lexAll(`"hello`)
		if len(toks) != 1 {
			t.Fatalf("token数量: got %d, want 1", len(toks))
		}
		if toks[0].Type != TokenError {
			t.Errorf("Type = %v, want TokenError", toks[0].Type)
		}
	})
}

// Test_LexerBoundary_EmptyStringThenOperator 空字符串后跟操作符
func Test_LexerBoundary_EmptyStringThenOperator(t *testing.T) {
	t.Run(`""+""`, func(t *testing.T) {
		toks := lexAll(`""+""`)
		assertTypes(t, toks, []TokenType{TokenString, TokenPlus, TokenString})
		if toks[0].Value != "" {
			t.Errorf("token[0] Value = %q, want empty", toks[0].Value)
		}
		if toks[2].Value != "" {
			t.Errorf("token[2] Value = %q, want empty", toks[2].Value)
		}
	})
}

// ========== 标识符和关键字测试 ==========

// Test_LexerBoundary_AllKeywords 所有关键字
func Test_LexerBoundary_AllKeywords(t *testing.T) {
	tests := []struct {
		kw  string
		typ TokenType
	}{
		{"fn", TokenFn},
		{"if", TokenIf},
		{"else", TokenElse},
		{"for", TokenFor},
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
		t.Run(tt.kw, func(t *testing.T) {
			toks := lexAll(tt.kw)
			if len(toks) != 1 {
				t.Fatalf("token数量: got %d, want 1", len(toks))
			}
			if toks[0].Type != tt.typ {
				t.Errorf("Type = %v, want %v", toks[0].Type, tt.typ)
			}
		})
	}
}

// Test_LexerBoundary_SimpleIdentifiers 普通标识符
func Test_LexerBoundary_SimpleIdentifiers(t *testing.T) {
	tests := []string{"x", "myVar", "_tmp", "a1", "camelCase"}
	for _, ident := range tests {
		t.Run(ident, func(t *testing.T) {
			toks := lexAll(ident)
			if len(toks) != 1 {
				t.Fatalf("token数量: got %d, want 1", len(toks))
			}
			if toks[0].Type != TokenIdent {
				t.Errorf("Type = %v, want TokenIdent", toks[0].Type)
			}
			if toks[0].Value != ident {
				t.Errorf("Value = %q, want %q", toks[0].Value, ident)
			}
		})
	}
}

// Test_LexerBoundary_UnderscoreLeadingIdentifiers 以下划线开头的标识符
func Test_LexerBoundary_UnderscoreLeadingIdentifiers(t *testing.T) {
	tests := []string{"_x", "_"}
	for _, ident := range tests {
		t.Run(ident, func(t *testing.T) {
			toks := lexAll(ident)
			if len(toks) != 1 {
				t.Fatalf("token数量: got %d, want 1", len(toks))
			}
			if toks[0].Type != TokenIdent {
				t.Errorf("Type = %v, want TokenIdent", toks[0].Type)
			}
		})
	}
}

// Test_LexerBoundary_IdentifierKeywordBoundary 标识符与关键字边界
func Test_LexerBoundary_IdentifierKeywordBoundary(t *testing.T) {
	tests := []string{"ifx", "myfor", "returnVal"}
	for _, ident := range tests {
		t.Run(ident, func(t *testing.T) {
			toks := lexAll(ident)
			if len(toks) != 1 {
				t.Fatalf("token数量: got %d, want 1", len(toks))
			}
			if toks[0].Type != TokenIdent {
				t.Errorf("%q应为TokenIdent, got %v", ident, toks[0].Type)
			}
		})
	}
}

// ========== 运算符Token测试 ==========

// Test_LexerBoundary_SingleCharOperators 单字符运算符
func Test_LexerBoundary_SingleCharOperators(t *testing.T) {
	tests := []struct {
		op  string
		typ TokenType
	}{
		{"+", TokenPlus},
		{"-", TokenMinus},
		{"*", TokenStar},
		{"/", TokenSlash},
		{"%", TokenPercent},
		{"^", TokenBitXor},
		{"&", TokenBitAnd},
		{"|", TokenBitOr},
		{"!", TokenNot},
	}
	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			toks := lexAll(tt.op)
			if len(toks) != 1 {
				t.Fatalf("token数量: got %d, want 1", len(toks))
			}
			if toks[0].Type != tt.typ {
				t.Errorf("Type = %v, want %v", toks[0].Type, tt.typ)
			}
		})
	}
}

// Test_LexerBoundary_DoubleCharOperators 双字符运算符
func Test_LexerBoundary_DoubleCharOperators(t *testing.T) {
	tests := []struct {
		op  string
		typ TokenType
	}{
		{"==", TokenEq},
		{"!=", TokenNeq},
		{"<=", TokenLe},
		{">=", TokenGe},
		{"&&", TokenAnd},
		{"||", TokenOr},
		{"<<", TokenLShift},
		{">>", TokenRShift},
	}
	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			toks := lexAll(tt.op)
			if len(toks) != 1 {
				t.Fatalf("token数量: got %d, want 1", len(toks))
			}
			if toks[0].Type != tt.typ {
				t.Errorf("Type = %v, want %v", toks[0].Type, tt.typ)
			}
		})
	}
}

// Test_LexerBoundary_TripleCharTypedAssign 三字符运算符:=>
func Test_LexerBoundary_TripleCharTypedAssign(t *testing.T) {
	t.Run(":=>", func(t *testing.T) {
		toks := lexAll(":=>")
		if len(toks) != 1 {
			t.Fatalf("token数量: got %d, want 1", len(toks))
		}
		if toks[0].Type != TokenTypedAssign {
			t.Errorf("Type = %v, want TokenTypedAssign", toks[0].Type)
		}
		if toks[0].Value != ":=>" {
			t.Errorf("Value = %q, want :=>", toks[0].Value)
		}
	})
}

// Test_LexerBoundary_AssignmentOperators 赋值运算符
func Test_LexerBoundary_AssignmentOperators(t *testing.T) {
	tests := []struct {
		op  string
		typ TokenType
	}{
		{":=", TokenAssign},
		{"+=", TokenPlusAssign},
		{"-=", TokenMinusAssign},
		{"*=", TokenStarAssign},
		{"/=", TokenSlashAssign},
	}
	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			toks := lexAll(tt.op)
			if len(toks) != 1 {
				t.Fatalf("token数量: got %d, want 1", len(toks))
			}
			if toks[0].Type != tt.typ {
				t.Errorf("Type = %v, want %v", toks[0].Type, tt.typ)
			}
		})
	}
}

// Test_LexerBoundary_ArrowOperator 箭头运算符
func Test_LexerBoundary_ArrowOperator(t *testing.T) {
	t.Run("=>", func(t *testing.T) {
		toks := lexAll("=>")
		if len(toks) != 1 {
			t.Fatalf("token数量: got %d, want 1", len(toks))
		}
		if toks[0].Type != TokenArrow {
			t.Errorf("Type = %v, want TokenArrow", toks[0].Type)
		}
		if toks[0].Value != "=>" {
			t.Errorf("Value = %q, want =>", toks[0].Value)
		}
	})
}

// Test_LexerBoundary_ColonOperator 冒号Token
func Test_LexerBoundary_ColonOperator(t *testing.T) {
	t.Run(":", func(t *testing.T) {
		toks := lexAll(":")
		if len(toks) != 1 {
			t.Fatalf("token数量: got %d, want 1", len(toks))
		}
		if toks[0].Type != TokenColon {
			t.Errorf("Type = %v, want TokenColon", toks[0].Type)
		}
	})
}

// Test_LexerBoundary_SemicolonToken 分号Token
func Test_LexerBoundary_SemicolonToken(t *testing.T) {
	t.Run(";", func(t *testing.T) {
		toks := lexAll(";")
		if len(toks) != 1 {
			t.Fatalf("token数量: got %d, want 1", len(toks))
		}
		if toks[0].Type != TokenSemicolon {
			t.Errorf("Type = %v, want TokenSemicolon", toks[0].Type)
		}
	})
}

// ========== 运算符边界测试 ==========

// Test_LexerBoundary_LtVsLe 单字符与双字符边界: < vs <=
func Test_LexerBoundary_LtVsLe(t *testing.T) {
	t.Run("lt_only", func(t *testing.T) {
		toks := lexAll("<")
		assertTypes(t, toks, []TokenType{TokenLt})
	})
	t.Run("le", func(t *testing.T) {
		toks := lexAll("<=")
		assertTypes(t, toks, []TokenType{TokenLe})
	})
}

// Test_LexerBoundary_GtVsGe 单字符与双字符边界: > vs >=
func Test_LexerBoundary_GtVsGe(t *testing.T) {
	t.Run("gt_only", func(t *testing.T) {
		toks := lexAll(">")
		assertTypes(t, toks, []TokenType{TokenGt})
	})
	t.Run("ge", func(t *testing.T) {
		toks := lexAll(">=")
		assertTypes(t, toks, []TokenType{TokenGe})
	})
}

// Test_LexerBoundary_EqVsAssign = vs ==
func Test_LexerBoundary_EqVsAssign(t *testing.T) {
	t.Run("assign_single", func(t *testing.T) {
		toks := lexAll("=")
		// 单个=产生TokenAssign(值="=")
		assertTypes(t, toks, []TokenType{TokenAssign})
	})
	t.Run("eq_double", func(t *testing.T) {
		toks := lexAll("==")
		assertTypes(t, toks, []TokenType{TokenEq})
	})
}

// Test_LexerBoundary_NotVsNeq ! vs !=
func Test_LexerBoundary_NotVsNeq(t *testing.T) {
	t.Run("not_only", func(t *testing.T) {
		toks := lexAll("!")
		assertTypes(t, toks, []TokenType{TokenNot})
	})
	t.Run("neq", func(t *testing.T) {
		toks := lexAll("!=")
		assertTypes(t, toks, []TokenType{TokenNeq})
	})
}

// Test_LexerBoundary_BitAndVsLogicalAnd & vs &&
func Test_LexerBoundary_BitAndVsLogicalAnd(t *testing.T) {
	t.Run("bitand", func(t *testing.T) {
		toks := lexAll("&")
		assertTypes(t, toks, []TokenType{TokenBitAnd})
	})
	t.Run("logical_and", func(t *testing.T) {
		toks := lexAll("&&")
		assertTypes(t, toks, []TokenType{TokenAnd})
	})
}

// Test_LexerBoundary_BitOrVsLogicalOr | vs ||
func Test_LexerBoundary_BitOrVsLogicalOr(t *testing.T) {
	t.Run("bitor", func(t *testing.T) {
		toks := lexAll("|")
		assertTypes(t, toks, []TokenType{TokenBitOr})
	})
	t.Run("logical_or", func(t *testing.T) {
		toks := lexAll("||")
		assertTypes(t, toks, []TokenType{TokenOr})
	})
}

// Test_LexerBoundary_ShiftVsLessThan < vs <<
func Test_LexerBoundary_ShiftVsLessThan(t *testing.T) {
	t.Run("lt", func(t *testing.T) {
		toks := lexAll("<")
		assertTypes(t, toks, []TokenType{TokenLt})
	})
	t.Run("lshift", func(t *testing.T) {
		toks := lexAll("<<")
		assertTypes(t, toks, []TokenType{TokenLShift})
	})
}

// Test_LexerBoundary_ShiftVsGreaterThan > vs >>
func Test_LexerBoundary_ShiftVsGreaterThan(t *testing.T) {
	t.Run("gt", func(t *testing.T) {
		toks := lexAll(">")
		assertTypes(t, toks, []TokenType{TokenGt})
	})
	t.Run("rshift", func(t *testing.T) {
		toks := lexAll(">>")
		assertTypes(t, toks, []TokenType{TokenRShift})
	})
}

// ========== 分隔符Token测试 ==========

// Test_LexerBoundary_Delimiters 分隔符
func Test_LexerBoundary_Delimiters(t *testing.T) {
	tests := []struct {
		delim string
		typ   TokenType
	}{
		{"(", TokenLParen},
		{")", TokenRParen},
		{"{", TokenLBrace},
		{"}", TokenRBrace},
		{"[", TokenLBracket},
		{"]", TokenRBracket},
		{",", TokenComma},
	}
	for _, tt := range tests {
		t.Run(tt.delim, func(t *testing.T) {
			toks := lexAll(tt.delim)
			if len(toks) != 1 {
				t.Fatalf("token数量: got %d, want 1", len(toks))
			}
			if toks[0].Type != tt.typ {
				t.Errorf("Type = %v, want %v", toks[0].Type, tt.typ)
			}
		})
	}
}

// ========== 注释Token测试 ==========

// Test_LexerBoundary_HashSingleLineComment 单行注释#(NextToken跳过)
func Test_LexerBoundary_HashSingleLineComment(t *testing.T) {
	t.Run("hash_comment_skipped", func(t *testing.T) {
		// NextToken跳过注释, 直接返回EOF
		toks := lexAll("# this is a comment")
		if len(toks) != 0 {
			t.Errorf("纯注释应返回0个token, got %d", len(toks))
		}
	})
}

// Test_LexerBoundary_DoubleSlashComment 单行注释//(NextToken跳过)
func Test_LexerBoundary_DoubleSlashComment(t *testing.T) {
	t.Run("slash_comment_skipped", func(t *testing.T) {
		toks := lexAll("// this is a comment")
		if len(toks) != 0 {
			t.Errorf("纯注释应返回0个token, got %d", len(toks))
		}
	})
}

// Test_LexerBoundary_MultilineComment 多行注释(NextToken跳过)
func Test_LexerBoundary_MultilineComment(t *testing.T) {
	t.Run("multiline_comment_skipped", func(t *testing.T) {
		toks := lexAll("/* comment */")
		if len(toks) != 0 {
			t.Errorf("纯注释应返回0个token, got %d", len(toks))
		}
	})
}

// Test_LexerBoundary_MultilineCommentCrossLine 多行注释跨行
func Test_LexerBoundary_MultilineCommentCrossLine(t *testing.T) {
	t.Run("cross_line", func(t *testing.T) {
		input := "/* line1\nline2\nline3 */ x := 1"
		toks := lexAll(input)
		// 注释跨3行, 之后是x := 1
		assertTypes(t, toks, []TokenType{TokenIdent, TokenAssign, TokenInt})
		// 行号应为第3行
		if toks[0].Line != 3 {
			t.Errorf("注释后首个token行号应为3, got %d", toks[0].Line)
		}
	})
}

// Test_LexerBoundary_UnclosedMultilineComment 未闭合多行注释产生错误
func Test_LexerBoundary_UnclosedMultilineComment(t *testing.T) {
	t.Run("unclosed", func(t *testing.T) {
		toks := lexAll("/* unclosed comment")
		if len(toks) != 1 {
			t.Fatalf("token数量: got %d, want 1", len(toks))
		}
		if toks[0].Type != TokenError {
			t.Errorf("Type = %v, want TokenError", toks[0].Type)
		}
	})
}

// Test_LexerBoundary_CommentThenCode 注释后跟代码
func Test_LexerBoundary_CommentThenCode(t *testing.T) {
	t.Run("hash_then_code", func(t *testing.T) {
		toks := lexAll("# comment\n42")
		assertTypes(t, toks, []TokenType{TokenInt})
		if toks[0].Value != "42" {
			t.Errorf("Value = %q, want \"42\"", toks[0].Value)
		}
	})
	t.Run("slash_then_code", func(t *testing.T) {
		toks := lexAll("// comment\n42")
		assertTypes(t, toks, []TokenType{TokenInt})
	})
}

// Test_LexerBoundary_CodeWithInterleavedComments 代码中穿插注释
func Test_LexerBoundary_CodeWithInterleavedComments(t *testing.T) {
	t.Run("interleaved", func(t *testing.T) {
		input := "x := 1 // inline\ny := 2"
		toks := lexAll(input)
		assertTypes(t, toks, []TokenType{
			TokenIdent, TokenAssign, TokenInt,
			TokenIdent, TokenAssign, TokenInt,
		})
	})
}

// ========== #fn指令测试 ==========

// Test_LexerBoundary_DefSimple 简单#fn指令
func Test_LexerBoundary_DefSimple(t *testing.T) {
	t.Run("simple_def", func(t *testing.T) {
		toks := lexAll("#fn name()")
		if len(toks) != 1 {
			t.Fatalf("token数量: got %d, want 1", len(toks))
		}
		if toks[0].Type != TokenDef {
			t.Errorf("Type = %v, want TokenDef", toks[0].Type)
		}
		if toks[0].Value != "#fn name()" {
			t.Errorf("Value = %q, want \"#fn name()\"", toks[0].Value)
		}
	})
}

// Test_LexerBoundary_DefWithReturn 带返回类型的#fn指令
func Test_LexerBoundary_DefWithReturn(t *testing.T) {
	t.Run("def_with_return", func(t *testing.T) {
		input := "#fn name(param)=>return"
		toks := lexAll(input)
		if len(toks) != 1 {
			t.Fatalf("token数量: got %d, want 1", len(toks))
		}
		if toks[0].Type != TokenDef {
			t.Errorf("Type = %v, want TokenDef", toks[0].Type)
		}
		if toks[0].Value != input {
			t.Errorf("Value = %q, want %q", toks[0].Value, input)
		}
	})
}

// Test_LexerBoundary_DefMultipleParams 多参数#fn指令
func Test_LexerBoundary_DefMultipleParams(t *testing.T) {
	t.Run("def_multi_params", func(t *testing.T) {
		input := "#fn name(param1, param2)=>return"
		toks := lexAll(input)
		if len(toks) != 1 {
			t.Fatalf("token数量: got %d, want 1", len(toks))
		}
		if toks[0].Type != TokenDef {
			t.Errorf("Type = %v, want TokenDef", toks[0].Type)
		}
		if toks[0].Value != input {
			t.Errorf("Value = %q, want %q", toks[0].Value, input)
		}
	})
}

// ========== 位置信息测试 ==========

// Test_LexerBoundary_SingleLinePosition 单行Token的Line和Column
func Test_LexerBoundary_SingleLinePosition(t *testing.T) {
	t.Run("position_in_line", func(t *testing.T) {
		toks := lexAll("x := 1")
		// x在Col=1, :=在Col=3, 1在Col=6
		if toks[0].Line != 1 || toks[0].Column != 1 {
			t.Errorf("x: Line=%d Col=%d, want Line=1 Col=1", toks[0].Line, toks[0].Column)
		}
		if toks[1].Line != 1 || toks[1].Column != 3 {
			t.Errorf(":=: Line=%d Col=%d, want Line=1 Col=3", toks[1].Line, toks[1].Column)
		}
		if toks[2].Line != 1 || toks[2].Column != 6 {
			t.Errorf("1: Line=%d Col=%d, want Line=1 Col=6", toks[2].Line, toks[2].Column)
		}
	})
}

// Test_LexerBoundary_MultilinePosition 多行输入行号递增
func Test_LexerBoundary_MultilinePosition(t *testing.T) {
	t.Run("line_increment", func(t *testing.T) {
		toks := lexAll("x := 1\ny := 2")
		// x在第1行, y在第2行
		if toks[0].Line != 1 {
			t.Errorf("x的行号应为1, got %d", toks[0].Line)
		}
		// y是第4个token(x, :=, 1, y)
		if toks[3].Line != 2 {
			t.Errorf("y的行号应为2, got %d", toks[3].Line)
		}
	})
}

// Test_LexerBoundary_TabColumnCalculation Tab的Column计算
func Test_LexerBoundary_TabColumnCalculation(t *testing.T) {
	t.Run("tab_indent", func(t *testing.T) {
		// Tab算1列
		toks := lexAll("\tx")
		if len(toks) != 1 {
			t.Fatalf("token数量: got %d, want 1", len(toks))
		}
		if toks[0].Column != 2 {
			t.Errorf("tab后x的列号应为2, got %d", toks[0].Column)
		}
	})
}

// Test_LexerBoundary_ColumnAfterNewline 换行后Column重置
func Test_LexerBoundary_ColumnAfterNewline(t *testing.T) {
	t.Run("column_reset", func(t *testing.T) {
		toks := lexAll("ab\ncd")
		// ab在Col=1, cd在Col=1(换行后重置)
		if toks[0].Column != 1 {
			t.Errorf("ab的列号应为1, got %d", toks[0].Column)
		}
		if toks[1].Column != 1 {
			t.Errorf("cd的列号应为1, got %d", toks[1].Column)
		}
	})
}

// ========== Token序列完整性测试 ==========

// Test_LexerBoundary_ComplexExpressionSequence 复杂表达式的完整Token序列
func Test_LexerBoundary_ComplexExpressionSequence(t *testing.T) {
	t.Run("complex_expr", func(t *testing.T) {
		input := `x := 10 + 20.5 * "hello"`
		toks := lexAll(input)
		expected := []struct {
			typ TokenType
			val string
		}{
			{TokenIdent, "x"},
			{TokenAssign, ":="},
			{TokenInt, "10"},
			{TokenPlus, "+"},
			{TokenFloat, "20.5"},
			{TokenStar, "*"},
			{TokenString, "hello"},
		}
		if len(toks) != len(expected) {
			t.Fatalf("token数量: got %d, want %d", len(toks), len(expected))
		}
		for i, want := range expected {
			if toks[i].Type != want.typ {
				t.Errorf("token[%d].Type = %v, want %v", i, toks[i].Type, want.typ)
			}
			if toks[i].Value != want.val {
				t.Errorf("token[%d].Value = %q, want %q", i, toks[i].Value, want.val)
			}
		}
	})
}

// Test_LexerBoundary_StatementSequence 语句序列的Token序列
func Test_LexerBoundary_StatementSequence(t *testing.T) {
	t.Run("statements", func(t *testing.T) {
		input := "if x > 1 { return x }"
		toks := lexAll(input)
		assertTypes(t, toks, []TokenType{
			TokenIf, TokenIdent, TokenGt, TokenInt,
			TokenLBrace, TokenReturn, TokenIdent, TokenRBrace,
		})
	})
}

// Test_LexerBoundary_TokenizeSlice Tokenize返回切片正确性
func Test_LexerBoundary_TokenizeSlice(t *testing.T) {
	t.Run("tokenize_result", func(t *testing.T) {
		l := NewLexer("a + b")
		tokens, err := l.Tokenize()
		assertNoError(t, err)
		if len(tokens) != 3 {
			t.Fatalf("token数量: got %d, want 3", len(tokens))
		}
		// 验证切片不包含EOF
		for i, tok := range tokens {
			if tok.Type == TokenEOF {
				t.Errorf("Tokenize结果[%d]不应包含EOF", i)
			}
		}
	})
}

// ========== 边界输入测试 ==========

// Test_LexerBoundary_EmptyInput 空字符串输入
func Test_LexerBoundary_EmptyInput(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		toks := lexAll("")
		if len(toks) != 0 {
			t.Errorf("空输入应返回0个token, got %d", len(toks))
		}
		// NextToken应直接返回EOF
		l := NewLexer("")
		tok := l.NextToken()
		if tok.Type != TokenEOF {
			t.Errorf("空输入NextToken应为EOF, got %v", tok.Type)
		}
	})
}

// Test_LexerBoundary_WhitespaceOnly 只有空白
func Test_LexerBoundary_WhitespaceOnly(t *testing.T) {
	inputs := []string{" ", "\t", "   ", "\r\r"}
	for _, input := range inputs {
		t.Run(fmt.Sprintf("%q", input), func(t *testing.T) {
			toks := lexAll(input)
			if len(toks) != 0 {
				t.Errorf("纯空白应返回0个token, got %d", len(toks))
			}
		})
	}
}

// Test_LexerBoundary_NewlineOnly 只有换行
func Test_LexerBoundary_NewlineOnly(t *testing.T) {
	t.Run("newline", func(t *testing.T) {
		toks := lexAll("\n\n\n")
		if len(toks) != 0 {
			t.Errorf("纯换行应返回0个token, got %d", len(toks))
		}
	})
}

// Test_LexerBoundary_CommentOnlyInput 只有注释
func Test_LexerBoundary_CommentOnlyInput(t *testing.T) {
	t.Run("hash_only", func(t *testing.T) {
		toks := lexAll("# only comment")
		if len(toks) != 0 {
			t.Errorf("纯注释应返回0个token, got %d", len(toks))
		}
	})
	t.Run("slash_only", func(t *testing.T) {
		toks := lexAll("// only comment")
		if len(toks) != 0 {
			t.Errorf("纯注释应返回0个token, got %d", len(toks))
		}
	})
	t.Run("multiline_only", func(t *testing.T) {
		toks := lexAll("/* only comment */")
		if len(toks) != 0 {
			t.Errorf("纯注释应返回0个token, got %d", len(toks))
		}
	})
}

// Test_LexerBoundary_MixedWhitespaceAndComments 混合空白和注释
func Test_LexerBoundary_MixedWhitespaceAndComments(t *testing.T) {
	t.Run("mixed", func(t *testing.T) {
		input := "  \n  # comment\n  \t  // another\n  /* block */  \n  "
		toks := lexAll(input)
		if len(toks) != 0 {
			t.Errorf("混合空白和注释应返回0个token, got %d", len(toks))
		}
	})
}

// Test_LexerBoundary_ConsecutiveSameOperators 连续相同运算符(引擎行为)
func Test_LexerBoundary_ConsecutiveSameOperators(t *testing.T) {
	t.Run("plus_plus", func(t *testing.T) {
		// ++产生两个独立的TokenPlus
		toks := lexAll("++")
		assertTypes(t, toks, []TokenType{TokenPlus, TokenPlus})
	})
	t.Run("minus_minus", func(t *testing.T) {
		// --产生两个独立的TokenMinus
		toks := lexAll("--")
		assertTypes(t, toks, []TokenType{TokenMinus, TokenMinus})
	})
}

// Test_LexerBoundary_ConsecutiveEqualSigns 连续等号(引擎行为)
func Test_LexerBoundary_ConsecutiveEqualSigns(t *testing.T) {
	t.Run("triple_equal", func(t *testing.T) {
		// ===产生TokenEq(==) + TokenAssign(=)
		toks := lexAll("===")
		assertTypes(t, toks, []TokenType{TokenEq, TokenAssign})
	})
}

// Test_LexerBoundary_TokenizeErrorPropagation Tokenize遇到错误立即返回
func Test_LexerBoundary_TokenizeErrorPropagation(t *testing.T) {
	t.Run("unclosed_string_tokenize", func(t *testing.T) {
		l := NewLexer(`"unclosed`)
		_, err := l.Tokenize()
		if err == nil {
			t.Error("未闭合字符串Tokenize应返回错误")
		}
	})
	t.Run("unclosed_comment_tokenize", func(t *testing.T) {
		l := NewLexer("/* unclosed")
		_, err := l.Tokenize()
		if err == nil {
			t.Error("未闭合多行注释Tokenize应返回错误")
		}
	})
	t.Run("invalid_char_tokenize", func(t *testing.T) {
		l := NewLexer("@")
		_, err := l.Tokenize()
		if err == nil {
			t.Error("无效字符Tokenize应返回错误")
		}
	})
}

// Test_LexerBoundary_ColonChain 冒号链式: : := :=> 混合
func Test_LexerBoundary_ColonChain(t *testing.T) {
	t.Run("colon_mix", func(t *testing.T) {
		toks := lexAll(": := :=>")
		assertTypes(t, toks, []TokenType{TokenColon, TokenAssign, TokenTypedAssign})
	})
}

// Test_LexerBoundary_EOFPosition EOF的Line和Column
func Test_LexerBoundary_EOFPosition(t *testing.T) {
	t.Run("eof_position", func(t *testing.T) {
		l := NewLexer("ab")
		l.NextToken() // 读取ab
		eof := l.NextToken()
		if eof.Type != TokenEOF {
			t.Fatalf("应为EOF, got %v", eof.Type)
		}
		if eof.Line != 1 {
			t.Errorf("EOF的行号应为1, got %d", eof.Line)
		}
	})
}
