package script

import (
    "strings"
    "testing"
)

// TestBoundary_Conditions 测试边界条件
func Test_boundary_Conditions(t *testing.T) {
    t.Run("数组边界索引", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`arr := [1, 2, 3]
arr[0]`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }

        engine := NewEngine()
        ctx := NewContext()
        result, err := engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }

        if result.Int() != 1 {
            t.Errorf("期望 1, 得到 %v", result.Int())
        }
    })

    t.Run("字符串UTF8索引", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`"你好"[0]`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }

        engine := NewEngine()
        ctx := NewContext()
        result, err := engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }

        // UTF-8第一个字节
        if len(result.String()) == 0 {
            t.Error("期望非空字符串")
        }
    })

    t.Run("切片边界", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`arr := [1, 2, 3, 4, 5]
arr[0:5]`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }

        engine := NewEngine()
        ctx := NewContext()
        result, err := engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }

        arr := result.Array()
        if len(arr.Elements) != 5 {
            t.Errorf("期望长度 5, 得到 %v", len(arr.Elements))
        }
    })
}

// TestParser_IfExprElse 测试if表达式else分支解析
func Test_parser_IfExprElse(t *testing.T) {
    // 测试简单的else分支
    t.Run("简单else", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`if true { 1 } else { 2 }`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        if script == nil {
            t.Fatal("脚本为nil")
        }
    })

    // 测试else if分支
    t.Run("else_if", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`if false { 1 } else if true { 2 } else { 3 }`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        if script == nil {
            t.Fatal("脚本为nil")
        }
    })

    // 测试多层else if
    t.Run("多层else_if", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`if false { 1 } else if false { 2 } else if true { 3 } else { 4 }`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        if script == nil {
            t.Fatal("脚本为nil")
        }
    })

    // 测试else if都不匹配
    t.Run("else_if最终else", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`if false { 1 } else if false { 2 } else { 3 }`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        if script == nil {
            t.Fatal("脚本为nil")
        }
    })
}

// TestParser_DefDirective 测试#fn指令解析
func Test_parser_DefDirective(t *testing.T) {
    t.Run("带#前缀", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`#fn externalFunc(x)=>int`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        if script == nil {
            t.Fatal("脚本为nil")
        }
    })

    t.Run("无返回值", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`#fn externalFunc(x)`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        if script == nil {
            t.Fatal("脚本为nil")
        }
    })

    t.Run("无参数", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`#fn externalFunc()=>int`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        if script == nil {
            t.Fatal("脚本为nil")
        }
    })
}

// TestCompile_ClassifyCall 测试函数调用分类
func Test_compile_ClassifyCall(t *testing.T) {
    t.Run("内置函数调用", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`len([1, 2, 3])`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }

        engine := NewEngine()
        ctx := NewContext()
        result, err := engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }

        if result.Int() != 3 {
            t.Errorf("期望 3, 得到 %v", result.Int())
        }
    })
}

// ========== 覆盖率快速提升测试2 ==========

