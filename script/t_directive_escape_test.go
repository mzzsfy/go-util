package script

import (
    "testing"
)

// ========== 指令格式、字符串转义和Token边界测试 ==========
// 参照 goja/otto 和 tengo 的测试模式

// ========== A. 指令格式测试 ==========

// Test_Dir_FnZeroParamReturnTypes 测试 #fn 零参数指令的各种返回类型
func Test_Dir_FnZeroParamReturnTypes(t *testing.T) {
    t.Run("返回int", func(t *testing.T) {
        result := runScriptWithFunc(t, `
            #fn f()=>int
            f()
        `, "f", func() int { return 42 })
        assertInt(t, result, 42)
    })
    t.Run("返回string", func(t *testing.T) {
        result := runScriptWithFunc(t, `
            #fn f()=>string
            f()
        `, "f", func() string { return "hello" })
        assertString(t, result, "hello")
    })
    t.Run("返回bool", func(t *testing.T) {
        result := runScriptWithFunc(t, `
            #fn f()=>bool
            f()
        `, "f", func() bool { return true })
        assertBool(t, result, true)
    })
    t.Run("返回float", func(t *testing.T) {
        result := runScriptWithFunc(t, `
            #fn f()=>float
            f()
        `, "f", func() float64 { return 3.14 })
        assertFloat(t, result, 3.14)
    })
}

// Test_Dir_FnWithParams 测试带参数的 #fn 指令
func Test_Dir_FnWithParams(t *testing.T) {
    t.Run("单参数", func(t *testing.T) {
        result := runScriptWithFunc(t, `
            #fn f(int)=>int
            f(10)
        `, "f", func(x int) int { return x * 2 })
        assertInt(t, result, 20)
    })
    t.Run("双参数", func(t *testing.T) {
        result := runScriptWithFunc(t, `
            #fn f(int, int)=>int
            f(10, 20)
        `, "f", func(a, b int) int { return a + b })
        assertInt(t, result, 30)
    })
    t.Run("三参数混合", func(t *testing.T) {
        result := runScriptWithFunc(t, `
            #fn f(int, string, bool)=>any
            f(1, "x", true)
        `, "f", func(a int, b string, c bool) interface{} {
            if c {
                return a + len(b)
            }
            return 0
        })
        assertInt(t, result, 2)
    })
}

// Test_Dir_FnTypeAlias 测试 #fn 使用类型别名
func Test_Dir_FnTypeAlias(t *testing.T) {
    t.Run("别名i", func(t *testing.T) {
        result := runScriptWithFunc(t, `
            #fn f(i)=>int
            f(5)
        `, "f", func(x int) int { return x * 3 })
        assertInt(t, result, 15)
    })
    t.Run("别名s", func(t *testing.T) {
        result := runScriptWithFunc(t, `
            #fn greet(s)=>s
            greet("World")
        `, "greet", func(name string) string { return "Hello, " + name })
        assertString(t, result, "Hello, World")
    })
    t.Run("别名f", func(t *testing.T) {
        result := runScriptWithFunc(t, `
            #fn scale(f)=>f
            scale(4.0)
        `, "scale", func(x float64) float64 { return x * 2.5 })
        assertFloat(t, result, 10.0)
    })
}

// Test_Dir_FnSpacing 测试 #fn 指令中空格的容错性
func Test_Dir_FnSpacing(t *testing.T) {
    t.Run("箭头前后有空格", func(t *testing.T) {
        result := runScriptWithFunc(t, `
            #fn f(int, int) => int
            f(10, 20)
        `, "f", func(a, b int) int { return a + b })
        assertInt(t, result, 30)
    })
    t.Run("无空格紧凑格式", func(t *testing.T) {
        result := runScriptWithFunc(t, `
            #fn f(int,int)=>int
            f(10, 20)
        `, "f", func(a, b int) int { return a + b })
        assertInt(t, result, 30)
    })
}

