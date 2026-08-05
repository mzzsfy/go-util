package script

import (
	"strings"
	"testing"
)

// ========== 错误和异常路径完整测试 ==========
// 测试脚本引擎中所有编译时错误和运行时错误的路径
// 包含throw语句、类型错误、除零、索引越界、函数调用错误、编译语法错误等

// ========== 辅助函数 ==========

// runRuntimeErrScript 编译并执行, 断言产生运行时错误
func runRuntimeErrScript(t *testing.T, input string) {
	t.Helper()
	parser := NewParser()
	script, err := parser.Compile(input)
	if err != nil {
		t.Fatalf("[%s] 编译失败(期望运行时错误): %v", input, err)
	}
	ctx := NewContext()
	engine := NewEngine()
	_, err = engine.Run(ctx, script)
	if err == nil {
		t.Errorf("[%s] 期望运行时错误, 但执行成功", input)
	}
}

// runRuntimeErrWithEngine 使用指定引擎执行, 断言产生运行时错误
func runRuntimeErrWithEngine(t *testing.T, input string, engine *Engine) {
	t.Helper()
	parser := NewParser()
	script, err := parser.Compile(input)
	if err != nil {
		t.Fatalf("[%s] 编译失败(期望运行时错误): %v", input, err)
	}
	ctx := NewContext()
	_, err = engine.Run(ctx, script)
	if err == nil {
		t.Errorf("[%s] 期望运行时错误, 但执行成功", input)
	}
}

// runCompileErrScript 断言产生编译错误
func runCompileErrScript(t *testing.T, input string) {
	t.Helper()
	parser := NewParser()
	_, err := parser.Compile(input)
	if err == nil {
		t.Errorf("[%s] 期望编译错误, 但编译成功", input)
	}
}

// ========== throw 语句测试 ==========

func Test_ErrorComplete_ThrowStringLiteral(t *testing.T) {
	t.Run("throw字符串字面量", func(t *testing.T) {
		runRuntimeErrScript(t, `throw "error"`)
	})
}

func Test_ErrorComplete_ThrowVariable(t *testing.T) {
	t.Run("throw变量", func(t *testing.T) {
		runRuntimeErrScript(t, "msg := \"err\"\nthrow msg")
	})
}

func Test_ErrorComplete_ThrowExpression(t *testing.T) {
	t.Run("throw表达式结果", func(t *testing.T) {
		runRuntimeErrScript(t, `throw "code: " + 123`)
	})
}

func Test_ErrorComplete_ThrowNumber(t *testing.T) {
	t.Run("throw数字", func(t *testing.T) {
		runRuntimeErrScript(t, `throw 42`)
	})
}

func Test_ErrorComplete_ThrowNil(t *testing.T) {
	t.Run("throw nil", func(t *testing.T) {
		runRuntimeErrScript(t, `throw nil`)
	})
}

func Test_ErrorComplete_ThrowBool(t *testing.T) {
	t.Run("throw布尔值", func(t *testing.T) {
		runRuntimeErrScript(t, `throw true`)
	})
}

func Test_ErrorComplete_ThrowInFunction(t *testing.T) {
	t.Run("函数中throw", func(t *testing.T) {
		runRuntimeErrScript(t, "fn f() { throw \"inner\" }\nf()")
	})
}

func Test_ErrorComplete_ThrowInCondition(t *testing.T) {
	t.Run("条件中throw", func(t *testing.T) {
		runRuntimeErrScript(t, `if true { throw "cond" }`)
	})
}

func Test_ErrorComplete_ThrowInLoop(t *testing.T) {
	t.Run("循环中throw", func(t *testing.T) {
		runRuntimeErrScript(t, "for i := 0; i < 3; i = i + 1 {\n  throw \"loop\"\n}")
	})
}

func Test_ErrorComplete_ThrowInNestedLoop(t *testing.T) {
	t.Run("嵌套循环中throw", func(t *testing.T) {
		input := "for i := 0; i < 2; i = i + 1 {\n" +
			"  for j := 0; j < 2; j = j + 1 {\n" +
			"    throw \"nested\"\n" +
			"  }\n" +
			"}"
		runRuntimeErrScript(t, input)
	})
}

// ========== 运行时错误 - 类型错误 ==========

func Test_ErrorComplete_NilArithmetic(t *testing.T) {
	t.Run("nil加nil", func(t *testing.T) {
		runRuntimeErrScript(t, `nil + nil`)
	})
	t.Run("nil乘整数", func(t *testing.T) {
		runRuntimeErrScript(t, `nil * 5`)
	})
}

