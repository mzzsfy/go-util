package script

import (
	"strings"
	"testing"
)

// ========== 错误消息和诊断质量测试 ==========
// 验证编译错误的位置准确性、消息内容完整性，以及运行时错误的诊断质量
// 所有测试围绕 CompileError 的 Line/Column/Message 和 RuntimeError 的 Message/StackTrace 展开

// ========== 辅助函数 ==========

// compileErr 编译并返回 CompileError，若编译成功或返回其他类型错误则标记失败
func compileErr(t *testing.T, code string) *CompileError {
	t.Helper()
	parser := NewParser()
	_, err := parser.Compile(code)
	if err == nil {
		t.Fatalf("[%q] 期望编译错误，但编译成功", code)
		return nil
	}
	ce, ok := err.(*CompileError)
	if !ok {
		t.Fatalf("[%q] 期望 *CompileError，得到 %T: %v", code, err, err)
		return nil
	}
	return ce
}

// runtimeErr 编译并运行，返回 RuntimeError，若运行成功或返回其他类型错误则标记失败
func runtimeErr(t *testing.T, code string) *RuntimeError {
	t.Helper()
	return runtimeErrWithCtx(t, code, NewContext())
}

// runtimeErrWithCtx 使用指定上下文编译并运行，返回 RuntimeError
func runtimeErrWithCtx(t *testing.T, code string, ctx *Context) *RuntimeError {
	t.Helper()
	parser := NewParser()
	script, err := parser.Compile(code)
	if err != nil {
		t.Fatalf("[%q] 编译失败(期望运行时错误): %v", code, err)
		return nil
	}
	engine := NewEngine()
	_, err = engine.Run(ctx, script)
	if err == nil {
		t.Fatalf("[%q] 期望运行时错误，但执行成功", code)
		return nil
	}
	re, ok := err.(*RuntimeError)
	if !ok {
		t.Fatalf("[%q] 期望 *RuntimeError，得到 %T: %v", code, err, err)
		return nil
	}
	return re
}

// runtimeErrWithEngine 使用指定引擎编译并运行，返回 RuntimeError
func runtimeErrWithEngine(t *testing.T, code string, engine *Engine) *RuntimeError {
	t.Helper()
	parser := NewParser()
	script, err := parser.Compile(code)
	if err != nil {
		t.Fatalf("[%q] 编译失败(期望运行时错误): %v", code, err)
		return nil
	}
	ctx := NewContext()
	_, err = engine.Run(ctx, script)
	if err == nil {
		t.Fatalf("[%q] 期望运行时错误，但执行成功", code)
		return nil
	}
	re, ok := err.(*RuntimeError)
	if !ok {
		t.Fatalf("[%q] 期望 *RuntimeError，得到 %T: %v", code, err, err)
		return nil
	}
	return re
}

// ========== 编译错误位置准确性 ==========

// Test_ErrorDiagnostic_CompilePos_Line1 第1行的错误应报告 Line=1
func Test_ErrorDiagnostic_CompilePos_Line1(t *testing.T) {
	t.Run("第1行未定义变量", func(t *testing.T) {
		ce := compileErr(t, "undefined_var")
		if ce.Line != 1 {
			t.Errorf("Line: 期望 1, 得到 %d", ce.Line)
		}
	})
	t.Run("第1行无效字符", func(t *testing.T) {
		ce := compileErr(t, "@")
		if ce.Line != 1 {
			t.Errorf("Line: 期望 1, 得到 %d", ce.Line)
		}
	})
}

// Test_ErrorDiagnostic_CompilePos_Line3 第3行的错误应报告 Line=3
func Test_ErrorDiagnostic_CompilePos_Line3(t *testing.T) {
	t.Run("第3行未定义变量", func(t *testing.T) {
		ce := compileErr(t, "x := 1\ny := 2\nundefined_var")
		if ce.Line != 3 {
			t.Errorf("Line: 期望 3, 得到 %d", ce.Line)
		}
	})
}

// Test_ErrorDiagnostic_CompilePos_ColumnAccuracy Column准确性
func Test_ErrorDiagnostic_CompilePos_ColumnAccuracy(t *testing.T) {
	t.Run("行首字符列号为1", func(t *testing.T) {
		ce := compileErr(t, "@")
		if ce.Column != 1 {
			t.Errorf("Column: 期望 1, 得到 %d", ce.Column)
		}
	})
	t.Run("表达式末尾位置", func(t *testing.T) {
		// "1 +" 共4字符，+后空格位置应为列4
		ce := compileErr(t, "1 +")
		if ce.Line != 1 {
			t.Errorf("Line: 期望 1, 得到 %d", ce.Line)
		}
		if ce.Column <= 0 {
			t.Errorf("Column 应为正数, 得到 %d", ce.Column)
		}
	})
}