// Test_Dir_FnInForLoop 测试 #fn 声明的函数在for循环中调用
func Test_Dir_FnInForLoop(t *testing.T) {
    result := runScriptWithFunc(t, `
        #fn double(int)=>int
        sum := 0
        for i := 1; i <= 5; i = i + 1 {
            sum = sum + double(i)
        }
        sum
    `, "double", func(x int) int { return x * 2 })
    // 2+4+6+8+10 = 30
    assertInt(t, result, 30)
}

// Test_Dir_FnInExpression 测试 #fn 声明的函数作为表达式的一部分
func Test_Dir_FnInExpression(t *testing.T) {
    ctx := NewContext()
    ctx.BindFunc("add", func(a, b int) int { return a + b })
    ctx.BindFunc("mul", func(a, b int) int { return a * b })
    result := runScriptWithContext(t, `
        #fn add(int, int)=>int
        #fn mul(int, int)=>int
        add(1, 2) + mul(3, 4)
    `, ctx)
    // 3 + 12 = 15
    assertInt(t, result, 15)
}

// Test_Dir_FnNestedCall 测试 #fn 声明的函数嵌套调用
func Test_Dir_FnNestedCall(t *testing.T) {
    ctx := NewContext()
    ctx.BindFunc("inc", func(x int) int { return x + 1 })
    ctx.BindFunc("dbl", func(x int) int { return x * 2 })
    result := runScriptWithContext(t, `
        #fn inc(int)=>int
        #fn dbl(int)=>int
        dbl(inc(inc(5)))
    `, ctx)
    // inc(5)=6, inc(6)=7, dbl(7)=14
    assertInt(t, result, 14)
}

// Test_Dir_TypeAliasEquivalence 测试类型别名与完整类型名等效
func Test_Dir_TypeAliasEquivalence(t *testing.T) {
    t.Run("int别名i等效", func(t *testing.T) {
        ctx := NewContext()
        ctx.BindFunc("f", func(x int) int { return x + 1 })
        r1 := runScriptWithContext(t, `#fn f(int)=>int`+"\n"+`f(10)`, ctx)
        ctx2 := NewContext()
        ctx2.BindFunc("f", func(x int) int { return x + 1 })
        r2 := runScriptWithContext(t, `#fn f(i)=>i`+"\n"+`f(10)`, ctx2)
        assertInt(t, r1, r2.Int())
    })
    t.Run("string别名s和str等效", func(t *testing.T) {
        ctx := NewContext()
        ctx.BindFunc("f", func(x string) string { return x + "!" })
        r1 := runScriptWithContext(t, `#fn f(s)=>s`+"\n"+`f("a")`, ctx)
        ctx2 := NewContext()
        ctx2.BindFunc("f", func(x string) string { return x + "!" })
        r2 := runScriptWithContext(t, `#fn f(str)=>str`+"\n"+`f("a")`, ctx2)
        assertString(t, r1, r2.String())
    })
    t.Run("float别名f等效", func(t *testing.T) {
        ctx := NewContext()
        ctx.BindFunc("f", func(x float64) float64 { return x * 2 })
        r1 := runScriptWithContext(t, `#fn f(f)=>f`+"\n"+`f(1.5)`, ctx)
        assertFloat(t, r1, 3.0)
    })
    t.Run("bool别名b等效", func(t *testing.T) {
        ctx := NewContext()
        ctx.BindFunc("f", func(x bool) bool { return !x })
        r1 := runScriptWithContext(t, `#fn f(b)=>b`+"\n"+`f(true)`, ctx)
        assertBool(t, r1, false)
    })
    t.Run("any类型", func(t *testing.T) {
        ctx := NewContext()
        ctx.BindFunc("f", func(x int) int { return x })
        r1 := runScriptWithContext(t, `#fn f(any)=>any`+"\n"+`f(42)`, ctx)
        assertInt(t, r1, 42)
    })
}

// ========== B. 字符串转义测试 ==========