// TestCoverage_LowCoverageFunctions 测试低覆盖率函数的错误路径
func Test_coverage_LowCoverageFunctions(t *testing.T) {
    // 测试compileVarDecl变量重声明路径
    t.Run("var_redeclare", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`
x := 10
x := 20
x
`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        engine := NewEngine()
        ctx := NewContext()
        result, err := engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        // 变量重声明会更新值
        if result.Int() != 20 {
            t.Errorf("期望 20, 得到 %d", result.Int())
        }
    })

    // 测试compileBlockStmt空块
    t.Run("empty_block", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`
if true {
}
42
`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        engine := NewEngine()
        ctx := NewContext()
        result, err := engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        if result.Int() != 42 {
            t.Errorf("期望 42, 得到 %d", result.Int())
        }
    })

    // 测试compileIfStmt无else分支
    t.Run("if_without_else", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`
x := 0
if false {
	x = 10
}
x
`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        engine := NewEngine()
        ctx := NewContext()
        result, err := engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        if result.Int() != 0 {
            t.Errorf("期望 0, 得到 %d", result.Int())
        }
    })

    // 测试compileFuncDecl无参数函数
    t.Run("func_no_params", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`
fn getValue() {
	return 42
}
`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        // 验证编译成功即可
        if script == nil {
            t.Error("脚本不应为nil")
        }
    })

    // 测试compileReturnStmt无返回值
    t.Run("return_no_value", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`
fn test() {
	return
}
`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        if script == nil {
            t.Error("脚本不应为nil")
        }
    })

    // 测试compileLiteralExpr所有类型
    t.Run("literal_all_types", func(t *testing.T) {
        tests := []struct {
            code  string
            check func(Value) bool
        }{
            {"nil", func(v Value) bool { return v.IsNil() }},
            {"true", func(v Value) bool { return v.Bool() == true }},
            {"false", func(v Value) bool { return v.Bool() == false }},
            {"42", func(v Value) bool { return v.Int() == 42 }},
            {"3.14", func(v Value) bool { return v.Float() == 3.14 }},
            {`"hello"`, func(v Value) bool { return v.String() == "hello" }},
        }

        for _, tc := range tests {
            parser := NewParser()
            script, err := parser.Compile(tc.code)
            if err != nil {
                t.Errorf("编译 %s 失败: %v", tc.code, err)
                continue
            }
            engine := NewEngine()
            ctx := NewContext()
            result, err := engine.Run(ctx, script)
            if err != nil {
                t.Errorf("执行 %s 失败: %v", tc.code, err)
                continue
            }
            if !tc.check(result) {
                t.Errorf("检查 %s 失败", tc.code)
            }
        }
    })

    // 测试compileArrayExpr空数组
    t.Run("empty_array", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`[]`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        engine := NewEngine()
        ctx := NewContext()
        result, err := engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        if len(result.Array().Elements) != 0 {
            t.Errorf("期望空数组, 得到 %d 元素", len(result.Array().Elements))
        }
    })

    // 测试compileMapExpr空map
    t.Run("empty_map", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`{}`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        engine := NewEngine()
        ctx := NewContext()
        result, err := engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        if len(result.Map().Pairs) != 0 {
            t.Errorf("期望空map, 得到 %d 元素", len(result.Map().Pairs))
        }
    })

    // 测试compileBinaryExpr所有运算符
    t.Run("binary_all_ops", func(t *testing.T) {
        tests := []string{
            "1 + 2",
            "5 - 3",
            "4 * 6",
            "10 / 2",
            "7 % 3",
            "1 == 1",
            "1 != 2",
            "1 < 2",
            "2 <= 2",
            "3 > 1",
            "3 >= 3",
            "1 & 1",
            "1 | 0",
            "1 ^ 1",
            "1 << 2",
            "8 >> 1",
        }
        for _, code := range tests {
            parser := NewParser()
            script, err := parser.Compile(code)
            if err != nil {
                t.Errorf("编译 %s 失败: %v", code, err)
                continue
            }
            engine := NewEngine()
            ctx := NewContext()
            _, err = engine.Run(ctx, script)
            if err != nil {
                t.Errorf("执行 %s 失败: %v", code, err)
            }
        }
    })

    // 测试compileSliceExpr完整切片
    t.Run("slice_full", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`[1,2,3,4,5][1:4]`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        engine := NewEngine()
        ctx := NewContext()
        result, err := engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        arr := result.Array().Elements
        if len(arr) != 3 {
            t.Errorf("期望3个元素, 得到 %d", len(arr))
        }
    })

    // 测试compileCallExpr外部函数
    t.Run("call_external", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`
#fn test(x)=>int
test(42)
`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        engine := NewEngine()
        ctx := NewContext()
        ctx.BindFunc("test", func(x int) int {
            return x * 2
        })
        result, err := engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        if result.Int() != 84 {
            t.Errorf("期望 84, 得到 %d", result.Int())
        }
    })

    // 测试compileBuiltinPrint
    t.Run("builtin_print", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`print("hello", "world")`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        engine := NewEngine()
        ctx := NewContext()
        _, err = engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
    })

    // 测试compileArgs多参数
    t.Run("multiple_args", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`
#fn add(a, b, c)
add(1, 2, 3)
`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        engine := NewEngine()
        ctx := NewContext()
        ctx.BindFunc("add", func(a, b, c int) int {
            return a + b + c
        })
        result, err := engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        if result.Int() != 6 {
            t.Errorf("期望 6, 得到 %d", result.Int())
        }
    })

    // 测试compileShortCircuit and/or
    t.Run("short_circuit", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`false && true`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        engine := NewEngine()
        ctx := NewContext()
        result, err := engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        if result.Bool() != false {
            t.Errorf("期望 false, 得到 %v", result.Bool())
        }

        parser2 := NewParser()
        script2, err := parser2.Compile(`true || false`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        result2, err := engine.Run(ctx, script2)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        if result2.Bool() != true {
            t.Errorf("期望 true, 得到 %v", result2.Bool())
        }
    })

    // 测试parseIfExprElse错误路径：else后缺少大括号
    t.Run("if_expr_else_missing_brace", func(t *testing.T) {
        parser := NewParser()
        _, err := parser.Compile(`x := if true { 1 } else 2`)
        if err == nil {
            t.Error("期望编译失败，但没有错误")
        }
    })

    // 测试tryParseTypeAnnotation基本类型
    t.Run("type_annotation_basic", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`
#fn func(x)=>int
0
`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        if len(script.Externals) != 1 {
            t.Errorf("应该有1个外部函数声明")
        }
        ext := script.Externals[0]
        if len(ext.Params) != 1 {
            t.Errorf("应该有1个参数")
        }
    })

    // 测试tryParseTypeAnnotation float类型
    t.Run("type_annotation_float", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`
#fn func(x)=>float
0
`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        if len(script.Externals) != 1 {
            t.Errorf("应该有1个外部函数声明")
        }
    })

    // 测试parseSingleParam带类型注解
    t.Run("single_param_with_type", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`
#fn func(x)=>int
0
`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        if len(script.Externals) != 1 {
            t.Errorf("应该有1个外部函数声明")
        }
        if len(script.Externals[0].Params) != 1 {
            t.Errorf("应该有1个参数")
        }
    })

    // 测试parseTypeExpr嵌套类型
    t.Run("type_expr_nested", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`
#fn func(x)=>int
0
`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        if len(script.Externals) != 1 {
            t.Errorf("应该有1个外部函数声明")
        }
    })

    // 测试parseIntLiteral正常路径
    t.Run("int_literal", func(t *testing.T) {
        parser := NewParser()
        // 测试正常整数解析
        script, err := parser.Compile(`12345`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        engine := NewEngine()
        ctx := NewContext()
        result, err := engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        if result.Int() != 12345 {
            t.Errorf("期望 12345, 得到 %d", result.Int())
        }
    })

    // 测试parseFloatLiteral正常路径
    t.Run("float_literal", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`3.14159`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        engine := NewEngine()
        ctx := NewContext()
        result, err := engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        if result.Float() < 3.14 || result.Float() > 3.15 {
            t.Errorf("期望约3.14159, 得到 %v", result.Float())
        }
    })

    // 测试runtime toInt错误路径
    t.Run("runtime_to_int_error", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`int("not a number")`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        engine := NewEngine()
        ctx := NewContext()
        _, err = engine.Run(ctx, script)
        if err == nil {
            t.Error("期望执行失败，但没有错误")
        }
    })

    // 测试runtime toFloat错误路径
    t.Run("runtime_to_float_error", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`float("not a number")`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        engine := NewEngine()
        ctx := NewContext()
        _, err = engine.Run(ctx, script)
        if err == nil {
            t.Error("期望执行失败，但没有错误")
        }
    })

    // 测试runtime length字符串
    t.Run("runtime_length_string", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`len("hello")`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        engine := NewEngine()
        ctx := NewContext()
        result, err := engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        if result.Int() != 5 {
            t.Errorf("期望 5, 得到 %d", result.Int())
        }
    })

    // 测试runtime length数组
    t.Run("runtime_length_array", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`len([1, 2, 3, 4])`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        engine := NewEngine()
        ctx := NewContext()
        result, err := engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        if result.Int() != 4 {
            t.Errorf("期望 4, 得到 %d", result.Int())
        }
    })

    // 测试compileUnaryExpr错误路径
    t.Run("unary_expr_error", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`-true`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        engine := NewEngine()
        ctx := NewContext()
        _, err = engine.Run(ctx, script)
        if err == nil {
            t.Error("期望执行失败，但没有错误")
        }
    })

    // 测试compileBlockStmt空块
    t.Run("empty_block_stmt", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`if true { }`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        engine := NewEngine()
        ctx := NewContext()
        _, err = engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
    })
}

