package script

import (
    "testing"
)

// ========== 内置函数完整测试 ==========
// 本文件测试脚本引擎的内置函数功能

// Test_Builtin_Print 测试print函数
func Test_Builtin_Print(t *testing.T) {
    t.Run("print编译", func(t *testing.T) {
        tests := []struct {
            name  string
            input string
        }{
            {"单个参数", `print("hello")`},
            {"多个参数", `print("a", "b", "c")`},
            {"表达式", `print(1 + 2)`},
        }

        for _, tt := range tests {
            t.Run(tt.name, func(t *testing.T) {
                parser := NewParser()
                _, err := parser.Compile(tt.input)
                if err != nil {
                    t.Errorf("print函数 %s 编译失败: %v", tt.name, err)
                }
            })
        }
    })

    t.Run("print各种类型", func(t *testing.T) {
        tests := []struct {
            name  string
            input string
        }{
            {"整数", `print(1)`},
            {"浮点", `print(1.5)`},
            {"字符串", `print("str")`},
            {"布尔", `print(true)`},
            {"nil", `print(nil)`},
        }

        for _, tt := range tests {
            t.Run(tt.name, func(t *testing.T) {
                parser := NewParser()
                _, err := parser.Compile(tt.input)
                if err != nil {
                    t.Errorf("print测试失败: %v", err)
                }
            })
        }
    })
}

// Test_Builtin_Len 测试len函数
func Test_Builtin_Len(t *testing.T) {
    t.Run("len数组", func(t *testing.T) {
        tests := []struct {
            name     string
            input    string
            expected int
        }{
            {"空数组", `len([])`, 0},
            {"单元素", `len([1])`, 1},
            {"多元素", `len([1, 2, 3, 4, 5])`, 5},
        }

        for _, tt := range tests {
            t.Run(tt.name, func(t *testing.T) {
                result, err := Eval(tt.input)
                if err != nil {
                    t.Errorf("len函数 %s 执行失败: %v", tt.name, err)
                    return
                }
                if result.Int() != tt.expected {
                    t.Errorf("%s: 期望 %d, 得到 %d", tt.name, tt.expected, result.Int())
                }
            })
        }
    })

    t.Run("len字符串", func(t *testing.T) {
        tests := []struct {
            name     string
            input    string
            expected int
        }{
            {"空字符串", `len("")`, 0},
            {"单字符", `len("a")`, 1},
            {"长字符串", `len("hello world")`, 11},
        }

        for _, tt := range tests {
            t.Run(tt.name, func(t *testing.T) {
                result, err := Eval(tt.input)
                if err != nil {
                    t.Errorf("len函数 %s 执行失败: %v", tt.name, err)
                    return
                }
                if result.Int() != tt.expected {
                    t.Errorf("%s: 期望 %d, 得到 %d", tt.name, tt.expected, result.Int())
                }
            })
        }
    })
}

// Test_Builtin_TypeConversion 测试类型转换函数
func Test_Builtin_TypeConversion(t *testing.T) {
    t.Run("int函数", func(t *testing.T) {
        tests := []struct {
            name     string
            input    string
            expected int
        }{
            {"正数", `int("123")`, 123},
            {"负数", `int("-456")`, -456},
            {"零", `int("0")`, 0},
        }

        for _, tt := range tests {
            t.Run(tt.name, func(t *testing.T) {
                result, err := Eval(tt.input)
                if err != nil {
                    t.Errorf("int函数 %s 执行失败: %v", tt.name, err)
                    return
                }
                if result.Int() != tt.expected {
                    t.Errorf("%s: 期望 %d, 得到 %d", tt.name, tt.expected, result.Int())
                }
            })
        }
    })

    t.Run("float函数", func(t *testing.T) {
        tests := []struct {
            name  string
            input string
        }{
            {"整数", `float("123")`},
            {"小数", `float("3.14159")`},
            {"科学计数", `float("1e10")`},
        }

        for _, tt := range tests {
            t.Run(tt.name, func(t *testing.T) {
                result, err := Eval(tt.input)
                if err != nil {
                    t.Errorf("float函数 %s 执行失败: %v", tt.name, err)
                    return
                }
                if result.Float() == 0 {
                    t.Errorf("%s: 期望非零浮点数, 得到 %f", tt.name, result.Float())
                }
            })
        }
    })

    t.Run("string函数", func(t *testing.T) {
        tests := []struct {
            name     string
            input    string
            expected string
        }{
            {"整数", `string(123)`, "123"},
            {"负数", `string(-456)`, "-456"},
            {"字符串", `string("already")`, "already"},
        }

        for _, tt := range tests {
            t.Run(tt.name, func(t *testing.T) {
                result, err := Eval(tt.input)
                if err != nil {
                    t.Errorf("string函数 %s 执行失败: %v", tt.name, err)
                    return
                }
                if result.String() != tt.expected {
                    t.Errorf("%s: 期望 %s, 得到 %s", tt.name, tt.expected, result.String())
                }
            })
        }
    })
}

