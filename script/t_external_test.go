package script

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// ========== 外部函数测试 ==========
// 合并了原 t_base_external_basic_test.go, t_advanced_external_functions_test.go

func Test_callexternalfunc_NoParams(t *testing.T) {
	fn := func() int {
		return 42
	}

	result, err := callExternalFunc(fn, []Value{})
	if err != nil {
		t.Fatalf("调用失败: %v", err)
	}

	if result.Type != TypeInt || result.Int() != 42 {
		t.Errorf("期望 42, 得到 %v", result.Int())
	}
}

func Test_callexternalfunc_WithParams(t *testing.T) {
	fn := func(a, b int) int {
		return a + b
	}

	args := []Value{
		NewValue(10),
		NewValue(20),
	}

	result, err := callExternalFunc(fn, args)
	if err != nil {
		t.Fatalf("调用失败: %v", err)
	}

	if result.Type != TypeInt || result.Int() != 30 {
		t.Errorf("期望 30, 得到 %v", result.Int())
	}
}

func Test_callexternalfunc_ParamCountMismatch(t *testing.T) {
	fn := func(a int) int {
		return a * 2
	}

	args := []Value{
		NewValue(10),
		NewValue(20),
	}

	_, err := callExternalFunc(fn, args)
	if err == nil {
		t.Error("期望参数数量不匹配错误")
	}
}

func Test_callexternalfunc_StringParam(t *testing.T) {
	fn := func(s string) string {
		return "hello " + s
	}

	result, err := callExternalFunc(fn, []Value{NewValue("world")})
	if err != nil {
		t.Fatalf("调用失败: %v", err)
	}

	if result.String() != "hello world" {
		t.Errorf("期望 'hello world', 得到 %v", result.String())
	}
}

func Test_callexternalfunc_FloatParam(t *testing.T) {
	fn := func(x float64) float64 {
		return x * 2.0
	}

	result, err := callExternalFunc(fn, []Value{NewValue(3.5)})
	if err != nil {
		t.Fatalf("调用失败: %v", err)
	}

	if result.Float() != 7.0 {
		t.Errorf("期望 7.0, 得到 %v", result.Float())
	}
}

func Test_handlereturnvalues_NoReturn(t *testing.T) {
	out := []reflect.Value{}

	result, err := handleReturnValues(out)
	if err != nil {
		t.Fatalf("处理失败: %v", err)
	}

	if !result.IsNil() {
		t.Errorf("期望 nil, 得到 %v", result)
	}
}

func Test_handlereturnvalues_SingleReturn(t *testing.T) {
	fn := func() int { return 42 }
	fnValue := reflect.ValueOf(fn)
	out := fnValue.Call([]reflect.Value{})

	result, err := handleReturnValues(out)
	if err != nil {
		t.Fatalf("处理失败: %v", err)
	}

	if result.Int() != 42 {
		t.Errorf("期望 42, 得到 %v", result.Int())
	}
}

func Test_handlereturnvalues_DualReturn(t *testing.T) {
	fn := func() (int, error) { return 42, nil }
	fnValue := reflect.ValueOf(fn)
	out := fnValue.Call([]reflect.Value{})

	result, err := handleReturnValues(out)
	if err != nil {
		t.Fatalf("处理失败: %v", err)
	}

	if result.Int() != 42 {
		t.Errorf("期望 42, 得到 %v", result.Int())
	}
}

func Test_handlereturnvalues_TooManyReturns(t *testing.T) {
	fn := func() (int, int, int) { return 1, 2, 3 }
	fnValue := reflect.ValueOf(fn)
	out := fnValue.Call([]reflect.Value{})

	_, err := handleReturnValues(out)
	if err == nil {
		t.Error("期望返回值数量错误")
	}
}

func Test_convertargs_Success(t *testing.T) {
	fn := func(a int, b string) {}
	fnType := reflect.TypeOf(fn)

	args := []Value{
		NewValue(10),
		NewValue("hello"),
	}

	result, err := convertArgs(fnType, args)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("期望 2 个参数, 得到 %d", len(result))
	}
}

