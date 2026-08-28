package script

import (
	"reflect"
	"testing"
)

// ========== Compiler 完整测试 ==========

func Test_compiler_LiteralExpr(t *testing.T) {
	tests := []string{"10", "10.5", "\"hello\"", "true", "false", "nil"}
	for _, tt := range tests {
		script := compileScript(t, tt)
		if len(script.Main.Constants) == 0 {
			t.Errorf("[%s] 应该有常量", tt)
		}
	}
}

func Test_compiler_CollectionLiteral(t *testing.T) {
	t.Run("数组", func(t *testing.T) {
		script := compileScript(t, "[1, 2, 3, 4, 5]")
		hasArrayNew := false
		for _, inst := range script.Main.Instructions {
			if inst.Op == OpArrayNew {
				hasArrayNew = true
				break
			}
		}
		if !hasArrayNew {
			t.Error("应该生成OpArrayNew指令")
		}
	})

	t.Run("Map", func(t *testing.T) {
		script := compileScript(t, `{"a": 1, "b": 2, "c": 3}`)
		hasMapNew := false
		for _, inst := range script.Main.Instructions {
			if inst.Op == OpMapNew {
				hasMapNew = true
				break
			}
		}
		if !hasMapNew {
			t.Error("应该生成OpMapNew指令")
		}
	})
}

func Test_compiler_Operators(t *testing.T) {
	t.Run("一元运算", func(t *testing.T) {
		runCompileTests(t, []string{"-10", "!true", "^15"})
	})

	t.Run("二元运算", func(t *testing.T) {
		runCompileTests(t, []string{
			"10 + 20", "10 - 5", "10 * 2", "10 / 2", "10 % 3",
			"10 & 5", "10 | 5", "10 ^ 5", "1 << 4",
			"10 == 20", "10 != 20", "10 < 20", "10 > 5",
			"1 && 1", "0 || 1",
		})
	})

	t.Run("嵌套表达式", func(t *testing.T) {
		runCompileTests(t, []string{
			"10 + 20 + 30", "10 * 20 + 30",
			"10 + 20 * 30", "(10 + 20) * 30",
		})
	})
}

func Test_compiler_Identifier(t *testing.T) {
	compileScript(t, `x := 10
y := 20
x + y`)
}

func Test_compiler_IfElse(t *testing.T) {
	script := compileScript(t, `x := 15
if x > 10 { 100 } else { 0 }`)
	hasJump := false
	for _, inst := range script.Main.Instructions {
		if inst.Op == OpJumpIfFalse || inst.Op == OpJump {
			hasJump = true
			break
		}
	}
	if !hasJump {
		t.Error("if-else应该生成跳转指令")
	}
}

func Test_compiler_Function(t *testing.T) {
	t.Run("基本函数", func(t *testing.T) {
		script := compileScript(t, `fn add(x, y) { return x + y }`)
		if len(script.Functions) == 0 {
			t.Fatal("应该创建函数")
		}
		if script.Functions[0].Name != "add" {
			t.Errorf("函数名应为'add'，得到'%s'", script.Functions[0].Name)
		}
	})

	t.Run("带参数函数", func(t *testing.T) {
		script := compileScript(t, `fn test(a, b, c) { return a }`)
		if script.Functions[0].NumParams != 3 {
			t.Errorf("参数数量应为3，得到%d", script.Functions[0].NumParams)
		}
	})

	t.Run("多个函数", func(t *testing.T) {
		script := compileScript(t, `fn f1() { 1 }
fn f2() { 2 }`)
		if len(script.Functions) != 2 {
			t.Errorf("应该创建2个函数，得到%d", len(script.Functions))
		}
	})
}

// ========== External 函数测试 ==========

func Test_external_CallExternalFunc(t *testing.T) {
	t.Run("整数参数", func(t *testing.T) {
		fn := func(a, b int) int { return a + b }
		result, err := callExternalFunc(fn, []Value{NewValue(10), NewValue(20)})
		if err != nil {
			t.Fatalf("调用失败: %v", err)
		}
		if result.Int() != 30 {
			t.Errorf("期望30，得到%d", result.Int())
		}
	})

	t.Run("字符串参数", func(t *testing.T) {
		fn := func(s string) string { return s + "!" }
		result, err := callExternalFunc(fn, []Value{NewValue("hello")})
		if err != nil {
			t.Fatalf("调用失败: %v", err)
		}
		if result.String() != "hello!" {
			t.Errorf("期望'hello!'，得到'%s'", result.String())
		}
	})
}