// Test_Builtin_Combinations 测试内置函数组合使用
func Test_Builtin_Combinations(t *testing.T) {
    t.Run("len和循环组合", func(t *testing.T) {
        result, err := Eval(`
arr := [1, 2, 3, 4, 5]
s := 0
for v := range arr {
    s = s + v
}
len(arr) + s
`)
        if err != nil {
            t.Errorf("组合使用失败: %v", err)
            return
        }
        // len(arr) = 5, s = 15, 总计 = 20
        if result.Int() != 20 {
            t.Errorf("期望 20, 得到 %d", result.Int())
        }
    })
}

// Test_Builtin_TypeOf 测试typeof函数
func Test_Builtin_TypeOf(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"整数类型", `typeof(123)`, "int"},
        {"浮点类型", `typeof(3.14)`, "float"},
        {"字符串类型", `typeof("hello")`, "string"},
        {"布尔类型", `typeof(true)`, "bool"},
        {"nil类型", `typeof(nil)`, "nil"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Eval(tt.input)
            if err != nil {
                t.Errorf("typeof函数 %s 执行失败: %v", tt.name, err)
                return
            }
            if result.String() != tt.expected {
                t.Errorf("%s: 期望 '%s', 得到 '%s'", tt.name, tt.expected, result.String())
            }
        })
    }
}

// Test_Builtin_EdgeCases 测试边界情况
func Test_Builtin_EdgeCases(t *testing.T) {
    t.Run("类型转换错误处理", func(t *testing.T) {
        t.Run("int函数转换失败", func(t *testing.T) {
            _, err := Eval(`int("abc")`)
            // 转换失败应返回错误
            if err == nil {
                t.Error("int(\"abc\")应返回错误")
            }
        })

        t.Run("float函数转换失败", func(t *testing.T) {
            _, err := Eval(`float("abc")`)
            // 转换失败应返回错误
            if err == nil {
                t.Error("float(\"abc\")应返回错误")
            }
        })

        t.Run("int函数布尔值转换", func(t *testing.T) {
            result, err := Eval(`int(true)`)
            if err != nil {
                t.Errorf("int(true)不应返回错误: %v", err)
                return
            }
            // true应转换为1
            if result.Int() != 1 {
                t.Errorf("int(true)应返回1, 得到 %d", result.Int())
            }
        })

        t.Run("int函数浮点数转换", func(t *testing.T) {
            result, err := Eval(`int(3.7)`)
            if err != nil {
                t.Errorf("int(3.7)不应返回错误: %v", err)
                return
            }
            // 浮点数应截断为整数
            if result.Int() != 3 {
                t.Errorf("int(3.7)应返回3, 得到 %d", result.Int())
            }
        })

        t.Run("float函数整数转换", func(t *testing.T) {
            result, err := Eval(`float(42)`)
            if err != nil {
                t.Errorf("float(42)不应返回错误: %v", err)
                return
            }
            // 整数应转换为浮点数
            if result.Float() != 42.0 {
                t.Errorf("float(42)应返回42.0, 得到 %f", result.Float())
            }
        })

        t.Run("string函数布尔值转换", func(t *testing.T) {
            result, err := Eval(`string(true)`)
            if err != nil {
                t.Errorf("string(true)不应返回错误: %v", err)
                return
            }
            // true应转换为"true"
            if result.String() != "true" {
                t.Errorf("string(true)应返回\"true\", 得到 %s", result.String())
            }
        })

        t.Run("string函数浮点数转换", func(t *testing.T) {
            result, err := Eval(`string(3.14)`)
            if err != nil {
                t.Errorf("string(3.14)不应返回错误: %v", err)
                return
            }
            // 浮点数应转换为字符串
            if result.String() != "3.14" {
                t.Errorf("string(3.14)应返回\"3.14\", 得到 %s", result.String())
            }
        })
    })

    t.Run("空输入", func(t *testing.T) {
        t.Run("len空数组", func(t *testing.T) {
            result, err := Eval(`len([])`)
            if err != nil {
                t.Errorf("len([]) 执行失败: %v", err)
                return
            }
            if result.Int() != 0 {
                t.Errorf("期望 0, 得到 %d", result.Int())
            }
        })

        t.Run("len空字符串", func(t *testing.T) {
            result, err := Eval(`len("")`)
            if err != nil {
                t.Errorf(`len("") 执行失败: %v`, err)
                return
            }
            if result.Int() != 0 {
                t.Errorf("期望 0, 得到 %d", result.Int())
            }
        })
    })
}