func Test_ErrorComplete_BoolArithmetic(t *testing.T) {
	t.Run("布尔加整数", func(t *testing.T) {
		runRuntimeErrScript(t, `true + 1`)
	})
}

func Test_ErrorComplete_ArrayArithmetic(t *testing.T) {
	t.Run("数组加整数", func(t *testing.T) {
		runRuntimeErrScript(t, `[] + 5`)
	})
}

func Test_ErrorComplete_MapArithmetic(t *testing.T) {
	t.Run("Map加整数", func(t *testing.T) {
		runRuntimeErrScript(t, `{"a": 1} + 5`)
	})
}

func Test_ErrorComplete_DeleteNonMap(t *testing.T) {
	t.Run("对整数delete", func(t *testing.T) {
		runRuntimeErrScript(t, `delete(5, "a")`)
	})
	t.Run("对数组delete", func(t *testing.T) {
		runRuntimeErrScript(t, `delete([1], "a")`)
	})
	t.Run("对nil delete", func(t *testing.T) {
		runRuntimeErrScript(t, `delete(nil, "a")`)
	})
}

func Test_ErrorComplete_PushNonArray(t *testing.T) {
	t.Run("对整数push", func(t *testing.T) {
		runRuntimeErrScript(t, `push(5, 1)`)
	})
	t.Run("对Map push", func(t *testing.T) {
		runRuntimeErrScript(t, `push({}, 1)`)
	})
}

func Test_ErrorComplete_IndexNonCollection(t *testing.T) {
	t.Run("对nil索引", func(t *testing.T) {
		runRuntimeErrScript(t, `nil[0]`)
	})
	t.Run("对bool索引", func(t *testing.T) {
		runRuntimeErrScript(t, `true[0]`)
	})
	t.Run("对int索引", func(t *testing.T) {
		runRuntimeErrScript(t, `42[0]`)
	})
}

// ========== 运行时错误 - 除零 ==========

func Test_ErrorComplete_DivisionByZero(t *testing.T) {
	t.Run("整数除零", func(t *testing.T) {
		runRuntimeErrScript(t, `5 / 0`)
	})
	t.Run("整数模零", func(t *testing.T) {
		runRuntimeErrScript(t, `5 % 0`)
	})
	t.Run("零除以零", func(t *testing.T) {
		runRuntimeErrScript(t, `0 / 0`)
	})
	t.Run("浮点除零", func(t *testing.T) {
		runRuntimeErrScript(t, `5.0 / 0.0`)
	})
	t.Run("负数除零", func(t *testing.T) {
		runRuntimeErrScript(t, `-5 / 0`)
	})
}

// ========== 运行时错误 - 索引越界(赋值) ==========

func Test_ErrorComplete_ArrayStoreOutOfBounds(t *testing.T) {
	t.Run("数组越界赋值", func(t *testing.T) {
		runRuntimeErrScript(t, "arr := [1, 2]\narr[100] = 5")
	})
	t.Run("数组负索引赋值", func(t *testing.T) {
		runRuntimeErrScript(t, "arr := [1, 2]\narr[-1] = 5")
	})
}

func Test_ErrorComplete_StringStoreIndex(t *testing.T) {
	t.Run("字符串索引赋值", func(t *testing.T) {
		runRuntimeErrScript(t, "s := \"hi\"\ns[0] = \"H\"")
	})
}

// ========== 运行时错误 - 索引越界(读取返回nil) ==========

func Test_ErrorComplete_IndexReadOutOfBoundsReturnsNil(t *testing.T) {
	t.Run("数组越界读取返回nil", func(t *testing.T) {
		result := runScript(t, "arr := [1, 2]\narr[100]")
		if !result.IsNil() {
			t.Errorf("期望nil, 得到 %v", result)
		}
	})
	t.Run("数组负索引读取返回nil", func(t *testing.T) {
		result := runScript(t, "arr := [1, 2]\narr[-1]")
		if !result.IsNil() {
			t.Errorf("期望nil, 得到 %v", result)
		}
	})
	t.Run("空数组读取返回nil", func(t *testing.T) {
		result := runScript(t, "arr := []\narr[0]")
		if !result.IsNil() {
			t.Errorf("期望nil, 得到 %v", result)
		}
	})
	t.Run("字符串越界读取返回nil", func(t *testing.T) {
		result := runScript(t, "s := \"hi\"\ns[100]")
		if !result.IsNil() {
			t.Errorf("期望nil, 得到 %v", result)
		}
	})
}

