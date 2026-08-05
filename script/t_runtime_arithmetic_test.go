package script

import (
    "testing"
)

// ========== 算术运算测试 ==========

func Test_runtime_Arithmetic(t *testing.T) {
    vm := newTestVM(t)

    tests := []struct {
        name     string
        a, b     Value
        op       string
        expected interface{}
    }{
        {"加法-int", NewValue(10), NewValue(20), "+", 30},
        {"加法-float", NewValue(10.5), NewValue(20.3), "+", 30.8},
        {"加法-string", NewValue("hello"), NewValue(" world"), "+", "hello world"},
        {"减法-int", NewValue(30), NewValue(10), "-", 20},
        {"乘法-int", NewValue(10), NewValue(5), "*", 50},
        {"除法-int", NewValue(20), NewValue(4), "/", 5},
        {"取模-int", NewValue(17), NewValue(5), "%", 2},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var result Value
            var err error

            switch tt.op {
            case "+":
                result, err = vm.add(tt.a, tt.b)
            case "-":
                result, err = vm.sub(tt.a, tt.b)
            case "*":
                result, err = vm.mul(tt.a, tt.b)
            case "/":
                result, err = vm.div(tt.a, tt.b)
            case "%":
                result, err = vm.mod(tt.a, tt.b)
            }

            if err != nil {
                t.Errorf("运算失败: %v", err)
                return
            }

            switch v := tt.expected.(type) {
            case int:
                if result.Int() != v {
                    t.Errorf("期望 %d, 得到 %d", v, result.Int())
                }
            case float64:
                if result.Float() != v {
                    t.Errorf("期望 %f, 得到 %f", v, result.Float())
                }
            case string:
                if result.String() != v {
                    t.Errorf("期望 %s, 得到 %s", v, result.String())
                }
            }
        })
    }
}

func Test_runtime_AddStringOther(t *testing.T) {
    vm := newTestVM(t)

    tests := []struct {
        name     string
        a, b     Value
        expected string
    }{
        {"string+int", NewValue("value:"), NewValue(42), "value:42"},
        {"string+float", NewValue("pi:"), NewValue(3.14), "pi:3.14"},
        {"string+bool", NewValue("flag:"), NewValue(true), "flag:true"},
        {"string+nil", NewValue("val:"), NewValue(nil), "val:nil"},
        {"int+string", NewValue(42), NewValue(" items"), "42 items"},
        {"float+string", NewValue(3.14), NewValue(" pi"), "3.14 pi"},
        {"bool+string", NewValue(true), NewValue(" flag"), "true flag"},
        {"nil+string", NewValue(nil), NewValue(" value"), "nil value"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := vm.add(tt.a, tt.b)
            if err != nil {
                t.Errorf("运算失败: %v", err)
                return
            }
            if result.String() != tt.expected {
                t.Errorf("期望 %s, 得到 %s", tt.expected, result.String())
            }
        })
    }
}

func Test_runtime_AddArrayArray(t *testing.T) {
    vm := newTestVM(t)

    arr1 := Value{Type: TypeArray, Data: &ArrayValue{Elements: []Value{NewValue(1), NewValue(2)}}}
    arr2 := Value{Type: TypeArray, Data: &ArrayValue{Elements: []Value{NewValue(3), NewValue(4)}}}

    result, err := vm.add(arr1, arr2)
    if err != nil {
        t.Errorf("数组连接失败: %v", err)
        return
    }

    resultArr := result.Array()
    if len(resultArr.Elements) != 4 {
        t.Errorf("期望数组长度4, 得到%d", len(resultArr.Elements))
    }
}

func Test_runtime_ConcatArrays(t *testing.T) {
    vm := newTestVM(t)

    arr1 := Value{Type: TypeArray, Data: &ArrayValue{Elements: []Value{NewValue(1), NewValue(2)}}}
    arr2 := Value{Type: TypeArray, Data: &ArrayValue{Elements: []Value{NewValue(3), NewValue(4)}}}

    result := vm.concatArrays(arr1, arr2)
    elements := result.Array().Elements

    if len(elements) != 4 {
        t.Errorf("期望4个元素，得到%d", len(elements))
    }
    if elements[0].Int() != 1 || elements[3].Int() != 4 {
        t.Errorf("数组连接结果不正确")
    }
}