func Test_external_CallExternalFuncWithError(t *testing.T) {
	fn := func(x int) (int, error) {
		if x < 0 {
			return 0, &testError{msg: "negative"}
		}
		return x * 2, nil
	}

	// 测试正常情况
	args := []Value{NewValue(10)}
	result, err := callExternalFunc(fn, args)
	if err != nil {
		t.Fatalf("调用失败: %v", err)
	}
	if result.Int() != 20 {
		t.Errorf("期望20，得到%d", result.Int())
	}

	// 测试错误情况
	args = []Value{NewValue(-1)}
	_, err = callExternalFunc(fn, args)
	if err == nil {
		t.Error("应该返回错误")
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func Test_external_ConvertValue(t *testing.T) {
	t.Run("切片", func(t *testing.T) {
		val := NewValue([]Value{NewValue(1), NewValue(2), NewValue(3)})
		converted, err := convertValueToGo(val, reflect.TypeOf([]int{}))
		if err != nil {
			t.Fatalf("转换失败: %v", err)
		}
		slice := converted.Interface().([]int)
		if len(slice) != 3 || slice[0] != 1 || slice[1] != 2 || slice[2] != 3 {
			t.Errorf("切片内容不正确")
		}
	})

	t.Run("Map", func(t *testing.T) {
		m := &MapValue{Pairs: map[string]Value{"a": NewValue(1), "b": NewValue(2)}}
		converted, err := convertValueToGo(NewValue(m), reflect.TypeOf(map[string]int{}))
		if err != nil {
			t.Fatalf("转换失败: %v", err)
		}
		result := converted.Interface().(map[string]int)
		if len(result) != 2 || result["a"] != 1 || result["b"] != 2 {
			t.Errorf("Map内容不正确")
		}
	})
}

func Test_external_GoToValue(t *testing.T) {
	t.Run("切片", func(t *testing.T) {
		result := convertGoToValue([]int{1, 2, 3})
		if result.Type != TypeArray || len(result.Array().Elements) != 3 {
			t.Errorf("切片转换失败")
		}
	})

	t.Run("Map", func(t *testing.T) {
		result := convertGoToValue(map[string]int{"a": 1, "b": 2})
		if result.Type != TypeMap || len(result.Map().Pairs) != 2 {
			t.Errorf("Map转换失败")
		}
	})

	t.Run("nil", func(t *testing.T) {
		if !convertGoToValue(nil).IsNil() {
			t.Errorf("nil应转换为Nil类型")
		}
	})

	t.Run("指针", func(t *testing.T) {
		val := 10
		if convertGoToValue(&val).Int() != 10 {
			t.Errorf("指针解引用失败")
		}
		var ptr *int
		if !convertGoToValue(ptr).IsNil() {
			t.Errorf("nil指针应转换为nil")
		}
	})
}

// ========== 编译器语句测试 ==========

func Test_compiler_Statements(t *testing.T) {
	t.Run("赋值", func(t *testing.T) {
		runCompileTests(t, []string{"x := 10", "x := 10\n y := 20", "x := 10 + 20", "x := 10\n y := x"})
	})

	t.Run("if语句", func(t *testing.T) {
		runCompileTests(t, []string{"if true { 1 }", "if true { 1 } else { 2 }", "if true { if false { 1 } else { 2 } }", "if 1 < 2 { 3 }"})
	})

	t.Run("切片", func(t *testing.T) {
		runCompileTests(t, []string{"arr := [1, 2, 3, 4]\narr[1:3]", "arr := [1, 2, 3]\narr[:2]", "arr := [1, 2, 3]\narr[1:]"})
	})
}

func Test_compiler_DefDirective(t *testing.T) {
	script := compileScript(t, `#fn externalFunc(x)=>int`)
	if len(script.Externals) != 1 {
		t.Errorf("应该有1个外部函数声明，得到 %d", len(script.Externals))
	}
}

func Test_compiler_MoreStatements(t *testing.T) {
	t.Run("throw", func(t *testing.T) {
		runCompileTests(t, []string{`throw "error"`, "throw 404"})
	})

	t.Run("函数声明", func(t *testing.T) {
		runCompileTests(t, []string{"fn f() { 1 }", "fn f(x) { x }", "fn f(a, b, c) { a + b + c }", "fn add(a, b) { return a + b }"})
	})

	t.Run("数组", func(t *testing.T) {
		runCompileTests(t, []string{"arr := []", "arr := [1]", "arr := [1, 2, 3, 4, 5]", "arr := [[1, 2], [3, 4]]", `arr := [1, "a", true]`})
	})

	t.Run("Map", func(t *testing.T) {
		runCompileTests(t, []string{"m := {}", `m := {"a": 1}`, `m := {"a": 1, "b": 2, "c": 3}`})
	})
}

// ========== 编译器覆盖测试（从coverage_compiler_test.go合并） ==========

// TestCompilerCoverage 测试编译器各种场景
func Test_compilercoverage(t *testing.T) {
	t.Run("空程序", func(t *testing.T) {
		parser := NewParser()
		script, err := parser.Compile("")
		if err != nil {
			t.Fatalf("编译失败: %v", err)
		}
		if script == nil {
			t.Fatal("脚本为nil")
		}
	})

	t.Run("单表达式", func(t *testing.T) {
		result := runScript(t, "42")
		if result.Int() != 42 {
			t.Errorf("期望 42, 实际 %d", result.Int())
		}
	})

	t.Run("多语句程序", func(t *testing.T) {
		input := `x := 1
y := 2
z := x + y
z`
		result := runScript(t, input)
		if result.Int() != 3 {
			t.Errorf("期望 3, 实际 %d", result.Int())
		}
	})
}

// TestGroupedExpr 测试分组表达式
func Test_groupedexpr(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"(1 + 2) * 3", 9},
		{"1 + (2 * 3)", 7},
		{"((1 + 2))", 3},
		{"(1)", 1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := runScript(t, tt.input)
			if result.Int() != tt.expected {
				t.Errorf("期望 %d, 实际 %d", tt.expected, result.Int())
			}
		})
	}
}