// ========== 运行时错误 - 函数调用 ==========

func Test_ErrorComplete_CallNonFunction(t *testing.T) {
	t.Run("调用int作为函数", func(t *testing.T) {
		runRuntimeErrScript(t, "x := 42\nx()")
	})
	t.Run("调用nil作为函数", func(t *testing.T) {
		runRuntimeErrScript(t, "x := nil\nx()")
	})
}

func Test_ErrorComplete_ArgumentCountMismatch(t *testing.T) {
	t.Run("参数过少", func(t *testing.T) {
		input := "fn f(a) { return a }\nf()"
		runRuntimeErrScript(t, input)
	})
	t.Run("参数过多", func(t *testing.T) {
		input := "fn f(a) { return a }\nf(1, 2)"
		runRuntimeErrScript(t, input)
	})
}

func Test_ErrorComplete_StackOverflow(t *testing.T) {
	t.Run("无限递归栈溢出", func(t *testing.T) {
		input := "fn rec(n) {\n  return rec(n + 1)\n}\nrec(0)"
		engine := NewEngine(WithMaxCallDepth(20))
		runRuntimeErrWithEngine(t, input, engine)
	})
}

// ========== 编译时错误 - 语法错误 ==========

func Test_ErrorComplete_IncompleteExpression(t *testing.T) {
	t.Run("不完整表达式", func(t *testing.T) {
		runCompileErrScript(t, `1 +`)
	})
}

func Test_ErrorComplete_IncompleteStatement(t *testing.T) {
	t.Run("不完整语句", func(t *testing.T) {
		runCompileErrScript(t, `x :=`)
	})
}

func Test_ErrorComplete_UnclosedParen(t *testing.T) {
	t.Run("未闭合括号", func(t *testing.T) {
		runCompileErrScript(t, `(1 + 2`)
	})
}

func Test_ErrorComplete_UnclosedString(t *testing.T) {
	t.Run("未闭合字符串", func(t *testing.T) {
		runCompileErrScript(t, `"hello`)
	})
}

func Test_ErrorComplete_UnclosedArray(t *testing.T) {
	t.Run("未闭合数组", func(t *testing.T) {
		runCompileErrScript(t, `[1, 2, 3`)
	})
}

func Test_ErrorComplete_UnclosedMap(t *testing.T) {
	t.Run("未闭合Map", func(t *testing.T) {
		runCompileErrScript(t, `{"a": 1`)
	})
}

func Test_ErrorComplete_InvalidCharacters(t *testing.T) {
	t.Run("非法字符@", func(t *testing.T) {
		runCompileErrScript(t, `@`)
	})
	t.Run("非法字符$", func(t *testing.T) {
		runCompileErrScript(t, `$`)
	})
	t.Run("非法字符?", func(t *testing.T) {
		runCompileErrScript(t, `?`)
	})
}

func Test_ErrorComplete_MultipleAssignment(t *testing.T) {
	t.Run("多重赋值不支持", func(t *testing.T) {
		runCompileErrScript(t, `a, b := 1, 2`)
	})
}

// ========== 编译时错误 - 类型/作用域 ==========

func Test_ErrorComplete_UndefinedVariable(t *testing.T) {
	t.Run("使用未定义变量", func(t *testing.T) {
		runCompileErrScript(t, `undefined_var + 1`)
	})
}

func Test_ErrorComplete_BreakOutsideLoop(t *testing.T) {
	t.Run("break在循环外", func(t *testing.T) {
		runCompileErrScript(t, `break`)
	})
}

func Test_ErrorComplete_ContinueOutsideLoop(t *testing.T) {
	t.Run("continue在循环外", func(t *testing.T) {
		runCompileErrScript(t, `continue`)
	})
}

func Test_ErrorComplete_SemicolonUnsupported(t *testing.T) {
	t.Run("分号作为语句分隔符不支持", func(t *testing.T) {
		runCompileErrScript(t, `x := 1; y := 2`)
	})
}

// ========== 错误消息验证 ==========

func Test_ErrorComplete_CompileErrorContainsLine(t *testing.T) {
	t.Run("编译错误包含行号", func(t *testing.T) {
		parser := NewParser()
		input := "x := 1\ny := 2\nundefined_var"
		_, err := parser.Compile(input)
		if err == nil {
			t.Fatal("期望编译错误")
		}
		// 第3行使用了未定义变量
		if !strings.Contains(err.Error(), "行") {
			t.Errorf("错误消息应包含行号信息, 得到: %v", err)
		}
	})
}