func Test_runtime_Neg_Complete(t *testing.T) {
    vm := newTestVM(t)

    tests := []struct {
        name     string
        input    Value
        expected interface{}
    }{
        {"int取负", NewValue(10), -10},
        {"int零取负", NewValue(0), 0},
        {"float取负", NewValue(3.14), -3.14},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := vm.neg(tt.input)
            if err != nil {
                t.Errorf("不应返回错误: %v", err)
            }
            switch v := tt.expected.(type) {
            case int:
                if result.Int() != v {
                    t.Errorf("期望%d, 得到%d", v, result.Int())
                }
            case float64:
                if result.Float() != v {
                    t.Errorf("期望%f, 得到%f", v, result.Float())
                }
            }
        })
    }
}

func Test_runtime_Bitwise(t *testing.T) {
    vm := newTestVM(t)

    tests := []struct {
        name     string
        a, b     Value
        op       string
        expected int
    }{
        {"位与", NewValue(0xFF), NewValue(0x0F), "&", 0x0F},
        {"位或", NewValue(0xF0), NewValue(0x0F), "|", 0xFF},
        {"位异或", NewValue(0xFF), NewValue(0x0F), "^", 0xF0},
        {"左移", NewValue(1), NewValue(4), "<<", 16},
        {"右移", NewValue(16), NewValue(2), ">>", 4},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var result Value
            var err error

            switch tt.op {
            case "&":
                result, err = vm.bitAnd(tt.a, tt.b)
            case "|":
                result, err = vm.bitOr(tt.a, tt.b)
            case "^":
                result, err = vm.bitXor(tt.a, tt.b)
            case "<<":
                result, err = vm.lshift(tt.a, tt.b)
            case ">>":
                result, err = vm.rshift(tt.a, tt.b)
            }

            if err != nil {
                t.Errorf("运算失败: %v", err)
                return
            }

            if result.Int() != tt.expected {
                t.Errorf("期望 %d, 得到 %d", tt.expected, result.Int())
            }
        })
    }
}

func Test_runtime_BitNot_Complete(t *testing.T) {
    vm := newTestVM(t)

    tests := []struct {
        name     string
        input    int
        expected int
    }{
        {"正常位取反", 0xFF, ^0xFF},
        {"零取反", 0, ^0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := vm.bitNot(NewValue(tt.input))
            if err != nil {
                t.Errorf("不应返回错误: %v", err)
            }
            if result.Int() != tt.expected {
                t.Errorf("期望%d, 得到%d", tt.expected, result.Int())
            }
        })
    }
}

func Test_runtime_TypeConversion(t *testing.T) {
    vm := newTestVM(t)

    result, err := vm.toInt(NewValue(3.14))
    if err != nil || result.Int() != 3 {
        t.Errorf("int转换失败: %v", err)
    }

    result, err = vm.toFloat(NewValue(10))
    if err != nil || result.Float() != 10.0 {
        t.Errorf("float转换失败: %v", err)
    }
}

func Test_runtime_ToInt_Complete(t *testing.T) {
    vm := newTestVM(t)

    tests := []struct {
        name     string
        input    Value
        expected int
    }{
        {"int直接返回", NewValue(42), 42},
        {"有效数字字符串", NewValue("123"), 123},
        {"负数字符串", NewValue("-456"), -456},
        {"bool true转1", NewValue(true), 1},
        {"bool false转0", NewValue(false), 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := vm.toInt(tt.input)
            if err != nil {
                t.Errorf("不应返回错误: %v", err)
            }
            if result.Int() != tt.expected {
                t.Errorf("期望%d, 得到%d", tt.expected, result.Int())
            }
        })
    }
}

func Test_runtime_ToFloat_Complete(t *testing.T) {
    vm := newTestVM(t)

    tests := []struct {
        name     string
        input    Value
        expected float64
    }{
        {"float直接返回", NewValue(3.14), 3.14},
        {"有效浮点字符串", NewValue("2.5"), 2.5},
        {"bool true转1.0", NewValue(true), 1.0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := vm.toFloat(tt.input)
            if err != nil {
                t.Errorf("不应返回错误: %v", err)
            }
            if result.Float() != tt.expected {
                t.Errorf("期望%f, 得到%f", tt.expected, result.Float())
            }
        })
    }
}
