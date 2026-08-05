package script

import (
    "reflect"
    "testing"
)

// ========== 集成测试 ==========
// 测试辅助函数已移至 testutil_test.go

func Test_compiler_VarDecl(t *testing.T) {
    tests := []string{"x := 10", "x := 10.5", "x := \"hello\"", "x := true", "x := nil"}
    for _, tt := range tests {
        compileScript(t, tt)
    }
}

func Test_compiler_ArrayDecl(t *testing.T) {
    script := compileScript(t, "[1, 2, 3]")
    if len(script.Main.Instructions) == 0 {
        t.Error("应该生成指令")
    }
}

func Test_compiler_MapDecl(t *testing.T) {
    script := compileScript(t, `{"a": 1, "b": 2}`)
    if len(script.Main.Instructions) == 0 {
        t.Error("应该生成指令")
    }
}

func Test_compiler_BinaryExpr(t *testing.T) {
    tests := []string{"10 + 20", "10 - 5", "10 * 2", "10 / 2", "10 % 3", "10 == 20", "10 != 20"}
    for _, tt := range tests {
        compileScript(t, tt)
    }
}

func Test_compiler_IfStatement(t *testing.T) {
    input := `x := 15
if x > 10 { y := 20 }`
    script := compileScript(t, input)
    hasJump := false
    for _, inst := range script.Main.Instructions {
        if inst.Op == OpJumpIfFalse || inst.Op == OpJump {
            hasJump = true
            break
        }
    }
    if !hasJump {
        t.Error("if语句应该生成跳转指令")
    }
}

func Test_vm_Constants(t *testing.T) {
    t.Run("整数", func(t *testing.T) { runIntTest(t, "10", 10) })
    t.Run("浮点数", func(t *testing.T) { runFloatTest(t, "10.5", 10.5) })
    t.Run("字符串", func(t *testing.T) { runStringTest(t, "\"hello\"", "hello") })
    t.Run("布尔真", func(t *testing.T) { runBoolTest(t, "true", true) })
    t.Run("布尔假", func(t *testing.T) { runBoolTest(t, "false", false) })
}

func Test_vm_Arithmetic(t *testing.T) {
    tests := []struct {
        Input    string
        Expected int
    }{
        {"10 + 20", 30}, {"30 - 10", 20}, {"5 * 4", 20}, {"20 / 4", 5}, {"17 % 5", 2},
    }
    RunIntTestsSimple(t, tests)
}

func Test_vm_Array(t *testing.T) {
    result := runScript(t, "[1, 2, 3]")
    if result.Type != TypeArray {
        t.Errorf("期望数组类型，得到 %v", result.Type)
    }
}

func Test_vm_Map(t *testing.T) {
    result := runScript(t, `{"a": 1, "b": 2}`)
    if result.Type != TypeMap {
        t.Errorf("期望Map类型，得到 %v", result.Type)
    }
}

func Test_engine_Options(t *testing.T) {
    engine := NewEngine(WithMaxCallDepth(512))
    if engine.maxDepth != 512 {
        t.Errorf("maxDepth应为512")
    }
}

func Test_value_Equal(t *testing.T) {
    tests := []struct {
        Name     string
        A, B     Value
        Expected bool
    }{
        {"整数相等", NewValue(10), NewValue(10), true},
        {"整数不等", NewValue(10), NewValue(20), false},
        {"浮点相等", NewValue(3.14), NewValue(3.14), true},
        {"字符串相等", NewValue("hello"), NewValue("hello"), true},
    }
    for _, tt := range tests {
        t.Run(tt.Name, func(t *testing.T) {
            if tt.A.Equal(tt.B) != tt.Expected {
                t.Errorf("Equal比较错误")
            }
        })
    }
}

// ========== 错误测试 ==========

func Test_parser_InvalidExpression(t *testing.T) {
    runErrorTest(t, "1 +")
}

func Test_parser_UnterminatedArray(t *testing.T) {
    runErrorTest(t, "[1, 2, 3")
}

func Test_throw_Statement(t *testing.T) {
    runRuntimeErrorTest(t, `throw "test error"`)
}

func Test_map_IndexAccess(t *testing.T) {
    runIntTest(t, `m = {"a": 1, "b": 2}
m["a"]`, 1)
}

func Test_array_IndexOutOfBounds(t *testing.T) {
    result := runScript(t, `arr := [1, 2, 3]
arr[10]`)
    if !result.IsNil() {
        t.Errorf("越界访问期望nil，得到 %v", result)
    }
}

func Test_bitwise_Not(t *testing.T) {
    runIntTest(t, "^5", ^5)
}

func Test_convertvalue_UnsupportedType(t *testing.T) {
    val := Value{Type: TypeFunction, Data: nil}
    _, err := convertValueToGo(val, reflect.TypeOf(0))
    if err == nil {
        t.Error("不支持的类型转换应返回错误")
    }
}