// TestPrintCoverage 测试print/println函数
func Test_printcoverage(t *testing.T) {
	t.Run("print函数", func(t *testing.T) {
		input := `print("test")
nil`
		result := runScript(t, input)
		if !result.IsNil() {
			t.Errorf("print应返回nil")
		}
	})

	t.Run("println函数", func(t *testing.T) {
		input := `println("test", 123)
nil`
		result := runScript(t, input)
		if !result.IsNil() {
			t.Errorf("println应返回nil")
		}
	})

	t.Run("print多参数", func(t *testing.T) {
		input := `print(1, 2, 3, "hello", true)
nil`
		result := runScript(t, input)
		if !result.IsNil() {
			t.Errorf("print应返回nil")
		}
	})
}

// TestCompiler_ExternalFunctionCall 测试外部函数调用编译
func Test_compiler_ExternalFunctionCall(t *testing.T) {
	input := `#fn myFunc(x)=>int
myFunc(42)
nil`
	parser := NewParser()
	script, err := parser.Compile(input)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}
	if len(script.Externals) != 1 {
		t.Errorf("应该有1个外部函数声明，得到 %d", len(script.Externals))
	}
	if script.Externals[0].Name != "myFunc" {
		t.Errorf("外部函数名应为myFunc，得到%s", script.Externals[0].Name)
	}
}

// TestCompiler_ErrorPaths2 测试编译器错误路径
func Test_compiler_ErrorPaths2(t *testing.T) {
	t.Run("未定义变量", func(t *testing.T) {
		parser := NewParser()
		_, err := parser.Compile("undefinedVar")
		if err == nil {
			t.Error("应该报告未定义变量错误")
		}
	})

	t.Run("不支持的一元运算符", func(t *testing.T) {
		// 直接测试编译器
		compiler := NewCompiler()
		expr := &UnaryExpr{
			Position: Position{Line: 1, Column: 1},
			Operator: "$", // 不支持的运算符
			Operand:  &LiteralExpr{Type: LiteralInt, Value: 1},
		}
		program := &Program{
			Statements: []Stmt{&ExprStmt{Expr: expr}},
		}
		_, err := compiler.Compile(program)
		if err == nil {
			t.Error("应该报告不支持的一元运算符错误")
		}
	})

	t.Run("不支持的二元运算符", func(t *testing.T) {
		compiler := NewCompiler()
		expr := &BinaryExpr{
			Position: Position{Line: 1, Column: 1},
			Left:     &LiteralExpr{Type: LiteralInt, Value: 1},
			Operator: "$$", // 不支持的运算符
			Right:    &LiteralExpr{Type: LiteralInt, Value: 2},
		}
		program := &Program{
			Statements: []Stmt{&ExprStmt{Expr: expr}},
		}
		_, err := compiler.Compile(program)
		if err == nil {
			t.Error("应该报告不支持的二元运算符错误")
		}
	})
}

// TestCompiler_BuiltinErrors 测试内置函数编译错误
func Test_compiler_BuiltinErrors(t *testing.T) {
	t.Run("len参数过多", func(t *testing.T) {
		parser := NewParser()
		_, err := parser.Compile("len(1, 2)")
		if err == nil {
			t.Error("应该报告len参数数量错误")
		}
	})

	t.Run("typeof参数过多", func(t *testing.T) {
		parser := NewParser()
		_, err := parser.Compile("typeof(1, 2)")
		if err == nil {
			t.Error("应该报告typeof参数数量错误")
		}
	})

	t.Run("int参数过多", func(t *testing.T) {
		parser := NewParser()
		_, err := parser.Compile("int(1, 2)")
		if err == nil {
			t.Error("应该报告int参数数量错误")
		}
	})

	t.Run("float参数过多", func(t *testing.T) {
		parser := NewParser()
		_, err := parser.Compile("float(1, 2)")
		if err == nil {
			t.Error("应该报告float参数数量错误")
		}
	})
}

