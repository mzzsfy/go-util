package script

import (
	"reflect"
	"strings"
	"testing"
)

// ========== 测试辅助函数 ==========
// 本文件包含所有测试共用的辅助函数和工具函数

// ========== 泛型测试辅助函数（Go 1.18+）==========

// runTypedTest 泛型测试函数，减少重复代码
func runTypedTest[T comparable](t *testing.T, input string, expected T, extract func(Value) T, typeName string) {
	t.Helper()
	result := runScript(t, input)
	if actual := extract(result); actual != expected {
		t.Errorf("[%s] 期望 %s 类型的值 %v，得到 %v", input, typeName, expected, actual)
	}
}

// RunTypedTestsSimple 泛型批量测试函数
func RunTypedTestsSimple[T comparable](t *testing.T, tests []struct {
	Input    string
	Expected T
}, extract func(Value) T, typeName string) {
	for _, tt := range tests {
		t.Run(tt.Input, func(t *testing.T) {
			runTypedTest(t, tt.Input, tt.Expected, extract, typeName)
		})
	}
}

// assertTyped 泛型断言函数，减少重复代码
func assertTyped[T comparable](t *testing.T, expected, actual T, msgAndArgs ...interface{}) {
	t.Helper()
	if expected != actual {
		t.Errorf("期望 %v，得到 %v. %v", expected, actual, msgAndArgs)
	}
}

// ========== 基础测试辅助函数 ==========

// runScript 执行脚本并返回结果
func runScript(t *testing.T, input string) Value {
	t.Helper()
	parser := NewParser()
	script, err := parser.Compile(input)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}
	ctx := NewContext()
	engine := NewEngine()
	result, err := engine.Run(ctx, script)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	return result
}

// runScriptWithContext 使用指定上下文执行脚本
func runScriptWithContext(t *testing.T, input string, ctx *Context) Value {
	t.Helper()
	parser := NewParser()
	script, err := parser.Compile(input)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}
	engine := NewEngine()
	result, err := engine.Run(ctx, script)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	return result
}

// runScriptWithBinds 使用绑定变量执行脚本
func runScriptWithBinds(t *testing.T, input string, binds map[string]interface{}) Value {
	t.Helper()
	ctx := NewContext()
	for name, val := range binds {
		ctx.BindValue(name, val)
	}
	return runScriptWithContext(t, input, ctx)
}

// runScriptWithFunc 使用绑定函数执行脚本
func runScriptWithFunc(t *testing.T, input string, name string, fn interface{}) Value {
	t.Helper()
	ctx := NewContext()
	ctx.BindFunc(name, fn)
	return runScriptWithContext(t, input, ctx)
}

// compileScript 编译脚本并返回编译结果
func compileScript(t *testing.T, input string) *CompiledScript {
	t.Helper()
	parser := NewParser()
	script, err := parser.Compile(input)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}
	return script
}

// runIntTest 执行脚本并验证整数结果
func runIntTest(t *testing.T, input string, expected int) {
	t.Helper()
	runTypedTest(t, input, expected, func(v Value) int { return v.Int() }, "int")
}

// runBoolTest 执行脚本并验证布尔结果
func runBoolTest(t *testing.T, input string, expected bool) {
	t.Helper()
	runTypedTest(t, input, expected, func(v Value) bool { return v.Bool() }, "bool")
}

// runFloatTest 执行脚本并验证浮点数结果
func runFloatTest(t *testing.T, input string, expected float64) {
	t.Helper()
	runTypedTest(t, input, expected, func(v Value) float64 { return v.Float() }, "float64")
}

// runStringTest 执行脚本并验证字符串结果
func runStringTest(t *testing.T, input string, expected string) {
	t.Helper()
	runTypedTest(t, input, expected, func(v Value) string { return v.String() }, "string")
}

// runErrorTest 验证脚本编译应该返回错误
func runErrorTest(t *testing.T, input string) {
	t.Helper()
	parser := NewParser()
	_, err := parser.Compile(input)
	if err == nil {
		t.Errorf("[%s] 期望编译错误，但编译成功", input)
	}
}

// runRuntimeErrorTest 验证脚本执行应该返回错误
func runRuntimeErrorTest(t *testing.T, input string) {
	t.Helper()
	parser := NewParser()
	script, err := parser.Compile(input)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}
	ctx := NewContext()
	engine := NewEngine()
	_, err = engine.Run(ctx, script)
	if err == nil {
		t.Errorf("[%s] 期望运行时错误，但执行成功", input)
	}
}

// RunIntTestsSimple 批量运行整数测试
func RunIntTestsSimple(t *testing.T, tests []struct {
	Input    string
	Expected int
}) {
	RunTypedTestsSimple(t, tests, func(v Value) int { return v.Int() }, "int")
}

