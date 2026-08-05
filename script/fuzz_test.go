package script

import (
	"testing"
	"time"
)

// Fuzz_Lexer 词法分析器 fuzzing, 任意输入不 panic
func Fuzz_Lexer(f *testing.F) {
	seeds := []string{
		"",
		"hello",
		"123 + 456",
		"fn f(a, b) -> int { return a + b }",
		"if true { 1 } else { 2 }",
		"for i := 0; i < 10; i = i + 1 { i }",
		"arr := [1, 2, 3]; arr[0]",
		`s := "hello"; s[0:3]`,
		"#fn test()=>int",
		"你好世界",
		"1 + 2 * 3 - 4 / 5 % 6",
		"x :=>int 42",
		"throw \"error\"",
		"{}[]()",
		"!!!???",
		"\n\t\r",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		lexer := NewLexer(input)
		for {
			tok := lexer.NextToken()
			if tok.Type == TokenEOF || tok.Type == TokenError {
				break
			}
		}
	})
}

// Fuzz_Parser 解析器 fuzzing, 编译失败可接受, panic 不行
func Fuzz_Parser(f *testing.F) {
	seeds := []string{
		"",
		"x := 1",
		"fn f() { }",
		"if { }",
		"for { }",
		"1 +",
		"arr[",
		"fn",
		"for for for",
		"{}{}{}",
		"a b c d e",
		"(((((((((())))))))))",
		"x :=>int 1\ny :=>string \"hello\"",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		parser := NewParser()
		defer returnParserToPool(parser)
		_, _ = parser.Compile(input)
	})
}

// Fuzz_Eval 完整执行 fuzzing, 带资源限制防止死循环
func Fuzz_Eval(f *testing.F) {
	seeds := []string{
		"1 + 2",
		"x := 10\nx * 2",
		"if true { 42 }",
		"fn add(a, b) { return a + b }\nadd(1, 2)",
		"for i := 0; i < 3; i = i + 1 { i }",
		"[1, 2, 3]",
		`{"a": 1, "b": 2}`,
		`"hello"[0:3]`,
		"1 == 1.0",
		"nil",
		"true && false || true",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		engine := NewEngine(WithMaxSteps(10000), WithTimeout(100*time.Millisecond))
		ctx := NewContext()
		parser := NewParser()
		defer returnParserToPool(parser)
		script, err := parser.Compile(input)
		if err != nil {
			return
		}
		_, _ = engine.Run(ctx, script)
	})
}

// Fuzz_StackOps 栈操作 fuzzing, 关注可能导致栈不平衡的输入
func Fuzz_StackOps(f *testing.F) {
	seeds := []string{
		"1 + 2 * 3",
		"f := fn(x) { return x }\nf(f(f(1)))",
		"arr := [1]; arr[0] = arr[0]",
		"x := 1; x += 2; x -= 1; x",
		"if 1 > 0 { if 2 > 1 { 42 } }",
		"a := [1,2]; b := [3,4]; a[0] + b[0]",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		engine := NewEngine(WithMaxSteps(10000), WithTimeout(100*time.Millisecond))
		ctx := NewContext()
		parser := NewParser()
		defer returnParserToPool(parser)
		script, err := parser.Compile(input)
		if err != nil {
			return
		}
		_, _ = engine.Run(ctx, script)
	})
}