// Test_Escape_Basic 测试基本转义字符
func Test_Escape_Basic(t *testing.T) {
    t.Run("换行符", func(t *testing.T) {
        runStringTest(t, `"\n"`, "\n")
    })
    t.Run("制表符", func(t *testing.T) {
        runStringTest(t, `"\t"`, "\t")
    })
    t.Run("回车符", func(t *testing.T) {
        runStringTest(t, `"\r"`, "\r")
    })
    t.Run("反斜杠", func(t *testing.T) {
        runStringTest(t, `"\\"`, "\\")
    })
    t.Run("双引号", func(t *testing.T) {
        runStringTest(t, `"\""`, "\"")
    })
    t.Run("组合转义", func(t *testing.T) {
        runStringTest(t, `"line1\nline2"`, "line1\nline2")
    })
}

// Test_Escape_InExpression 测试转义在表达式中的行为
func Test_Escape_InExpression(t *testing.T) {
    t.Run("带转义的字符串拼接", func(t *testing.T) {
        runStringTest(t, `"a\n" + "b"`, "a\nb")
    })
    t.Run("带转义的字符串长度", func(t *testing.T) {
        runIntTest(t, `len("a\nb")`, 3)
    })
    t.Run("带转义的字符串比较", func(t *testing.T) {
        runBoolTest(t, `"\n" == "\n"`, true)
    })
    t.Run("typeof结果为string", func(t *testing.T) {
        runStringTest(t, `typeof("\n")`, "string")
    })
}

// Test_Escape_UnknownPreserved 测试未知转义保留原样
func Test_Escape_UnknownPreserved(t *testing.T) {
    // \x 不是已知转义,保留为 \x 两个字符
    runStringTest(t, `"\x"`, "\\x")
}

// ========== C. 数字字面量测试 ==========

// Test_NumLit_Hex 测试十六进制数字
func Test_NumLit_Hex(t *testing.T) {
    tests := []struct {
        Input    string
        Expected int
    }{
        {"0xFF", 255},
        {"0x0", 0},
        {"0xABCD", 43981},
        {"0XFF", 255},     // 大写X
        {"0xff_00", 65280}, // 带下划线
    }
    RunIntTestsSimple(t, tests)
}

// Test_NumLit_Underscore 测试下划线分隔符
func Test_NumLit_Underscore(t *testing.T) {
    tests := []struct {
        Input    string
        Expected int
    }{
        {"1_000", 1000},
        {"1_000_000", 1000000},
        {"0xFF_FF", 65535},
    }
    RunIntTestsSimple(t, tests)
}

// Test_NumLit_Float 测试浮点数字面量
func Test_NumLit_Float(t *testing.T) {
    tests := []struct {
        Input    string
        Expected float64
    }{
        {"0.0", 0.0},
        {"0.5", 0.5},
        {"1.0", 1.0},
        {"3.14", 3.14},
        {"0.00_1", 0.001},
        {"3.14_15", 3.1415},
    }
    RunFloatTestsSimple(t, tests)
}

// ========== D. Token边界测试 ==========

// Test_Tok_ArithmeticOps 测试算术运算符token
func Test_Tok_ArithmeticOps(t *testing.T) {
    tests := []struct {
        Input    string
        Expected int
    }{
        {"5 + 3", 8},
        {"5 - 3", 2},
        {"5 * 3", 15},
        {"6 / 3", 2},
        {"7 % 3", 1},
    }
    RunIntTestsSimple(t, tests)
}

// Test_Tok_ComparisonOps 测试比较运算符
func Test_Tok_ComparisonOps(t *testing.T) {
    tests := []struct {
        Input    string
        Expected bool
    }{
        {"3 == 3", true},
        {"3 != 4", true},
        {"3 < 4", true},
        {"3 <= 3", true},
        {"4 > 3", true},
        {"4 >= 4", true},
    }
    RunBoolTestsSimple(t, tests)
}

// Test_Tok_LogicalOps 测试逻辑运算符
func Test_Tok_LogicalOps(t *testing.T) {
    tests := []struct {
        Input    string
        Expected bool
    }{
        {"true && true", true},
        {"true || false", true},
        {"!false", true},
    }
    RunBoolTestsSimple(t, tests)
}

