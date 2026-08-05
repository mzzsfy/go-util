package script

import (
	"testing"
)

// ========== Parser表达式与语句深度测试 ==========
// 本文件深度测试解析器的表达式和语句解析能力
// 使用 Validate 进行纯语法解析测试(不触发语义检查)
// 使用 compileScript/runIntTest/runBoolTest 进行编译与执行测试

// validateOK 断言源码可以通过语法解析(Validate)
func validateOK(t *testing.T, input string) {
	t.Helper()
	parser := NewParser()
	if err := parser.Validate(input); err != nil {
		t.Errorf("解析失败 [%s]: %v", input, err)
	}
}

// validateErr 断言源码在语法解析阶段(Validate)应该报错
func validateErr(t *testing.T, input string) {
	t.Helper()
	parser := NewParser()
	if err := parser.Validate(input); err == nil {
		t.Errorf("期望解析错误, 但成功通过: [%s]", input)
	}
}

// ========== 字面量表达式解析 ==========

func Test_ParserDeep_LiteralInt(t *testing.T) {
	t.Run("十进制整数", func(t *testing.T) {
		runIntTest(t, "42", 42)
	})
	t.Run("零", func(t *testing.T) {
		runIntTest(t, "0", 0)
	})
	t.Run("十六进制", func(t *testing.T) {
		runIntTest(t, "0xFF", 255)
	})
	t.Run("大整数", func(t *testing.T) {
		runIntTest(t, "1000000", 1000000)
	})
}

func Test_ParserDeep_LiteralFloat(t *testing.T) {
	t.Run("浮点数", func(t *testing.T) {
		runFloatTest(t, "3.14", 3.14)
	})
	t.Run("浮点运算", func(t *testing.T) {
		runFloatTest(t, "1.5 + 2.5", 4.0)
	})
}

func Test_ParserDeep_LiteralString(t *testing.T) {
	t.Run("字符串", func(t *testing.T) {
		runStringTest(t, `"hello"`, "hello")
	})
	t.Run("空字符串", func(t *testing.T) {
		runStringTest(t, `""`, "")
	})
}

func Test_ParserDeep_LiteralBool(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		runBoolTest(t, "true", true)
	})
	t.Run("false", func(t *testing.T) {
		runBoolTest(t, "false", false)
	})
}

func Test_ParserDeep_LiteralNil(t *testing.T) {
	t.Run("nil字面量", func(t *testing.T) {
		validateOK(t, "nil")
	})
	t.Run("nil赋值", func(t *testing.T) {
		validateOK(t, "x := nil")
	})
}

// ========== 标识符表达式解析 ==========

func Test_ParserDeep_Ident(t *testing.T) {
	t.Run("简单标识符声明", func(t *testing.T) {
		validateOK(t, "x := 1")
	})
	t.Run("标识符引用", func(t *testing.T) {
		validateOK(t, "x := 1\ny := x")
	})
	t.Run("下划线标识符", func(t *testing.T) {
		validateOK(t, "_x := 1")
	})
}

// ========== 数组与Map字面量解析 ==========

func Test_ParserDeep_ArrayLiteral(t *testing.T) {
	t.Run("空数组", func(t *testing.T) {
		validateOK(t, "[]")
	})
	t.Run("单元素数组", func(t *testing.T) {
		validateOK(t, "[1]")
	})
	t.Run("多元素数组", func(t *testing.T) {
		validateOK(t, "[1, 2, 3]")
	})
	t.Run("数组中表达式元素", func(t *testing.T) {
		validateOK(t, "[1 + 2, 3 * 4]")
	})
	t.Run("嵌套数组", func(t *testing.T) {
		validateOK(t, "[[1, 2], [3, 4]]")
	})
}

func Test_ParserDeep_MapLiteral(t *testing.T) {
	t.Run("空Map", func(t *testing.T) {
		validateOK(t, "{}")
	})
	t.Run("单元素Map", func(t *testing.T) {
		validateOK(t, `{"a": 1}`)
	})
	t.Run("多元素Map", func(t *testing.T) {
		validateOK(t, `{"a": 1, "b": 2}`)
	})
	t.Run("Map表达式值", func(t *testing.T) {
		validateOK(t, `{"a": 1 + 2}`)
	})
	t.Run("嵌套Map", func(t *testing.T) {
		validateOK(t, `{"outer": {"inner": 1}}`)
	})
}

