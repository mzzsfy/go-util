package script

import (
	"testing"
)

// ========== Lexer边界测试 ==========

func Test_Lexer_HexNumbers(t *testing.T) {
	tests := []struct {
		input string
		toks  []TokenType
	}{
		{"0xFF", []TokenType{TokenInt}},
		{"0XAB", []TokenType{TokenInt}},
		{"0xabcdef", []TokenType{TokenInt}},
		{"0x10 + 0x20", []TokenType{TokenInt, TokenPlus, TokenInt}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			tokens, err := l.Tokenize()
			assertNoError(t, err)
			if len(tokens) != len(tt.toks) {
				t.Fatalf("token数量: got %d, want %d", len(tokens), len(tt.toks))
			}
			for i, want := range tt.toks {
				if tokens[i].Type != want {
					t.Errorf("token[%d].Type = %v, want %v", i, tokens[i].Type, want)
				}
			}
		})
	}
}

func Test_Lexer_NumberUnderscore(t *testing.T) {
	inputs := []string{
		"1_000",
		"1_000_000",
		"3.14_15",
		"0xFF_FF",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			l := NewLexer(input)
			tokens, err := l.Tokenize()
			assertNoError(t, err)
			if len(tokens) != 1 {
				t.Errorf("应解析为单个token, got %d", len(tokens))
			}
		})
	}
}

func Test_Lexer_FloatEdgeCases(t *testing.T) {
	tests := []struct {
		input string
		typ   TokenType
	}{
		{"0.0", TokenFloat},
		{"0.5", TokenFloat},
		{"123.456", TokenFloat},
		{"0", TokenInt},
		{"123", TokenInt},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			tokens, err := l.Tokenize()
			assertNoError(t, err)
			if len(tokens) != 1 {
				t.Fatalf("got %d tokens, want 1", len(tokens))
			}
			if tokens[0].Type != tt.typ {
				t.Errorf("token type = %v, want %v", tokens[0].Type, tt.typ)
			}
		})
	}
}

func Test_Lexer_StringEscapes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"\n"`, "\n"},
		{`"\t"`, "\t"},
		{`"\r"`, "\r"},
		{`"\\"`, "\\"},
		{`"\""`, "\""},
		{`"hello\nworld"`, "hello\nworld"},
		{`"tab\there"`, "tab\there"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := NewLexer(tt.input)
			tokens, err := l.Tokenize()
			assertNoError(t, err)
			if len(tokens) != 1 {
				t.Fatalf("got %d tokens, want 1", len(tokens))
			}
			if tokens[0].Value != tt.expected {
				t.Errorf("value = %q, want %q", tokens[0].Value, tt.expected)
			}
		})
	}
}

func Test_Lexer_UnknownEscape(t *testing.T) {
	// 未识别的转义序列应保留原样
	l := NewLexer(`"\x"`)
	tokens, err := l.Tokenize()
	assertNoError(t, err)
	if len(tokens) != 1 {
		t.Fatalf("got %d tokens, want 1", len(tokens))
	}
	// \x 不是有效转义，应保留为 \x
	if tokens[0].Value != "\\x" {
		t.Errorf("未知转义应保留原样, got %q", tokens[0].Value)
	}
}

func Test_Lexer_MultilineComment(t *testing.T) {
	input := `/* comment */ x := 1`
	l := NewLexer(input)
	tokens, err := l.Tokenize()
	assertNoError(t, err)
	// 注释应被过滤，只留下 x := 1
	if len(tokens) != 3 {
		t.Errorf("注释后应剩3个token, got %d", len(tokens))
	}
}

func Test_Lexer_NestedMultilineComment(t *testing.T) {
	// 多行注释中包含换行
	input := "/* line1\nline2\nline3 */ x := 1"
	l := NewLexer(input)
	tokens, err := l.Tokenize()
	assertNoError(t, err)
	if len(tokens) != 3 {
		t.Errorf("多行注释后应剩3个token, got %d", len(tokens))
	}
	// 行号应正确（注释跨3行，token在第3行）
	if tokens[0].Line != 3 {
		t.Errorf("注释后token行号应为3, got %d", tokens[0].Line)
	}
}

func Test_Lexer_UnclosedMultilineComment(t *testing.T) {
	l := NewLexer("/* unclosed comment")
	_, err := l.Tokenize()
	if err == nil {
		t.Error("未闭合的多行注释应返回错误")
	}
}

func Test_Lexer_UnclosedString(t *testing.T) {
	l := NewLexer(`"unclosed string`)
	_, err := l.Tokenize()
	if err == nil {
		t.Error("未闭合的字符串应返回错误")
	}
}

func Test_Lexer_AllOperators(t *testing.T) {
	tests := []struct {
		op  string
		typ TokenType
	}{
		{"+", TokenPlus},
		{"-", TokenMinus},
		{"*", TokenStar},
		{"/", TokenSlash},
		{"%", TokenPercent},
		{"==", TokenEq},
		{"!=", TokenNeq},
		{"<", TokenLt},
		{">", TokenGt},
		{"<=", TokenLe},
		{">=", TokenGe},
		{"&&", TokenAnd},
		{"||", TokenOr},
		{"!", TokenNot},
		{"&", TokenBitAnd},
		{"|", TokenBitOr},
		{"^", TokenBitXor},
		{"<<", TokenLShift},
		{">>", TokenRShift},
		{":=", TokenAssign},
		{"+=", TokenPlusAssign},
		{"-=", TokenMinusAssign},
		{"*=", TokenStarAssign},
		{"/=", TokenSlashAssign},
		{"=>", TokenArrow},
		{":", TokenColon},
		{":=>", TokenTypedAssign},
	}

	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			l := NewLexer(tt.op)
			tokens, err := l.Tokenize()
			assertNoError(t, err)
			if len(tokens) != 1 {
				t.Fatalf("got %d tokens, want 1", len(tokens))
			}
			if tokens[0].Type != tt.typ {
				t.Errorf("type = %v, want %v", tokens[0].Type, tt.typ)
			}
		})
	}
}