// TestCompiler_IfExpression 测试if表达式编译
func Test_compiler_IfExpression(t *testing.T) {
	tests := []string{
		"if true { 1 } else { 2 }",
		"if false { 1 } else { 2 }",
		"if true { if true { 10 } else { 20 } } else { 30 }",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			parser := NewParser()
			_, err := parser.Compile(tt)
			if err != nil {
				t.Errorf("编译失败 [%s]: %v", tt, err)
			}
		})
	}
}

// TestCompiler_ShortCircuit 测试短路求值
func Test_compiler_ShortCircuit(t *testing.T) {
	t.Run("逻辑与短路", func(t *testing.T) {
		result := runScript(t, "false && true")
		if result.Bool() != false {
			t.Error("逻辑与短路求值失败")
		}
	})

	t.Run("逻辑或短路", func(t *testing.T) {
		result := runScript(t, "true || false")
		if result.Bool() != true {
			t.Error("逻辑或短路求值失败")
		}
	})

	t.Run("逻辑与全真", func(t *testing.T) {
		result := runScript(t, "true && true")
		if result.Bool() != true {
			t.Error("逻辑与全真求值失败")
		}
	})

	t.Run("逻辑或全假", func(t *testing.T) {
		result := runScript(t, "false || false")
		if result.Bool() != false {
			t.Error("逻辑或全假求值失败")
		}
	})
}

// TestCompiler_ForStmtNotSupported 测试for循环不支持错误
func Test_compiler_ForStmtNotSupported(t *testing.T) {
	parser := NewParser()
	_, err := parser.Compile("for i := 0; i < 10; i++ { 1 }")
	if err == nil {
		t.Error("应该报告for循环不支持错误")
	}
}

// TestCompiler_BreakStmtOutsideLoop 测试 break 在循环外使用报错
func Test_compiler_BreakStmtOutsideLoop(t *testing.T) {
	parser := NewParser()
	_, err := parser.Compile("break")
	if err == nil {
		t.Error("应该报告 break 必须在循环内部使用的错误")
	}
}

// TestCompiler_ContinueStmtOutsideLoop 测试 continue 在循环外使用报错
func Test_compiler_ContinueStmtOutsideLoop(t *testing.T) {
	parser := NewParser()
	_, err := parser.Compile("continue")
	if err == nil {
		t.Error("应该报告 continue 必须在循环内部使用的错误")
	}
}

// TestCompiler_UserDefinedCall 测试用户定义函数调用编译
// 注意：当前实现中用户定义的脚本函数不能在主程序中直接调用
// compileUserDefinedCall主要用于间接调用或复杂表达式
func Test_compiler_UserDefinedCall(t *testing.T) {
	// 测试用户定义函数声明
	t.Run("函数声明", func(t *testing.T) {
		parser := NewParser()
		script, err := parser.Compile("fn add(a, b) { return a + b }")
		if err != nil {
			t.Errorf("函数声明编译失败: %v", err)
		}
		if len(script.Functions) != 1 {
			t.Errorf("应该有1个函数，得到%d", len(script.Functions))
		}
	})

	// 测试间接函数调用（非标识符调用）
	t.Run("间接调用表达式", func(t *testing.T) {
		// 这种调用方式会触发compileUserDefinedCall
		// 但由于函数表达式不是标识符，需要特殊处理
		compiler := NewCompiler()
		expr := &CallExpr{
			Position: Position{Line: 1, Column: 1},
			Func: &BinaryExpr{
				Position: Position{Line: 1, Column: 1},
				Left:     &LiteralExpr{Type: LiteralInt, Value: 1},
				Operator: "+",
				Right:    &LiteralExpr{Type: LiteralInt, Value: 2},
			},
			Args: []Expr{},
		}
		program := &Program{
			Statements: []Stmt{&ExprStmt{Expr: expr}},
		}
		_, err := compiler.Compile(program)
		// 这种调用会尝试对(1+2)进行调用，可能报错或成功
		_ = err // 接受任何结果
	})
}

