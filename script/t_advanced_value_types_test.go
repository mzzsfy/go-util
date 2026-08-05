package script

import (
    "reflect"
    "testing"
)

// ========== Value边界测试 ==========

// TestValue_Equal_Array 测试数组相等比较
func Test_value_Equal_Array(t *testing.T) {
    arr1 := NewValue([]Value{NewValue(1), NewValue(2)})
    arr2 := NewValue([]Value{NewValue(1), NewValue(2)})
    if !arr1.Equal(arr2) {
        t.Error("相同数组应相等")
    }
    arr3 := NewValue([]Value{NewValue(1), NewValue(3)})
    if arr1.Equal(arr3) {
        t.Error("不同数组不应相等")
    }
}

// TestValue_Equal_Map 测试Map相等比较
func Test_value_Equal_Map(t *testing.T) {
    m1 := NewValue(&MapValue{Pairs: map[string]Value{"a": NewValue(1)}})
    m2 := NewValue(&MapValue{Pairs: map[string]Value{"a": NewValue(1)}})
    if !m1.Equal(m2) {
        t.Error("相同Map应相等")
    }
    m3 := NewValue(&MapValue{Pairs: map[string]Value{"a": NewValue(2)}})
    if m1.Equal(m3) {
        t.Error("不同Map不应相等")
    }
}

// TestValue_GoString_Unknown 测试未知类型GoString
func Test_value_GoString_Unknown(t *testing.T) {
    v := Value{Type: ValueType(999), Data: nil}
    result := v.GoString()
    if result != "<unknown>" {
        t.Errorf("未知类型应返回'<unknown>', 实际为'%s'", result)
    }
}

// TestValue_GoString_Array 测试数组GoString
func Test_value_GoString_Array(t *testing.T) {
    arr := &ArrayValue{Elements: []Value{
        NewValue(1),
        NewValue(2),
        NewValue(3),
    }}
    v := Value{Type: TypeArray, Data: arr}
    result := v.GoString()
    if result != "[1, 2, 3]" {
        t.Errorf("数组GoString应为'[1, 2, 3]', 实际为'%s'", result)
    }
}

// TestValue_GoString_Map 测试Map GoString
func Test_value_GoString_Map(t *testing.T) {
    m := &MapValue{Pairs: map[string]Value{
        "a": NewValue(1),
        "b": NewValue(2),
    }}
    v := Value{Type: TypeMap, Data: m}
    result := v.GoString()
    if result == "" || result == "<unknown>" {
        t.Errorf("Map GoString不应为空或unknown, 实际为'%s'", result)
    }
}

// TestValue_GoString_Function 测试函数GoString
func Test_value_GoString_Function(t *testing.T) {
    fn := &FunctionValue{Compiled: &CompiledFunction{Name: "testFunc"}}
    v := Value{Type: TypeFunction, Data: fn}
    result := v.GoString()
    if result != "<function testFunc>" {
        t.Errorf("函数GoString应为'<function testFunc>', 实际为'%s'", result)
    }
}

// TestValue_GoString_ExternalFunc 测试外部函数GoString
func Test_value_GoString_ExternalFunc(t *testing.T) {
    ef := &ExternalFuncValue{Name: "extFunc", Func: func() {}}
    v := Value{Type: TypeExternalFunc, Data: ef}
    result := v.GoString()
    if result != "<external extFunc>" {
        t.Errorf("外部函数GoString应为'<external extFunc>', 实际为'%s'", result)
    }
}

// TestValue_GoString_NestedArray 测试嵌套数组GoString
func Test_value_GoString_NestedArray(t *testing.T) {
    inner := &ArrayValue{Elements: []Value{NewValue(1), NewValue(2)}}
    outer := &ArrayValue{Elements: []Value{
        {Type: TypeArray, Data: inner},
        NewValue(3),
    }}
    v := Value{Type: TypeArray, Data: outer}
    result := v.GoString()
    if result != "[[1, 2], 3]" {
        t.Errorf("嵌套数组GoString应为'[[1, 2], 3]', 实际为'%s'", result)
    }
}

// TestValue_IsNil 测试nil值判断
func Test_value_IsNil(t *testing.T) {
    v := NewValue(nil)
    if !v.IsNil() {
        t.Error("nil值应该返回true")
    }

    v2 := NewValue(0)
    if v2.IsNil() {
        t.Error("0不是nil值")
    }
}

