package script

import (
	"testing"
)

// ========== 条件表达式和分支逻辑完整测试 ==========
// 本文件测试 if/if-else/if-else if-else 链的所有细节
// 引擎行为: 非 bool 条件通过 truthiness 判断(nil/0/0.0/""为假, 其余为真)

// ========== 基本条件 ==========

// Test_ConditionalLogic_BasicIf_True if true 条件成立时执行then块
func Test_ConditionalLogic_BasicIf_True(t *testing.T) {
	t.Run("if true取then", func(t *testing.T) {
			runIntTest(t, `if true { 1 }`, 1)
	})
	t.Run("if true带表达式", func(t *testing.T) {
			runIntTest(t, `if true { 10 + 20 }`, 30)
	})
}

// Test_ConditionalLogic_BasicIf_False if false 条件不成立时不执行then块
func Test_ConditionalLogic_BasicIf_False(t *testing.T) {
	t.Run("if false不修改变量", func(t *testing.T) {
			runIntTest(t, `x := 0
if false { x = 1 }
x`, 0)
	})
	t.Run("if false无else保持默认", func(t *testing.T) {
			runIntTest(t, `x := 99
if false { x = 1 }
x`, 99)
	})
}

// Test_ConditionalLogic_BasicIf_ComparisonTrue 比较条件为真
func Test_ConditionalLogic_BasicIf_ComparisonTrue(t *testing.T) {
	t.Run("1大于0", func(t *testing.T) {
			runIntTest(t, `if 1 > 0 { 1 }`, 1)
	})
	t.Run("5大于等于5", func(t *testing.T) {
			runIntTest(t, `if 5 >= 5 { 1 }`, 1)
	})
}

// Test_ConditionalLogic_BasicIf_ComparisonFalse 比较条件为假
func Test_ConditionalLogic_BasicIf_ComparisonFalse(t *testing.T) {
	t.Run("1小于0不执行", func(t *testing.T) {
			runIntTest(t, `x := 0
if 1 < 0 { x = 1 }
x`, 0)
	})
	t.Run("3等于4不执行", func(t *testing.T) {
			runIntTest(t, `x := 0
if 3 == 4 { x = 1 }
x`, 0)
	})
}

// Test_ConditionalLogic_BasicIf_MultipleStatements if块内多条语句
func Test_ConditionalLogic_BasicIf_MultipleStatements(t *testing.T) {
	t.Run("then块多条语句", func(t *testing.T) {
			runIntTest(t, `x := 3
if x > 0 {
    a := x + 1
    b := a * 2
    b
}`, 8)
	})
	t.Run("then块累加变量", func(t *testing.T) {
			runIntTest(t, `x := 10
if x > 5 {
    x = x + 1
    x = x + 2
    x
}`, 13)
	})
}

// ========== if-else ==========

// Test_ConditionalLogic_IfElse_TrueBranch if true时取then分支
func Test_ConditionalLogic_IfElse_TrueBranch(t *testing.T) {
	t.Run("条件为真取then", func(t *testing.T) {
			runIntTest(t, `if true { 1 } else { 2 }`, 1)
	})
	t.Run("比较为真取then", func(t *testing.T) {
			runIntTest(t, `if 5 > 3 { 10 } else { 20 }`, 10)
	})
}

// Test_ConditionalLogic_IfElse_FalseBranch if false时取else分支
func Test_ConditionalLogic_IfElse_FalseBranch(t *testing.T) {
	t.Run("条件为假取else", func(t *testing.T) {
			runIntTest(t, `if false { 1 } else { 2 }`, 2)
	})
	t.Run("比较为假取else", func(t *testing.T) {
			runIntTest(t, `if 3 > 5 { 10 } else { 20 }`, 20)
	})
}

// Test_ConditionalLogic_IfElse_MultiStatements if-else中都有多条语句
func Test_ConditionalLogic_IfElse_MultiStatements(t *testing.T) {
	t.Run("then块多语句", func(t *testing.T) {
			runIntTest(t, `if true {
    a := 1
    b := 2
    a + b
} else {
    0
}`, 3)
	})
	t.Run("else块多语句", func(t *testing.T) {
			runIntTest(t, `if false {
    0
} else {
    a := 10
    b := 20
    a + b
}`, 30)
	})
}

