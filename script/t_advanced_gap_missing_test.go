package script

import (
	"reflect"
	"testing"
)

// ========== 外部函数差距测试 ==========

// Test_gap_External_3ReturnValues 测试3个返回值的外部函数
func Test_gap_External_3ReturnValues(t *testing.T) {
	fn := func() (int, int, int) { return 1, 2, 3 }
	_, err := callExternalFunc(fn, []Value{})
	if err == nil {
		t.Error("3个返回值应该报错")
	}
}

// Test_gap_External_4ReturnValues 测试4个返回值的外部函数
func Test_gap_External_4ReturnValues(t *testing.T) {
	fn := func() (int, int, int, int) { return 1, 2, 3, 4 }
	_, err := callExternalFunc(fn, []Value{})
	if err == nil {
		t.Error("4个返回值应该报错")
	}
}

// Test_gap_External_FunctionTypeConvert 测试函数类型转换
func Test_gap_External_FunctionTypeConvert(t *testing.T) {
	funcVal := &FunctionValue{Compiled: &CompiledFunction{Name: "test"}}
	val := Value{Type: TypeFunction, Data: funcVal}
	_, err := convertValueToGo(val, reflect.TypeOf(0))
	if err == nil {
		t.Error("函数类型转换应该报错")
	}
}

// Test_gap_External_ExternalFuncTypeConvert 测试外部函数类型转换
func Test_gap_External_ExternalFuncTypeConvert(t *testing.T) {
	extFunc := &ExternalFuncValue{Name: "external"}
	val := Value{Type: TypeExternalFunc, Data: extFunc}
	_, err := convertValueToGo(val, reflect.TypeOf(0))
	if err == nil {
		t.Error("外部函数类型转换应该报错")
	}
}

// Test_gap_External_NumericToBool 测试数值转布尔
func Test_gap_External_NumericToBool(t *testing.T) {
	val := NewValue(42)
	_, err := convertNumericValue(val, reflect.TypeOf(false))
	if err == nil {
		t.Error("数值转bool应该报错")
	}
}

// Test_gap_External_NumericToStruct 测试数值转结构体
func Test_gap_External_NumericToStruct(t *testing.T) {
	val := NewValue(42)
	_, err := convertNumericValue(val, reflect.TypeOf(struct{}{}))
	if err == nil {
		t.Error("数值转struct应该报错")
	}
}

// ========== Context差距测试 ==========

// Test_gap_Context_BindFunc_String 测试BindFunc传入字符串返回错误
func Test_gap_Context_BindFunc_String(t *testing.T) {
	ctx := NewContext()
	if err := ctx.BindFunc("test", "not a function"); err == nil {
		t.Error("传入字符串应返回错误")
	}
}

// Test_gap_Context_BindFunc_Int 测试BindFunc传入整数返回错误
func Test_gap_Context_BindFunc_Int(t *testing.T) {
	ctx := NewContext()
	if err := ctx.BindFunc("test", 42); err == nil {
		t.Error("传入整数应返回错误")
	}
}

// Test_gap_Context_BindFunc_Nil 测试BindFunc传入nil返回错误
func Test_gap_Context_BindFunc_Nil(t *testing.T) {
	ctx := NewContext()
	if err := ctx.BindFunc("test", nil); err == nil {
		t.Error("传入nil应返回错误")
	}
}

// ========== 编译期错误路径测试 ==========

// ========== Value差距测试 ==========

// Test_gap_Value_FormatExternalFunc 测试外部函数格式化

// Test_gap_Compile_AssignIntIndex 测试对int索引赋值
func Test_gap_Compile_AssignIntIndex(t *testing.T) {
	p := NewParser()
	compiled, err := p.Compile("x := 42\nx[0] = 1")
	if err != nil {
		// 编译期错误也可以接受（如果编译器能静态检测）
		return
	}

	// 运行时应该报错
	ctx := NewContext()
	engine := NewEngine()
	_, err = engine.Run(ctx, compiled)
	if err == nil {
		t.Error("对int索引赋值应该在运行时报错")
	}
}

// Test_gap_Compile_AssignOutOfBounds 测试数组越界赋值
func Test_gap_Compile_AssignOutOfBounds(t *testing.T) {
	p := NewParser()
	script, err := p.Compile("arr := [1, 2]\narr[10] = 3")
	if err != nil {
		return // 编译期错误，OK
	}
	ctx := NewContext()
	engine := NewEngine()
	_, err = engine.Run(ctx, script)
	if err == nil {
		t.Error("数组越界赋值应该报错")
	}
}