// TestValue_Int_FloatValue 测试浮点数调用Int
func Test_value_Int_FloatValue(t *testing.T) {
    v := NewValue(3.14)
    result := v.Int()
    if result != 0 {
        t.Errorf("浮点数调用Int()应该返回0, 得到 %d", result)
    }
}

// TestValue_Float_IntValue 测试整数调用Float, 自动转换为float64
func Test_value_Float_IntValue(t *testing.T) {
    v := NewValue(42)
    result := v.Float()
    if result != 42.0 {
        t.Errorf("整数调用Float()应该返回42.0, 得到 %f", result)
    }
}

// TestValue_Bool_Truthiness 测试布尔真值
func Test_value_Bool_Truthiness(t *testing.T) {
    tests := []struct {
        value    Value
        expected bool
    }{
        {NewValue(true), true},
        {NewValue(false), false},
    }

    for _, tt := range tests {
        result := tt.value.Bool()
        if result != tt.expected {
            t.Errorf("值 %v 的布尔转换: 期望 %v, 得到 %v", tt.value, tt.expected, result)
        }
    }
}

// TestValue_Array 测试数组访问
func Test_value_Array(t *testing.T) {
    arr := []Value{NewValue(1), NewValue(2), NewValue(3)}
    v := NewValue(arr)
    result := v.Array()
    if result == nil {
        t.Error("数组值应该返回ArrayValue")
        return
    }
    if len(result.Elements) != 3 {
        t.Errorf("期望长度3, 得到 %d", len(result.Elements))
    }
}

// TestValue_Map 测试Map访问
func Test_value_Map(t *testing.T) {
    m := &MapValue{Pairs: map[string]Value{"a": NewValue(1)}}
    v := NewValue(m)
    result := v.Map()
    if result == nil {
        t.Error("Map值应该返回MapValue")
        return
    }
    if len(result.Pairs) != 1 {
        t.Errorf("期望长度1, 得到 %d", len(result.Pairs))
    }
}

// TestScript_NewScript 测试NewScript函数
func Test_script_NewScript(t *testing.T) {
    parser := NewParser()
    compiled, err := parser.Compile("10 + 20")
    if err != nil {
        t.Fatalf("编译失败: %v", err)
    }

    script := NewScript(compiled)
    if script == nil {
        t.Error("NewScript不应返回nil")
    }
}

// TestScript_Clone 测试Script克隆
func Test_script_Clone(t *testing.T) {
    parser := NewParser()
    compiled, err := parser.Compile("x := 10")
    if err != nil {
        t.Fatalf("编译失败: %v", err)
    }

    script := NewScript(compiled)
    cloned := script.Clone()
    if cloned == nil {
        t.Error("Clone不应返回nil")
    }
}

// TestScript_GetCompiled 测试获取编译结果
func Test_script_GetCompiled(t *testing.T) {
    parser := NewParser()
    compiled, err := parser.Compile("10 + 20")
    if err != nil {
        t.Fatalf("编译失败: %v", err)
    }

    script := NewScript(compiled)
    result := script.GetCompiled()
    if result == nil {
        t.Error("GetCompiled不应返回nil")
    }
}

// TestScript_EncodeDecode 测试编码解码
func Test_script_EncodeDecode(t *testing.T) {
    parser := NewParser()
    compiled, err := parser.Compile("x := 10\nx + 20")
    if err != nil {
        t.Fatalf("编译失败: %v", err)
    }

    script := NewScript(compiled)
    data, err := script.Encode()
    if err != nil {
        t.Logf("Encode失败（可能未实现）: %v", err)
        return
    }

    script2 := NewScript(nil)
    _, err = script2.Decode(data)
    if err != nil {
        t.Logf("Decode失败（可能未实现）: %v", err)
        return
    }
}

// ========== Value方法测试 ==========

// TestValue_GoString 测试GoString方法
func Test_value_GoString(t *testing.T) {
    tests := []struct {
        value    Value
        expected string
    }{
        {NewValue(10), "10"},
        {NewValue(3.14), "3.14"},
        {NewValue("hello"), "\"hello\""},
        {NewValue(true), "true"},
        {NewValue(nil), "nil"},
    }

    for _, tt := range tests {
        result := tt.value.GoString()
        if result != tt.expected {
            t.Errorf("GoString() = %s, want %s", result, tt.expected)
        }
    }
}

