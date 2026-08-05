package script

import (
	"testing"
)

// ========== Compiler指令生成验证测试 ==========

func Test_Compiler_GeneratesConst(t *testing.T) {
	script := compileScript(t, "42")
	hasConst := false
	for _, inst := range script.Main.Instructions {
		if inst.Op == OpConst {
			hasConst = true
			break
		}
	}
	if !hasConst {
		t.Error("字面量42应生成OpConst指令")
	}
}

func Test_Compiler_GeneratesAdd(t *testing.T) {
	script := compileScript(t, "1 + 2")
	hasAdd := false
	for _, inst := range script.Main.Instructions {
		if inst.Op == OpAdd {
			hasAdd = true
			break
		}
	}
	if !hasAdd {
		t.Error("1+2应生成OpAdd指令")
	}
}

func Test_Compiler_GeneratesJump(t *testing.T) {
	script := compileScript(t, "if true { 1 }")
	hasJump := false
	for _, inst := range script.Main.Instructions {
		if inst.Op == OpJumpIfFalse || inst.Op == OpJump {
			hasJump = true
			break
		}
	}
	if !hasJump {
		t.Error("if应生成跳转指令")
	}
}

func Test_Compiler_GeneratesReturn(t *testing.T) {
	script := compileScript(t, `
		fn test() { return 1 }
	`)
	// 函数声明应生成指令
	if len(script.Main.Instructions) == 0 {
		t.Error("函数声明应生成指令")
	}
}

func Test_Compiler_ConstantDedup(t *testing.T) {
	script := compileScript(t, "1 + 1 + 1")
	// 常量1应只出现一次（去重）
	count := 0
	for _, c := range script.Main.Constants {
		if c.Type == TypeInt && c.Int() == 1 {
			count++
		}
	}
	if count != 1 {
		t.Errorf("常量1应去重为1次, got %d", count)
	}
}

func Test_Compiler_ArrayNew(t *testing.T) {
	script := compileScript(t, "[1, 2, 3]")
	hasArrayNew := false
	for _, inst := range script.Main.Instructions {
		if inst.Op == OpArrayNew {
			hasArrayNew = true
			break
		}
	}
	if !hasArrayNew {
		t.Error("[1,2,3]应生成OpArrayNew指令")
	}
}

func Test_Compiler_MapNew(t *testing.T) {
	script := compileScript(t, `{"a": 1}`)
	hasMapNew := false
	for _, inst := range script.Main.Instructions {
		if inst.Op == OpMapNew {
			hasMapNew = true
			break
		}
	}
	if !hasMapNew {
		t.Error(`{"a":1}应生成OpMapNew指令`)
	}
}

// ========== Compiler错误路径测试 ==========

func Test_Compiler_EmptyProgram(t *testing.T) {
	script := compileScript(t, "")
	// 空程序应该能正常编译
	if script == nil {
		t.Error("空程序编译结果不应为nil")
	}
}

func Test_Compiler_OnlyComments(t *testing.T) {
	// 只有注释的程序等同于空程序，应该能正常编译
	script := compileScript(t, "# this is a comment")
	if script == nil {
		t.Error("只有注释的程序编译结果不应为nil")
	}
}

func Test_Compiler_DeepNesting(t *testing.T) {
	// 深度嵌套表达式
	compileScript(t, "(((((1 + 2)))))")
}

func Test_Compiler_LongExpression(t *testing.T) {
	compileScript(t, "1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 9 + 10")
}

func Test_Compiler_LargeArray(t *testing.T) {
	input := "["
	for i := 0; i < 100; i++ {
		if i > 0 {
			input += ", "
		}
		input += "0"
	}
	input += "]"
	compileScript(t, input)
}

// ========== Script执行流程测试 ==========

func Test_Run_VoidExpression(t *testing.T) {
	// 不产生值的表达式
	result := runScript(t, `x := 1`)
	// 结果可能是nil或最后一个表达式的值
	_ = result
}

func Test_Run_LastExpressionReturn(t *testing.T) {
	// 脚本最后一个表达式作为返回值
	runIntTest(t, "1", 1)
	runIntTest(t, "1\n2", 2)
	runIntTest(t, "x := 10\nx", 10)
}

func Test_Run_StringResult(t *testing.T) {
	runStringTest(t, `"hello"`, "hello")
	runStringTest(t, `"a" + "b"`, "ab")
}

func Test_Run_BoolResult(t *testing.T) {
	runBoolTest(t, "true", true)
	runBoolTest(t, "false", false)
	runBoolTest(t, "1 == 1", true)
}

