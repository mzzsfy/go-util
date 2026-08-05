package script

import (
    "testing"
)

// ========== 运行时错误路径测试 ==========

// runRuntimeErrorCheck 编译并执行代码，验证应产生运行时错误
func runRuntimeErrorCheck(t *testing.T, code string) {
    t.Helper()
    parser := NewParser()
    script, err := parser.Compile(code)
    if err != nil {
        t.Errorf("编译失败: %v", err)
        return
    }
    engine := NewEngine()
    ctx := NewContext()
    _, err = engine.Run(ctx, script)
    if err == nil {
        t.Error("期望运行时错误")
    }
}

func Test_runtime_ErrorPaths_DivisionByZero(t *testing.T) {
    tests := []struct {
        name  string
        input string
    }{
        {"int除零", "1 / 0"},
        {"float除零", "1.0 / 0.0"},
        {"int模零", "5 % 0"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            runRuntimeErrorCheck(t, tt.input)
        })
    }
}

func Test_runtime_ErrorPaths_InvalidTypeOperations(t *testing.T) {
    tests := []struct {
        name  string
        input string
    }{
        {"加法-数组+整数", "[] + 5"},
        {"减法-字符串-字符串", `"a" - "b"`},
        {"取负-字符串", `-"hello"`},
        {"取负-nil", "-nil"},
        {"位与-浮点数", "1.5 & 2"},
        {"位或-字符串", `"a" | "b"`},
        {"位异或-浮点数", "1.5 ^ 2"},
        {"位取反-字符串", `^"a"`},
        {"左移-浮点数", "1 << 2.5"},
        {"右移-浮点数", "4 >> 1.5"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            runRuntimeErrorCheck(t, tt.input)
        })
    }
}

func Test_runtime_ErrorPaths_Comparison(t *testing.T) {
    tests := []struct {
        name  string
        input string
    }{
        {"数组比较", "[1] < [2]"},
        {"nil与int比较", "nil < 5"},
        {"bool与string比较", `true < "a"`},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            runRuntimeErrorCheck(t, tt.input)
        })
    }
}

func Test_runtime_ErrorPaths_TypeConversion(t *testing.T) {
    tests := []struct {
        name  string
        input string
    }{
        {"int转换数组", "int([])"},
        {"float转换Map", `float({"a":1})`},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            runRuntimeErrorCheck(t, tt.input)
        })
    }
}

func Test_runtime_Error_ArithmeticUnsupported(t *testing.T) {
    tests := []struct {
        name string
        code string
    }{
        // 加法错误
        {"布尔加整数", `true + 5`},
        {"nil加整数", `nil + 5`},
        {"布尔加数组", `true + [1]`},
        // 减法错误
        {"字符串减整数", `"hello" - 5`},
        {"数组减数组", `[1] - [2]`},
        {"布尔减布尔", `true - false`},
        // 乘法错误
        {"字符串乘字符串", `"a" * "b"`},
        {"数组乘整数", `[1] * 2`},
        {"布尔乘布尔", `true * false`},
        // 取模错误
        {"浮点数模整数", `5.5 % 2`},
        {"整数模浮点数", `10 % 3.0`},
        {"浮点数模浮点数", `5.5 % 2.5`},
        {"字符串模字符串", `"a" % "b"`},
        {"布尔模布尔", `true % false`},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            runRuntimeErrorCheck(t, tt.code)
        })
    }
}

func Test_runtime_Error_NegUnsupported(t *testing.T) {
    tests := []struct {
        name string
        code string
    }{
        {"字符串取负", `-"hello"`},
        {"布尔取负", `-true`},
        {"nil取负", `-nil`},
        {"数组取负", `-[1, 2]`},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            runRuntimeErrorCheck(t, tt.code)
        })
    }
}