// Test_ConditionalLogic_IfElse_ExpressionInElse else块中的表达式
func Test_ConditionalLogic_IfElse_ExpressionInElse(t *testing.T) {
	t.Run("else块算术表达式", func(t *testing.T) {
			runIntTest(t, `if false { 0 } else { 3 * 4 + 1 }`, 13)
	})
	t.Run("else块函数调用", func(t *testing.T) {
			runIntTest(t, `fn double(n) {
    n * 2
}
if false { 0 } else { double(21) }`, 42)
	})
}

// ========== if-else if-else 链 ==========

// Test_ConditionalLogic_ElseIfChain_FirstMatch 第一个条件满足
func Test_ConditionalLogic_ElseIfChain_FirstMatch(t *testing.T) {
	runIntTest(t, `x := 5
if x > 0 { 1 } else if x > 10 { 2 } else { 3 }`, 1)
}

// Test_ConditionalLogic_ElseIfChain_SecondMatch 第二个条件满足
func Test_ConditionalLogic_ElseIfChain_SecondMatch(t *testing.T) {
	runIntTest(t, `x := 15
if x > 20 { 1 } else if x > 10 { 2 } else { 3 }`, 2)
}

// Test_ConditionalLogic_ElseIfChain_LastMatch 最后else条件满足
func Test_ConditionalLogic_ElseIfChain_LastMatch(t *testing.T) {
	runIntTest(t, `x := 100
if x < 0 { 1 } else if x < 50 { 2 } else { 3 }`, 3)
}

// Test_ConditionalLogic_ElseIfChain_NoneMatch 都不满足走else
func Test_ConditionalLogic_ElseIfChain_NoneMatch(t *testing.T) {
	runIntTest(t, `x := -1
if x > 10 { 1 } else if x > 5 { 2 } else if x > 0 { 3 } else { 4 }`, 4)
}

// Test_ConditionalLogic_ElseIfChain_LongChain 多个else if链
func Test_ConditionalLogic_ElseIfChain_LongChain(t *testing.T) {
	runIntTest(t, `x := 42
if x == 0 { 0 } else if x == 10 { 1 } else if x == 20 { 2 } else if x == 30 { 3 } else if x == 40 { 4 } else if x == 42 { 5 } else { 6 }`, 5)
}

// ========== 条件表达式类型 ==========

// Test_ConditionalLogic_Cond_ComparisonOps 所有比较运算符作为条件
func Test_ConditionalLogic_Cond_ComparisonOps(t *testing.T) {
	t.Run("大于", func(t *testing.T) {
	runIntTest(t, `if 5 > 3 { 1 } else { 0 }`, 1)
	})
	t.Run("小于", func(t *testing.T) {
	runIntTest(t, `if 3 < 5 { 1 } else { 0 }`, 1)
	})
	t.Run("等于", func(t *testing.T) {
	runIntTest(t, `if 3 == 3 { 1 } else { 0 }`, 1)
	})
	t.Run("不等", func(t *testing.T) {
	runIntTest(t, `if 3 != 4 { 1 } else { 0 }`, 1)
	})
	t.Run("大于等于", func(t *testing.T) {
	runIntTest(t, `if 5 >= 5 { 1 } else { 0 }`, 1)
	})
	t.Run("小于等于", func(t *testing.T) {
	runIntTest(t, `if 5 <= 5 { 1 } else { 0 }`, 1)
	})
}

// Test_ConditionalLogic_Cond_LogicalOps 逻辑运算符作为条件
func Test_ConditionalLogic_Cond_LogicalOps(t *testing.T) {
	t.Run("逻辑与都真", func(t *testing.T) {
	runIntTest(t, `if true && true { 1 } else { 0 }`, 1)
	})
	t.Run("逻辑与有假", func(t *testing.T) {
	runIntTest(t, `if true && false { 1 } else { 0 }`, 0)
	})
	t.Run("逻辑或有真", func(t *testing.T) {
	runIntTest(t, `if true || false { 1 } else { 0 }`, 1)
	})
	t.Run("逻辑或都假", func(t *testing.T) {
	runIntTest(t, `if false || false { 1 } else { 0 }`, 0)
	})
	t.Run("逻辑非真", func(t *testing.T) {
	runIntTest(t, `if !false { 1 } else { 0 }`, 1)
	})
	t.Run("逻辑非假", func(t *testing.T) {
	runIntTest(t, `if !true { 1 } else { 0 }`, 0)
	})
}