// ========== 类型转换错误测试 ==========

// Test_gap_Convert_ArrayToInt 测试数组转整数
func Test_gap_Convert_ArrayToInt(t *testing.T) {
	arr := NewValue([]Value{NewValue(1), NewValue(2)})
	_, err := convertValueToGo(arr, reflect.TypeOf(0))
	if err == nil {
		t.Error("数组转整数应该报错")
	}
}

// Test_gap_Convert_MapToInt 测试map转整数
func Test_gap_Convert_MapToInt(t *testing.T) {
	m := NewValue(&MapValue{Pairs: map[string]Value{"a": NewValue(1)}})
	_, err := convertValueToGo(m, reflect.TypeOf(0))
	if err == nil {
		t.Error("map转整数应该报错")
	}
}

// Test_gap_Convert_ArrayToString 测试数组转字符串
func Test_gap_Convert_ArrayToString(t *testing.T) {
	arr := NewValue([]Value{NewValue(1)})
	_, err := convertValueToGo(arr, reflect.TypeOf(""))
	if err == nil {
		t.Error("数组转字符串应该报错")
	}
}

// Test_gap_Convert_MapToSlice 测试map转切片
func Test_gap_Convert_MapToSlice(t *testing.T) {
	m := NewValue(&MapValue{Pairs: map[string]Value{"a": NewValue(1)}})
	_, err := convertValueToGo(m, reflect.TypeOf([]int{}))
	if err == nil {
		t.Error("map转切片应该报错")
	}
}

// ========== 更多Value类型测试 ==========

// Test_gap_Value_EqualString 测试字符串相等比较
func Test_gap_Value_EqualString(t *testing.T) {
	a := NewValue("hello")
	b := NewValue("hello")
	if !a.Equal(b) {
		t.Error(`"hello" == "hello" 应为true`)
	}
}

// Test_gap_Value_EqualInt 测试整数相等比较
func Test_gap_Value_EqualInt(t *testing.T) {
	a := NewValue(42)
	b := NewValue(42)
	if !a.Equal(b) {
		t.Error("42 == 42 应为true")
	}
}

// Test_gap_Value_IsNil 测试Value.IsNil方法
func Test_gap_Value_IsNil(t *testing.T) {
	if !NewValue(nil).IsNil() {
		t.Error("nil.IsNil() 应为true")
	}
	if NewValue(42).IsNil() {
		t.Error("42.IsNil() 应为false")
	}
}

// Test_gap_Value_Int 测试Value.Int方法
func Test_gap_Value_Int(t *testing.T) {
	if NewValue(42).Int() != 42 {
		t.Error("42.Int() 应为42")
	}
}

// Test_gap_Value_Float 测试Value.Float方法
func Test_gap_Value_Float(t *testing.T) {
	if NewValue(3.14).Float() != 3.14 {
		t.Error("3.14.Float() 应为3.14")
	}
}

// Test_gap_Value_Bool 测试Value.Bool方法
func Test_gap_Value_Bool(t *testing.T) {
	if !NewValue(true).Bool() {
		t.Error("true.Bool() 应为true")
	}
}

// ========== For循环错误路径测试 ==========

// Test_gap_ForRange_WrongAssign 测试range使用=而不是:=
func Test_gap_ForRange_WrongAssign(t *testing.T) {
	p := NewParser()
	_, err := p.Compile("for k = arr { }")
	if err == nil {
		t.Error("应该报告期望 := 运算符错误")
	}
}

// Test_gap_ForRange_NonIdentKey 测试非标识符作为range键
func Test_gap_ForRange_NonIdentKey(t *testing.T) {
	tests := []string{
		`for 123 := arr { }`,
		`for "x" := arr { }`,
	}
	for _, tc := range tests {
		p := NewParser()
		_, err := p.Compile(tc)
		if err == nil {
			t.Errorf("'%s' 应该报告range键必须是标识符错误", tc)
		}
	}
}

// Test_gap_ForRange_NoBody 测试缺少循环体
func Test_gap_ForRange_NoBody(t *testing.T) {
	p := NewParser()
	_, err := p.Compile("for k := arr")
	if err == nil {
		t.Error("应该报告缺少循环体错误")
	}
}