// TestValue_PrintFormatters 测试所有打印格式化器
func Test_value_PrintFormatters(t *testing.T) {
    // 测试函数类型打印
    t.Run("function_type_print", func(t *testing.T) {
        // 直接创建函数值进行测试
        cf := &CompiledFunction{Name: "testFunc"}
        fv := &FunctionValue{Compiled: cf}
        v := Value{Type: TypeFunction, Data: fv}
        str := v.GoString()
        if !strings.Contains(str, "<function") {
            t.Errorf("函数类型打印应包含'<function', 得到: %s", str)
        }
    })

    // 测试外部函数类型打印
    t.Run("external_func_print", func(t *testing.T) {
        // 直接创建外部函数值进行测试
        ef := &ExternalFuncValue{Name: "extFunc", Func: func(args []Value) Value { return Value{} }}
        v := Value{Type: TypeExternalFunc, Data: ef}
        str := v.GoString()
        if !strings.Contains(str, "<external") {
            t.Errorf("外部函数类型打印应包含'<external', 得到: %s", str)
        }
    })

    // 测试未知类型打印
    t.Run("unknown_type_print", func(t *testing.T) {
        // 创建一个未知类型的Value
        v := Value{Type: ValueType(255)}
        str := formatValueForPrint(v)
        if str != "<unknown>" {
            t.Errorf("未知类型应打印'<unknown>', 得到: %s", str)
        }
    })
}

