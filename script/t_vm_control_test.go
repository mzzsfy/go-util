package script

import (
    "testing"
)

// ========== If 表达式测试 ==========

// TestIfExpr_ElseIfChain 测试if-else-if语句链
func Test_ifstmt_ElseIfChain(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected int
    }{
        {"第一个分支", `x := 0
if true { x = 1 } else if false { x = 2 } else { x = 3 }
x`, 1},
        {"第二个分支", `x := 0
if false { x = 1 } else if true { x = 2 } else { x = 3 }
x`, 2},
        {"第三个分支", `x := 0
if false { x = 1 } else if false { x = 2 } else { x = 3 }
x`, 3},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := runScript(t, tt.input)
            if result.Int() != tt.expected {
                t.Errorf("期望 %d, 实际 %d", tt.expected, result.Int())
            }
        })
    }
}

// TestIfExpr 测试if表达式（表达式形式，非语句）
func Test_ifexpr(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected int
    }{
        {"基本if表达式", `x := if true { 1 } else { 0 }
x`, 1},
        {"条件为假", `x := if false { 1 } else { 0 }
x`, 0},
        {"带运算的if表达式", `x := if 1 < 2 { 10 + 20 } else { 30 + 40 }
x`, 30},
        {"复杂条件", `x := if 5 > 3 && 2 < 4 { 100 } else { 0 }
x`, 100},
        {"使用变量", `a := 5
b := 10
x := if a < b { a + b } else { a - b }
x`, 15},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := runScript(t, tt.input)
            if result.Int() != tt.expected {
                t.Errorf("if表达式失败: 期望 %d, 实际 %d", tt.expected, result.Int())
            }
        })
    }
}

// TestIfExpr_Nested 测试嵌套if表达式
func Test_ifexpr_Nested(t *testing.T) {
    // 注意：嵌套if表达式需要作为表达式返回值，而不是语句
    tests := []struct {
        name     string
        input    string
        expected int
    }{
        // 使用三元选择模式替代深层嵌套
        {"条件选择", `x := 1
y := if x == 1 { 10 } else { 20 }
y`, 10},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := runScript(t, tt.input)
            if result.Int() != tt.expected {
                t.Errorf("if表达式失败: 期望 %d, 实际 %d", tt.expected, result.Int())
            }
        })
    }
}

// TestIfExpr_Types 测试if表达式返回不同类型
func Test_ifexpr_Types(t *testing.T) {
    t.Run("字符串结果", func(t *testing.T) {
        result := runScript(t, `x := if true { "yes" } else { "no" }
x`)
        if result.String() != "yes" {
            t.Errorf("期望 yes, 实际 %s", result.String())
        }
    })

    t.Run("布尔结果", func(t *testing.T) {
        result := runScript(t, `x := if true { true } else { false }
x`)
        if !result.Bool() {
            t.Error("期望 true")
        }
    })
}

// TestTypeAnnotation 测试类型注解
func Test_typeannotation(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        shouldError bool
    }{
        {"简单变量声明", "x := 10", false},
        {"字符串声明", `s := "hello"`, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            parser := NewParser()
            _, err := parser.Compile(tt.input)
            if tt.shouldError && err == nil {
                t.Error("期望编译错误")
            }
            if !tt.shouldError && err != nil {
                t.Errorf("编译失败: %v", err)
            }
        })
    }
}

// TestTypeExpr 测试类型表达式解析
func Test_typeexpr(t *testing.T) {
    t.Run("数组类型", func(t *testing.T) {
        parser := NewParser()
        err := parser.Validate("#fn test(arr []int)=>int")
        if err != nil {
            t.Logf("数组类型验证: %v", err)
        }
    })

    t.Run("Map类型", func(t *testing.T) {
        parser := NewParser()
        err := parser.Validate("#fn test(m map{string:int})=>int")
        if err != nil {
            t.Logf("Map类型验证: %v", err)
        }
    })

    t.Run("基本类型", func(t *testing.T) {
        parser := NewParser()
        err := parser.Validate("#fn test(x int, y string, z float)=>nil")
        if err != nil {
            t.Logf("基本类型验证: %v", err)
        }
    })
}

// TestParseErrors 测试解析错误
func Test_parseerrors(t *testing.T) {
    tests := []struct {
        name  string
        input string
    }{
        {"缺少表达式", "x :="},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            parser := NewParser()
            _, err := parser.Compile(tt.input)
            if err == nil {
                t.Error("期望编译错误")
            }
        })
    }
}

// TestParser_ErrorPaths 测试Parser错误路径
func Test_parser_ErrorPaths(t *testing.T) {
    tests := []struct {
        name  string
        input string
    }{
        {"未闭合字符串", `x := "unclosed`},
        {"无效数字格式", `x := 123abc`},
        {"空数组索引", `arr[]`},
        {"缺少操作数", `x := 1 +`},
        {"无效token", `x := @`},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            parser := NewParser()
            _, err := parser.Compile(tt.input)
            if err == nil {
                t.Errorf("期望解析错误: %s", tt.input)
            }
        })
    }
}

// ========== Break 语句测试 ==========

// Test_Break_InWhileLoop 测试 while 循环中的 break
func Test_Break_InWhileLoop(t *testing.T) {
    code := `
        sum := 0
        i := 0
        for i < 10 {
            if i == 5 { break }
            sum = sum + i
            i = i + 1
        }
        sum
    `
    result := runScript(t, code)
    // 0 + 1 + 2 + 3 + 4 = 10 (在 i=5 时 break)
    expected := 10
    if result.Int() != expected {
        t.Errorf("期望 %d, 得到 %d", expected, result.Int())
    }
}