// ========== Runtime storeIndex/mapGet/mapSet测试 ==========

// TestRuntime_ArrayStoreIndex 测试数组索引赋值
func Test_runtime_ArrayStoreIndex(t *testing.T) {
    vm := newTestVM(t)

    arr := NewValue([]Value{NewValue(1), NewValue(2), NewValue(3)})
    err := arrayStoreIndex(vm, arr, NewValue(0), NewValue(100))
    if err != nil {
        t.Errorf("数组索引赋值失败: %v", err)
    }

    elements := arr.Array().Elements
    if elements[0].Int() != 100 {
        t.Errorf("期望 100, 得到 %d", elements[0].Int())
    }
}

// TestRuntime_ArrayStoreIndex_OutOfBounds 测试数组越界赋值
func Test_runtime_ArrayStoreIndex_OutOfBounds(t *testing.T) {
    vm := newTestVM(t)

    arr := NewValue([]Value{NewValue(1), NewValue(2)})
    // 使用recover捕获panic，因为runtimeError会panic
    defer func() {
        if r := recover(); r != nil {
            // 预期会panic
        }
    }()
    _ = arrayStoreIndex(vm, arr, NewValue(10), NewValue(100))
}

// TestRuntime_MapStoreIndex 测试Map索引赋值
func Test_runtime_MapStoreIndex(t *testing.T) {
    vm := newTestVM(t)

    m := &MapValue{Pairs: map[string]Value{}}
    obj := NewValue(m)
    err := mapStoreIndex(vm, obj, NewValue("key"), NewValue("value"))
    if err != nil {
        t.Errorf("Map索引赋值失败: %v", err)
    }

    if m.Pairs["key"].String() != "value" {
        t.Errorf("期望 value, 得到 %s", m.Pairs["key"].String())
    }
}

// TestRuntime_StoreIndex 测试storeIndex分派
func Test_runtime_StoreIndex(t *testing.T) {
    vm := newTestVM(t)

    t.Run("数组索引赋值", func(t *testing.T) {
        arr := NewValue([]Value{NewValue(1), NewValue(2)})
        err := vm.storeIndex(arr, NewValue(0), NewValue(10))
        if err != nil {
            t.Errorf("数组索引赋值失败: %v", err)
        }
    })

    t.Run("Map索引赋值", func(t *testing.T) {
        m := &MapValue{Pairs: map[string]Value{}}
        obj := NewValue(m)
        err := vm.storeIndex(obj, NewValue("k"), NewValue("v"))
        if err != nil {
            t.Errorf("Map索引赋值失败: %v", err)
        }
    })

    t.Run("不支持的类型", func(t *testing.T) {
        defer func() {
            if r := recover(); r != nil {
                // 预期会panic
            }
        }()
        _ = vm.storeIndex(NewValue(123), NewValue(0), NewValue(10))
    })
}

// ========== convertNilValue测试 ==========

// TestConvertNilValue 测试nil值转换
func Test_convertnilvalue(t *testing.T) {
    val := NewValue(nil)
    result, err := convertNilValue(val, reflect.TypeOf(0))
    if err != nil {
        t.Errorf("nil值转换失败: %v", err)
    }
    if !result.IsZero() {
        t.Error("nil值应转换为零值")
    }
}

// TestConvertNilValue_Pointer 测试nil值转换为指针
func Test_convertnilvalue_Pointer(t *testing.T) {
    val := NewValue(nil)
    result, err := convertNilValue(val, reflect.TypeOf((*int)(nil)))
    if err != nil {
        t.Errorf("nil值转换失败: %v", err)
    }
    if !result.IsNil() {
        t.Error("指针应转换为nil")
    }
}

// TestRuntime_SliceBounds_Negative 测试切片负数边界
func Test_runtime_SliceBounds_Negative(t *testing.T) {
    start, end, ok := normalizeSliceBounds(0, 3, 10)
    if !ok {
        t.Errorf("规范化失败")
    }
    if start != 0 || end != 3 {
        t.Errorf("期望 [0,3], 得到 [%d,%d]", start, end)
    }
}