// Test_ErrorDiagnostic_CompilePos_MultiLine 多行脚本定位错误行
func Test_ErrorDiagnostic_CompilePos_MultiLine(t *testing.T) {
	t.Run("第4行错误", func(t *testing.T) {
		code := "x := 1\ny := 2\nz := 3\nundefined_var"
		ce := compileErr(t, code)
		if ce.Line != 4 {
			t.Errorf("Line: 期望 4, 得到 %d", ce.Line)
		}
	})
	t.Run("第2行错误", func(t *testing.T) {
		code := "x := 1\nundefined_var"
		ce := compileErr(t, code)
		if ce.Line != 2 {
			t.Errorf("Line: 期望 2, 得到 %d", ce.Line)
		}
	})
}

// Test_ErrorDiagnostic_CompilePos_AfterBlankLine 空行后错误的位置
func Test_ErrorDiagnostic_CompilePos_AfterBlankLine(t *testing.T) {
	t.Run("两个空行后错误在第3行", func(t *testing.T) {
		ce := compileErr(t, "\n\nundefined_var")
		if ce.Line != 3 {
			t.Errorf("Line: 期望 3, 得到 %d", ce.Line)
		}
	})
	t.Run("一个空行后错误在第2行", func(t *testing.T) {
		ce := compileErr(t, "\nundefined_var")
		if ce.Line != 2 {
			t.Errorf("Line: 期望 2, 得到 %d", ce.Line)
		}
	})
}

// ========== 编译错误消息内容 ==========

// Test_ErrorDiagnostic_CompileMsg_UndefinedVar 未定义变量错误包含变量名
func Test_ErrorDiagnostic_CompileMsg_UndefinedVar(t *testing.T) {
	t.Run("变量名出现在消息中", func(t *testing.T) {
		ce := compileErr(t, "myUndefinedVar")
		if !strings.Contains(ce.Message, "myUndefinedVar") {
			t.Errorf("错误消息应包含变量名 'myUndefinedVar'\n实际: %s", ce.Message)
		}
	})
	t.Run("不同变量名出现在消息中", func(t *testing.T) {
		ce := compileErr(t, "anotherVar")
		if !strings.Contains(ce.Message, "anotherVar") {
			t.Errorf("错误消息应包含变量名 'anotherVar'\n实际: %s", ce.Message)
		}
	})
}

// Test_ErrorDiagnostic_CompileMsg_UnclosedString 未闭合字符串错误包含位置信息
func Test_ErrorDiagnostic_CompileMsg_UnclosedString(t *testing.T) {
	t.Run("未闭合字符串消息", func(t *testing.T) {
		ce := compileErr(t, `"unclosed`)
		if !strings.Contains(ce.Message, "字符串") {
			t.Errorf("错误消息应包含 '字符串'\n实际: %s", ce.Message)
		}
		if !strings.Contains(ce.Message, "闭合") || !strings.Contains(ce.Message, "引号") {
			t.Errorf("错误消息应提示闭合/引号\n实际: %s", ce.Message)
		}
	})
}

// Test_ErrorDiagnostic_CompileMsg_UnclosedDelimiters 未闭合括号错误包含位置
func Test_ErrorDiagnostic_CompileMsg_UnclosedDelimiters(t *testing.T) {
	t.Run("未闭合圆括号", func(t *testing.T) {
		ce := compileErr(t, "(1 + 2")
		if !strings.Contains(ce.Message, ")") {
			t.Errorf("错误消息应包含 ')'\n实际: %s", ce.Message)
		}
	})
	t.Run("未闭合方括号", func(t *testing.T) {
		ce := compileErr(t, "[1, 2, 3")
		if !strings.Contains(ce.Message, "]") {
			t.Errorf("错误消息应包含 ']'\n实际: %s", ce.Message)
		}
	})
	t.Run("未闭合花括号", func(t *testing.T) {
		ce := compileErr(t, `{"a": 1`)
		if !strings.Contains(ce.Message, "}") {
			t.Errorf("错误消息应包含 '}'\n实际: %s", ce.Message)
		}
	})
}