// Test_ConditionalLogic_Cond_CompoundLogic 复合逻辑条件
func Test_ConditionalLogic_Cond_CompoundLogic(t *testing.T) {
	t.Run("括号分组与运算", func(t *testing.T) {
	runIntTest(t, `if (3 > 1) && (5 < 10) { 1 } else { 0 }`, 1)
	})
	t.Run("括号分组或运算", func(t *testing.T) {
	runIntTest(t, `if (3 > 5) || (1 < 2) { 1 } else { 0 }`, 1)
	})
	t.Run("非运算与比较组合", func(t *testing.T) {
	runIntTest(t, `if !(3 > 4) { 1 } else { 0 }`, 1)
	})
}

// Test_ConditionalLogic_Cond_FunctionReturn 函数返回值作为条件
func Test_ConditionalLogic_Cond_FunctionReturn(t *testing.T) {
	result := runScript(t, `fn isPositive(n) {
    n > 0
}
x := 5
if isPositive(x) { 1 } else { 0 }`)
	if result.Int() != 1 {
		t.Errorf("期望 1, 得到 %d", result.Int())
	}
}

// Test_ConditionalLogic_Cond_VariableAsCondition 变量作为条件
func Test_ConditionalLogic_Cond_VariableAsCondition(t *testing.T) {
	t.Run("bool变量true", func(t *testing.T) {
	runIntTest(t, `b := true
if b { 1 } else { 0 }`, 1)
	})
	t.Run("bool变量false", func(t *testing.T) {
	runIntTest(t, `b := false
if b { 1 } else { 0 }`, 0)
	})
}

// ========== 非 bool 条件行为 ==========
// 引擎行为: Value通过IsTruthy判断, nil/0/0.0/""为假, 非0数值/非空字符串/array/map/function为真

// Test_ConditionalLogic_NonBool_Int 整数条件
func Test_ConditionalLogic_NonBool_Int(t *testing.T) {
	t.Run("非零整数条件为真", func(t *testing.T) {
	runIntTest(t, `if 1 { 1 } else { 2 }`, 1)
	})
}

// Test_ConditionalLogic_NonBool_Zero 零条件
func Test_ConditionalLogic_NonBool_Zero(t *testing.T) {
	t.Run("零整数条件为假", func(t *testing.T) {
	runIntTest(t, `if 0 { 1 } else { 2 }`, 2)
	})
}

// Test_ConditionalLogic_NonBool_String 字符串条件
func Test_ConditionalLogic_NonBool_String(t *testing.T) {
	t.Run("非空字符串条件为真", func(t *testing.T) {
	runIntTest(t, `if "hello" { 1 } else { 2 }`, 1)
	})
}

// Test_ConditionalLogic_NonBool_EmptyString 空字符串条件
func Test_ConditionalLogic_NonBool_EmptyString(t *testing.T) {
	t.Run("空字符串条件为假", func(t *testing.T) {
	runIntTest(t, `if "" { 1 } else { 2 }`, 2)
	})
}

// Test_ConditionalLogic_NonBool_Nil nil条件
func Test_ConditionalLogic_NonBool_Nil(t *testing.T) {
	t.Run("nil条件为假", func(t *testing.T) {
	runIntTest(t, `if nil { 1 } else { 2 }`, 2)
	})
}

// Test_ConditionalLogic_NonBool_Array 数组条件
func Test_ConditionalLogic_NonBool_Array(t *testing.T) {
	t.Run("非空数组条件为真", func(t *testing.T) {
	runIntTest(t, `if [1, 2] { 1 } else { 2 }`, 1)
	})
}

// Test_ConditionalLogic_NonBool_Map Map条件
func Test_ConditionalLogic_NonBool_Map(t *testing.T) {
	t.Run("非空Map条件为真", func(t *testing.T) {
	runIntTest(t, `if {"a": 1} { 1 } else { 2 }`, 1)
	})
	t.Run("空Map条件为真", func(t *testing.T) {
	runIntTest(t, `if {} { 1 } else { 2 }`, 1)
	})
}