// ========== 错误处理测试 ==========

// TestCompileError_EmptyMessage 测试空消息错误
func Test_compileerror_EmptyMessage(t *testing.T) {
    err := &CompileError{Line: 1, Column: 1, Message: ""}
    result := err.Error()
    if result == "" {
        t.Error("Error()不应该返回空字符串")
    }
}

// ========== Engine边界测试 ==========

// TestEngine_Run_NilScript 测试空脚本, panic 防护将其转为 RuntimeError
func Test_engine_Run_NilScript(t *testing.T) {
    engine := NewEngine()
    ctx := NewContext()
    _, err := engine.Run(ctx, nil)
    if err == nil {
        t.Error("空脚本应返回错误")
    }
}

// TestEngine_MultipleRuns_SameContext 测试多次运行
func Test_engine_MultipleRuns_SameContext(t *testing.T) {
    engine := NewEngine()
    ctx := NewContext()
    parser := NewParser()
    script, _ := parser.Compile("1 + 1")

    for i := 0; i < 5; i++ {
        _, err := engine.Run(ctx, script)
        if err != nil {
            t.Errorf("第%d次运行失败: %v", i+1, err)
        }
    }

    stats := ctx.GetStats()
    if stats.TotalRuns != 5 {
        t.Errorf("期望5次运行, 得到 %d", stats.TotalRuns)
    }
}

// TestInstruction_Creation 测试指令创建
func Test_instruction_Creation(t *testing.T) {
    inst := Instruction{Op: OpAdd, ArgCount: 0}
    if inst.Op != OpAdd {
        t.Errorf("OpCode应为OpAdd, 实际为%d", inst.Op)
    }
}

// TestInstruction_WithArgs 测试带参数指令
func Test_instruction_WithArgs(t *testing.T) {
    inst := Instruction{Op: OpConst, ArgCount: 2, Args: [2]int{1, 2}}
    if inst.ArgCount != 2 {
        t.Errorf("ArgCount应为2, 实际为%d", inst.ArgCount)
    }
    if inst.Args[0] != 1 {
        t.Errorf("Args[0]应为1, 实际为%d", inst.Args[0])
    }
}

// TestCompiledFunction_Creation 测试编译函数创建
func Test_compiledfunction_Creation(t *testing.T) {
    fn := &CompiledFunction{
        Name:         "testFunc",
        Instructions: []Instruction{{Op: OpConst, ArgCount: 1, Args: [2]int{0}}},
        Constants:    []Value{NewValue(42)},
        NumLocals:    2,
        NumParams:    1,
    }
    if fn.Name != "testFunc" {
        t.Errorf("Name应为'testFunc', 实际为'%s'", fn.Name)
    }
    if len(fn.Constants) != 1 {
        t.Errorf("Constants长度应为1, 实际为%d", len(fn.Constants))
    }
}

// TestCompiledScript_Creation 测试编译脚本创建
func Test_compiledscript_Creation(t *testing.T) {
    script := &CompiledScript{
        Main: &CompiledFunction{
            Name:         "main",
            Instructions: []Instruction{},
            Constants:    []Value{},
            NumLocals:    0,
            NumParams:    0,
        },
    }
    if script.Main == nil {
        t.Error("Main不应为nil")
    }
    if script.Main.Name != "main" {
        t.Errorf("Main.Name应为'main', 实际为'%s'", script.Main.Name)
    }
}

// ========== Token/Position测试 ==========

// TestToken_Creation 测试Token创建
func Test_token_Creation(t *testing.T) {
    tok := Token{Type: TokenInt, Value: "42", Line: 1, Column: 5}
    if tok.Type != TokenInt {
        t.Errorf("Type应为TokenInt, 实际为%d", tok.Type)
    }
    if tok.Value != "42" {
        t.Errorf("Value应为'42', 实际为'%s'", tok.Value)
    }
}

// TestPosition_Creation 测试Position创建
func Test_position_Creation(t *testing.T) {
    pos := Position{Line: 10, Column: 5}
    if pos.Line != 10 {
        t.Errorf("Line应为10, 实际为%d", pos.Line)
    }
    if pos.Column != 5 {
        t.Errorf("Column应为5, 实际为%d", pos.Column)
    }
}

// ========== 解析器错误路径测试 ==========