// ========== 混合场景测试 ==========

func Test_Mixed_ArithmeticAndComparison(t *testing.T) {
	runBoolTest(t, "(1 + 2) * 3 == 9", true)
	runBoolTest(t, "2 * (3 + 4) > 10", true)
	runBoolTest(t, "100 / 10 < 20", true)
}

func Test_Mixed_ArrayAndArithmetic(t *testing.T) {
	runIntTest(t, "[1, 2, 3][0] + [4, 5, 6][2]", 7)
	runIntTest(t, "len([1, 2, 3]) + len([4, 5])", 5)
}

func Test_Mixed_StringAndNumber(t *testing.T) {
	// 字符串和数字不应自动转换
	// string() 内建函数应能转换
	result := runScript(t, `string(42)`)
	if result.String() != "42" {
		t.Errorf("got %q, want %q", result.String(), "42")
	}
}

func Test_Mixed_FunctionAndArray(t *testing.T) {
	runIntTest(t, `
		fn first(arr) {
			return arr[0]
		}
		first([10, 20, 30])
	`, 10)
}

func Test_Mixed_FunctionAndMap(t *testing.T) {
	result := runScript(t, `
		fn getKey(m, k) {
			return m[k]
		}
		getKey({"a": 1, "b": 2}, "a")
	`)
	if result.Int() != 1 {
		t.Errorf("got %d, want 1", result.Int())
	}
}

func Test_Mixed_IfAndFunction(t *testing.T) {
	runIntTest(t, `
		fn abs(x) {
			if x < 0 {
				return 0 - x
			}
			return x
		}
		abs(-5) + abs(3)
	`, 8)
}

func Test_Mixed_LoopAndArray(t *testing.T) {
	result := runScript(t, `
		arr := [10, 20, 30, 40, 50]
		sum := 0
		for v := range arr {
			sum = sum + v
		}
		sum
	`)
	if result.Int() != 150 {
		t.Errorf("got %d, want 150", result.Int())
	}
}

func Test_Mixed_LoopAndFunction(t *testing.T) {
	runIntTest(t, `
		fn fib(n) {
			if n <= 1 { return n }
			return fib(n - 1) + fib(n - 2)
		}
		total := 0
		for i := 0; i < 5; i = i + 1 {
			total = total + fib(i)
		}
		total
	`, 7) // fib(0..4) = 0+1+1+2+3 = 7
}

// ========== Token String方法测试 ==========

func Test_TokenType_String(t *testing.T) {
	// 所有token类型都应有可读名称
	tests := []struct {
		typ   TokenType
		nonEmpty bool
	}{
		{TokenEOF, true},
		{TokenInt, true},
		{TokenString, true},
		{TokenPlus, true},
		{TokenMinus, true},
		{TokenAssign, true},
		{TokenFn, true},
		{TokenIf, true},
		{TokenFor, true},
	}

	for _, tt := range tests {
		t.Run(tt.typ.String(), func(t *testing.T) {
			s := tt.typ.String()
			if s == "" {
				t.Errorf("TokenType(%d).String()不应为空", tt.typ)
			}
		})
	}
}

func Test_TokenType_String_Unknown(t *testing.T) {
	// 未知类型不应panic
	s := TokenType(99999).String()
	if s == "" {
		t.Error("未知TokenType的String()不应为空")
	}
}

// ========== Position测试 ==========

func Test_Position_TokenHasPosition(t *testing.T) {
	l := NewLexer("x := 1")
	tokens, _ := l.Tokenize()

	for _, tok := range tokens {
		if tok.Line < 1 {
			t.Errorf("token %v 行号应>=1, got %d", tok.Type, tok.Line)
		}
		if tok.Column < 1 {
			t.Errorf("token %v 列号应>=1, got %d", tok.Type, tok.Column)
		}
	}
}

func Test_Position_MultilineCorrect(t *testing.T) {
	input := "x := 1\ny := 2\nz := 3"
	l := NewLexer(input)
	tokens, _ := l.Tokenize()

	// 第1行的x
	if tokens[0].Line != 1 {
		t.Errorf("token[0]行号应为1, got %d", tokens[0].Line)
	}

	// 第2行的y (tokens: x, :=, 1, y)
	if tokens[3].Line != 2 {
		t.Errorf("token[3]行号应为2, got %d", tokens[3].Line)
	}

	// 第3行的z (tokens: x, :=, 1, y, :=, 2, z)
	if tokens[6].Line != 3 {
		t.Errorf("token[6]行号应为3, got %d", tokens[6].Line)
	}
}