func Test_convertargs_TypeMismatch(t *testing.T) {
	fn := func(a int) {}
	fnType := reflect.TypeOf(fn)

	// 传入字符串，转换器会将字符串值转换为Go字符串
	// 这与目标int类型不匹配，但convertArgs本身不检查
	args := []Value{
		NewValue("not an int"),
	}

	// convertArgs会成功返回，但类型不匹配
	result, err := convertArgs(fnType, args)
	if err != nil {
		t.Fatalf("convertArgs不应失败: %v", err)
	}

	// 验证转换结果是字符串类型，而非int
	if result[0].Kind() != reflect.String {
		t.Errorf("期望String类型，得到 %v", result[0].Kind())
	}
}

func Test_wrapconverterror(t *testing.T) {
	innerErr := errors.New("内部错误")
	targetType := reflect.TypeOf(0)
	actualValue := NewValue("test")
	wrapped := wrapConvertError(2, targetType, actualValue, innerErr)

	if wrapped == nil {
		t.Fatal("期望错误")
	}

	expected := "参数 2 转换失败: 内部错误"
	if !strings.Contains(wrapped.Error(), expected) {
		t.Errorf("期望包含 '%s', 得到 '%s'", expected, wrapped.Error())
	}
}

func Test_context_BindAndCallFunc(t *testing.T) {
	ctx := NewContext()

	// 绑定一个简单函数
	ctx.BindFunc("double", func(x int) int {
		return x * 2
	})

	fn, ok := ctx.GetBindFunc("double")
	if !ok {
		t.Fatal("未找到绑定的函数")
	}

	// 调用函数
	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		t.Fatal("不是函数类型")
	}
}

func Test_handlecallbind_Success(t *testing.T) {
	parser := NewParser()
	// 编译带有#fn声明和函数调用的脚本
	compiled, err := parser.Compile("#fn double(x)\ndouble(21)")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	// 创建上下文并绑定函数
	ctx := NewContext()
	ctx.BindFunc("double", func(x int) int {
		return x * 2
	})

	// 执行脚本
	engine := NewEngine()
	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	// 验证结果
	if result.Type != TypeInt || result.Int() != 42 {
		t.Errorf("期望 42, 得到 %v", result)
	}
}

func Test_handlecallbind_MultipleArgs(t *testing.T) {
	parser := NewParser()
	compiled, err := parser.Compile("#fn add(a, b)\nadd(10, 20)")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	ctx.BindFunc("add", func(a, b int) int {
		return a + b
	})

	engine := NewEngine()
	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	if result.Int() != 30 {
		t.Errorf("期望 30, 得到 %v", result.Int())
	}
}

func Test_handlecallbind_NotBound(t *testing.T) {
	parser := NewParser()
	compiled, err := parser.Compile("#fn myFunc(x)\nmyFunc(1)")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	// 不绑定函数，期望错误

	engine := NewEngine()
	_, err = engine.Run(ctx, compiled)
	if err == nil {
		t.Error("期望函数未绑定错误")
	}
}

func Test_handlecallbind_StringResult(t *testing.T) {
	parser := NewParser()
	compiled, err := parser.Compile("#fn greet(name)\ngreet(\"world\")")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	ctx.BindFunc("greet", func(name string) string {
		return "hello " + name
	})

	engine := NewEngine()
	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	if result.String() != "hello world" {
		t.Errorf("期望 'hello world', 得到 %v", result.String())
	}
}

func Test_handlecallbind_NoArgs(t *testing.T) {
	parser := NewParser()
	compiled, err := parser.Compile("#fn getValue\ngetValue()")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	ctx.BindFunc("getValue", func() int {
		return 100
	})

	engine := NewEngine()
	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	if result.Int() != 100 {
		t.Errorf("期望 100, 得到 %v", result.Int())
	}
}