func Test_ParserDeep_MixedLiteral(t *testing.T) {
	t.Run("数组和Map混合", func(t *testing.T) {
		validateOK(t, `[{"a": 1}, [1, 2]]`)
	})
	t.Run("Map中嵌套数组", func(t *testing.T) {
		validateOK(t, `{"data": [1, 2, 3]}`)
	})
	t.Run("数组中嵌套Map", func(t *testing.T) {
		validateOK(t, `[{"k": "v"}, {"k2": "v2"}]`)
	})
	t.Run("深层混合嵌套", func(t *testing.T) {
		validateOK(t, `[{"a": [1, {"b": 2}]}]`)
	})
}

// ========== 一元运算解析 ==========

func Test_ParserDeep_UnaryNeg(t *testing.T) {
	t.Run("负数", func(t *testing.T) {
		runIntTest(t, "-5", -5)
	})
	t.Run("负数乘法", func(t *testing.T) {
		runIntTest(t, "-2 * 3", -6)
	})
	t.Run("负括号", func(t *testing.T) {
		runIntTest(t, "-(2 + 3)", -5)
	})
	t.Run("双重负号", func(t *testing.T) {
		runIntTest(t, "-(-5)", 5)
	})
}

func Test_ParserDeep_UnaryNot(t *testing.T) {
	t.Run("非true", func(t *testing.T) {
		runBoolTest(t, "!true", false)
	})
	t.Run("非false", func(t *testing.T) {
		runBoolTest(t, "!false", true)
	})
	t.Run("非比较", func(t *testing.T) {
		runBoolTest(t, "!(1 > 2)", true)
	})
	t.Run("非与非运算混合", func(t *testing.T) {
		runBoolTest(t, "!true && false", false)
	})
}

func Test_ParserDeep_UnaryBitXor(t *testing.T) {
	t.Run("异或前缀", func(t *testing.T) {
		validateOK(t, "^5")
	})
}

// ========== 二元算术运算解析 ==========

func Test_ParserDeep_BinaryArith(t *testing.T) {
	t.Run("加法", func(t *testing.T) {
		runIntTest(t, "1 + 2", 3)
	})
	t.Run("减法", func(t *testing.T) {
		runIntTest(t, "10 - 3", 7)
	})
	t.Run("乘法", func(t *testing.T) {
		runIntTest(t, "4 * 5", 20)
	})
	t.Run("除法", func(t *testing.T) {
		runIntTest(t, "20 / 4", 5)
	})
	t.Run("取模", func(t *testing.T) {
		runIntTest(t, "10 % 3", 1)
	})
}

func Test_ParserDeep_BinaryPrecedence(t *testing.T) {
	t.Run("乘法优先于加法", func(t *testing.T) {
		runIntTest(t, "1 + 2 * 3", 7)
	})
	t.Run("括号改变优先级", func(t *testing.T) {
		runIntTest(t, "(1 + 2) * 3", 9)
	})
	t.Run("混合运算", func(t *testing.T) {
		runIntTest(t, "2 + 3 * 4 - 5", 9)
	})
}

func Test_ParserDeep_ChainedArith(t *testing.T) {
	t.Run("链式加法", func(t *testing.T) {
		runIntTest(t, "1 + 2 + 3 + 4", 10)
	})
	t.Run("链式减法左结合", func(t *testing.T) {
		runIntTest(t, "10 - 3 - 2", 5)
	})
	t.Run("链式乘除", func(t *testing.T) {
		runIntTest(t, "2 * 3 * 4 / 2", 12)
	})
	t.Run("长链运算", func(t *testing.T) {
		runIntTest(t, "1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 9", 45)
	})
}