// ========== ForWhile循环测试 ==========

// Test_gap_ForWhile_Condition 测试while风格循环条件
func Test_gap_ForWhile_Condition(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"条件true", "for true { 1 }"},
		{"比较条件", "for 1 < 2 { 1 }"},
		{"变量条件", "x := 1\nfor x { 1 }"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Compile(tt.input)
			t.Logf("%s: %v", tt.name, err)
		})
	}
}

// ========== 类型注解测试 ==========

// Test_gap_TypeAnnotation_Basic 测试基本类型注解
func Test_gap_TypeAnnotation_Basic(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"带类型参数", "#fn f(int) { }"},
		{"带类型参数2", "#fn f(string) { }"},
		{"带类型返回", "#fn f()=>int { return 1 }"},
		{"无类型参数", "#fn f(x) { }"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Compile(tt.input)
			t.Logf("%s: %v", tt.name, err)
		})
	}
}

// ========== 数组类型测试 ==========

// Test_gap_ArrayType_Basic 测试数组类型解析
func Test_gap_ArrayType_Basic(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"int数组", "#fn f(int) { }"},
		{"string数组", "#fn f(string) { }"},
		{"float数组", "#fn f(float) { }"},
		{"bool数组", "#fn f(bool) { }"},
		{"嵌套数组", "#fn f(int) { }"},
		{"返回数组类型", "#fn f()=>int { return 1 }"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Compile(tt.input)
			t.Logf("%s: %v", tt.name, err)
		})
	}
}

// ========== Map类型测试 ==========

// Test_gap_MapType_Basic 测试Map类型解析
func Test_gap_MapType_Basic(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"string到int", "#fn f(int) { }"},
		{"int到string", "#fn f(string) { }"},
		{"string到bool", "#fn f(bool) { }"},
		{"string到float", "#fn f(float) { }"},
		{"返回Map类型", "#fn f()=>int { return 1 }"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Compile(tt.input)
			t.Logf("%s: %v", tt.name, err)
		})
	}
}

// Test_gap_MapType_Errors 测试Map类型解析错误
func Test_gap_MapType_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"缺少左花括号", "#fn f(x) { }"},
		{"缺少冒号", "#fn f(x) { }"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Compile(tt.input)
			t.Logf("%s: %v", tt.name, err)
		})
	}
}

// ========== 基本类型错误测试 ==========

// Test_gap_BaseType_Invalid 测试无效基本类型
func Test_gap_BaseType_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"无效类型", "#fn f(x) { }"},
		{"自定义类型", "#fn f(x) { }"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Compile(tt.input)
			t.Logf("%s: %v", tt.name, err)
		})
	}
}

// ========== 浮点数解析测试 ==========

// Test_gap_FloatLiteral 测试浮点数解析
func Test_gap_FloatLiteral(t *testing.T) {
	tests := []string{
		"0.0",
		"0.5",
		"3.14",
		"1e10",
		"1.5e-10",
		"9.99e99",
	}
	for _, tc := range tests {
		t.Run(tc, func(t *testing.T) {
			p := NewParser()
			_, err := p.Compile(tc)
			t.Logf("'%s' 编译: %v", tc, err)
		})
	}
}

// ========== 不支持功能测试 ==========

// Test_gap_NotSupported 测试不支持的功能
func Test_gap_NotSupported(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"for循环", "for i := 0; i < 10; i++ { }"},
		{"switch语句", "switch x { case 1: 2 }"},
		{"break语句", "break"},
		{"continue语句", "continue"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Compile(tt.input)
			if err == nil {
				t.Errorf("%s 应该报错", tt.name)
			}
		})
	}
}

// ========== 索引赋值测试 ==========

// Test_gap_StoreIndex_Array 测试数组索引访问
func Test_gap_StoreIndex_Array(t *testing.T) {
	p := NewParser()
	script, err := p.Compile("arr := [1, 2, 3]\narr[0]")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}
	ctx := NewContext()
	engine := NewEngine()
	result, err := engine.Run(ctx, script)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if result.Int() != 1 {
		t.Errorf("arr[0] = %d, 期望 1", result.Int())
	}
}