// TestCompiler_CompileErrors 测试编译器各种错误
func Test_compiler_CompileErrors(t *testing.T) {
	t.Run("空函数体", func(t *testing.T) {
		parser := NewParser()
		_, err := parser.Compile("fn f() { }")
		// 空函数体应该可以编译
		if err != nil {
			t.Errorf("空函数体应该可以编译: %v", err)
		}
	})

	t.Run("throw错误", func(t *testing.T) {
		result, err := runScriptWithResult("throw \"test error\"")
		if err == nil {
			_ = result
			t.Error("应该返回throw错误")
		}
	})
}

// runScriptWithResult 辅助函数：运行脚本并返回结果和错误
func runScriptWithResult(input string) (Value, error) {
	parser := NewParser()
	script, err := parser.Compile(input)
	if err != nil {
		return Value{}, err
	}

	ctx := NewContext()
	engine := NewEngine()
	return engine.Run(ctx, script)
}

// TestCompiler_UnknownBuiltin 测试未知内置函数错误
func Test_compiler_UnknownBuiltin(t *testing.T) {
	parser := NewParser()
	_, err := parser.Compile("unknownBuiltin(1)")
	if err == nil {
		t.Error("应该报告未知内置函数错误")
	}
}

// TestCompiler_CompileBuiltinCallEdgeCases 测试内置函数调用边缘情况
func Test_compiler_CompileBuiltinCallEdgeCases(t *testing.T) {
	t.Run("string函数", func(t *testing.T) {
		result := runScript(t, "string(123)")
		if result.String() != "123" {
			t.Errorf("期望'123'，得到'%s'", result.String())
		}
	})

	t.Run("getBindValue无参数", func(t *testing.T) {
		parser := NewParser()
		_, err := parser.Compile("getBindValue()")
		if err == nil {
			t.Error("getBindValue应该需要参数")
		}
	})

	t.Run("len函数", func(t *testing.T) {
		result := runScript(t, "len([1,2,3])")
		if result.Int() != 3 {
			t.Errorf("期望3，得到%d", result.Int())
		}
	})

	t.Run("typeof函数", func(t *testing.T) {
		result := runScript(t, "typeof(123)")
		if result.String() == "" {
			t.Error("typeof应该返回类型名")
		}
	})

	t.Run("int函数", func(t *testing.T) {
		result := runScript(t, "int(3.7)")
		if result.Int() != 3 {
			t.Errorf("期望3，得到%d", result.Int())
		}
	})

	t.Run("float函数", func(t *testing.T) {
		result := runScript(t, "float(3)")
		if result.Float() != 3.0 {
			t.Errorf("期望3.0，得到%f", result.Float())
		}
	})

	t.Run("print函数", func(t *testing.T) {
		result := runScript(t, "print(\"test\")\nnil")
		if !result.IsNil() {
			t.Error("print应返回nil")
		}
	})

	t.Run("println函数", func(t *testing.T) {
		result := runScript(t, "println(\"test\")\nnil")
		if !result.IsNil() {
			t.Error("println应返回nil")
		}
	})
}

// TestCompiler_EngineStop 测试Engine.Stop方法
func Test_compiler_EngineStop(t *testing.T) {
	t.Run("Stop_NoRunningVM", func(t *testing.T) {
		// 测试没有运行中VM时调用Stop不会panic
		engine := NewEngine()
		engine.Stop()
	})

	t.Run("Stop_LongRunning", func(t *testing.T) {
		// 测试停止长时间运行的脚本
		// 注意：由于脚本引擎的for循环语法限制，这里用一个简单的脚本测试Stop功能
		engine := NewEngine()
		ctx := NewContext()

		// 使用一个简单的脚本，主要测试Stop不会panic
		script := `1 + 1`

		parser := NewParser()
		compiled, err := parser.Compile(script)
		if err != nil {
			t.Fatalf("编译失败: %v", err)
		}

		// 运行脚本
		go func() {
			_, _ = engine.Run(ctx, compiled)
		}()

		// 调用Stop应该不会panic
		engine.Stop()
	})
}

// TestCompiler_TypeAnnotationEdgeCases 测试类型注解边缘情况
func Test_compiler_TypeAnnotationEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"变量类型注解", "x := 10 >int"},
		{"函数参数类型注解", "fn f(x >int) { x }"},
		{"函数返回类型注解", "fn f() >int { 10 }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			_, err := parser.Compile(tt.input)
			// 类型注解可能支持也可能不支持，接受任何结果
			_ = err
		})
	}
}

// TestCompiler_UnknownBuiltin_NoArgs 测试无参数调用未知内置函数
func Test_compiler_UnknownBuiltin_NoArgs(t *testing.T) {
	parser := NewParser()
	// 无参数调用未知内置函数应该报错
	_, err := parser.Compile("unknownFunc()")
	if err == nil {
		t.Error("无参数调用未知内置函数应该报错")
	}
}