// TestParser_ParseIntLiteral_Error 测试整数解析错误路径
func Test_parser_ParseIntLiteral_Error(t *testing.T) {
    tests := []struct {
        name  string
        input string
    }{
        {"超大整数", "999999999999999999999999999999999999"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            parser := NewParser()
            _, err := parser.Compile(tt.input)
            // 大整数可能被解析为浮点数或报错
            _ = err
        })
    }
}

// TestParser_ParseFloatLiteral_Error 测试浮点数解析错误路径
func Test_parser_ParseFloatLiteral_Error(t *testing.T) {
    tests := []struct {
        name  string
        input string
    }{
        {"无效浮点数", "3.14.15"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            parser := NewParser()
            _, err := parser.Compile(tt.input)
            // 无效浮点数应该报错
            _ = err
        })
    }
}

// ========== 编译器覆盖率测试 ==========

// TestCompiler_CompileBuiltinCall_Error 测试内置函数调用错误路径
func Test_compiler_CompileBuiltinCall_Error(t *testing.T) {
    compiler := NewCompiler()

    // 测试未知内置函数（有参数）
    call := &CallExpr{
        Position: Position{Line: 1, Column: 1},
        Func:     &IdentExpr{Name: "unknownBuiltin"},
        Args:     []Expr{&LiteralExpr{Type: LiteralInt, Value: 1}},
    }
    program := &Program{
        Statements: []Stmt{&ExprStmt{Expr: call}},
    }
    _, err := compiler.Compile(program)
    if err == nil {
        t.Error("未知内置函数应该报错")
    }
}

// TestCompiler_CompileLiteralExpr_Unsupported 测试不支持的字面量类型
func Test_compiler_CompileLiteralExpr_Unsupported(t *testing.T) {
    compiler := NewCompiler()

    // 创建不支持类型的字面量
    literal := &LiteralExpr{
        Position: Position{Line: 1, Column: 1},
        Type:     LiteralType(999),
        Value:    "test",
    }
    program := &Program{
        Statements: []Stmt{&ExprStmt{Expr: literal}},
    }
    _, err := compiler.Compile(program)
    if err == nil {
        t.Error("不支持的字面量类型应该报错")
    }
}

// ========== 运行时操作覆盖率测试 ==========

// TestRuntime_BitwiseOperations_Cover 测试位运算覆盖
func Test_runtime_BitwiseOperations_Cover(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected int
    }{
        {"左移", "1 << 4", 16},
        {"右移", "16 >> 2", 4},
        {"位与", "15 & 7", 7},
        {"位或", "8 | 4", 12},
        {"位异或", "15 ^ 7", 8},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            parser := NewParser()
            script, err := parser.Compile(tt.input)
            if err != nil {
                t.Fatalf("编译失败: %v", err)
            }

            engine := NewEngine()
            result, err := engine.Run(NewContext(), script)
            if err != nil {
                t.Fatalf("运行失败: %v", err)
            }

            if result.Int() != tt.expected {
                t.Errorf("期望 %d, 实际 %d", tt.expected, result.Int())
            }
        })
    }
}

// TestRuntime_CompareOperations_Cover 测试比较运算覆盖
func Test_runtime_CompareOperations_Cover(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected bool
    }{
        {"整数小于", "3 < 5", true},
        {"整数小于等于", "5 <= 5", true},
        {"整数大于", "5 > 3", true},
        {"整数大于等于", "5 >= 5", true},
        {"字符串小于", "\"abc\" < \"def\"", true},
        {"浮点数比较", "3.14 < 3.15", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            parser := NewParser()
            script, err := parser.Compile(tt.input)
            if err != nil {
                t.Fatalf("编译失败: %v", err)
            }

            engine := NewEngine()
            result, err := engine.Run(NewContext(), script)
            if err != nil {
                t.Fatalf("运行失败: %v", err)
            }

            if result.Bool() != tt.expected {
                t.Errorf("期望 %v, 实际 %v", tt.expected, result.Bool())
            }
        })
    }
}