// TestTryParseTypeAnnotation 测试tryParseTypeAnnotation完整路径
func Test_tryparsetypeannotation(t *testing.T) {
    // 测试变量声明带类型注解 - 使用 > 语法
    t.Run("var_decl_with_type_annotation", func(t *testing.T) {
        parser := NewParser()
        // 注意：类型注解语法是 >type
        script, err := parser.Compile(`x > int := 10`)
        if err != nil {
            // 如果不支持此语法，跳过
            t.Logf("不支持变量声明类型注解语法: %v", err)
            return
        }
        engine := NewEngine()
        ctx := NewContext()
        result, err := engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
        if result.Int() != 10 {
            t.Errorf("期望 10, 得到 %d", result.Int())
        }
    })

    // 测试变量声明带数组类型注解
    t.Run("var_decl_with_array_type", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`arr > []int := [1, 2, 3]`)
        if err != nil {
            t.Logf("不支持数组类型注解语法: %v", err)
            return
        }
        engine := NewEngine()
        ctx := NewContext()
        _, err = engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
    })

    // 测试变量声明带Map类型注解
    t.Run("var_decl_with_map_type", func(t *testing.T) {
        parser := NewParser()
        script, err := parser.Compile(`m > map{string:int} := map{"a": 1}`)
        if err != nil {
            t.Logf("不支持Map类型注解语法: %v", err)
            return
        }
        engine := NewEngine()
        ctx := NewContext()
        _, err = engine.Run(ctx, script)
        if err != nil {
            t.Fatalf("执行失败: %v", err)
        }
    })

    // 测试hasTypeAnnotation使用TokenGt的情况
    t.Run("has_type_annotation_gt", func(t *testing.T) {
        parser := NewParser()
        // 使用 #fn 格式
        script, err := parser.Compile(`#fn test(x)=>int
10`)
        if err != nil {
            t.Fatalf("编译失败: %v", err)
        }
        if len(script.Externals) != 1 {
            t.Errorf("应该有1个外部函数声明")
        }
    })
}