// Test_ErrorDiagnostic_CompileMsg_InvalidChar 无效字符错误包含字符本身
func Test_ErrorDiagnostic_CompileMsg_InvalidChar(t *testing.T) {
	chars := []string{"@", "$", "?"}
	for _, ch := range chars {
		t.Run("非法字符"+ch, func(t *testing.T) {
			ce := compileErr(t, ch)
			if !strings.Contains(ce.Message, ch) {
				t.Errorf("错误消息应包含字符 %q\n实际: %s", ch, ce.Message)
			}
		})
	}
}

// Test_ErrorDiagnostic_CompileMsg_SyntaxReadable 语法错误消息可读性好
func Test_ErrorDiagnostic_CompileMsg_SyntaxReadable(t *testing.T) {
	t.Run("break在循环外消息可读", func(t *testing.T) {
		ce := compileErr(t, "break")
		if !strings.Contains(ce.Message, "break") {
			t.Errorf("错误消息应包含 'break'\n实际: %s", ce.Message)
		}
		if !strings.Contains(ce.Message, "循环") {
			t.Errorf("错误消息应说明需在循环中使用\n实际: %s", ce.Message)
		}
	})
	t.Run("continue在循环外消息可读", func(t *testing.T) {
		ce := compileErr(t, "continue")
		if !strings.Contains(ce.Message, "continue") {
			t.Errorf("错误消息应包含 'continue'\n实际: %s", ce.Message)
		}
		if !strings.Contains(ce.Message, "循环") {
			t.Errorf("错误消息应说明需在循环中使用\n实际: %s", ce.Message)
		}
	})
}

// ========== 运行时错误 - 算术 ==========

// Test_ErrorDiagnostic_RT_Arithmetic_DivZero 除零错误消息
func Test_ErrorDiagnostic_RT_Arithmetic_DivZero(t *testing.T) {
	t.Run("整数除零", func(t *testing.T) {
		re := runtimeErr(t, `5 / 0`)
		if !strings.Contains(re.Message, "除零") {
			t.Errorf("消息应包含 '除零'\n实际: %s", re.Message)
		}
	})
	t.Run("浮点除零", func(t *testing.T) {
		re := runtimeErr(t, `5.0 / 0.0`)
		if !strings.Contains(re.Message, "除零") {
			t.Errorf("消息应包含 '除零'\n实际: %s", re.Message)
		}
	})
}

// Test_ErrorDiagnostic_RT_Arithmetic_TypeError 类型错误消息
func Test_ErrorDiagnostic_RT_Arithmetic_TypeError(t *testing.T) {
	t.Run("nil加整数包含类型名", func(t *testing.T) {
		re := runtimeErr(t, `nil + 5`)
		if !strings.Contains(re.Message, "类型错误") {
			t.Errorf("消息应包含 '类型错误'\n实际: %s", re.Message)
		}
		if !strings.Contains(re.Message, "nil") {
			t.Errorf("消息应包含操作数类型 'nil'\n实际: %s", re.Message)
		}
		if !strings.Contains(re.Message, "int") {
			t.Errorf("消息应包含操作数类型 'int'\n实际: %s", re.Message)
		}
	})
	t.Run("数组加整数包含类型名", func(t *testing.T) {
		re := runtimeErr(t, `[] + 5`)
		if !strings.Contains(re.Message, "类型错误") {
			t.Errorf("消息应包含 '类型错误'\n实际: %s", re.Message)
		}
	})
}

// Test_ErrorDiagnostic_RT_Arithmetic_ModZero 模零错误消息
func Test_ErrorDiagnostic_RT_Arithmetic_ModZero(t *testing.T) {
	t.Run("整数模零", func(t *testing.T) {
		re := runtimeErr(t, `5 % 0`)
		if !strings.Contains(re.Message, "取模") || !strings.Contains(re.Message, "零") {
			t.Errorf("消息应包含 '取模' 和 '零'\n实际: %s", re.Message)
		}
	})
}

// Test_ErrorDiagnostic_RT_Arithmetic_BitOp 负数位运算消息(浮点位运算)
func Test_ErrorDiagnostic_RT_Arithmetic_BitOp(t *testing.T) {
	t.Run("浮点数位运算包含类型信息", func(t *testing.T) {
		re := runtimeErr(t, `1.5 & 2`)
		if !strings.Contains(re.Message, "位运算") {
			t.Errorf("消息应包含 '位运算'\n实际: %s", re.Message)
		}
		if !strings.Contains(re.Message, "整数") {
			t.Errorf("消息应说明只支持整数\n实际: %s", re.Message)
		}
	})
}

// ========== 运行时错误 - 索引 ==========