// Test_ConditionalLogic_NonBool_NotOperator 非bool值的取反运算
func Test_ConditionalLogic_NonBool_NotOperator(t *testing.T) {
	// 取反基于truthiness: 真值取反为false, 假值取反为true
	t.Run("非零整数取反为false", func(t *testing.T) {
	runBoolTest(t, `!1`, false)
	})
	t.Run("零取反为true", func(t *testing.T) {
	runBoolTest(t, `!0`, true)
	})
	t.Run("nil取反为true", func(t *testing.T) {
	runBoolTest(t, `!nil`, true)
	})
	t.Run("非空字符串取反为false", func(t *testing.T) {
	runBoolTest(t, `!"hello"`, false)
	})
}

// ========== 嵌套 if ==========

// Test_ConditionalLogic_NestedIf_IfInIf if嵌套if
func Test_ConditionalLogic_NestedIf_IfInIf(t *testing.T) {
	runIntTest(t, `x := 3
result := 0
if x > 0 {
    if x > 1 { result = 1 } else { result = 2 }
} else {
    result = 3
}
result`, 1)
}

// Test_ConditionalLogic_NestedIf_IfElseInIf if嵌套if-else
func Test_ConditionalLogic_NestedIf_IfElseInIf(t *testing.T) {
	runIntTest(t, `x := 5
result := 0
if x > 0 {
    if x > 10 { result = 1 } else { result = 2 }
} else {
    result = 3
}
result`, 2)
}

// Test_ConditionalLogic_NestedIf_IfInIfElse if-else嵌套if
func Test_ConditionalLogic_NestedIf_IfInIfElse(t *testing.T) {
	runIntTest(t, `x := 0
result := 0
if x > 0 {
    result = 1
} else {
    if x == 0 { result = 2 } else { result = 3 }
}
result`, 2)
}

// Test_ConditionalLogic_NestedIf_IfElseInIfElse if-else嵌套if-else
func Test_ConditionalLogic_NestedIf_IfElseInIfElse(t *testing.T) {
	runIntTest(t, `x := 0
y := 10
result := 0
if x > 0 {
    if y > 5 { result = 1 } else { result = 2 }
} else {
    if y > 5 { result = 3 } else { result = 4 }
}
result`, 3)
}

// Test_ConditionalLogic_NestedIf_Deep3Levels 深层嵌套3层以上
func Test_ConditionalLogic_NestedIf_Deep3Levels(t *testing.T) {
	runIntTest(t, `x := 5
r := 0
if x > 0 {
    if x > 3 {
        if x > 7 { r = 1 } else { r = 2 }
    } else {
        r = 3
    }
} else {
    r = 4
}
r`, 2)
}

// ========== if 与表达式 ==========

// Test_ConditionalLogic_IfBlock_VariableDecl if块中定义变量
func Test_ConditionalLogic_IfBlock_VariableDecl(t *testing.T) {
	t.Run("if块中声明并使用变量", func(t *testing.T) {
	runIntTest(t, `x := 10
if x > 5 {
    y := x * 2
    z := y + 1
    z
} else {
    0
}`, 21)
	})
}

// Test_ConditionalLogic_IfBlock_ArrayOp if块中数组操作
func Test_ConditionalLogic_IfBlock_ArrayOp(t *testing.T) {
	t.Run("if块中push数组", func(t *testing.T) {
	runIntTest(t, `arr := [1, 2]
if true { arr = push(arr, 3) }
len(arr)`, 3)
	})
	t.Run("if块中修改数组元素", func(t *testing.T) {
	runIntTest(t, `arr := [1, 2, 3]
if arr[0] == 1 { arr[0] = 99 }
arr[0]`, 99)
	})
}

// Test_ConditionalLogic_IfBlock_MapOp if块中Map操作
func Test_ConditionalLogic_IfBlock_MapOp(t *testing.T) {
	t.Run("if块中读取Map值", func(t *testing.T) {
	runIntTest(t, `m := {"a": 1, "b": 2}
if m["a"] == 1 { m["b"] } else { 0 }`, 2)
	})
	t.Run("if块中修改Map值", func(t *testing.T) {
	runIntTest(t, `m := {"key": 10}
if true { m["key"] = 42 }
m["key"]`, 42)
	})
}