// Test_Tok_BitwiseOps 测试位运算符
func Test_Tok_BitwiseOps(t *testing.T) {
    tests := []struct {
        Input    string
        Expected int
    }{
        {"0xFF & 0x0F", 0x0F},
        {"0xF0 | 0x0F", 0xFF},
        {"0xFF ^ 0x0F", 0xF0},
        {"1 << 4", 16},
        {"256 >> 2", 64},
    }
    RunIntTestsSimple(t, tests)
}

// Test_Tok_CompoundAssign 测试复合赋值运算符
func Test_Tok_CompoundAssign(t *testing.T) {
    // lexer识别 += -= *= /= token, 编译器已支持
    cases := []struct{ src string; expected int }{
        {"x := 10\nx += 5\nx", 15},
        {"x := 10\nx -= 3\nx", 7},
        {"x := 10\nx *= 3\nx", 30},
        {"x := 10\nx /= 2\nx", 5},
    }
    for _, tc := range cases {
        t.Run(tc.src, func(t *testing.T) {
            runIntTest(t, tc.src, tc.expected)
        })
    }
}

// Test_Tok_AssignmentOps 测试赋值运算符
func Test_Tok_AssignmentOps(t *testing.T) {
    t.Run("声明赋值:=", func(t *testing.T) {
        runIntTest(t, "x := 42\nx", 42)
    })
    t.Run("普通赋值=", func(t *testing.T) {
        runIntTest(t, "x := 10\nx = 20\nx", 20)
    })
}

// Test_Tok_Delimiters 测试分隔符在表达式中的正确解析
func Test_Tok_Delimiters(t *testing.T) {
    t.Run("括号", func(t *testing.T) {
        runIntTest(t, "(1 + 2) * 3", 9)
    })
    t.Run("花括号map字面量", func(t *testing.T) {
        runIntTest(t, `{"a": 1, "b": 2}["a"]`, 1)
    })
    t.Run("方括号", func(t *testing.T) {
        runIntTest(t, "[10, 20, 30][1]", 20)
    })
    t.Run("逗号分隔参数", func(t *testing.T) {
        runIntTest(t, "len([1, 2, 3, 4, 5])", 5)
    })
    t.Run("分号在for循环中", func(t *testing.T) {
        runIntTest(t, "sum := 0\nfor i := 0; i < 3; i = i + 1 {\nsum = sum + 1\n}\nsum", 3)
    })
}

// Test_Tok_KeywordsAsPrefix 测试关键字作为标识符前缀不被误判
func Test_Tok_KeywordsAsPrefix(t *testing.T) {
    tests := []struct {
        name  string
        input string
        expected int
    }{
        {"fnName不以fn解析", "fnName := 42\nfnName", 42},
        {"ifElse不以if解析", "ifElse := 7\nifElse", 7},
        {"forLoop不以for解析", "forLoop := 3\nforLoop", 3},
        {"returnVal不以return解析", "returnVal := 99\nreturnVal", 99},
        {"breakPoint不以break解析", "breakPoint := 5\nbreakPoint", 5},
        {"continueLoop不以continue解析", "continueLoop := 8\ncontinueLoop", 8},
        {"throwErr不以throw解析", "throwErr := 1\nthrowErr", 1},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            runIntTest(t, tt.input, tt.expected)
        })
    }
}

// Test_Tok_KeywordValues 测试关键字字面量值
func Test_Tok_KeywordValues(t *testing.T) {
    t.Run("true", func(t *testing.T) {
        runBoolTest(t, "true", true)
    })
    t.Run("false", func(t *testing.T) {
        runBoolTest(t, "false", false)
    })
    t.Run("nil等于nil", func(t *testing.T) {
        runBoolTest(t, "nil == nil", true)
    })
}

// Test_Tok_LexerKeywordIdentification 测试关键字在lexer层正确识别
func Test_Tex_LexerKeywordIdentification(t *testing.T) {
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
        // 注意: map在lexer层是TokenIdent,不是关键字
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
                t.Errorf("keyword %q: type = %v, want %v", tt.kw, tokens[0].Type, tt.typ)
            }
        })
    }
}