// Test_ErrorDiagnostic_RT_Index_TypeError 类型错误索引(对int做索引)
func Test_ErrorDiagnostic_RT_Index_TypeError(t *testing.T) {
	t.Run("对int索引报错", func(t *testing.T) {
		re := runtimeErr(t, `42[0]`)
		if !strings.Contains(re.Message, "类型错误") {
			t.Errorf("消息应包含 '类型错误'\n实际: %s", re.Message)
		}
		if !strings.Contains(re.Message, "int") {
			t.Errorf("消息应包含 'int'\n实际: %s", re.Message)
		}
	})
}

// Test_ErrorDiagnostic_RT_Index_Nil nil索引错误
func Test_ErrorDiagnostic_RT_Index_Nil(t *testing.T) {
	t.Run("对nil索引报错", func(t *testing.T) {
		re := runtimeErr(t, `nil[0]`)
		if !strings.Contains(re.Message, "nil") {
			t.Errorf("消息应包含 'nil'\n实际: %s", re.Message)
		}
		if !strings.Contains(re.Message, "索引") {
			t.Errorf("消息应包含 '索引'\n实际: %s", re.Message)
		}
	})
}

// Test_ErrorDiagnostic_RT_Index_OutOfBoundsAssign 数组越界赋值报错
func Test_ErrorDiagnostic_RT_Index_OutOfBoundsAssign(t *testing.T) {
	t.Run("数组越界赋值报错", func(t *testing.T) {
		re := runtimeErr(t, "arr := [1, 2]\narr[100] = 5")
		if !strings.Contains(re.Message, "越界") {
			t.Errorf("消息应包含 '越界'\n实际: %s", re.Message)
		}
	})
}

// ========== 运行时错误 - 函数 ==========

// Test_ErrorDiagnostic_RT_Func_CallNonFunction 调用非函数错误
func Test_ErrorDiagnostic_RT_Func_CallNonFunction(t *testing.T) {
	t.Run("调用int作为函数", func(t *testing.T) {
		re := runtimeErr(t, "x := 42\nx()")
		if !strings.Contains(re.Message, "函数") {
			t.Errorf("消息应包含 '函数'\n实际: %s", re.Message)
		}
	})
	t.Run("调用nil作为函数", func(t *testing.T) {
		re := runtimeErr(t, "x := nil\nx()")
		if !strings.Contains(re.Message, "函数") {
			t.Errorf("消息应包含 '函数'\n实际: %s", re.Message)
		}
	})
}

// Test_ErrorDiagnostic_RT_Func_ArgCount 参数不匹配错误
func Test_ErrorDiagnostic_RT_Func_ArgCount(t *testing.T) {
	t.Run("参数过少", func(t *testing.T) {
		re := runtimeErr(t, "fn f(a) { return a }\nf()")
		if !strings.Contains(re.Message, "参数") {
			t.Errorf("消息应包含 '参数'\n实际: %s", re.Message)
		}
		// 消息应包含函数名和期望参数个数
		if !strings.Contains(re.Message, "f") {
			t.Errorf("消息应包含函数名 'f'\n实际: %s", re.Message)
		}
	})
	t.Run("参数过多", func(t *testing.T) {
		re := runtimeErr(t, "fn f(a) { return a }\nf(1, 2)")
		if !strings.Contains(re.Message, "参数") {
			t.Errorf("消息应包含 '参数'\n实际: %s", re.Message)
		}
	})
}

// Test_ErrorDiagnostic_RT_Func_StackOverflow 调用栈溢出包含深度信息
func Test_ErrorDiagnostic_RT_Func_StackOverflow(t *testing.T) {
	t.Run("递归栈溢出包含深度", func(t *testing.T) {
		engine := NewEngine(WithMaxCallDepth(10))
		re := runtimeErrWithEngine(t, "fn rec(n) {\n  return rec(n + 1)\n}\nrec(0)", engine)
		if !strings.Contains(re.Message, "栈溢出") {
			t.Errorf("消息应包含 '栈溢出'\n实际: %s", re.Message)
		}
		if !strings.Contains(re.Message, "10") {
			t.Errorf("消息应包含最大深度 '10'\n实际: %s", re.Message)
		}
	})
}

// ========== 运行时错误 - 绑定值 ==========

// Test_ErrorDiagnostic_RT_BindValue_NotFound 未绑定值错误包含值名
func Test_ErrorDiagnostic_RT_BindValue_NotFound(t *testing.T) {
	t.Run("未绑定值包含名称", func(t *testing.T) {
		ctx := NewContext()
		re := runtimeErrWithCtx(t, `x :=>int getBindValue("missingValue")`, ctx)
		if !strings.Contains(re.Message, "missingValue") {
			t.Errorf("消息应包含值名 'missingValue'\n实际: %s", re.Message)
		}
		if !strings.Contains(re.Message, "绑定") {
			t.Errorf("消息应包含 '绑定'\n实际: %s", re.Message)
		}
	})
}