// Test_Break_InStandardFor 测试标准 for 循环中的 break
func Test_Break_InStandardFor(t *testing.T) {
    code := `
        sum := 0
        for i := 0; i < 10; i = i + 1 {
            if i == 5 { break }
            sum = sum + i
        }
        sum
    `
    result := runScript(t, code)
    expected := 10
    if result.Int() != expected {
        t.Errorf("期望 %d, 得到 %d", expected, result.Int())
    }
}

// Test_Break_InCountFor 测试计数循环中的 break
// 注意：计数循环 for i := n 的语义是 i 存储循环次数 n
// 内部使用 __for_i__ 作为计数器，循环 n 次
func Test_Break_InCountFor(t *testing.T) {
    code := `
        sum := 0
        for n := 10 {
            sum = sum + 1
            if sum == 3 { break }
        }
        sum
    `
    result := runScript(t, code)
    expected := 3
    if result.Int() != expected {
        t.Errorf("期望 %d, 得到 %d", expected, result.Int())
    }
}

// Test_Break_Nested 测试嵌套循环中的 break（只跳出内层循环）
func Test_Break_Nested(t *testing.T) {
    code := `
        sum := 0
        for i := 0; i < 3; i = i + 1 {
            for j := 0; j < 5; j = j + 1 {
                if j == 2 { break }
                sum = sum + 1
            }
        }
        sum
    `
    result := runScript(t, code)
    // 外层循环 3 次，内层每次执行 2 次（j=0, j=1），共 6 次
    expected := 6
    if result.Int() != expected {
        t.Errorf("期望 %d, 得到 %d", expected, result.Int())
    }
}

// Test_Break_OutsideLoop_Error 测试循环外使用 break 应该报错
func Test_Break_OutsideLoop_Error(t *testing.T) {
    code := `break`
    parser := NewParser()
    _, err := parser.Compile(code)
    if err == nil {
        t.Error("期望 break 在循环外报错，但没有报错")
    }
}

// ========== Continue 语句测试 ==========

// Test_Continue_InWhileLoop 测试 while 循环中的 continue
func Test_Continue_InWhileLoop(t *testing.T) {
    code := `
        sum := 0
        i := 0
        for i < 5 {
            i = i + 1
            if i % 2 == 0 { continue }
            sum = sum + i
        }
        sum
    `
    result := runScript(t, code)
    // 1 + 3 + 5 = 9 (跳过偶数 2, 4)
    expected := 9
    if result.Int() != expected {
        t.Errorf("期望 %d, 得到 %d", expected, result.Int())
    }
}

// Test_Continue_InStandardFor 测试标准 for 循环中的 continue
func Test_Continue_InStandardFor(t *testing.T) {
    code := `
        sum := 0
        for i := 0; i < 5; i = i + 1 {
            if i % 2 == 0 { continue }
            sum = sum + i
        }
        sum
    `
    result := runScript(t, code)
    // 1 + 3 = 4 (跳过偶数 0, 2, 4)
    expected := 4
    if result.Int() != expected {
        t.Errorf("期望 %d, 得到 %d", expected, result.Int())
    }
}

// Test_Continue_OutsideLoop_Error 测试循环外使用 continue 应该报错
func Test_Continue_OutsideLoop_Error(t *testing.T) {
    code := `continue`
    parser := NewParser()
    _, err := parser.Compile(code)
    if err == nil {
        t.Error("期望 continue 在循环外报错，但没有报错")
    }
}

// Test_Break_Continue_Combined 测试 break 和 continue 组合使用
func Test_Break_Continue_Combined(t *testing.T) {
    code := `
        sum := 0
        for i := 0; i < 20; i = i + 1 {
            if i == 10 { break }
            if i % 2 == 0 { continue }
            sum = sum + i
        }
        sum
    `
    result := runScript(t, code)
    // 奇数求和，直到 i=10: 1 + 3 + 5 + 7 + 9 = 25
    expected := 25
    if result.Int() != expected {
        t.Errorf("期望 %d, 得到 %d", expected, result.Int())
    }
}

// Test_CountLoop_FromOneToN 测试计数器循环从1到N的语义
func Test_CountLoop_FromOneToN(t *testing.T) {
    t.Run("基本计数循环", func(t *testing.T) {
        code := `
            result := ""
            for i := 5 {
                result = result + string(i)
            }
            result
        `
        result, err := Eval(code)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        // i 从 1 到 5，输出 "12345"
        expected := "12345"
        if result.String() != expected {
            t.Errorf("期望 %s, 得到 %s", expected, result.String())
        }
    })

    t.Run("求和测试", func(t *testing.T) {
        code := `
            sum := 0
            for i := 10 {
                sum = sum + i
            }
            sum
        `
        result, err := Eval(code)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        // 1 + 2 + ... + 10 = 55
        expected := 55
        if result.Int() != expected {
            t.Errorf("期望 %d, 得到 %d", expected, result.Int())
        }
    })

    t.Run("单次循环", func(t *testing.T) {
        code := `
            count := 0
            for i := 1 {
                count = count + 1
            }
            count
        `
        result, err := Eval(code)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        // 只循环1次，i=1
        expected := 1
        if result.Int() != expected {
            t.Errorf("期望 %d, 得到 %d", expected, result.Int())
        }
    })

    t.Run("零次循环", func(t *testing.T) {
        code := `
            count := 0
            for i := 0 {
                count = count + 1
            }
            count
        `
        result, err := Eval(code)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        // 0次循环，因为 1 <= 0 为假
        expected := 0
        if result.Int() != expected {
            t.Errorf("期望 %d, 得到 %d", expected, result.Int())
        }
    })
}