// Test_ConditionalLogic_IfBlock_FuncCall if块中函数调用
func Test_ConditionalLogic_IfBlock_FuncCall(t *testing.T) {
	t.Run("if块中调用用户函数", func(t *testing.T) {
	runIntTest(t, `fn add(a, b) {
    a + b
}
if true { add(3, 4) } else { 0 }`, 7)
	})
}

// Test_ConditionalLogic_IfBlock_Loop if块中循环
func Test_ConditionalLogic_IfBlock_Loop(t *testing.T) {
	t.Run("if块中for循环累加", func(t *testing.T) {
	runIntTest(t, `sum := 0
n := 5
if n > 0 {
    for i := 1; i <= n; i = i + 1 {
        sum = sum + i
    }
    sum
} else {
    0
}`, 15)
	})
	t.Run("if块中break跳出循环", func(t *testing.T) {
	runIntTest(t, `sum := 0
limit := 10
if limit > 0 {
    for i := 1; i <= limit; i = i + 1 {
        if i > 5 { break }
        sum = sum + i
    }
    sum
} else {
    0
}`, 15)
	})
}

// ========== if 与 return ==========

// Test_ConditionalLogic_Return_InIf if中return
func Test_ConditionalLogic_Return_InIf(t *testing.T) {
	result := runScript(t, `fn check(n) {
    if n > 0 { return 1 }
    return -1
}
check(5)`)
	if result.Int() != 1 {
		t.Errorf("期望 1, 得到 %d", result.Int())
	}
}

// Test_ConditionalLogic_Return_InElse else中return
func Test_ConditionalLogic_Return_InElse(t *testing.T) {
	result := runScript(t, `fn check(n) {
    if n > 0 {
        return 1
    } else {
        return -1
    }
}
check(-3)`)
	if result.Int() != -1 {
		t.Errorf("期望 -1, 得到 %d", result.Int())
	}
}

// Test_ConditionalLogic_Return_InElseIfChain if-else if链中return
func Test_ConditionalLogic_Return_InElseIfChain(t *testing.T) {
	result := runScript(t, `fn classify(n) {
    if n < 0 { return 1 }
    else if n == 0 { return 2 }
    else if n < 10 { return 3 }
    else { return 4 }
}
classify(5)`)
	if result.Int() != 3 {
		t.Errorf("期望 3, 得到 %d", result.Int())
	}
}

// Test_ConditionalLogic_Return_MultiplePaths 函数中多个return路径
func Test_ConditionalLogic_Return_MultiplePaths(t *testing.T) {
	result := runScript(t, `fn abs(n) {
    if n >= 0 { return n }
    return 0 - n
}
abs(0 - 5)`)
	if result.Int() != 5 {
		t.Errorf("期望 5, 得到 %d", result.Int())
	}
}

// Test_ConditionalLogic_Return_CodeAfterReturn 条件return后的代码
func Test_ConditionalLogic_Return_CodeAfterReturn(t *testing.T) {
	t.Run("条件return不满足返回默认值", func(t *testing.T) {
		result := runScript(t, `fn test(n) {
    if n > 0 { return 1 }
    return 0
}
test(-1)`)
		if result.Int() != 0 {
			t.Errorf("期望 0, 得到 %d", result.Int())
		}
	})
	t.Run("多个条件return都不满足", func(t *testing.T) {
		result := runScript(t, `fn test(n) {
    if n > 0 { return 1 }
    if n < 0 { return -1 }
    return 0
}
test(0)`)
		if result.Int() != 0 {
			t.Errorf("期望 0, 得到 %d", result.Int())
		}
	})
}

// ========== 条件中的复杂表达式 ==========

// Test_ConditionalLogic_ComplexExpr_Arithmetic 算术表达式条件
func Test_ConditionalLogic_ComplexExpr_Arithmetic(t *testing.T) {
	t.Run("加法比较条件", func(t *testing.T) {
	runIntTest(t, `if 1 + 2 > 2 { 1 } else { 0 }`, 1)
	})
	t.Run("乘法比较条件", func(t *testing.T) {
	runIntTest(t, `if 2 * 3 > 5 { 1 } else { 0 }`, 1)
	})
}

// Test_ConditionalLogic_ComplexExpr_Parens 带括号的复杂条件
func Test_ConditionalLogic_ComplexExpr_Parens(t *testing.T) {
	t.Run("括号改变优先级", func(t *testing.T) {
	runIntTest(t, `if (1 + 2) * 3 > 8 { 1 } else { 0 }`, 1)
	})
	t.Run("嵌套括号", func(t *testing.T) {
	runIntTest(t, `if ((1 + 2) * 3) > 8 { 1 } else { 0 }`, 1)
	})
}