// Test_Builtin_Push 测试push函数
func Test_Builtin_Push(t *testing.T) {
    t.Run("基本追加", func(t *testing.T) {
        code := `
            arr := [1, 2, 3]
            arr2 := push(arr, 4)
            arr2
        `
        result, err := Eval(code)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        arr := result.Array()
        if len(arr.Elements) != 4 {
            t.Errorf("期望长度 4, 得到 %d", len(arr.Elements))
        }
        if arr.Elements[3].Int() != 4 {
            t.Errorf("期望最后一个元素为 4, 得到 %d", arr.Elements[3].Int())
        }
    })

    t.Run("原数组不变", func(t *testing.T) {
        code := `
            arr := [1, 2, 3]
            arr2 := push(arr, 4)
            len(arr)
        `
        result, err := Eval(code)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        if result.Int() != 3 {
            t.Errorf("期望原数组长度 3, 得到 %d", result.Int())
        }
    })

    t.Run("链式追加", func(t *testing.T) {
        code := `
            arr := []
            arr = push(arr, 1)
            arr = push(arr, 2)
            arr = push(arr, 3)
            len(arr)
        `
        result, err := Eval(code)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        if result.Int() != 3 {
            t.Errorf("期望长度 3, 得到 %d", result.Int())
        }
    })

    t.Run("追加不同类型", func(t *testing.T) {
        code := `
            arr := [1, "hello"]
            arr2 := push(arr, true)
            arr2
        `
        result, err := Eval(code)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        arr := result.Array()
        if len(arr.Elements) != 3 {
            t.Errorf("期望长度 3, 得到 %d", len(arr.Elements))
        }
        if !arr.Elements[2].Bool() {
            t.Errorf("期望最后一个元素为 true")
        }
    })

    t.Run("空数组追加", func(t *testing.T) {
        code := `
            arr := []
            arr2 := push(arr, 1)
            len(arr2)
        `
        result, err := Eval(code)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        if result.Int() != 1 {
            t.Errorf("期望长度 1, 得到 %d", result.Int())
        }
    })

    t.Run("快速排序场景", func(t *testing.T) {
        // 模拟README中的快速排序示例
        code := `
            arr := [3, 1, 4, 1, 5]
            less := []
            greater := []

            for v := range arr {
                if v < 3 {
                    less = push(less, v)
                } else {
                    greater = push(greater, v)
                }
            }

            len(less) + len(greater)
        `
        result, err := Eval(code)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        if result.Int() != 5 {
            t.Errorf("期望总长度 5, 得到 %d", result.Int())
        }
    })

    t.Run("错误处理-非数组", func(t *testing.T) {
        code := `push(123, 1)`
        _, err := Eval(code)
        if err == nil {
            t.Error("期望返回错误，但执行成功")
        }
    })

    t.Run("错误处理-参数数量错误", func(t *testing.T) {
        code := `push([1, 2])`
        _, err := Eval(code)
        if err == nil {
            t.Error("期望返回错误，但执行成功")
        }
    })
}