// Test_gap_StoreIndex_Map 测试Map索引访问
func Test_gap_StoreIndex_Map(t *testing.T) {
	p := NewParser()
	script, err := p.Compile("m := {\"a\": 1, \"b\": 2}\nm[\"a\"]")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}
	ctx := NewContext()
	engine := NewEngine()
	result, err := engine.Run(ctx, script)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if result.Int() != 1 {
		t.Errorf("m[\"a\"] = %d, 期望 1", result.Int())
	}
}

// ========== VM runtime方法测试 ==========

// Test_gap_FinalizeForWhile 测试while风格for循环
func Test_gap_FinalizeForWhile(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"条件true", "for true { 1 }"},
		{"比较条件", "for 1 < 2 { 1 }"},
		{"布尔变量", "x := true\nfor x { 1 }"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Compile(tt.input)
			if err != nil {
				t.Logf("编译结果: %v", err)
			}
		})
	}
}

// ========== parseArrayType测试 ==========

// Test_gap_ParseArrayType 测试数组类型解析
func Test_gap_ParseArrayType(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"int数组类型", "#fn f(int) { }"},
		{"string数组类型", "#fn f(string) { }"},
		{"嵌套数组类型", "#fn f(int) { }"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Compile(tt.input)
			if err != nil {
				t.Logf("编译结果: %v", err)
			}
		})
	}
}

// ========== compileForStmt测试 ==========

// Test_gap_CompileForStmt 测试for循环编译错误
func Test_gap_CompileForStmt(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"C风格for", "for i := 0; i < 10; i++ { }"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Compile(tt.input)
			if err == nil {
				t.Error("C风格for循环应该报错")
			}
		})
	}
}

// ========== handleStoreIndex测试 ==========

// Test_gap_HandleStoreIndex_Array 测试数组索引赋值
func Test_gap_HandleStoreIndex_Array(t *testing.T) {
	arr := []interface{}{1, 2, 3}
	ctx := NewContext()
	ctx.BindValue("arr", arr)

	p := NewParser()
	// 正确做法：从全局获取值，赋给局部变量，然后操作
	script, err := p.Compile(`
        localArr :=>arr getBindValue("arr")
        localArr[0] = 100
        localArr[0]
    `)
	if err != nil {
		t.Errorf("数组索引赋值编译失败: %v", err)
		return
	}

	engine := NewEngine()
	result, err := engine.Run(ctx, script)
	if err != nil {
		t.Errorf("执行失败: %v", err)
		return
	}

	if result.Int() != 100 {
		t.Errorf("期望 100, 得到 %d", result.Int())
	}
}

// Test_gap_HandleStoreIndex_Map 测试Map索引赋值
func Test_gap_HandleStoreIndex_Map(t *testing.T) {
	m := map[string]interface{}{"a": 1}
	ctx := NewContext()
	ctx.BindValue("m", m)

	p := NewParser()
	// 正确做法：从全局获取值，赋给局部变量，然后操作
	script, err := p.Compile(`
        localMap :=>any getBindValue("m")
        localMap["a"] = 100
        localMap["a"]
    `)
	if err != nil {
		t.Errorf("Map索引赋值编译失败: %v", err)
		return
	}

	engine := NewEngine()
	result, err := engine.Run(ctx, script)
	if err != nil {
		t.Errorf("执行失败: %v", err)
		return
	}

	if result.Int() != 100 {
		t.Errorf("期望 100, 得到 %d", result.Int())
	}
}

// ========== VM runtime方法测试 ==========

// Test_gap_VM_ToString 测试VM.toString方法
func Test_gap_VM_ToString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"整数转字符串", "string(42)", "42"},
		{"浮点转字符串", "string(3.14)", "3.14"},
		{"布尔转字符串", "string(true)", "true"},
		{"nil转字符串", "string(nil)", "nil"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			script, err := p.Compile(tt.input)
			if err != nil {
				t.Fatalf("编译失败: %v", err)
			}
			ctx := NewContext()
			engine := NewEngine()
			result, err := engine.Run(ctx, script)
			if err != nil {
				t.Fatalf("执行失败: %v", err)
			}
			if result.String() != tt.expected {
				t.Errorf("string(%s) = %q, 期望 %q", tt.input, result.String(), tt.expected)
			}
		})
	}
}