func Test_Lexer_AllDelimiters(t *testing.T) {
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
		{";", TokenSemicolon},
	}

	for _, tt := range tests {
		t.Run(tt.delim, func(t *testing.T) {
			l := NewLexer(tt.delim)
			tokens, err := l.Tokenize()
			assertNoError(t, err)
			if len(tokens) != 1 {
				t.Fatalf("got %d tokens, want 1", len(tokens))
			}
			if tokens[0].Type != tt.typ {
				t.Errorf("type = %v, want %v", tokens[0].Type, tt.typ)
			}
		})
	}
}

func Test_Lexer_AllKeywords(t *testing.T) {
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
			l := NewLexer(tt.kw)
			tokens, err := l.Tokenize()
			assertNoError(t, err)
			if len(tokens) != 1 {
				t.Fatalf("got %d tokens, want 1", len(tokens))
			}
			if tokens[0].Type != tt.typ {
				t.Errorf("type = %v, want %v", tokens[0].Type, tt.typ)
			}
		})
	}
}

func Test_Lexer_IdentifierVsKeyword(t *testing.T) {
	// 包含关键字子串的标识符不应被误判
	tests := []string{
		"myFunc",
		"ifElse",
		"forLoop",
		"returnVal",
		"breakPoint",
		"continueLoop",
		"throwErr",
		"null",
		"truely",
		"falsey",
		"function",
	}

	for _, ident := range tests {
		t.Run(ident, func(t *testing.T) {
			l := NewLexer(ident)
			tokens, err := l.Tokenize()
			assertNoError(t, err)
			if len(tokens) != 1 {
				t.Fatalf("got %d tokens, want 1", len(tokens))
			}
			if tokens[0].Type != TokenIdent {
				t.Errorf("%q应是TokenIdent, got %v", ident, tokens[0].Type)
			}
		})
	}
}

func Test_Lexer_SingleLineComment(t *testing.T) {
	inputs := []string{
		"# comment\nx := 1",
		"// comment\nx := 1",
	}

	for _, input := range inputs {
		t.Run(input[:10], func(t *testing.T) {
			l := NewLexer(input)
			tokens, err := l.Tokenize()
			assertNoError(t, err)
			if len(tokens) != 3 {
				t.Errorf("注释后应剩3个token, got %d", len(tokens))
			}
		})
	}
}

func Test_Lexer_InvalidChar(t *testing.T) {
	invalids := []string{"@", "$", "?", "`", "~"}

	for _, ch := range invalids {
		t.Run(ch, func(t *testing.T) {
			l := NewLexer(ch)
			_, err := l.Tokenize()
			if err == nil {
				t.Errorf("字符 %q 应产生错误", ch)
			}
		})
	}
}

func Test_Lexer_EmptyInput(t *testing.T) {
	l := NewLexer("")
	tokens, err := l.Tokenize()
	assertNoError(t, err)
	if len(tokens) != 0 {
		t.Errorf("空输入应返回0个token, got %d", len(tokens))
	}
}

func Test_Lexer_WhitespaceOnly(t *testing.T) {
	inputs := []string{" ", "\t", "\n", "\r\n", "   \n\t  "}
	for _, input := range inputs {
		l := NewLexer(input)
		tokens, err := l.Tokenize()
		assertNoError(t, err)
		if len(tokens) != 0 {
			t.Errorf("纯空白输入 %q 应返回0个token, got %d", input, len(tokens))
		}
	}
}

func Test_Lexer_Directive(t *testing.T) {
	input := "#fn add(int, int) => int"
	l := NewLexer(input)
	tokens, err := l.Tokenize()
	assertNoError(t, err)
	if len(tokens) != 1 {
		t.Fatalf("got %d tokens, want 1", len(tokens))
	}
	if tokens[0].Type != TokenDef {
		t.Errorf("type = %v, want TokenDef", tokens[0].Type)
	}
}

func Test_Lexer_MixedTokens(t *testing.T) {
	input := `x := 10 + 20.5 * "hello"`
	l := NewLexer(input)
	tokens, err := l.Tokenize()
	assertNoError(t, err)

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

	if len(tokens) != len(expected) {
		t.Fatalf("got %d tokens, want %d", len(tokens), len(expected))
	}

	for i, want := range expected {
		if tokens[i].Type != want.typ {
			t.Errorf("token[%d].Type = %v, want %v", i, tokens[i].Type, want.typ)
		}
		if tokens[i].Value != want.val {
			t.Errorf("token[%d].Value = %q, want %q", i, tokens[i].Value, want.val)
		}
	}
}

func Test_Lexer_TokenPosition(t *testing.T) {
	input := "x := 1\ny := 2"
	l := NewLexer(input)
	tokens, err := l.Tokenize()
	assertNoError(t, err)

	// x在第1行
	if tokens[0].Line != 1 {
		t.Errorf("x的行号应为1, got %d", tokens[0].Line)
	}
	// y在第2行
	yIdx := 3 // x, :=, 1, y
	if tokens[yIdx].Line != 2 {
		t.Errorf("y的行号应为2, got %d", tokens[yIdx].Line)
	}
}