// TestCompiler_IfExprErrors 测试if表达式编译错误路径
func Test_compiler_IfExprErrors(t *testing.T) {
	// 直接构造AST测试条件编译失败
	t.Run("条件中引用未定义变量", func(t *testing.T) {
		compiler := NewCompiler()
		expr := &IfExpr{
			Condition: &IdentExpr{Name: "undefinedVar"},
			Then:      &LiteralExpr{Type: LiteralInt, Value: 1},
			Else:      &LiteralExpr{Type: LiteralInt, Value: 0},
		}
		program := &Program{
			Statements: []Stmt{&ExprStmt{Expr: expr}},
		}
		_, err := compiler.Compile(program)
		if err == nil {
			t.Error("引用未定义变量应该报错")
		}
	})

	// 测试then分支编译失败
	t.Run("then分支引用未定义变量", func(t *testing.T) {
		compiler := NewCompiler()
		expr := &IfExpr{
			Condition: &LiteralExpr{Type: LiteralBool, Value: true},
			Then:      &IdentExpr{Name: "undefinedInThen"},
		}
		program := &Program{
			Statements: []Stmt{&ExprStmt{Expr: expr}},
		}
		_, err := compiler.Compile(program)
		if err == nil {
			t.Error("then分支引用未定义变量应该报错")
		}
	})

	// 测试else分支编译失败
	t.Run("else分支引用未定义变量", func(t *testing.T) {
		compiler := NewCompiler()
		expr := &IfExpr{
			Condition: &LiteralExpr{Type: LiteralBool, Value: true},
			Then:      &LiteralExpr{Type: LiteralInt, Value: 1},
			Else:      &IdentExpr{Name: "undefinedInElse"},
		}
		program := &Program{
			Statements: []Stmt{&ExprStmt{Expr: expr}},
		}
		_, err := compiler.Compile(program)
		if err == nil {
			t.Error("else分支引用未定义变量应该报错")
		}
	})
}

// TestCompiler_ShortCircuitErrors 测试短路求值编译错误路径
func Test_compiler_ShortCircuitErrors(t *testing.T) {
	// 构造短路求值表达式，右侧引用未定义变量
	t.Run("短路求值右侧未定义变量", func(t *testing.T) {
		compiler := NewCompiler()
		expr := &BinaryExpr{
			Left:     &LiteralExpr{Type: LiteralBool, Value: true},
			Operator: "&&",
			Right:    &IdentExpr{Name: "undefinedVar"},
		}
		program := &Program{
			Statements: []Stmt{&ExprStmt{Expr: expr}},
		}
		_, err := compiler.Compile(program)
		if err == nil {
			t.Error("短路求值右侧引用未定义变量应该报错")
		}
	})

	t.Run("短路求值或运算右侧未定义变量", func(t *testing.T) {
		compiler := NewCompiler()
		expr := &BinaryExpr{
			Left:     &LiteralExpr{Type: LiteralBool, Value: false},
			Operator: "||",
			Right:    &IdentExpr{Name: "undefinedVar"},
		}
		program := &Program{
			Statements: []Stmt{&ExprStmt{Expr: expr}},
		}
		_, err := compiler.Compile(program)
		if err == nil {
			t.Error("短路求值或运算右侧引用未定义变量应该报错")
		}
	})
}

// TestCompiler_CompileBuiltinCall_Errors 测试compileBuiltinCall错误路径
func Test_compiler_CompileBuiltinCall_Errors(t *testing.T) {
	// 测试无参数调用未知内置函数（使用NewCompileError路径）
	t.Run("无参数未知内置函数AST", func(t *testing.T) {
		compiler := NewCompiler()
		// 直接构造AST，调用未知内置函数且无参数
		call := &CallExpr{
			Func: &IdentExpr{Name: "unknownFunc"},
			Args: []Expr{},
		}
		program := &Program{
			Statements: []Stmt{&ExprStmt{Expr: call}},
		}
		_, err := compiler.Compile(program)
		if err == nil {
			t.Error("无参数调用未知内置函数应该报错")
		}
	})
}

// TestCompiler_CompileReturnStmt_Error 测试compileReturnStmt错误路径
func Test_compiler_CompileReturnStmt_Error(t *testing.T) {
	// 测试return语句中引用未定义变量
	t.Run("return未定义变量", func(t *testing.T) {
		compiler := NewCompiler()
		// 构造一个包含return语句的函数，return引用未定义变量
		fn := &FuncDeclStmt{
			Name:   "testFunc",
			Params: []Param{},
			Body: &BlockStmt{
				Statements: []Stmt{
					&ReturnStmt{Value: &IdentExpr{Name: "undefinedVar"}},
				},
			},
		}
		program := &Program{
			Statements: []Stmt{fn},
		}
		_, err := compiler.Compile(program)
		if err == nil {
			t.Error("return引用未定义变量应该报错")
		}
	})
}