// Test_ConditionalLogic_ComplexExpr_CompareChain 比较链
func Test_ConditionalLogic_ComplexExpr_CompareChain(t *testing.T) {
	t.Run("递减链", func(t *testing.T) {
	runIntTest(t, `if 5 > 3 && 3 > 2 && 2 > 1 { 1 } else { 0 }`, 1)
	})
	t.Run("链中有一个不满足", func(t *testing.T) {
	runIntTest(t, `if 5 > 3 && 3 > 4 && 2 > 1 { 1 } else { 0 }`, 0)
	})
}

// Test_ConditionalLogic_ComplexExpr_MixedOps 混合运算符条件
func Test_ConditionalLogic_ComplexExpr_MixedOps(t *testing.T) {
	t.Run("非运算与比较组合", func(t *testing.T) {
	runIntTest(t, `done := false
count := 3
max := 10
if !done && count < max { 1 } else { 0 }`, 1)
	})
	t.Run("或运算与非运算组合", func(t *testing.T) {
	runIntTest(t, `done := true
if done || !done { 1 } else { 0 }`, 1)
	})
}

// Test_ConditionalLogic_ComplexExpr_ArrayLen 数组长度条件
func Test_ConditionalLogic_ComplexExpr_ArrayLen(t *testing.T) {
	t.Run("数组长度大于零", func(t *testing.T) {
	runIntTest(t, `arr := [1, 2, 3]
if len(arr) > 0 { 1 } else { 0 }`, 1)
	})
	t.Run("空数组长度等于零", func(t *testing.T) {
	runIntTest(t, `arr := []
if len(arr) == 0 { 1 } else { 0 }`, 1)
	})
}

// ========== if 与索引访问 ==========

// Test_ConditionalLogic_IndexAccess_Array 条件中使用数组索引
func Test_ConditionalLogic_IndexAccess_Array(t *testing.T) {
	t.Run("数组索引比较条件", func(t *testing.T) {
	runIntTest(t, `arr := [10, 20, 30]
if arr[0] > 5 { arr[1] } else { arr[2] }`, 20)
	})
	t.Run("数组索引等于比较", func(t *testing.T) {
	runIntTest(t, `arr := [5, 10, 15]
if arr[1] == 10 { 1 } else { 0 }`, 1)
	})
}

// Test_ConditionalLogic_IndexAccess_Map 条件中使用Map访问
func Test_ConditionalLogic_IndexAccess_Map(t *testing.T) {
	t.Run("Map值等于比较", func(t *testing.T) {
	runIntTest(t, `m := {"key": 42}
if m["key"] == 42 { 1 } else { 0 }`, 1)
	})
	t.Run("Map值不等比较", func(t *testing.T) {
	runIntTest(t, `m := {"a": 1, "b": 2}
if m["a"] != m["b"] { 1 } else { 0 }`, 1)
	})
}

// Test_ConditionalLogic_IndexAccess_Nested 嵌套索引条件
func Test_ConditionalLogic_IndexAccess_Nested(t *testing.T) {
	t.Run("嵌套Map索引比较", func(t *testing.T) {
	runIntTest(t, `m := {"outer": {"inner": 42}}
if m["outer"]["inner"] > 0 { 1 } else { 0 }`, 1)
	})
	t.Run("数组中Map索引比较", func(t *testing.T) {
	runIntTest(t, `arr := [{"v": 10}, {"v": 20}]
if arr[1]["v"] == 20 { 1 } else { 0 }`, 1)
	})
}

// ========== if 中的赋值和副作用 ==========

// Test_ConditionalLogic_SideEffect_ConditionFunc if条件中函数修改状态
func Test_ConditionalLogic_SideEffect_ConditionFunc(t *testing.T) {
	called := 0
	ctx := NewContext()
	ctx.BindFunc("counter", func() int {
		called++
		return called
	})
	parser := NewParser()
	script, err := parser.Compile(`#fn counter()=>int
if counter() > 0 && counter() > 0 { 1 } else { 0 }`)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}
	engine := NewEngine()
	result, err := engine.Run(ctx, script)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}
	if called != 2 {
		t.Errorf("期望 counter 被调用 2 次, 实际 %d 次", called)
	}
	if result.Int() != 1 {
		t.Errorf("期望 1, 得到 %d", result.Int())
	}
}