func Test_ParserDeep_GroupedExpr(t *testing.T) {
	t.Run("简单括号", func(t *testing.T) {
		runIntTest(t, "(1 + 2)", 3)
	})
	t.Run("括号乘法", func(t *testing.T) {
		runIntTest(t, "(2 + 3) * (4 + 5)", 45)
	})
	t.Run("深层嵌套括号", func(t *testing.T) {
		runIntTest(t, "((((1 + 2))))", 3)
	})
	t.Run("多层混合括号", func(t *testing.T) {
		runIntTest(t, "((1 + 2) * (3 + 4) - 5) / 2", 8)
	})
}

// ========== 比较运算解析 ==========

func Test_ParserDeep_Comparison(t *testing.T) {
	t.Run("小于", func(t *testing.T) {
		runBoolTest(t, "1 < 2", true)
	})
	t.Run("小于等于", func(t *testing.T) {
		runBoolTest(t, "2 <= 2", true)
	})
	t.Run("大于", func(t *testing.T) {
		runBoolTest(t, "3 > 2", true)
	})
	t.Run("大于等于", func(t *testing.T) {
		runBoolTest(t, "2 >= 3", false)
	})
}

func Test_ParserDeep_Equality(t *testing.T) {
	t.Run("相等", func(t *testing.T) {
		runBoolTest(t, "1 == 1", true)
	})
	t.Run("不等", func(t *testing.T) {
		runBoolTest(t, "1 != 2", true)
	})
	t.Run("算术后比较", func(t *testing.T) {
		runBoolTest(t, "1 + 1 == 2", true)
	})
}

// ========== 逻辑运算解析 ==========

func Test_ParserDeep_Logical(t *testing.T) {
	t.Run("逻辑与", func(t *testing.T) {
		runBoolTest(t, "true && false", false)
	})
	t.Run("逻辑或", func(t *testing.T) {
		runBoolTest(t, "true || false", true)
	})
	t.Run("非与混合", func(t *testing.T) {
		runBoolTest(t, "!true && false", false)
	})
	t.Run("非或混合", func(t *testing.T) {
		runBoolTest(t, "!false || false", true)
	})
	t.Run("逻辑优先级", func(t *testing.T) {
		// AND 优先于 OR
		runBoolTest(t, "true || false && false", true)
	})
	t.Run("比较与逻辑混合", func(t *testing.T) {
		runBoolTest(t, "1 > 0 && 2 > 1", true)
	})
	t.Run("复杂逻辑组合", func(t *testing.T) {
		runBoolTest(t, "(1 < 2) && (3 > 4) || !(5 == 6)", true)
	})
}

// ========== 位运算解析 ==========

func Test_ParserDeep_Bitwise(t *testing.T) {
	t.Run("位与", func(t *testing.T) {
		runIntTest(t, "6 & 3", 2)
	})
	t.Run("位或", func(t *testing.T) {
		runIntTest(t, "6 | 3", 7)
	})
	t.Run("位异或", func(t *testing.T) {
		runIntTest(t, "6 ^ 3", 5)
	})
	t.Run("左移", func(t *testing.T) {
		runIntTest(t, "1 << 4", 16)
	})
	t.Run("右移", func(t *testing.T) {
		runIntTest(t, "256 >> 2", 64)
	})
	t.Run("位运算优先级", func(t *testing.T) {
		// & 优先于 |
		runIntTest(t, "1 | 2 & 3", 3)
	})
	t.Run("移位与算术混合", func(t *testing.T) {
		// + 优先于 <<
		runIntTest(t, "1 << 2 + 1", 8)
	})
}

// ========== 索引与切片解析 ==========

func Test_ParserDeep_Index(t *testing.T) {
	t.Run("数组索引", func(t *testing.T) {
		runIntTest(t, "[10, 20, 30][1]", 20)
	})
	t.Run("Map索引", func(t *testing.T) {
		runIntTest(t, `{"a": 1, "b": 2}["a"]`, 1)
	})
}

func Test_ParserDeep_NestedIndex(t *testing.T) {
	t.Run("嵌套数组索引", func(t *testing.T) {
		runIntTest(t, "[[1, 2], [3, 4]][0][1]", 2)
	})
	t.Run("嵌套Map索引", func(t *testing.T) {
		runIntTest(t, `{"a": {"b": 42}}["a"]["b"]`, 42)
	})
	t.Run("混合索引数组到Map", func(t *testing.T) {
		runIntTest(t, `[{"k": 99}][0]["k"]`, 99)
	})
	t.Run("混合索引Map到数组", func(t *testing.T) {
		runIntTest(t, `{"arr": [7, 8]}["arr"][1]`, 8)
	})
}

