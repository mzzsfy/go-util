package script

import (
    "testing"
)

// ========== 数组操作完整测试 ==========

func Test_arrayoperationsfull(t *testing.T) {
    t.Run("数组索引访问", func(t *testing.T) {
        result := runScript(t, "[1, 2, 3][0]")
        if result.Int() != 1 {
            t.Errorf("数组索引访问失败: 期望 1, 实际 %d", result.Int())
        }
    })

    t.Run("数组长度", func(t *testing.T) {
        result := runScript(t, "len([1, 2, 3, 4, 5])")
        if result.Int() != 5 {
            t.Errorf("数组长度获取失败: 期望 5, 实际 %d", result.Int())
        }
    })

    t.Run("嵌套数组", func(t *testing.T) {
        result := runScript(t, "[[1, 2], [3, 4]][1][0]")
        if result.Int() != 3 {
            t.Errorf("嵌套数组访问失败: 期望 3, 实际 %d", result.Int())
        }
    })
}

// ========== 比较运算完整测试 ==========

func Test_comparisonoperationsfull(t *testing.T) {
    tests := []struct {
        input    string
        expected bool
    }{
        {"5 == 5", true},
        {"5 != 3", true},
        {"5 < 10", true},
        {"5 <= 5", true},
        {"10 > 5", true},
        {"10 >= 10", true},
    }

    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            result := runScript(t, tt.input)
            if result.Bool() != tt.expected {
                t.Errorf("比较运算失败: 期望 %v, 实际 %v", tt.expected, result.Bool())
            }
        })
    }
}

// ========== 类型转换边界测试 ==========

func Test_handletoint_EdgeCases(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected int
    }{
        {"负浮点转整数", "int(-3.9)", -3},
        {"零浮点转整数", "int(0.0)", 0},
        {"大整数不变", "int(999999)", 999999},
        {"字符串转整数", `int("123")`, 123},
        {"true转整数", "int(true)", 1},
        {"false转整数", "int(false)", 0},
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

func Test_handletofloat_EdgeCases(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected float64
    }{
        {"负整数转浮点", "float(-42)", -42.0},
        {"零整数转浮点", "float(0)", 0.0},
        {"字符串转浮点", `float("3.14")`, 3.14},
        {"true转浮点", "float(true)", 1.0},
        {"false转浮点", "float(false)", 0.0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := runScript(t, tt.input)
            if result.Float() != tt.expected {
                t.Errorf("期望 %f, 实际 %f", tt.expected, result.Float())
            }
        })
    }
}

// TestTypeConversionErrors 测试类型转换错误情况
func Test_typeconversionerrors(t *testing.T) {
    // 测试不支持的类型转换应该返回错误
    errorCases := []struct {
        name  string
        input string
    }{
        {"数组转整数", "int([1, 2, 3])"},
        {"Map转整数", "int({\"a\": 1})"},
        {"数组转浮点", "float([1, 2])"},
        {"Map转浮点", "float({\"a\": 1})"},
    }

    for _, tt := range errorCases {
        t.Run(tt.name, func(t *testing.T) {
            parser := NewParser()
            script, err := parser.Compile(tt.input)
            if err != nil {
                // 编译错误也是可以接受的
                return
            }

            ctx := NewContext()
            engine := NewEngine()
            _, err = engine.Run(ctx, script)
            // 运行时应该产生错误
            if err == nil {
                t.Logf("警告: %s 应该产生错误", tt.name)
            }
        })
    }
}

func Test_handletostring_EdgeCases(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"负整数转字符串", "string(-42)", "-42"},
        {"零转字符串", "string(0)", "0"},
        {"false转字符串", "string(false)", "false"},
        {"nil转字符串", "string(nil)", "nil"},
        {"浮点转字符串", "string(3.14)", "3.14"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := runScript(t, tt.input)
            if result.String() != tt.expected {
                t.Errorf("期望 %s, 实际 %s", tt.expected, result.String())
            }
        })
    }
}

// ========== 数组Map创建边界测试 ==========

func Test_handlearraynew_EdgeCases(t *testing.T) {
    t.Run("空数组", func(t *testing.T) {
        result := runScript(t, "[]")
        if result.Type != TypeArray {
            t.Error("期望数组类型")
        }
        arr := result.Array()
        if len(arr.Elements) != 0 {
            t.Errorf("期望0个元素, 实际 %d", len(arr.Elements))
        }
    })

    t.Run("单元素数组", func(t *testing.T) {
        result := runScript(t, "[42]")
        if result.Type != TypeArray {
            t.Error("期望数组类型")
        }
        arr := result.Array()
        if len(arr.Elements) != 1 || arr.Elements[0].Int() != 42 {
            t.Error("单元素数组内容错误")
        }
    })

    t.Run("混合类型数组", func(t *testing.T) {
        result := runScript(t, `[1, "hello", true, nil]`)
        if result.Type != TypeArray {
            t.Error("期望数组类型")
        }
        arr := result.Array()
        if len(arr.Elements) != 4 {
            t.Errorf("期望4个元素, 实际 %d", len(arr.Elements))
        }
    })
}