// Test_ConditionalLogic_SideEffect_ModifyVar if块中修改变量后影响后续判断
func Test_ConditionalLogic_SideEffect_ModifyVar(t *testing.T) {
	runIntTest(t, `x := 5
y := 0
if x > 0 {
    y = x * 2
}
if y > 5 { 1 } else { 0 }`, 1)
}

// Test_ConditionalLogic_SideEffect_SequentialIfs 多个if按顺序检查
func Test_ConditionalLogic_SideEffect_SequentialIfs(t *testing.T) {
	runIntTest(t, `x := 5
result := 0
if x > 1 { result = result + 1 }
if x > 2 { result = result + 1 }
if x > 3 { result = result + 1 }
if x > 10 { result = result + 1 }
result`, 3)
}

// ========== switch 模式模拟 ==========

// Test_ConditionalLogic_SwitchPattern_ByValue 用if-else if模拟switch按值分发
func Test_ConditionalLogic_SwitchPattern_ByValue(t *testing.T) {
	t.Run("值分发匹配", func(t *testing.T) {
	runStringTest(t, `x := 2
result := ""
if x == 1 { result = "one" } else if x == 2 { result = "two" } else if x == 3 { result = "three" } else { result = "other" }
result`, "two")
	})
	t.Run("值分发default", func(t *testing.T) {
	runStringTest(t, `x := 99
result := ""
if x == 1 { result = "one" } else if x == 2 { result = "two" } else { result = "other" }
result`, "other")
	})
}

// Test_ConditionalLogic_SwitchPattern_Range 范围判断(分数等级)
func Test_ConditionalLogic_SwitchPattern_Range(t *testing.T) {
	t.Run("A级", func(t *testing.T) {
	runStringTest(t, `score := 95
grade := ""
if score >= 90 { grade = "A" } else if score >= 80 { grade = "B" } else if score >= 70 { grade = "C" } else if score >= 60 { grade = "D" } else { grade = "F" }
grade`, "A")
	})
	t.Run("C级", func(t *testing.T) {
	runStringTest(t, `score := 75
grade := ""
if score >= 90 { grade = "A" } else if score >= 80 { grade = "B" } else if score >= 70 { grade = "C" } else if score >= 60 { grade = "D" } else { grade = "F" }
grade`, "C")
	})
	t.Run("F级", func(t *testing.T) {
	runStringTest(t, `score := 55
grade := ""
if score >= 90 { grade = "A" } else if score >= 80 { grade = "B" } else if score >= 70 { grade = "C" } else if score >= 60 { grade = "D" } else { grade = "F" }
grade`, "F")
	})
}

// Test_ConditionalLogic_SwitchPattern_TypeCheck 类型判断模式
func Test_ConditionalLogic_SwitchPattern_TypeCheck(t *testing.T) {
	t.Run("typeof判断int", func(t *testing.T) {
	runIntTest(t, `x := 42
if typeof(x) == "int" { 1 } else { 0 }`, 1)
	})
	t.Run("typeof判断string", func(t *testing.T) {
	runIntTest(t, `x := "hello"
if typeof(x) == "string" { 1 } else { 0 }`, 1)
	})
	t.Run("typeof多分支", func(t *testing.T) {
	runStringTest(t, `x := true
result := ""
if typeof(x) == "int" { result = "number" } else if typeof(x) == "string" { result = "text" } else if typeof(x) == "bool" { result = "flag" } else { result = "unknown" }
result`, "flag")
	})
}

// ========== 短路求值 ==========

// Test_ConditionalLogic_ShortCircuit_AndFalse false && f() 不调用右侧
func Test_ConditionalLogic_ShortCircuit_AndFalse(t *testing.T) {
	// false && (1/0>0): 短路使右侧不被求值, 避免除零错误
	runBoolTest(t, `false && (1 / 0 > 0)`, false)
}

// Test_ConditionalLogic_ShortCircuit_OrTrue true || f() 不调用右侧
func Test_ConditionalLogic_ShortCircuit_OrTrue(t *testing.T) {
	// true || (1/0>0): 短路使右侧不被求值, 避免除零错误
	runBoolTest(t, `true || (1 / 0 > 0)`, true)
}