func Test_ParserDeep_Slice(t *testing.T) {
	t.Run("数组切片", func(t *testing.T) {
		validateOK(t, "arr := [1,2,3,4,5]\narr[1:3]")
	})
	t.Run("字符串切片", func(t *testing.T) {
		validateOK(t, `"hello"[0:3]`)
	})
	t.Run("省略起始切片", func(t *testing.T) {
		validateOK(t, "arr := [1,2,3]\narr[:2]")
	})
	t.Run("省略结束切片", func(t *testing.T) {
		validateOK(t, "arr := [1,2,3]\narr[1:]")
	})
	t.Run("链式切片后索引", func(t *testing.T) {
		validateOK(t, "arr := [1,2,3,4,5]\narr[0:3][1]")
	})
}

// ========== 函数调用解析 ==========

func Test_ParserDeep_FunctionCall(t *testing.T) {
	t.Run("无参数调用", func(t *testing.T) {
		runIntTest(t, "fn f() { return 1 }\nf()", 1)
	})
	t.Run("单参数调用", func(t *testing.T) {
		runIntTest(t, "fn f(x) { return x }\nf(42)", 42)
	})
	t.Run("多参数调用", func(t *testing.T) {
		runIntTest(t, "fn f(a, b, c) { return a + b + c }\nf(1, 2, 3)", 6)
	})
}

func Test_ParserDeep_NestedCall(t *testing.T) {
	t.Run("嵌套函数调用", func(t *testing.T) {
		runIntTest(t, "fn f(x) { return x }\nf(f(42))", 42)
	})
	t.Run("三层嵌套调用", func(t *testing.T) {
		// h(5)=10, g(10)=11, f(11)=11
		runIntTest(t, `
			fn f(x) { return x }
			fn g(x) { return x + 1 }
			fn h(x) { return x * 2 }
			f(g(h(5)))
		`, 11)
	})
	t.Run("表达式作为参数", func(t *testing.T) {
		runIntTest(t, "fn f(x) { return x }\nf(1 + 2 * 3)", 7)
	})
	t.Run("数组作为参数", func(t *testing.T) {
		validateOK(t, "fn f(x) { return x }\nf([1, 2, 3])")
	})
	t.Run("Map作为参数", func(t *testing.T) {
		validateOK(t, `fn f(x) { return x }`+"\n"+`f({"a": 1})`)
	})
}

// ========== 赋值解析 ==========

func Test_ParserDeep_AssignSimple(t *testing.T) {
	t.Run("简单赋值", func(t *testing.T) {
		runIntTest(t, "x := 1\nx", 1)
	})
	t.Run("表达式赋值", func(t *testing.T) {
		runIntTest(t, "x := 1 + 2 * 3\nx", 7)
	})
	t.Run("变量赋值", func(t *testing.T) {
		runIntTest(t, "x := 1\ny := x\ny", 1)
	})
}

func Test_ParserDeep_AssignCollection(t *testing.T) {
	t.Run("数组赋值", func(t *testing.T) {
		validateOK(t, "arr := [1, 2, 3]")
	})
	t.Run("Map赋值", func(t *testing.T) {
		validateOK(t, `m := {"a": 1}`)
	})
	t.Run("嵌套赋值", func(t *testing.T) {
		validateOK(t, `data := {"items": [1, 2]}`)
	})
}

func Test_ParserDeep_AssignIndex(t *testing.T) {
	t.Run("数组索引赋值", func(t *testing.T) {
		validateOK(t, "arr := [1,2,3]\narr[0] = 99")
	})
	t.Run("Map索引赋值", func(t *testing.T) {
		validateOK(t, `m := {"a": 1}`+"\n"+`m["key"] = "value"`)
	})
	t.Run("嵌套索引赋值", func(t *testing.T) {
		validateOK(t, "arr := [[1,2],[3,4]]\narr[0][1] = 99")
	})
}