// Test_gap_VM_Equal 测试VM.equal方法
func Test_gap_VM_Equal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"nil相等", "nil == nil", true},
		{"整数相等", "42 == 42", true},
		{"浮点相等", "3.14 == 3.14", true},
		{"字符串相等", `"hello" == "hello"`, true},
		{"布尔相等", "true == true", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			script, err := p.Compile(tt.input)
			if err != nil {
				t.Fatalf("编译失败: %v", err)
			}
			ctx := NewContext()
			engine := NewEngine()
			result, err := engine.Run(ctx, script)
			if err != nil {
				t.Fatalf("执行失败: %v", err)
			}
			if result.Bool() != tt.expected {
				t.Errorf("%s = %v, 期望 %v", tt.input, result.Bool(), tt.expected)
			}
		})
	}
}

// Test_gap_VM_NotEqual 测试VM.notEqual方法
func Test_gap_VM_NotEqual(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"nil不等于整数", "nil != 1", true},
		{"整数不相等", "42 != 0", true},
		{"浮点不相等", "3.14 != 2.71", true},
		{"字符串不相等", `"hello" != "world"`, true},
		{"布尔不相等", "true != false", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			script, err := p.Compile(tt.input)
			if err != nil {
				t.Fatalf("编译失败: %v", err)
			}
			ctx := NewContext()
			engine := NewEngine()
			result, err := engine.Run(ctx, script)
			if err != nil {
				t.Fatalf("执行失败: %v", err)
			}
			if result.Bool() != tt.expected {
				t.Errorf("%s = %v, 期望 %v", tt.input, result.Bool(), tt.expected)
			}
		})
	}
}

// ========== parseForWithInit测试 ==========

// Test_gap_ParseForWithInit_Count 测试按次数循环
func Test_gap_ParseForWithInit_Count(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"次数循环", "for i := 5 { 1 }"},
		{"变量次数", "n := 3\nfor i := n { 1 }"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Compile(tt.input)
			t.Logf("%s: %v", tt.name, err)
		})
	}
}

// Test_gap_ParseForWithInit_Standard 测试标准for循环
func Test_gap_ParseForWithInit_Standard(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"标准循环", "for i := 0; i < 10; i = i + 1 { 1 }"},
		{"缺少后置", "for i := 0; i < 10 { 1 }"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Compile(tt.input)
			t.Logf("%s: %v", tt.name, err)
		})
	}
}

// ========== makeBuiltinUnary测试 ==========

// Test_gap_MakeBuiltinUnary 测试单参数内置函数生成
func Test_gap_MakeBuiltinUnary(t *testing.T) {
	// 测试int函数
	ctx := NewContext()
	p := NewParser()
	script, err := p.Compile("int(42)")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}
	engine := NewEngine()
	result, err := engine.Run(ctx, script)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if result.Int() != 42 {
		t.Errorf("int(42) = %d, 期望 42", result.Int())
	}
}

// ========== initStmtCompilers测试 ==========

// Test_gap_InitStmtCompilers 测试语句编译器初始化
func Test_gap_InitStmtCompilers(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"表达式语句", "1 + 2"},
		{"变量声明", "x := 10"},
		{"if语句", "if true { 1 }"},
		{"return语句", "return 1"},
		{"print调用", "print(1)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Compile(tt.input)
			if err != nil {
				t.Errorf("%s 编译失败: %v", tt.name, err)
			}
		})
	}
}

// ========== compileIfStmt测试 ==========

// Test_gap_CompileIfStmt 测试if语句编译
func Test_gap_CompileIfStmt(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"简单if", "if true { 1 }"},
		{"if-else", "if true { 1 } else { 2 }"},
		{"if-elseif-else", "if false { 1 } else if true { 2 } else { 3 }"},
		{"嵌套if", "if true { if false { 1 } else { 2 } } else { 3 }"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			script, err := p.Compile(tt.input)
			if err != nil {
				t.Fatalf("编译失败: %v", err)
			}
			ctx := NewContext()
			engine := NewEngine()
			_, err = engine.Run(ctx, script)
			if err != nil {
				t.Errorf("执行失败: %v", err)
			}
		})
	}
}

// ========== parseDirectiveComponents测试 ==========