// RunBoolTestsSimple 批量运行布尔测试
func RunBoolTestsSimple(t *testing.T, tests []struct {
	Input    string
	Expected bool
}) {
	RunTypedTestsSimple(t, tests, func(v Value) bool { return v.Bool() }, "bool")
}

// RunFloatTestsSimple 批量运行浮点数测试
func RunFloatTestsSimple(t *testing.T, tests []struct {
	Input    string
	Expected float64
}) {
	RunTypedTestsSimple(t, tests, func(v Value) float64 { return v.Float() }, "float64")
}

// RunStringTestsSimple 批量运行字符串测试
func RunStringTestsSimple(t *testing.T, tests []struct {
	Input    string
	Expected string
}) {
	RunTypedTestsSimple(t, tests, func(v Value) string { return v.String() }, "string")
}

// assertValueEqual 断言两个值相等
func assertValueEqual(t *testing.T, expected, actual Value, msgAndArgs ...interface{}) {
	t.Helper()
	if !expected.Equal(actual) {
		t.Errorf("期望 %v，得到 %v. %v", expected, actual, msgAndArgs)
	}
}

// assertValueDeepEqual 深度断言两个值相等
func assertValueDeepEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("期望 %v，得到 %v", expected, actual)
	}
}

// assertNoError 断言没有错误
func assertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		t.Errorf("期望无错误，但得到: %v. %v", err, msgAndArgs)
	}
}

// assertError 断言有错误
func assertError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err == nil {
		t.Errorf("期望有错误，但没有错误. %v", msgAndArgs)
	}
}

// assertType 断言值的类型
func assertType(t *testing.T, val Value, expectedType ValueType) {
	t.Helper()
	if val.Type != expectedType {
		t.Errorf("期望类型 %v，得到 %v", expectedType, val.Type)
	}
}

// assertInt 断言整数值
func assertInt(t *testing.T, val Value, expected int) {
	t.Helper()
	assertTyped(t, expected, val.Int(), "值类型: int")
}

// assertBool 断言布尔值
func assertBool(t *testing.T, val Value, expected bool) {
	t.Helper()
	assertTyped(t, expected, val.Bool(), "值类型: bool")
}

// assertString 断言字符串值
func assertString(t *testing.T, val Value, expected string) {
	t.Helper()
	assertTyped(t, expected, val.String(), "值类型: string")
}

// assertFloat 断言浮点数值
func assertFloat(t *testing.T, val Value, expected float64) {
	t.Helper()
	assertTyped(t, expected, val.Float(), "值类型: float64")
}

// assertNil 断言值为nil
func assertNil(t *testing.T, val Value) {
	t.Helper()
	if !val.IsNil() {
		t.Errorf("期望 nil，得到 %v", val)
	}
}

// assertNotNil 断言值不为nil
func assertNotNil(t *testing.T, val Value) {
	t.Helper()
	if val.IsNil() {
		t.Errorf("期望非 nil 值")
	}
}

// assertLen 断言长度
func assertLen(t *testing.T, val interface{}, expectedLen int) {
	t.Helper()
	v := reflect.ValueOf(val)
	if v.Len() != expectedLen {
		t.Errorf("期望长度 %d，得到 %d", expectedLen, v.Len())
	}
}

// assertContains 断言字符串包含子串
func assertContains(t *testing.T, str, substr string) {
	t.Helper()
	if !contains(str, substr) {
		t.Errorf("期望 %q 包含 %q", str, substr)
	}
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// runCompileTests 批量编译测试，验证代码可以成功编译
func runCompileTests(t *testing.T, inputs []string) {
	t.Helper()
	for _, input := range inputs {
		compileScript(t, input)
	}
}

// runIntTestWithBinds 使用绑定变量执行整数测试
func runIntTestWithBinds(t *testing.T, input string, expected int, binds map[string]interface{}) {
	t.Helper()
	result := runScriptWithBinds(t, input, binds)
	if result.Int() != expected {
		t.Errorf("[%s] 期望 %d，得到 %d", input, expected, result.Int())
	}
}

// runBoolTestWithBinds 使用绑定变量执行布尔测试
func runBoolTestWithBinds(t *testing.T, input string, expected bool, binds map[string]interface{}) {
	t.Helper()
	result := runScriptWithBinds(t, input, binds)
	if result.Bool() != expected {
		t.Errorf("[%s] 期望 %v，得到 %v", input, expected, result.Bool())
	}
}

// newTestVM 创建测试用的VM实例
// 统一VM创建方式，便于后续维护和修改
func newTestVM(t *testing.T) *VM {
	t.Helper()
	ctx := NewContext()
	return NewVM(ctx, 256)
}

// newTestVMWithContext 使用指定上下文创建测试用的VM实例
func newTestVMWithContext(t *testing.T, ctx *Context) *VM {
	t.Helper()
	return NewVM(ctx, 256)
}