func Test_ParserDeep_TypedAssign(t *testing.T) {
	t.Run("int类型声明", func(t *testing.T) {
		runIntTest(t, "x :=>int 10\nx", 10)
	})
	t.Run("float类型声明", func(t *testing.T) {
		runFloatTest(t, "x :=>float 3.14\nx", 3.14)
	})
	t.Run("string类型声明", func(t *testing.T) {
		runStringTest(t, `x :=>string "hello"`+"\nx", "hello")
	})
	t.Run("bool类型声明", func(t *testing.T) {
		runBoolTest(t, "x :=>bool true\nx", true)
	})
	t.Run("类型声明与表达式", func(t *testing.T) {
		runIntTest(t, "x :=>int 10 + 20\nx", 30)
	})
	t.Run("混合声明", func(t *testing.T) {
		runIntTest(t, "x :=>int 10\ny := 20\nx + y", 30)
	})
}

func Test_ParserDeep_IfExpr(t *testing.T) {
	t.Run("if表达式基本", func(t *testing.T) {
		runIntTest(t, "x := if true { 1 } else { 2 }\nx", 1)
	})
	t.Run("if表达式条件为假", func(t *testing.T) {
		runIntTest(t, "x := if false { 1 } else { 2 }\nx", 2)
	})
	t.Run("if表达式带运算", func(t *testing.T) {
		runIntTest(t, "x := if 3 > 2 { 10 } else { 20 }\nx", 10)
	})
}

func Test_ParserDeep_CompoundAssign(t *testing.T) {
	// 复合赋值运算符由解析器和编译器均支持
	t.Run("加等于", func(t *testing.T) {
		validateOK(t, "x := 5\nx += 3")
	})
	t.Run("减等于", func(t *testing.T) {
		validateOK(t, "x := 5\nx -= 2")
	})
	t.Run("乘等于", func(t *testing.T) {
		validateOK(t, "x := 5\nx *= 3")
	})
	t.Run("除等于", func(t *testing.T) {
		validateOK(t, "x := 20\nx /= 4")
	})
}

func Test_ParserDeep_Associativity(t *testing.T) {
	t.Run("减法左结合", func(t *testing.T) {
		// (10-3)-2=5
		runIntTest(t, "10 - 3 - 2", 5)
	})
	t.Run("除法左结合", func(t *testing.T) {
		// (100/5)/2=10
		runIntTest(t, "100 / 5 / 2", 10)
	})
	t.Run("混合左右结合", func(t *testing.T) {
		// ((2+3)-1)*2/4=2
		runIntTest(t, "2 + 3 - 1 * 2 / 4", 5)
	})
}

// ========== 表达式与语句解析 ==========

func Test_ParserDeep_ExprStmt(t *testing.T) {
	t.Run("表达式语句", func(t *testing.T) {
		runIntTest(t, "1 + 2", 3)
	})
	t.Run("多条表达式语句", func(t *testing.T) {
		runIntTest(t, "1 + 2\n3 + 4", 7)
	})
}

func Test_ParserDeep_VarDeclStmt(t *testing.T) {
	t.Run("整数声明", func(t *testing.T) {
		runIntTest(t, "x := 42\nx", 42)
	})
	t.Run("布尔声明", func(t *testing.T) {
		runBoolTest(t, "x := true\nx", true)
	})
	t.Run("nil声明", func(t *testing.T) {
		validateOK(t, "x := nil")
	})
}

func Test_ParserDeep_IfStmt(t *testing.T) {
	t.Run("简单if", func(t *testing.T) {
		validateOK(t, "if true { 1 }")
	})
	t.Run("if-else", func(t *testing.T) {
		validateOK(t, "if true { 1 } else { 2 }")
	})
	t.Run("if-else if-else", func(t *testing.T) {
		validateOK(t, "x := 5\nif x > 10 { 1 } else if x > 0 { 2 } else { 3 }")
	})
	t.Run("多层else-if链", func(t *testing.T) {
		validateOK(t, "x := 5\nif x > 10 { 1 } else if x > 8 { 2 } else if x > 5 { 3 } else if x > 0 { 4 } else { 5 }")
	})
}