func Test_runtime_Error_BitOpsUnsupported(t *testing.T) {
    tests := []struct {
        name string
        code string
    }{
        {"浮点数位与", `1.5 & 2`},
        {"字符串位或", `"a" | "b"`},
        {"浮点数位异或", `1.5 ^ 2`},
        {"浮点数位取反", `^3.14`},
        {"字符串左移", `"a" << 2`},
        {"字符串右移", `"a" >> 2`},
        {"布尔位与", `true & false`},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            runRuntimeErrorCheck(t, tt.code)
        })
    }
}

func Test_runtime_Error_CompareUnsupported(t *testing.T) {
    tests := []struct {
        name string
        code string
    }{
        {"数组小于数组", `[1] < [2]`},
        {"nil小于整数", `nil < 5`},
        {"布尔小于字符串", `true < "a"`},
        {"整数大于数组", `1 > [1]`},
        {"数组小于等于数组", `[1] <= [2]`},
        {"布尔大于等于字符串", `true >= "a"`},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            runRuntimeErrorCheck(t, tt.code)
        })
    }
}

func Test_runtime_Error_IndexUnsupported(t *testing.T) {
    tests := []struct {
        name string
        code string
    }{
        {"整数索引", `42[0]`},
        {"浮点数索引", `3.14[0]`},
        {"布尔索引", `true[0]`},
        {"nil索引", `nil[0]`},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            runRuntimeErrorCheck(t, tt.code)
        })
    }
}

func Test_runtime_Error_SliceUnsupported(t *testing.T) {
    tests := []struct {
        name string
        code string
    }{
        {"整数切片", `42[0:2]`},
        {"浮点数切片", `3.14[0:2]`},
        {"布尔切片", `true[0:1]`},
        {"nil切片", `nil[0:1]`},
        {"Map切片", `{}[0:1]`},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            runRuntimeErrorCheck(t, tt.code)
        })
    }
}

func Test_runtime_Error_ArrayBoundary(t *testing.T) {
    // 测试数组越界读取返回nil而非错误
    result := runScript(t, `arr := [1, 2]
arr[10]`)
    // 越界读取返回nil
    if !result.IsNil() {
        t.Errorf("期望nil，得到 %v", result)
    }
}

// Test_runtime_StackOverflow 递归超过MaxCallDepth应报错
// 当前handleCall不检查MaxDepth, 递归会正常完成而非报错
func Test_runtime_StackOverflow(t *testing.T) {
	parser := NewParser()
	script, err := parser.Compile(`
		fn rec(n) {
			if n <= 0 { return 0 }
			return rec(n - 1)
		}
		rec(100)
	`)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}
	engine := NewEngine(WithMaxCallDepth(10))
	ctx := NewContext()
	_, err = engine.Run(ctx, script)
	if err == nil {
		t.Error("递归超过MaxCallDepth应返回调用栈溢出错误")
	}
}

// Test_runtime_RecursionWithinLimit 正常递归(未超过MaxCallDepth)应成功
func Test_runtime_RecursionWithinLimit(t *testing.T) {
	parser := NewParser()
	script, err := parser.Compile(`
		fn sum(n) {
			if n <= 0 { return 0 }
			return n + sum(n - 1)
		}
		sum(50)
	`)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}
	engine := NewEngine(WithMaxCallDepth(100))
	ctx := NewContext()
	result, err := engine.Run(ctx, script)
	if err != nil {
		t.Fatalf("正常递归不应报错: %v", err)
	}
	// sum(50) = 50+49+...+1 = 1275
	if result.Int() != 1275 {
		t.Errorf("期望1275, 得到%d", result.Int())
	}
}

func Test_runtime_Error_TypeConversion(t *testing.T) {
    tests := []struct {
        name string
        code string
    }{
        {"数组转整数", `int([1, 2])`},
        {"Map转整数", `int({})`},
        {"数组转浮点", `float([1, 2])`},
        {"Map转浮点", `float({})`},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            runRuntimeErrorCheck(t, tt.code)
        })
    }
}