// Test_ErrorDiagnostic_RT_BindValue_Typo 绑定值名称拼写错误的诊断
func Test_ErrorDiagnostic_RT_BindValue_Typo(t *testing.T) {
	t.Run("拼写错误的绑定值名", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindValue("correctName", 42)
		re := runtimeErrWithCtx(t, `x :=>int getBindValue("correctNam")`, ctx)
		if !strings.Contains(re.Message, "correctNam") {
			t.Errorf("消息应包含使用的值名\n实际: %s", re.Message)
		}
	})
}

// ========== 运行时错误 - 绑定函数 ==========

// Test_ErrorDiagnostic_RT_BindFunc_NotFound 未绑定函数错误包含函数名
func Test_ErrorDiagnostic_RT_BindFunc_NotFound(t *testing.T) {
	t.Run("未绑定函数包含名称", func(t *testing.T) {
		ctx := NewContext()
		re := runtimeErrWithCtx(t, "#fn missingFn() => int\nmissingFn()", ctx)
		if !strings.Contains(re.Message, "missingFn") {
			t.Errorf("消息应包含函数名 'missingFn'\n实际: %s", re.Message)
		}
		if !strings.Contains(re.Message, "绑定") {
			t.Errorf("消息应包含 '绑定'\n实际: %s", re.Message)
		}
	})
}

// ========== 运行时错误 - 类型转换 ==========

// Test_ErrorDiagnostic_RT_Convert_IntFail int()转换失败消息
func Test_ErrorDiagnostic_RT_Convert_IntFail(t *testing.T) {
	t.Run("int转换数组失败", func(t *testing.T) {
		re := runtimeErr(t, `int([])`)
		if !strings.Contains(re.Message, "类型转换错误") {
			t.Errorf("消息应包含 '类型转换错误'\n实际: %s", re.Message)
		}
		if !strings.Contains(re.Message, "整数") {
			t.Errorf("消息应包含目标类型 '整数'\n实际: %s", re.Message)
		}
		if !strings.Contains(re.Message, "array") {
			t.Errorf("消息应包含源类型 'array'\n实际: %s", re.Message)
		}
	})
}

// Test_ErrorDiagnostic_RT_Convert_FloatFail float()转换失败消息
func Test_ErrorDiagnostic_RT_Convert_FloatFail(t *testing.T) {
	t.Run("float转换Map失败", func(t *testing.T) {
		re := runtimeErr(t, `float({})`)
		if !strings.Contains(re.Message, "类型转换错误") {
			t.Errorf("消息应包含 '类型转换错误'\n实际: %s", re.Message)
		}
		if !strings.Contains(re.Message, "浮点") {
			t.Errorf("消息应包含目标类型 '浮点'\n实际: %s", re.Message)
		}
		if !strings.Contains(re.Message, "map") {
			t.Errorf("消息应包含源类型 'map'\n实际: %s", re.Message)
		}
	})
}

// ========== 运行时错误 - 堆栈追踪 ==========

// Test_ErrorDiagnostic_RT_StackTrace_SingleCall 单层函数调用的堆栈追踪
func Test_ErrorDiagnostic_RT_StackTrace_SingleCall(t *testing.T) {
	t.Run("单层调用堆栈至少2帧", func(t *testing.T) {
		re := runtimeErr(t, "fn f() { throw \"err\" }\nf()")
		// 应至少包含 f 和 main 两帧
		if len(re.StackTrace) < 2 {
			t.Errorf("堆栈追踪应至少2帧，得到 %d 帧: %v", len(re.StackTrace), re.StackTrace)
		}
	})
}

// Test_ErrorDiagnostic_RT_StackTrace_MultiLayer 多层嵌套调用的堆栈追踪
func Test_ErrorDiagnostic_RT_StackTrace_MultiLayer(t *testing.T) {
	t.Run("三层嵌套调用堆栈", func(t *testing.T) {
		code := "fn f() { throw \"deep\" }\n" +
			"fn g() { return f() }\n" +
			"fn h() { return g() }\n" +
			"h()"
		re := runtimeErr(t, code)
		// 应包含 f, g, h, main 四帧
		if len(re.StackTrace) < 4 {
			t.Errorf("堆栈追踪应至少4帧，得到 %d 帧: %v", len(re.StackTrace), re.StackTrace)
		}
		// 验证堆栈中包含函数名
		traceJoined := strings.Join(re.StackTrace, "\n")
		for _, name := range []string{"f", "g", "h", "main"} {
			if !strings.Contains(traceJoined, name) {
				t.Errorf("堆栈追踪应包含函数名 '%s'\n堆栈: %s", name, traceJoined)
			}
		}
	})
}