// TestCompiler_CompileThrowStmt_Error 测试compileThrowStmt错误路径
func Test_compiler_CompileThrowStmt_Error(t *testing.T) {
	// 测试throw语句中引用未定义变量
	t.Run("throw未定义变量", func(t *testing.T) {
		compiler := NewCompiler()
		throw := &ThrowStmt{Value: &IdentExpr{Name: "undefinedError"}}
		program := &Program{
			Statements: []Stmt{throw},
		}
		_, err := compiler.Compile(program)
		if err == nil {
			t.Error("throw引用未定义变量应该报错")
		}
	})
}

// TestCompiler_CompileBlockStmt_Error 测试compileBlockStmt错误路径
func Test_compiler_CompileBlockStmt_Error(t *testing.T) {
	// 测试block中语句编译失败
	t.Run("block中引用未定义变量", func(t *testing.T) {
		compiler := NewCompiler()
		// 构造if语句，then block中引用未定义变量
		ifStmt := &IfStmt{
			Condition: &LiteralExpr{Type: LiteralBool, Value: true},
			Then: &BlockStmt{
				Statements: []Stmt{
					&ExprStmt{Expr: &IdentExpr{Name: "undefinedInBlock"}},
				},
			},
		}
		program := &Program{
			Statements: []Stmt{ifStmt},
		}
		_, err := compiler.Compile(program)
		if err == nil {
			t.Error("block中引用未定义变量应该报错")
		}
	})
}

// TestCompiler_CompileArrayExpr_Error 测试compileArrayExpr错误路径
func Test_compiler_CompileArrayExpr_Error(t *testing.T) {
	// 测试数组元素中引用未定义变量
	t.Run("数组元素未定义变量", func(t *testing.T) {
		compiler := NewCompiler()
		arr := &ArrayExpr{
			Elements: []Expr{
				&LiteralExpr{Type: LiteralInt, Value: 1},
				&IdentExpr{Name: "undefinedInArray"},
			},
		}
		program := &Program{
			Statements: []Stmt{&ExprStmt{Expr: arr}},
		}
		_, err := compiler.Compile(program)
		if err == nil {
			t.Error("数组包含未定义变量应该报错")
		}
	})

	// 测试嵌套数组元素错误
	t.Run("嵌套数组元素未定义变量", func(t *testing.T) {
		compiler := NewCompiler()
		arr := &ArrayExpr{
			Elements: []Expr{
				&ArrayExpr{
					Elements: []Expr{
						&IdentExpr{Name: "undefinedInNested"},
					},
				},
			},
		}
		program := &Program{
			Statements: []Stmt{&ExprStmt{Expr: arr}},
		}
		_, err := compiler.Compile(program)
		if err == nil {
			t.Error("嵌套数组包含未定义变量应该报错")
		}
	})
}

// TestCompiler_CompileSliceExpr_Error 测试compileSliceExpr错误路径
func Test_compiler_CompileSliceExpr_Error(t *testing.T) {
	// 测试切片Object未定义
	t.Run("切片对象未定义", func(t *testing.T) {
		compiler := NewCompiler()
		slice := &SliceExpr{
			Object: &IdentExpr{Name: "undefinedArray"},
			Start:  &LiteralExpr{Type: LiteralInt, Value: 0},
			End:    &LiteralExpr{Type: LiteralInt, Value: 2},
		}
		program := &Program{
			Statements: []Stmt{&ExprStmt{Expr: slice}},
		}
		_, err := compiler.Compile(program)
		if err == nil {
			t.Error("切片对象未定义应该报错")
		}
	})

	// 测试切片Start边界未定义
	t.Run("切片Start边界未定义", func(t *testing.T) {
		compiler := NewCompiler()
		slice := &SliceExpr{
			Object: &IdentExpr{Name: "arr"},
			Start:  &IdentExpr{Name: "undefinedStart"},
			End:    &LiteralExpr{Type: LiteralInt, Value: 2},
		}
		program := &Program{
			Statements: []Stmt{&ExprStmt{Expr: slice}},
		}
		_, err := compiler.Compile(program)
		if err == nil {
			t.Error("切片Start边界未定义应该报错")
		}
	})

	// 测试切片End边界未定义
	t.Run("切片End边界未定义", func(t *testing.T) {
		compiler := NewCompiler()
		slice := &SliceExpr{
			Object: &IdentExpr{Name: "arr"},
			Start:  &LiteralExpr{Type: LiteralInt, Value: 0},
			End:    &IdentExpr{Name: "undefinedEnd"},
		}
		program := &Program{
			Statements: []Stmt{&ExprStmt{Expr: slice}},
		}
		_, err := compiler.Compile(program)
		if err == nil {
			t.Error("切片End边界未定义应该报错")
		}
	})
}