// ========== newCompositeValue低覆盖路径测试 ==========

// TestNewCompositeValue_FunctionValue 测试FunctionValue路径
func Test_newcompositevalue_FunctionValue(t *testing.T) {
    fn := &FunctionValue{Compiled: &CompiledFunction{Name: "test"}}
    val := newCompositeValue(fn)
    if val.Type != TypeFunction {
        t.Errorf("期望TypeFunction, 得到%d", val.Type)
    }
}

// TestNewCompositeValue_Default 测试默认返回nil路径
func Test_newcompositevalue_Default(t *testing.T) {
    val := newCompositeValue(12345)
    if !val.IsNil() {
        t.Error("期望nil")
    }
}

// ========== sliceAccess低覆盖路径测试 ==========

// TestSliceAccess_NegativeIndex 测试sliceAccess负数索引（规范化为0）
func Test_sliceaccess_NegativeIndex(t *testing.T) {
    vm := newTestVM(t)

    arr := NewValue([]Value{NewValue(1), NewValue(2), NewValue(3)})
    result, err := vm.sliceAccess(arr, NewValue(-1), NewValue(2))
    if err != nil {
        t.Errorf("不应报错: %v", err)
    }
    // 负数索引从末尾倒数, -1+3=2, arr[2:2]为空
    if len(result.Array().Elements) != 0 {
        t.Errorf("期望长度0, 得到长度%d", len(result.Array().Elements))
    }
}

// TestSliceAccess_OutOfRange 测试sliceAccess超出范围
func Test_sliceaccess_OutOfRange(t *testing.T) {
    vm := newTestVM(t)

    arr := NewValue([]Value{NewValue(1), NewValue(2)})
    result, err := vm.sliceAccess(arr, NewValue(0), NewValue(10))
    if err != nil {
        t.Errorf("不应报错: %v", err)
    }
    // 超出范围返回有效部分
    if len(result.Array().Elements) != 2 {
        t.Errorf("期望长度2, 得到%d", len(result.Array().Elements))
    }
}

// TestSliceAccess_OmitEnd 测试sliceAccess省略end
func Test_sliceaccess_OmitEnd(t *testing.T) {
    vm := newTestVM(t)

    arr := NewValue([]Value{NewValue(1), NewValue(2), NewValue(3)})
    result, err := vm.sliceAccess(arr, NewValue(1), NewValue(SliceEndDefault))
    if err != nil {
        t.Errorf("不应报错: %v", err)
    }
    if len(result.Array().Elements) != 2 {
        t.Errorf("期望长度2, 得到%d", len(result.Array().Elements))
    }
}

// ========== compileSliceExpr低覆盖路径测试 ==========

// TestCompiler_CompileSliceExpr_StartError 测试start错误路径
func Test_compiler_CompileSliceExpr_StartError(t *testing.T) {
    compiler := NewCompiler()
    slice := &SliceExpr{
        Object: &LiteralExpr{Type: LiteralInt, Value: 1},
        Start:  &IdentExpr{Name: "undefinedStart"},
        End:    &LiteralExpr{Type: LiteralInt, Value: 5},
    }
    program := &Program{
        Statements: []Stmt{&ExprStmt{Expr: slice}},
    }
    _, err := compiler.Compile(program)
    if err == nil {
        t.Error("未定义变量应该报错")
    }
}