// Test_ErrorDiagnostic_RT_StackTrace_Recursion 递归调用中的堆栈追踪
func Test_ErrorDiagnostic_RT_StackTrace_Recursion(t *testing.T) {
	t.Run("递归调用堆栈包含多帧同名函数", func(t *testing.T) {
		code := "fn rec(n) {\n" +
			"  if n <= 0 { throw \"bottom\" }\n" +
			"  return rec(n - 1)\n" +
			"}\n" +
			"rec(3)"
		re := runtimeErr(t, code)
		// rec 被调用多次，堆栈中应出现多次 rec
		count := strings.Count(strings.Join(re.StackTrace, "\n"), "rec")
		if count < 2 {
			t.Errorf("递归堆栈应出现多次 'rec'，实际 %d 次: %v", count, re.StackTrace)
		}
	})
}

// Test_ErrorDiagnostic_RT_StackTrace_InLoop 循环中的堆栈追踪
func Test_ErrorDiagnostic_RT_StackTrace_InLoop(t *testing.T) {
	t.Run("循环中throw堆栈包含main帧", func(t *testing.T) {
		code := "for i := 0; i < 3; i = i + 1 {\n  throw \"loop\"\n}"
		re := runtimeErr(t, code)
		traceJoined := strings.Join(re.StackTrace, "\n")
		if !strings.Contains(traceJoined, "main") {
			t.Errorf("循环中throw堆栈应包含 'main'\n堆栈: %s", traceJoined)
		}
	})
}

// Test_ErrorDiagnostic_RT_StackTrace_Format 堆栈追踪格式正确
func Test_ErrorDiagnostic_RT_StackTrace_Format(t *testing.T) {
	t.Run("堆栈帧包含函数名和位置", func(t *testing.T) {
		re := runtimeErr(t, "fn myFunc() { throw \"fmt\" }\nmyFunc()")
		for i, frame := range re.StackTrace {
			if !strings.Contains(frame, "at") {
				t.Errorf("第%d帧缺少 'at': %s", i, frame)
			}
			// 每帧应包含函数名或ip信息
			if !strings.Contains(frame, "ip") && !strings.Contains(frame, "myFunc") && !strings.Contains(frame, "main") {
				t.Errorf("第%d帧缺少位置信息: %s", i, frame)
			}
		}
	})
}

// ========== throw 消息 ==========

// Test_ErrorDiagnostic_Throw_String throw字符串的错误消息
func Test_ErrorDiagnostic_Throw_String(t *testing.T) {
	t.Run("throw字符串字面量", func(t *testing.T) {
		re := runtimeErr(t, `throw "custom error"`)
		if !strings.Contains(re.Message, "custom error") {
			t.Errorf("消息应包含 throw 的字符串\n实际: %s", re.Message)
		}
	})
}

// Test_ErrorDiagnostic_Throw_Expression throw表达式的消息
func Test_ErrorDiagnostic_Throw_Expression(t *testing.T) {
	t.Run("throw数字", func(t *testing.T) {
		re := runtimeErr(t, `throw 42`)
		if !strings.Contains(re.Message, "42") {
			t.Errorf("消息应包含 throw 的数字\n实际: %s", re.Message)
		}
	})
	t.Run("throw表达式结果", func(t *testing.T) {
		re := runtimeErr(t, `throw "code: " + 123`)
		if !strings.Contains(re.Message, "code") {
			t.Errorf("消息应包含表达式结果\n实际: %s", re.Message)
		}
	})
}

// Test_ErrorDiagnostic_Throw_InFunction throw在函数中的堆栈追踪
func Test_ErrorDiagnostic_Throw_InFunction(t *testing.T) {
	t.Run("函数中throw包含函数帧", func(t *testing.T) {
		re := runtimeErr(t, "fn throwError() { throw \"fn err\" }\nthrowError()")
		traceJoined := strings.Join(re.StackTrace, "\n")
		if !strings.Contains(traceJoined, "throwError") {
			t.Errorf("堆栈应包含函数名 'throwError'\n堆栈: %s", traceJoined)
		}
	})
}