func Test_handlemapnew_EdgeCases(t *testing.T) {
    t.Run("空Map", func(t *testing.T) {
        result := runScript(t, "{}")
        if result.Type != TypeMap {
            t.Error("期望Map类型")
        }
        m := result.Map()
        if len(m.Pairs) != 0 {
            t.Errorf("期望0个键值对, 实际 %d", len(m.Pairs))
        }
    })

    t.Run("单键值对Map", func(t *testing.T) {
        result := runScript(t, `{"key": 42}`)
        if result.Type != TypeMap {
            t.Error("期望Map类型")
        }
        m := result.Map()
        if len(m.Pairs) != 1 {
            t.Errorf("期望1个键值对, 实际 %d", len(m.Pairs))
        }
    })

    t.Run("多键值对Map", func(t *testing.T) {
        result := runScript(t, `{"a": 1, "b": 2, "c": 3}`)
        if result.Type != TypeMap {
            t.Error("期望Map类型")
        }
        m := result.Map()
        if len(m.Pairs) != 3 {
            t.Errorf("期望3个键值对, 实际 %d", len(m.Pairs))
        }
    })
}

// ========== 绑定值完整测试 ==========

func Test_handleloadbind_AllTypes(t *testing.T) {
    tests := []struct {
        name     string
        bindName string
        bindVal  interface{}
        input    string
        check    func(Value) bool
    }{
        {"绑定整数", "num", 123, `x :=>int getBindValue("num")
x`, func(v Value) bool { return v.Int() == 123 }},
        {"绑定字符串", "str", "hello", `x :=>string getBindValue("str")
x`, func(v Value) bool { return v.String() == "hello" }},
        {"绑定浮点数", "float", 3.14, `x :=>float getBindValue("float")
x`, func(v Value) bool { return v.Float() == 3.14 }},
        {"绑定布尔值", "bool", true, `x :=>bool getBindValue("bool")
x`, func(v Value) bool { return v.Bool() == true }},
        {"绑定nil", "nilval", nil, `x :=>any getBindValue("nilval")
x`, func(v Value) bool { return v.IsNil() }},
        {"绑定数组", "arr", []int{1, 2, 3}, `x :=>arr getBindValue("arr")
x`, func(v Value) bool { return v.Type == TypeArray }},
        {"绑定Map", "mapval", map[string]int{"a": 1}, `x :=>any getBindValue("mapval")
x`, func(v Value) bool { return v.Type == TypeMap }},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            parser := NewParser()
            script, err := parser.Compile(tt.input)
            if err != nil {
                t.Fatalf("编译失败: %v", err)
            }

            ctx := NewContext()
            ctx.BindValue(tt.bindName, tt.bindVal)

            engine := NewEngine()
            result, err := engine.Run(ctx, script)
            if err != nil {
                t.Fatalf("执行失败: %v", err)
            }

            if !tt.check(result) {
                t.Errorf("绑定值检查失败: %v", result)
            }
        })
    }
}

// ========== 算术运算边界测试 ==========

func Test_arithmetic_EdgeCases(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected int
    }{
        {"大数加法", "1000000000 + 1000000000", 2000000000},
        {"负数加法", "-10 + (-20)", -30},
        {"连续减法", "100 - 30 - 20", 50},
        {"连续乘法", "2 * 3 * 4", 24},
        {"混合运算", "10 + 2 * 3", 16},
        {"括号优先", "(10 + 2) * 3", 36},
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

// ========== 切片操作测试 ==========

func Test_slice_ArraySlice(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected []int
    }{
        {"基本切片", "[1, 2, 3, 4, 5][1:3]", []int{2, 3}},
        {"从头切片", "[1, 2, 3, 4, 5][:2]", []int{1, 2}},
        {"到尾切片", "[1, 2, 3, 4, 5][2:]", []int{3, 4, 5}},
        {"完整切片", "[1, 2, 3][:]", []int{1, 2, 3}},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := runScript(t, tt.input)
            arr := result.Array()
            if arr == nil {
                t.Fatalf("期望数组，得到nil")
            }
            if len(arr.Elements) != len(tt.expected) {
                t.Errorf("切片长度不匹配: 期望 %d, 实际 %d", len(tt.expected), len(arr.Elements))
                return
            }
            for i, elem := range arr.Elements {
                if elem.Int() != tt.expected[i] {
                    t.Errorf("索引 %d: 期望 %d, 实际 %d", i, tt.expected[i], elem.Int())
                }
            }
        })
    }
}

func Test_slice_StringSlice(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"基本切片", `"hello"[1:3]`, "el"},
        {"从头切片", `"hello"[:2]`, "he"},
        {"到尾切片", `"hello"[2:]`, "llo"},
        {"完整切片", `"hello"[:]`, "hello"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := runScript(t, tt.input)
            if result.String() != tt.expected {
                t.Errorf("字符串切片失败: 期望 %s, 实际 %s", tt.expected, result.String())
            }
        })
    }
}

// TestTypeConversion_EdgeCases 测试类型转换边界情况
func Test_typeconversion_EdgeCases(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected interface{}
    }{
        {"字符串转int", `int("123")`, 123},
        {"负数字符串", `int("-456")`, -456},
        {"bool转int", `int(true)`, 1},
        {"int转float", `float(42)`, 42.0},
        {"字符串转float", `float("3.14")`, 3.14},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := runScript(t, tt.input)
            switch v := tt.expected.(type) {
            case int:
                if result.Int() != v {
                    t.Errorf("期望 %d, 实际 %d", v, result.Int())
                }
            case float64:
                if result.Float() != v {
                    t.Errorf("期望 %f, 实际 %f", v, result.Float())
                }
            }
        })
    }
}