func Test_ParserDeep_ForStmt(t *testing.T) {
	t.Run("标准for循环", func(t *testing.T) {
		validateOK(t, "for i := 0; i < 10; i = i + 1 { 1 }")
	})
	t.Run("条件while循环", func(t *testing.T) {
		validateOK(t, "for true { 1 }")
	})
	t.Run("无限循环", func(t *testing.T) {
		validateOK(t, "for { break }")
	})
	t.Run("计数循环", func(t *testing.T) {
		validateOK(t, "for i := 3 { break }")
	})
}

func Test_ParserDeep_ForRange(t *testing.T) {
	t.Run("单变量range数组", func(t *testing.T) {
		validateOK(t, "for k := range [1,2,3] { 1 }")
	})
	t.Run("双变量range数组", func(t *testing.T) {
		validateOK(t, "for k, v := range [1,2,3] { 1 }")
	})
	t.Run("range Map", func(t *testing.T) {
		validateOK(t, `for k, v := range {"a":1} { 1 }`)
	})
	t.Run("隐式range", func(t *testing.T) {
		validateOK(t, "for k := [1,2,3] { 1 }")
	})
}

func Test_ParserDeep_FuncDecl(t *testing.T) {
	t.Run("无参函数", func(t *testing.T) {
		validateOK(t, "fn f() { return 1 }")
	})
	t.Run("单参函数", func(t *testing.T) {
		validateOK(t, "fn f(x) { return x }")
	})
	t.Run("多参函数", func(t *testing.T) {
		validateOK(t, "fn f(a, b, c) { return a }")
	})
	t.Run("空函数体", func(t *testing.T) {
		validateOK(t, "fn f() { }")
	})
}

func Test_ParserDeep_ReturnStmt(t *testing.T) {
	t.Run("return值", func(t *testing.T) {
		runIntTest(t, "fn f() { return 42 }\nf()", 42)
	})
	t.Run("return表达式", func(t *testing.T) {
		runIntTest(t, "fn f() { return 1 + 2 }\nf()", 3)
	})
	t.Run("return变量", func(t *testing.T) {
		runIntTest(t, "fn f() { x := 5\nreturn x }\nf()", 5)
	})
	t.Run("空return", func(t *testing.T) {
		validateOK(t, "fn f() { return }")
	})
}

func Test_ParserDeep_BreakContinue(t *testing.T) {
	t.Run("break语句", func(t *testing.T) {
		validateOK(t, "for true { break }")
	})
	t.Run("continue语句", func(t *testing.T) {
		validateOK(t, "for true { continue }")
	})
	t.Run("嵌套循环中的break", func(t *testing.T) {
		validateOK(t, "for i := 0; i < 3; i = i + 1 {\nfor j := 0; j < 3; j = j + 1 {\nbreak\n}\n}")
	})
}

func Test_ParserDeep_ThrowStmt(t *testing.T) {
	t.Run("throw字符串", func(t *testing.T) {
		validateOK(t, `throw "error"`)
	})
	t.Run("throw整数", func(t *testing.T) {
		validateOK(t, "throw 42")
	})
	t.Run("throw表达式", func(t *testing.T) {
		validateOK(t, "throw 1 + 2")
	})
}

// ========== 控制流嵌套 ==========

func Test_ParserDeep_NestedControlFlow(t *testing.T) {
	t.Run("if嵌套if", func(t *testing.T) {
		validateOK(t, "x := 5\nif x > 0 {\nif x > 3 { 1 }\n}")
	})
	t.Run("for嵌套for", func(t *testing.T) {
		validateOK(t, "for i := 0; i < 3; i = i + 1 {\nfor j := 0; j < 3; j = j + 1 {\ni + j\n}\n}")
	})
	t.Run("for嵌套if", func(t *testing.T) {
		validateOK(t, "for i := 0; i < 5; i = i + 1 {\nif i > 2 { i }\n}")
	})
	t.Run("if嵌套for", func(t *testing.T) {
		validateOK(t, "x := 1\nif x > 0 {\nfor i := 0; i < 3; i = i + 1 {\ni\n}\n}")
	})
	t.Run("函数定义嵌套在循环中", func(t *testing.T) {
		validateOK(t, "for i := 0; i < 3; i = i + 1 {\nfn g(x) { return x }\ng(i)\n}")
	})
	t.Run("多层调用嵌套", func(t *testing.T) {
		// c(5)=10, b(10)=11, a(11)=11
		runIntTest(t, `
			fn a(x) { return x }
			fn b(x) { return x + 1 }
			fn c(x) { return x * 2 }
			a(b(c(5)))
		`, 11)
	})
	t.Run("深层if-else嵌套", func(t *testing.T) {
		validateOK(t, "x := 3\nif x > 10 { 1 } else { if x > 5 { 2 } else { if x > 0 { 3 } else { 4 } } }")
	})
}