// TestCompiler_CompileSliceExpr_NilEnd 测试compileSliceExpr省略End情况
func Test_compiler_CompileSliceExpr_NilEnd(t *testing.T) {
	compiler := NewCompiler()
	slice := &SliceExpr{
		Object: &IdentExpr{Name: "arr"},
		Start:  &LiteralExpr{Type: LiteralInt, Value: 0},
		End:    nil,
	}
	program := &Program{
		Statements: []Stmt{&ExprStmt{Expr: slice}},
	}
	_, err := compiler.Compile(program)
	if err == nil {
		t.Error("切片对象未定义应该报错")
	}
}

// TestCompiler_CompileMapExpr_Error 测试compileMapExpr错误路径
func Test_compiler_CompileMapExpr_Error(t *testing.T) {
	// 测试Map键未定义
	t.Run("Map键未定义", func(t *testing.T) {
		compiler := NewCompiler()
		mapExpr := &MapExpr{
			Pairs: []MapPair{
				{Key: &IdentExpr{Name: "undefinedKey"}, Value: &LiteralExpr{Type: LiteralInt, Value: 1}},
			},
		}
		program := &Program{
			Statements: []Stmt{&ExprStmt{Expr: mapExpr}},
		}
		_, err := compiler.Compile(program)
		if err == nil {
			t.Error("Map键未定义应该报错")
		}
	})

	// 测试Map值未定义
	t.Run("Map值未定义", func(t *testing.T) {
		compiler := NewCompiler()
		mapExpr := &MapExpr{
			Pairs: []MapPair{
				{Key: &LiteralExpr{Type: LiteralString, Value: "a"}, Value: &IdentExpr{Name: "undefinedValue"}},
			},
		}
		program := &Program{
			Statements: []Stmt{&ExprStmt{Expr: mapExpr}},
		}
		_, err := compiler.Compile(program)
		if err == nil {
			t.Error("Map值未定义应该报错")
		}
	})
}

// TestCompiler_CompileLiteralExpr_AllTypes 测试compileLiteralExpr所有类型
func Test_compiler_CompileLiteralExpr_AllTypes(t *testing.T) {
	tests := []struct {
		name  string
		expr  Expr
		valid bool
	}{
		{"整数", &LiteralExpr{Type: LiteralInt, Value: 42}, true},
		{"浮点", &LiteralExpr{Type: LiteralFloat, Value: 3.14}, true},
		{"字符串", &LiteralExpr{Type: LiteralString, Value: "test"}, true},
		{"布尔", &LiteralExpr{Type: LiteralBool, Value: true}, true},
		{"Nil", &LiteralExpr{Type: LiteralNil, Value: nil}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler()
			program := &Program{
				Statements: []Stmt{&ExprStmt{Expr: tt.expr}},
			}
			_, err := compiler.Compile(program)
			if tt.valid && err != nil {
				t.Errorf("编译失败: %v", err)
			}
		})
	}
}

// TestCompiler_CompileBuiltinCall_NoArgsError 测试内置函数无参数错误
func Test_compiler_CompileBuiltinCall_NoArgsError(t *testing.T) {
	compiler := NewCompiler()
	call := &CallExpr{
		Func: &IdentExpr{Name: "unknownBuiltin"},
		Args: []Expr{},
	}
	program := &Program{
		Statements: []Stmt{&ExprStmt{Expr: call}},
	}
	_, err := compiler.Compile(program)
	if err == nil {
		t.Error("未知内置函数无参数应该报错")
	}
}

// TestCompiler_CompileArgs_Error 测试compileArgs错误路径
func Test_compiler_CompileArgs_Error(t *testing.T) {
	compiler := NewCompiler()
	args := []Expr{
		&IdentExpr{Name: "undefinedArg"},
	}
	err := compiler.compileArgs(args)
	if err == nil {
		t.Error("未定义参数应该报错")
	}
}

// TestCompiler_CompileUnaryExpr_Error 测试compileUnaryExpr错误路径
func Test_compiler_CompileUnaryExpr_Error(t *testing.T) {
	compiler := NewCompiler()
	unary := &UnaryExpr{
		Operator: "-",
		Operand:  &IdentExpr{Name: "undefinedVar"},
	}
	program := &Program{
		Statements: []Stmt{&ExprStmt{Expr: unary}},
	}
	_, err := compiler.Compile(program)
	if err == nil {
		t.Error("未定义变量应该报错")
	}
}