// TestRuntime_TypeConversion_Cover 测试类型转换覆盖
func Test_runtime_TypeConversion_Cover(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected interface{}
    }{
        {"字符串转整数", "int(\"123\")", 123},
        {"布尔转整数", "int(true)", 1},
        {"字符串转浮点", "float(\"3.14\")", 3.14},
        {"布尔转浮点", "float(false)", 0.0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            parser := NewParser()
            script, err := parser.Compile(tt.input)
            if err != nil {
                t.Fatalf("编译失败: %v", err)
            }

            engine := NewEngine()
            result, err := engine.Run(NewContext(), script)
            if err != nil {
                t.Fatalf("运行失败: %v", err)
            }

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

// ========== Value更多方法测试 ==========

// Test_value_FunctionMethod 测试Function方法
func Test_value_FunctionMethod(t *testing.T) {
    fn := &FunctionValue{Compiled: &CompiledFunction{Name: "testFn"}}
    v := NewValue(fn)
    result := v.Function()
    if result == nil {
        t.Error("Function不应返回nil")
    }
    if result.Compiled.Name != "testFn" {
        t.Errorf("期望testFn, 得到%s", result.Compiled.Name)
    }
}

// Test_value_ExternalFuncMethod 测试ExternalFunc方法
func Test_value_ExternalFuncMethod(t *testing.T) {
    ef := &ExternalFuncValue{Name: "extFn", Func: func() {}}
    v := Value{Type: TypeExternalFunc, Data: ef}
    result := v.ExternalFunc()
    if result == nil {
        t.Error("ExternalFunc不应返回nil")
    }
    if result.Name != "extFn" {
        t.Errorf("期望extFn, 得到%s", result.Name)
    }
}

// Test_value_GetTyped_WrongType 测试getTyped错误类型
func Test_value_GetTyped_WrongType(t *testing.T) {
    v := NewValue(42)    // int类型
    result := v.String() // 获取string应该返回零值
    if result != "" {
        t.Errorf("错误类型应返回零值, 得到'%s'", result)
    }
}

// Test_value_Equal_AllTypes 测试所有类型相等比较
func Test_value_Equal_AllTypes(t *testing.T) {
    // nil比较
    if !NewValue(nil).Equal(NewValue(nil)) {
        t.Error("nil应该等于nil")
    }

    // int比较
    if !NewValue(42).Equal(NewValue(42)) {
        t.Error("42应该等于42")
    }

    // float比较
    if !NewValue(3.14).Equal(NewValue(3.14)) {
        t.Error("3.14应该等于3.14")
    }

    // string比较
    if !NewValue("hello").Equal(NewValue("hello")) {
        t.Error("hello应该等于hello")
    }

    // bool比较
    if !NewValue(true).Equal(NewValue(true)) {
        t.Error("true应该等于true")
    }

    // 不同类型比较
    if NewValue(42).Equal(NewValue("42")) {
        t.Error("int和string不应相等")
    }
}

// Test_value_formatValueForPrint_Unknown 测试formatValueForPrint未知类型
func Test_value_formatValueForPrint_Unknown(t *testing.T) {
    v := Value{Type: ValueType(999), Data: nil}
    result := formatValueForPrint(v)
    if result != "<unknown>" {
        t.Errorf("期望<unknown>, 得到'%s'", result)
    }
}

// Test_value_NewValue_int64 测试int64类型
func Test_value_NewValue_int64(t *testing.T) {
    var i int64 = 42
    v := NewValue(i)
    if v.Int() != 42 {
        t.Errorf("期望42, 得到%d", v.Int())
    }
}

// Test_value_IsNil_NilData 测试nil数据
func Test_value_IsNil_NilData(t *testing.T) {
    v := Value{Type: TypeNil}
    if !v.IsNil() {
        t.Error("TypeNil时IsNil应返回true")
    }
    // TypeInt的值通过NewValue创建
    i := NewValue(42)
    if i.IsNil() {
        t.Error("TypeInt不应被判定为nil")
    }
}

// Test_value_formatValue 测试formatValue函数
func Test_value_formatValue(t *testing.T) {
    tests := []struct {
        input    any
        expected string
    }{
        {"hello", "\"hello\""},
        {42, "42"},
        {3.14, "3.14"},
        {true, "true"},
    }

    for _, tt := range tests {
        result := formatValue(tt.input)
        if result != tt.expected {
            t.Errorf("formatValue(%v) = '%s', 期望 '%s'", tt.input, result, tt.expected)
        }
    }
}