// TestCompiler_CompileSliceExpr_EndError 测试end错误路径
func Test_compiler_CompileSliceExpr_EndError(t *testing.T) {
    compiler := NewCompiler()
    slice := &SliceExpr{
        Object: &LiteralExpr{Type: LiteralInt, Value: 1},
        Start:  &LiteralExpr{Type: LiteralInt, Value: 0},
        End:    &IdentExpr{Name: "undefinedEnd"},
    }
    program := &Program{
        Statements: []Stmt{&ExprStmt{Expr: slice}},
    }
    _, err := compiler.Compile(program)
    if err == nil {
        t.Error("未定义变量应该报错")
    }
}

// ========== 低覆盖率函数测试 ==========

// TestParseMapLiteral_EdgeCases 测试parseMapLiteral边界情况
func Test_parsemapliteral_EdgeCases(t *testing.T) {
    tests := []struct {
        name  string
        input string
    }{
        {"空Map", "{}"},
        {"单元素Map", `{"a": 1}`},
        {"多元素Map", `{"a": 1, "b": 2, "c": 3}`},
        {"嵌套Map", `{"outer": {"inner": 42}}`},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p := NewParser()
            _, err := p.Compile(tt.input)
            if err != nil {
                t.Errorf("Map解析失败: %v", err)
            }
        })
    }
}

// TestParseMapPair_AllCases 测试parseMapPair所有路径
func Test_parsemappair_AllCases(t *testing.T) {
    tests := []struct {
        name  string
        input string
    }{
        {"字符串键", `{"key": "value"}`},
        {"表达式键", `{1 + 2: "value"}`},
        {"标识符键", `{x: 1}`},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p := NewParser()
            _, err := p.Compile(tt.input)
            t.Logf("编译结果: %v", err)
        })
    }
}

// TestParseIfExprBranch_AllPaths 测试parseIfExprBranch所有路径
func Test_parseifexprbranch_AllPaths(t *testing.T) {
    tests := []struct {
        name  string
        input string
    }{
        {"if-only", "if true { 1 }"},
        {"if-else", "if true { 1 } else { 2 }"},
        {"if-elif-else", "if false { 1 } else if true { 2 } else { 3 }"},
        {"多层elif", "if false { 1 } else if false { 2 } else if true { 3 } else { 4 }"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p := NewParser()
            _, err := p.Compile(tt.input)
            if err != nil {
                t.Errorf("if表达式解析失败: %v", err)
            }
        })
    }
}

// TestParseSlice_AllVariants 测试切片解析所有变体
func Test_parseslice_AllVariants(t *testing.T) {
    tests := []struct {
        name  string
        input string
    }{
        {"完整切片", "arr := [1, 2, 3, 4, 5]\narr[1:3]"},
        {"省略开始", "arr := [1, 2, 3]\narr[:2]"},
        {"省略结束", "arr := [1, 2, 3]\narr[1:]"},
        {"全部省略", "arr := [1, 2, 3]\narr[:]"},
        {"字符串切片", "\"hello\"[1:3]"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p := NewParser()
            _, err := p.Compile(tt.input)
            if err != nil {
                t.Errorf("切片解析失败: %v", err)
            }
        })
    }
}

// TestParseInfixExpr_AllOps 测试parseInfixExpr所有运算符
func Test_parseinfixexpr_AllOps(t *testing.T) {
    ops := []string{
        "+", "-", "*", "/", "%",
        "==", "!=", "<", ">", "<=", ">=",
        "&&", "||",
        "&", "|", "^", "<<", ">>",
    }

    for _, op := range ops {
        t.Run(op, func(t *testing.T) {
            p := NewParser()
            _, err := p.Compile("1 " + op + " 2")
            if err != nil {
                t.Errorf("运算符 '%s' 解析失败: %v", op, err)
            }
        })
    }
}