// Test_ConditionalLogic_ShortCircuit_SafeAccess 安全访问模式
func Test_ConditionalLogic_ShortCircuit_SafeAccess(t *testing.T) {
	t.Run("nil安全访问不报错", func(t *testing.T) {
	runBoolTest(t, `a := nil
a != nil && a[0] > 0`, false)
	})
	t.Run("非nil安全访问正常求值", func(t *testing.T) {
	runBoolTest(t, `a := [1, 2]
a != nil && a[0] > 0`, true)
	})
}

// Test_ConditionalLogic_ShortCircuit_AvoidError 短路避免运行时错误
func Test_ConditionalLogic_ShortCircuit_AvoidError(t *testing.T) {
	t.Run("左侧false避免右侧除零", func(t *testing.T) {
		// 如果不短路, 1/0会触发运行时错误
		result := runScript(t, `false && (1 / 0 > 0)`)
		if result.Bool() != false {
			t.Errorf("期望 false, 得到 %v", result.Bool())
		}
	})
	t.Run("左侧true避免右侧除零", func(t *testing.T) {
		result := runScript(t, `true || (1 / 0 > 0)`)
		if result.Bool() != true {
			t.Errorf("期望 true, 得到 %v", result.Bool())
		}
	})
}

// ========== 条件边界 ==========

// Test_ConditionalLogic_Boundary_Equal 相等比较边界
func Test_ConditionalLogic_Boundary_Equal(t *testing.T) {
	t.Run("相等时取then", func(t *testing.T) {
	runIntTest(t, `if 5 == 5 { 1 } else { 0 }`, 1)
	})
	t.Run("不等时取else", func(t *testing.T) {
	runIntTest(t, `if 5 == 6 { 1 } else { 0 }`, 0)
	})
}

// Test_ConditionalLogic_Boundary_Compare 大小比较边界
func Test_ConditionalLogic_Boundary_Compare(t *testing.T) {
	t.Run("正好等于边界取then(大于等于)", func(t *testing.T) {
	runIntTest(t, `if 10 >= 10 { 1 } else { 0 }`, 1)
	})
	t.Run("正好等于边界取then(小于等于)", func(t *testing.T) {
	runIntTest(t, `if 10 <= 10 { 1 } else { 0 }`, 1)
	})
	t.Run("正好等于边界取else(大于)", func(t *testing.T) {
	runIntTest(t, `if 10 > 10 { 1 } else { 0 }`, 0)
	})
	t.Run("正好等于边界取else(小于)", func(t *testing.T) {
	runIntTest(t, `if 10 < 10 { 1 } else { 0 }`, 0)
	})
}

// Test_ConditionalLogic_Boundary_Zero 0值判断
func Test_ConditionalLogic_Boundary_Zero(t *testing.T) {
	t.Run("0等于0", func(t *testing.T) {
	runIntTest(t, `if 0 == 0 { 1 } else { 0 }`, 1)
	})
	t.Run("0不等于1", func(t *testing.T) {
	runIntTest(t, `if 0 != 1 { 1 } else { 0 }`, 1)
	})
	t.Run("0小于1", func(t *testing.T) {
	runIntTest(t, `if 0 < 1 { 1 } else { 0 }`, 1)
	})
}

// Test_ConditionalLogic_Boundary_EmptyCollection 空集合判断
func Test_ConditionalLogic_Boundary_EmptyCollection(t *testing.T) {
	t.Run("空数组长度为0", func(t *testing.T) {
	runIntTest(t, `arr := []
if len(arr) == 0 { 1 } else { 0 }`, 1)
	})
	t.Run("非空数组长度大于0", func(t *testing.T) {
	runIntTest(t, `arr := [1]
if len(arr) > 0 { 1 } else { 0 }`, 1)
	})
	t.Run("空Map长度为0", func(t *testing.T) {
	runIntTest(t, `m := {}
if len(m) == 0 { 1 } else { 0 }`, 1)
	})
	t.Run("非空Map长度大于0", func(t *testing.T) {
	runIntTest(t, `m := {"a": 1}
if len(m) > 0 { 1 } else { 0 }`, 1)
	})
}