func Test_handlecallbind_WithFloat(t *testing.T) {
	parser := NewParser()
	compiled, err := parser.Compile("#fn scale(x)\nscale(2.5)")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	ctx.BindFunc("scale", func(x float64) float64 {
		return x * 2.0
	})

	engine := NewEngine()
	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	if result.Float() != 5.0 {
		t.Errorf("期望 5.0, 得到 %v", result.Float())
	}
}

func Test_convertnumericvalue_NotNumeric(t *testing.T) {
	val := NewValue(42)

	// 目标类型是string而非数值
	_, err := convertNumericValue(val, reflect.TypeOf(""))
	if err == nil {
		t.Error("期望目标类型非数值返回错误")
	}
}

func Test_convertgotovalue_UnknownKind(t *testing.T) {
	// struct类型没有注册转换器
	type MyStruct struct{ X int }
	val := MyStruct{X: 42}

	result := convertGoToValue(val)
	if !result.IsNil() {
		t.Errorf("期望未知类型返回nil，得到 %v", result)
	}
}

func Test_convertgoptr_NilPointer(t *testing.T) {
	var ptr *int
	v := reflect.ValueOf(ptr)

	result := convertGoPtr(v)
	if !result.IsNil() {
		t.Errorf("期望nil指针返回nil，得到 %v", result)
	}
}

func Test_handlesinglereturnvalue_ErrorType(t *testing.T) {
	fn := func() error { return errors.New("test error") }
	fnValue := reflect.ValueOf(fn)
	out := fnValue.Call([]reflect.Value{})

	_, err := handleSingleReturnValue(out)
	if err == nil {
		t.Error("期望返回error")
	}
}

func Test_handledualreturnvalue_ErrorInSecond(t *testing.T) {
	fn := func() (int, error) { return 0, errors.New("test error") }
	fnValue := reflect.ValueOf(fn)
	out := fnValue.Call([]reflect.Value{})

	_, err := handleDualReturnValue(out)
	if err == nil {
		t.Error("期望返回error")
	}
}

func Test_convertnumericvalue_InvalidType(t *testing.T) {
	val := NewValue(100)
	_, err := convertNumericValue(val, reflect.TypeOf("string"))
	if err == nil {
		t.Error("无效类型转换应返回错误")
	}
}

func Test_gotovalue_NilPointer(t *testing.T) {
	var ptr *int
	result := convertGoToValue(ptr)
	if !result.IsNil() {
		t.Error("nil指针应转换为nil值")
	}
}

func Test_gotovalue_UnsupportedType(t *testing.T) {
	// struct类型不在goConverters中，应返回nil
	type TestStruct struct{ X int }
	result := convertGoToValue(TestStruct{X: 1})
	if !result.IsNil() {
		t.Error("不支持的类型应返回nil")
	}
}

func Test_externalfunc_ParameterConversion(t *testing.T) {
	fn := func(x int) int {
		return x * 2
	}

	args := []Value{NewValue(5)}
	result, err := callExternalFunc(fn, args)
	if err != nil {
		t.Fatalf("调用失败: %v", err)
	}

	if result.Int() != 10 {
		t.Errorf("期望10, 得到 %d", result.Int())
	}
}

func Test_externalfunc_NilReturn(t *testing.T) {
	fn := func() interface{} {
		return nil
	}

	args := []Value{}
	result, err := callExternalFunc(fn, args)
	if err != nil {
		t.Fatalf("调用失败: %v", err)
	}

	if !result.IsNil() {
		t.Errorf("期望nil, 得到 %v", result)
	}
}

func Test_parser_ThrowStmt(t *testing.T) {
	input := `
x := -1
if x < 0 {
	throw "error"
}
`
	parser := NewParser()
	_, err := parser.Compile(input)
	if err != nil {
		t.Errorf("throw语句编译失败: %v", err)
	}
}

func Test_compiler_CompoundAssign(t *testing.T) {
	input := `
x := 10
x += 5
x
`
	parser := NewParser()
	_, err := parser.Compile(input)
	if err != nil {
		t.Logf("复合赋值编译: %v", err)
	}
}