func Test_ErrorComplete_RuntimeErrorContainsStackTrace(t *testing.T) {
	t.Run("运行时错误包含调用栈", func(t *testing.T) {
		parser := NewParser()
		script, err := parser.Compile("fn f() { throw \"test\" }\nf()")
		if err != nil {
			t.Fatalf("编译失败: %v", err)
		}
		engine := NewEngine()
		ctx := NewContext()
		_, err = engine.Run(ctx, script)
		if err == nil {
			t.Fatal("期望运行时错误")
		}
		// 函数中throw应包含调用栈
		if !strings.Contains(err.Error(), "调用栈") {
			t.Errorf("错误消息应包含调用栈, 得到: %v", err)
		}
	})
}

func Test_ErrorComplete_ErrorMessageReadable(t *testing.T) {
	t.Run("除零错误消息可读", func(t *testing.T) {
		parser := NewParser()
		script, err := parser.Compile(`5 / 0`)
		if err != nil {
			t.Fatalf("编译失败: %v", err)
		}
		engine := NewEngine()
		ctx := NewContext()
		_, err = engine.Run(ctx, script)
		if err == nil {
			t.Fatal("期望运行时错误")
		}
		if !strings.Contains(err.Error(), "除零") {
			t.Errorf("错误消息应包含'除零', 得到: %v", err)
		}
	})
	t.Run("类型错误消息可读", func(t *testing.T) {
		parser := NewParser()
		script, err := parser.Compile(`nil + nil`)
		if err != nil {
			t.Fatalf("编译失败: %v", err)
		}
		engine := NewEngine()
		ctx := NewContext()
		_, err = engine.Run(ctx, script)
		if err == nil {
			t.Fatal("期望运行时错误")
		}
		if !strings.Contains(err.Error(), "类型错误") {
			t.Errorf("错误消息应包含'类型错误', 得到: %v", err)
		}
	})
}

// ========== 复杂错误场景 ==========

func Test_ErrorComplete_NestedFunctionThrow(t *testing.T) {
	t.Run("嵌套函数中throw", func(t *testing.T) {
		input := "fn f() { throw \"always\" }\n" +
			"fn g() { return f() }\n" +
			"g()"
		runRuntimeErrScript(t, input)
	})
}

func Test_ErrorComplete_DeepRecursionThrow(t *testing.T) {
	t.Run("多层递归后throw", func(t *testing.T) {
		input := "fn rec(n) {\n" +
			"  if n <= 0 { throw \"deep\" }\n" +
			"  return rec(n - 1)\n" +
			"}\n" +
			"rec(5)"
		parser := NewParser()
		script, err := parser.Compile(input)
		if err != nil {
			t.Fatalf("编译失败: %v", err)
		}
		engine := NewEngine()
		ctx := NewContext()
		_, err = engine.Run(ctx, script)
		if err == nil {
			t.Fatal("期望运行时错误")
		}
		// 多层递归后throw, 调用栈应包含多层rec
		if !strings.Contains(err.Error(), "rec") {
			t.Errorf("错误消息应包含函数名'rec', 得到: %v", err)
		}
	})
}

func Test_ErrorComplete_ConditionalThrowInLoop(t *testing.T) {
	t.Run("循环中条件throw", func(t *testing.T) {
		input := "for i := 0; i < 5; i = i + 1 {\n" +
			"  if i == 2 { throw \"got\" }\n" +
			"}"
		runRuntimeErrScript(t, input)
	})
}

func Test_ErrorComplete_FunctionReturnThrowResult(t *testing.T) {
	t.Run("函数返回throw结果传播错误", func(t *testing.T) {
		input := "fn f() { throw \"propagated\" }\n" +
			"fn g() { return f() }\n" +
			"fn h() { return g() }\n" +
			"h()"
		parser := NewParser()
		script, err := parser.Compile(input)
		if err != nil {
			t.Fatalf("编译失败: %v", err)
		}
		engine := NewEngine()
		ctx := NewContext()
		_, err = engine.Run(ctx, script)
		if err == nil {
			t.Fatal("期望运行时错误")
		}
		// 错误消息应包含原始throw的值
		if !strings.Contains(err.Error(), "propagated") {
			t.Errorf("错误消息应包含'propagated', 得到: %v", err)
		}
	})
}