// ========== 解析错误 ==========

func Test_ParserDeep_ErrUnclosed(t *testing.T) {
	t.Run("未闭合数组", func(t *testing.T) {
		validateErr(t, "[1, 2")
	})
	t.Run("未闭合Map", func(t *testing.T) {
		validateErr(t, `{"a": 1`)
	})
	t.Run("未闭合括号", func(t *testing.T) {
		validateErr(t, "(1 + 2")
	})
	t.Run("索引缺少右括号", func(t *testing.T) {
		validateErr(t, "arr[0")
	})
	t.Run("函数参数缺少右括号", func(t *testing.T) {
		validateErr(t, "fn f( { return 1 }")
	})
}

func Test_ParserDeep_ErrInvalidSyntax(t *testing.T) {
	t.Run("缺少条件", func(t *testing.T) {
		validateErr(t, "if { 1 }")
	})
	t.Run("无效表达式双运算符", func(t *testing.T) {
		validateErr(t, "1 + + 2")
	})
	t.Run("缺少右操作数", func(t *testing.T) {
		validateErr(t, "1 *")
	})
	t.Run("非法token序列", func(t *testing.T) {
		validateErr(t, "if else")
	})
	t.Run("参数列表双逗号", func(t *testing.T) {
		validateErr(t, "f(1, , 2)")
	})
	t.Run("else无前置if", func(t *testing.T) {
		validateErr(t, "else")
	})
	t.Run("类型声明缺少类型", func(t *testing.T) {
		validateErr(t, "x :=> 10")
	})
	t.Run("throw缺少值", func(t *testing.T) {
		validateErr(t, "throw")
	})
}

func Test_ParserDeep_ErrInvalidAssign(t *testing.T) {
	t.Run("对字面量赋值", func(t *testing.T) {
		runErrorTest(t, "1 + 2 = 3")
	})
}

// ========== 复杂程序解析 ==========

func Test_ParserDeep_ComplexProgram(t *testing.T) {
	t.Run("完整脚本", func(t *testing.T) {
		validateOK(t, `
			x := 10
			fn factorial(n) {
				if n <= 1 { return 1 }
				return n * factorial(n - 1)
			}
			result := factorial(x)
			if result > 0 { 1 } else { 0 }
		`)
	})
	t.Run("多个函数定义", func(t *testing.T) {
		validateOK(t, `
			fn add(a, b) { return a + b }
			fn sub(a, b) { return a - b }
			fn mul(a, b) { return a * b }
			add(1, 2)
			sub(3, 1)
			mul(2, 4)
		`)
	})
	t.Run("深层嵌套结构", func(t *testing.T) {
		validateOK(t, `
			data := {
				"list": [1, 2, {
					"nested": [3, 4]
				}]
			}
			for k, v := range data {
				k
			}
		`)
	})
	t.Run("复杂表达式组合", func(t *testing.T) {
		// 10 + 3 * 2 - 3 % 2 + 10 / 3 = 10+6-1+3 = 18
		runIntTest(t, `
			x := 10
			y := 3
			x + y * 2 - y % 2 + x / y
		`, 18)
	})
	t.Run("循环中调用函数", func(t *testing.T) {
		runIntTest(t, `
			fn double(x) { return x * 2 }
			sum := 0
			for i := 0; i < 5; i = i + 1 {
				sum = sum + double(i)
			}
			sum
		`, 20)
	})
}