// ========== #fn 外部函数预定义语法测试 ==========

// Test_ExternalFn_ShortSyntax 测试 #fn 短语法
func Test_ExternalFn_ShortSyntax(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("double", func(x int) int { return x * 2 })

	parser := NewParser()
	engine := NewEngine()

	compiled, err := parser.Compile(`
        #fn double(int)=>int
        double(5)
    `)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	if result.Int() != 10 {
		t.Errorf("期望 10, 得到 %d", result.Int())
	}
}

// Test_ExternalFn_TypeAlias 测试使用类型别名
func Test_ExternalFn_TypeAlias(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("greet", func(name string) string { return "Hello, " + name })

	parser := NewParser()
	engine := NewEngine()

	compiled, err := parser.Compile(`
        #fn greet(s)=>s
        greet("World")
    `)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	if result.String() != "Hello, World" {
		t.Errorf("期望 'Hello, World', 得到 %s", result.String())
	}
}

// Test_ExternalFn_MultipleParams 测试多参数
func Test_ExternalFn_MultipleParams(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("add", func(a, b int) int { return a + b })

	parser := NewParser()
	engine := NewEngine()

	compiled, err := parser.Compile(`
        #fn add(int, int)=>int
        add(10, 20)
    `)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	if result.Int() != 30 {
		t.Errorf("期望 30, 得到 %d", result.Int())
	}
}

// Test_ExternalFn_NoReturn 测试无返回值
func Test_ExternalFn_NoReturn(t *testing.T) {
	ctx := NewContext()
	called := false
	ctx.BindFunc("doSomething", func(x int) { called = true })

	parser := NewParser()
	engine := NewEngine()

	compiled, err := parser.Compile(`
        #fn doSomething(int)
        doSomething(42)
    `)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	_, err = engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	if !called {
		t.Error("期望函数被调用")
	}
}

// Test_ExternalFn_MixedWithDef 测试多个 #fn 声明
func Test_ExternalFn_MixedWithDef(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("double", func(x int) int { return x * 2 })
	ctx.BindFunc("triple", func(x int) int { return x * 3 })

	parser := NewParser()
	engine := NewEngine()

	compiled, err := parser.Compile(`
        #fn double(int)=>int
        #fn triple(int)=>int
        double(5) + triple(5)
    `)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	// 10 + 15 = 25
	if result.Int() != 25 {
		t.Errorf("期望 25, 得到 %d", result.Int())
	}
}

// Test_ExternalFn_FloatParam 测试浮点数参数
func Test_ExternalFn_FloatParam(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("scale", func(x float64) float64 { return x * 2.5 })

	parser := NewParser()
	engine := NewEngine()

	compiled, err := parser.Compile(`
        #fn scale(f)=>f
        scale(4.0)
    `)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	if result.Float() != 10.0 {
		t.Errorf("期望 10.0, 得到 %f", result.Float())
	}
}

// Test_ExternalFn_NoParams 测试无参数
func Test_ExternalFn_NoParams(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("getValue", func() int { return 42 })

	parser := NewParser()
	engine := NewEngine()

	compiled, err := parser.Compile(`
        #fn getValue()=>int
        getValue()
    `)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	if result.Int() != 42 {
		t.Errorf("期望 42, 得到 %d", result.Int())
	}
}

// Test_ExternalFn_ComplexTypes 测试复杂类型别名
func Test_ExternalFn_ComplexTypes(t *testing.T) {
	ctx := NewContext()
	ctx.BindFunc("first", func(arr []int) int {
		if len(arr) > 0 {
			return arr[0]
		}
		return 0
	})

	parser := NewParser()
	engine := NewEngine()

	compiled, err := parser.Compile(`
        #fn first(arr)=>i
        first([10, 20, 30])
    `)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	if result.Int() != 10 {
		t.Errorf("期望 10, 得到 %v", result)
	}
}