// Test_ErrorDiagnostic_Throw_InLoop throw在循环中的堆栈追踪
func Test_ErrorDiagnostic_Throw_InLoop(t *testing.T) {
	t.Run("循环中throw消息包含thrown值", func(t *testing.T) {
		code := "for i := 0; i < 5; i = i + 1 {\n  if i == 2 { throw \"got2\" }\n}"
		re := runtimeErr(t, code)
		if !strings.Contains(re.Message, "got2") {
			t.Errorf("消息应包含 thrown 值 'got2'\n实际: %s", re.Message)
		}
	})
}

// ========== 错误消息格式一致性 ==========

// Test_ErrorDiagnostic_Format_AllChinese 所有错误消息都有中文描述
func Test_ErrorDiagnostic_Format_AllChinese(t *testing.T) {
	t.Run("编译错误含中文", func(t *testing.T) {
		ce := compileErr(t, "undefined_var")
		// 至少包含一个常见中文关键词
		hasChinese := strings.Contains(ce.Message, "错误") ||
			strings.Contains(ce.Message, "问题") ||
			strings.Contains(ce.Message, "变量") ||
			strings.Contains(ce.Message, "语法")
		if !hasChinese {
			t.Errorf("编译错误消息应包含中文描述\n实际: %s", ce.Message)
		}
	})
	t.Run("运行时错误含中文", func(t *testing.T) {
		re := runtimeErr(t, `5 / 0`)
		hasChinese := strings.Contains(re.Message, "错误") ||
			strings.Contains(re.Message, "问题") ||
			strings.Contains(re.Message, "除零")
		if !hasChinese {
			t.Errorf("运行时错误消息应包含中文描述\n实际: %s", re.Message)
		}
	})
}

// Test_ErrorDiagnostic_Format_ConsistentStructure 错误消息结构一致
func Test_ErrorDiagnostic_Format_ConsistentStructure(t *testing.T) {
	t.Run("详细运行时错误包含问题前缀", func(t *testing.T) {
		re := runtimeErr(t, `5 / 0`)
		if !strings.Contains(re.Message, "→") {
			t.Errorf("详细错误消息应使用 '→' 结构化前缀\n实际: %s", re.Message)
		}
		if !strings.Contains(re.Message, ErrorPrefixProblem) {
			t.Errorf("错误消息应包含 '→ 问题：' 前缀\n实际: %s", re.Message)
		}
	})
	t.Run("绑定值错误包含建议前缀", func(t *testing.T) {
		ctx := NewContext()
		re := runtimeErrWithCtx(t, `x :=>int getBindValue("nv")`, ctx)
		if !strings.Contains(re.Message, "→") {
			t.Errorf("绑定值错误应使用 '→' 结构化前缀\n实际: %s", re.Message)
		}
	})
}

// Test_ErrorDiagnostic_Format_ReasonableLength 错误消息长度合理
func Test_ErrorDiagnostic_Format_ReasonableLength(t *testing.T) {
	t.Run("编译错误消息非空", func(t *testing.T) {
		ce := compileErr(t, "undefined_var")
		if len(ce.Message) < 10 {
			t.Errorf("编译错误消息过短(应至少10字符): %q", ce.Message)
		}
	})
	t.Run("运行时错误消息非空", func(t *testing.T) {
		re := runtimeErr(t, `nil + 5`)
		if len(re.Message) < 10 {
			t.Errorf("运行时错误消息过短(应至少10字符): %q", re.Message)
		}
	})
}

// ========== 错误恢复 ==========

// Test_ErrorDiagnostic_Recovery_CompileAfterError 编译错误后可以编译新脚本
func Test_ErrorDiagnostic_Recovery_CompileAfterError(t *testing.T) {
	t.Run("编译错误后新Parser可正常编译", func(t *testing.T) {
		// 第一次编译失败
		parser1 := NewParser()
		_, err := parser1.Compile("undefined_var")
		if err == nil {
			t.Fatal("期望第一次编译失败")
		}
		// 第二次编译成功(纯表达式才产生返回值)
		parser2 := NewParser()
		script, err := parser2.Compile("1 + 2")
		if err != nil {
			t.Fatalf("编译错误后应能正常编译新脚本: %v", err)
		}
		engine := NewEngine()
		ctx := NewContext()
		result, err := engine.Run(ctx, script)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		if result.Int() != 3 {
			t.Errorf("结果: 期望 3, 得到 %d", result.Int())
		}
	})
}