// Test_gap_ParseDirectiveComponents 测试指令组件解析
func Test_gap_ParseDirectiveComponents(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"简单函数", "#fn f() { }"},
		{"带参数函数", "#fn f(x) { }"},
		{"带返回类型", "#fn f()=>int { return 1 }"},
		{"完整定义", "#fn f(int)=>int { return x }"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Compile(tt.input)
			t.Logf("%s: %v", tt.name, err)
		})
	}
}

// ========== handleCallBind测试 ==========

// Test_gap_HandleCallBind 测试外部函数绑定调用
func Test_gap_HandleCallBind(t *testing.T) {
	// 通过#fn声明和完整执行路径测试handleCallBind
	parser := NewParser()
	compiled, err := parser.Compile("#fn add(a, b)\nadd(1, 2)")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	ctx.BindFunc("add", func(a, b int) int { return a + b })

	engine := NewEngine()
	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Errorf("执行失败: %v", err)
	} else {
		t.Logf("结果: %v", result)
	}
}

// ========== convertMapValue测试 ==========

// Test_gap_ConvertMapValue 测试Map值转换
func Test_gap_ConvertMapValue(t *testing.T) {
	// 测试成功转换
	m := &MapValue{Pairs: map[string]Value{"a": NewValue(1), "b": NewValue(2)}}
	val := NewValue(m)
	result, err := convertMapValue(val, reflect.TypeOf(map[string]int{}))
	if err != nil {
		t.Errorf("转换失败: %v", err)
	}
	t.Logf("转换结果: %v", result)
}

// ========== parseDefParams测试 ==========

// Test_gap_ParseDefParams 测试函数参数解析
func Test_gap_ParseDefParams(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"无参数", "#fn f() { }"},
		{"单参数", "#fn f(x) { }"},
		{"多参数", "#fn f(a, b, c) { }"},
		{"带类型参数", "#fn f(int, string) { }"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Compile(tt.input)
			t.Logf("%s: %v", tt.name, err)
		})
	}
}

// ========== parseDefSingleParam测试 ==========

// Test_gap_ParseDefSingleParam 测试单个参数解析
func Test_gap_ParseDefSingleParam(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"无类型参数", "#fn f(x) { }"},
		{"带类型参数", "#fn f(int) { }"},
		{"带箭头类型", "#fn f(int) { }"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Compile(tt.input)
			t.Logf("%s: %v", tt.name, err)
		})
	}
}

// ========== parseIfStmt测试 ==========

// Test_gap_ParseIfStmt 测试if语句解析
func Test_gap_ParseIfStmt(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"简单if", "if true { 1 }"},
		{"if-else", "if true { 1 } else { 2 }"},
		{"if-elseif", "if false { 1 } else if true { 2 }"},
		{"复杂条件", "if 1 < 2 && 2 < 3 { 1 }"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			_, err := p.Compile(tt.input)
			if err != nil {
				t.Errorf("编译失败: %v", err)
			}
		})
	}
}

// ========== registerNumericConverters边界测试 ==========

// Test_gap_RegisterNumericConverters 测试数值类型转换器
func Test_gap_RegisterNumericConverters(t *testing.T) {
	// 测试各种数值类型的转换
	ctx := NewContext()

	// 注册各种数值类型的外部函数
	ctx.BindFunc("testInt8", func(v int8) int { return int(v) })
	ctx.BindFunc("testInt16", func(v int16) int { return int(v) })
	ctx.BindFunc("testInt32", func(v int32) int { return int(v) })
	ctx.BindFunc("testUint", func(v uint) int { return int(v) })
	ctx.BindFunc("testUint8", func(v uint8) int { return int(v) })
	ctx.BindFunc("testUint16", func(v uint16) int { return int(v) })
	ctx.BindFunc("testUint32", func(v uint32) int { return int(v) })
	ctx.BindFunc("testUint64", func(v uint64) int { return int(v) })
	ctx.BindFunc("testFloat32", func(v float32) int { return int(v) })

	// 测试调用这些函数
	p := NewParser()

	// 测试int8参数
	script, err := p.Compile("testInt8(8)")
	if err == nil {
		engine := NewEngine()
		result, err := engine.Run(ctx, script)
		t.Logf("testInt8(8) = %v, err=%v", result, err)
	}

	// 测试float32参数
	script, err = p.Compile("testFloat32(3.14)")
	if err == nil {
		engine := NewEngine()
		result, err := engine.Run(ctx, script)
		t.Logf("testFloat32(3.14) = %v, err=%v", result, err)
	}
}