// Test_ErrorDiagnostic_Recovery_RuntimeAfterError 运行时错误后VM可重用
func Test_ErrorDiagnostic_Recovery_RuntimeAfterError(t *testing.T) {
	t.Run("运行时错误后引擎可执行新脚本", func(t *testing.T) {
		engine := NewEngine()
		ctx := NewContext()

		// 第一次执行产生运行时错误
		parser1 := NewParser()
		script1, _ := parser1.Compile(`5 / 0`)
		_, err := engine.Run(ctx, script1)
		if err == nil {
			t.Fatal("期望第一次执行产生运行时错误")
		}

		// 第二次执行正常脚本(纯表达式才产生返回值)
		parser2 := NewParser()
		script2, err := parser2.Compile("10 + 20")
		if err != nil {
			t.Fatalf("编译失败: %v", err)
		}
		result, err := engine.Run(ctx, script2)
		if err != nil {
			t.Fatalf("运行时错误后引擎应可重用: %v", err)
		}
		if result.Int() != 30 {
			t.Errorf("结果: 期望 30, 得到 %d", result.Int())
		}
	})
}

// Test_ErrorDiagnostic_Recovery_MultipleErrorsNotInterfering 多个错误不相互影响
func Test_ErrorDiagnostic_Recovery_MultipleErrorsNotInterfering(t *testing.T) {
	t.Run("连续两次编译错误互不影响", func(t *testing.T) {
		// 第一次编译错误 - 未定义变量
		ce1 := compileErr(t, "firstUndefined")
		if !strings.Contains(ce1.Message, "firstUndefined") {
			t.Errorf("第一次错误消息应包含 'firstUndefined'\n实际: %s", ce1.Message)
		}
		// 第二次编译错误 - 无效字符
		ce2 := compileErr(t, "@")
		if !strings.Contains(ce2.Message, "@") {
			t.Errorf("第二次错误消息应包含 '@'\n实际: %s", ce2.Message)
		}
		// 确保第二次错误不包含第一次的变量名
		if strings.Contains(ce2.Message, "firstUndefined") {
			t.Errorf("第二次错误不应包含第一次的变量名\n实际: %s", ce2.Message)
		}
	})
}

// ========== 综合诊断质量 ==========

// Test_ErrorDiagnostic_Quality_ErrorFunc 检查Error()方法输出格式
func Test_ErrorDiagnostic_Quality_ErrorFunc(t *testing.T) {
	t.Run("CompileError.Error()包含行号", func(t *testing.T) {
		ce := compileErr(t, "x := 1\nundefined_var")
		errStr := ce.Error()
		if !strings.Contains(errStr, "行") {
			t.Errorf("Error()输出应包含行号\n实际: %s", errStr)
		}
		if !strings.Contains(errStr, "编译错误") {
			t.Errorf("Error()输出应包含 '编译错误'\n实际: %s", errStr)
		}
	})
	t.Run("RuntimeError.Error()包含调用栈", func(t *testing.T) {
		re := runtimeErr(t, "fn f() { throw \"test\" }\nf()")
		errStr := re.Error()
		if !strings.Contains(errStr, "运行时错误") {
			t.Errorf("Error()输出应包含 '运行时错误'\n实际: %s", errStr)
		}
		if !strings.Contains(errStr, "调用栈") {
			t.Errorf("Error()输出应包含 '调用栈'\n实际: %s", errStr)
		}
	})
}

// Test_ErrorDiagnostic_Quality_StackTraceNotEmpty 运行时错误堆栈追踪非空
func Test_ErrorDiagnostic_Quality_StackTraceNotEmpty(t *testing.T) {
	t.Run("简单运行时错误有堆栈追踪", func(t *testing.T) {
		re := runtimeErr(t, `5 / 0`)
		if len(re.StackTrace) == 0 {
			t.Error("运行时错误应有非空调用栈")
		}
		if len(re.StackTrace) > 0 && !strings.Contains(re.StackTrace[0], "main") {
			t.Errorf("顶层错误堆栈首帧应包含 'main'\n堆栈: %v", re.StackTrace)
		}
	})
}

// Test_ErrorDiagnostic_Quality_MessagePropagation 错误消息在调用链中传播
func Test_ErrorDiagnostic_Quality_MessagePropagation(t *testing.T) {
	t.Run("多层调用后throw消息保留原值", func(t *testing.T) {
		code := "fn f() { throw \"propagated\" }\n" +
			"fn g() { return f() }\n" +
			"fn h() { return g() }\n" +
			"h()"
		re := runtimeErr(t, code)
		if !strings.Contains(re.Message, "propagated") {
			t.Errorf("throw消息应传播到顶层\n实际: %s", re.Message)
		}
	})
}
