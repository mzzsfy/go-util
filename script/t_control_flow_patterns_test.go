package script

import (
	"strconv"
	"testing"
)

// ========== 控制流模式测试 ==========
// 参照 expr-lang 的控制流测试和 gopher-lua 的控制结构测试
// 本文件测试 if-else-if 链、嵌套if、for循环、break/continue、throw 和混合控制流

// ---------- if-else if-else 链 ----------

// Test_CF_IfElseIfChain 多分支条件选择
func Test_CF_IfElseIfChain(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			"命中第一分支",
			`x := 95
r := 0
if x >= 90 { r = 1 } else if x >= 80 { r = 2 } else { r = 3 }
r`,
			1,
		},
		{
			"命中第二分支",
			`x := 85
r := 0
if x >= 90 { r = 1 } else if x >= 80 { r = 2 } else { r = 3 }
r`,
			2,
		},
		{
			"命中else分支",
			`x := 50
r := 0
if x >= 90 { r = 1 } else if x >= 80 { r = 2 } else { r = 3 }
r`,
			3,
		},
		{
			"五层分支",
			`score := 75
grade := 0
if score >= 90 { grade = 5 } else if score >= 80 { grade = 4 } else if score >= 70 { grade = 3 } else if score >= 60 { grade = 2 } else { grade = 1 }
grade`,
			3,
		},
		{
			"边界值精确匹配",
			`x := 80
r := 0
if x >= 90 { r = 1 } else if x >= 80 { r = 2 } else { r = 3 }
r`,
			2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// Test_CF_IfElseIfShortCircuit 条件从上到下短路
func Test_CF_IfElseIfShortCircuit(t *testing.T) {
	t.Run("短路不执行后续比较", func(t *testing.T) {
		// x > 0 为false, 短路后不执行 y/x
		runBoolTest(t, `
x := 0
y := 10
result := x > 0 && y > 5
result`, false)
	})

	t.Run("短路后走else", func(t *testing.T) {
		runIntTest(t, `
x := 0
y := 10
r := 0
if x > 0 && y > 5 { r = 1 } else { r = 2 }
r`, 2)
	})

	t.Run("or短路跳过除零", func(t *testing.T) {
		// true || (1/0 > 0) 不执行除零
		runBoolTest(t, `
x := 0
result := true || (1 / x > 0)
result`, true)
	})
}

// Test_CF_IfElseIfAllFail 所有条件都不满足时走else
func Test_CF_IfElseIfAllFail(t *testing.T) {
	runIntTest(t, `
n := -1
r := 0
if n > 10 { r = 1 } else if n > 20 { r = 2 } else if n > 30 { r = 3 } else { r = 99 }
r`, 99)
}

// ---------- 嵌套if ----------

// Test_CF_NestedIfBasic if内嵌套if
func Test_CF_NestedIfBasic(t *testing.T) {
	runIntTest(t, `
x := 10
y := 20
result := 0
if x > 5 {
    if y > 15 {
        result = 1
    }
} else {
    result = 3
}
result`, 1)
}

// Test_CF_NestedIfElse if内嵌套if-else
func Test_CF_NestedIfElse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			"内层走if",
			`
x := 10
y := 20
result := 0
if x > 5 {
    if y > 15 { result = 1 } else { result = 2 }
} else {
    result = 3
}
result`,
			1,
		},
		{
			"内层走else",
			`
x := 10
y := 5
result := 0
if x > 5 {
    if y > 15 { result = 1 } else { result = 2 }
} else {
    result = 3
}
result`,
			2,
		},
		{
			"外层走else",
			`
x := 3
y := 20
result := 0
if x > 5 {
    if y > 15 { result = 1 } else { result = 2 }
} else {
    result = 3
}
result`,
			3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// Test_CF_DeepNestedConditions 多层嵌套条件
func Test_CF_DeepNestedConditions(t *testing.T) {
	runIntTest(t, `
a := 1
b := 2
c := 3
d := 4
result := 0
if a > 0 {
    if b > 1 {
        if c > 2 {
            if d > 3 { result = 100 } else { result = 200 }
        } else { result = 300 }
    } else { result = 400 }
} else { result = 500 }
result`, 100)
}

// ---------- for循环模式 ----------

// Test_CF_StandardForLoop 标准for循环
func Test_CF_StandardForLoop(t *testing.T) {
	runIntTest(t, `
sum := 0
for i := 0; i < 10; i = i + 1 {
    sum = sum + i
}
sum`, 45)
}

// Test_CF_ConditionForLoop 条件for循环
func Test_CF_ConditionForLoop(t *testing.T) {
	runIntTest(t, `
i := 0
sum := 0
for i < 5 {
    sum = sum + i
    i = i + 1
}
sum`, 10)
}

// Test_CF_InfiniteForLoopBreak 无限for循环 + break
func Test_CF_InfiniteForLoopBreak(t *testing.T) {
	runIntTest(t, `
i := 0
for {
    i = i + 1
    if i >= 10 { break }
}
i`, 10)
}

// Test_CF_CountForLoop 计数for循环 for i := n
func Test_CF_CountForLoop(t *testing.T) {
	// for i := n 循环执行n次
	runIntTest(t, `
sum := 0
for i := 5 {
    sum = sum + 1
}
sum`, 5)
}

// Test_CF_RangeForLoop range遍历
func Test_CF_RangeForLoop(t *testing.T) {
	runIntTest(t, `
arr := [10, 20, 30, 40, 50]
sum := 0
for v := range arr {
    sum = sum + v
}
sum`, 150)
}

// Test_CF_RangeTwoVars 双变量range
func Test_CF_RangeTwoVars(t *testing.T) {
	runIntTest(t, `
arr := [10, 20, 30]
sum := 0
for i, v := range arr {
    sum = sum + i * 100 + v
}
sum`, 360)
}

// Test_CF_ForModifyCondition for循环中修改条件变量
func Test_CF_ForModifyCondition(t *testing.T) {
	runIntTest(t, `
sum := 0
for i := 0; i < 100; i = i + 1 {
    if sum > 20 { break }
    sum = sum + i
}
sum`, 21)
}

// Test_CF_ForBreakEarlyExit for循环中通过break提前退出
func Test_CF_ForBreakEarlyExit(t *testing.T) {
	runIntTest(t, `
found := 0
for i := 0; i < 100; i = i + 1 {
    if i == 7 {
        found = i
        break
    }
}
found`, 7)
}

// Test_CF_ForContinue for循环中continue跳过迭代
func Test_CF_ForContinue(t *testing.T) {
	runIntTest(t, `
sum := 0
for i := 0; i < 10; i = i + 1 {
    if i == 3 { continue }
    if i == 5 { continue }
    sum = sum + i
}
sum`, 37) // 0+1+2+4+6+7+8+9 = 37
}

// ---------- break/continue复杂模式 ----------

// Test_CF_BreakWithCondition break带条件
func Test_CF_BreakWithCondition(t *testing.T) {
	runIntTest(t, `
arr := [1, 2, 3, 4, 5, 6, 7, 8]
result := 0
for v := range arr {
    if v > 5 { break }
    result = result + v
}
result`, 15) // 1+2+3+4+5 = 15
}

// Test_CF_ContinueWithCondition continue带条件
func Test_CF_ContinueWithCondition(t *testing.T) {
	runIntTest(t, `
arr := [1, 2, 3, 4, 5, 6]
result := 0
for v := range arr {
    if v % 2 == 0 { continue }
    result = result + v
}
result`, 9) // 1+3+5 = 9
}

// Test_CF_NestedLoopBreak 嵌套循环中的break只跳出内层
func Test_CF_NestedLoopBreak(t *testing.T) {
	runIntTest(t, `
sum := 0
for i := 0; i < 3; i = i + 1 {
    for j := 0; j < 3; j = j + 1 {
        if j == 1 { break }
        sum = sum + 1
    }
}
sum`, 3)
}

// Test_CF_NestedLoopContinue 嵌套循环中的continue
func Test_CF_NestedLoopContinue(t *testing.T) {
	runIntTest(t, `
sum := 0
for i := 0; i < 3; i = i + 1 {
    for j := 0; j < 3; j = j + 1 {
        if j == 1 { continue }
        sum = sum + 1
    }
}
sum`, 6)
}

// Test_CF_BreakLoopVarState break后循环变量状态
func Test_CF_BreakLoopVarState(t *testing.T) {
	runIntTest(t, `
i := 0
for i = 0; i < 100; i = i + 1 {
    if i == 42 { break }
}
i`, 42)
}

// ---------- throw错误处理 ----------

// Test_CF_ThrowString throw字符串
func Test_CF_ThrowString(t *testing.T) {
	runRuntimeErrorTest(t, `throw "error message"`)
}

// Test_CF_ThrowNumber throw数字
func Test_CF_ThrowNumber(t *testing.T) {
	runRuntimeErrorTest(t, `throw 404`)
}

// Test_CF_ThrowInFunction throw在函数内
func Test_CF_ThrowInFunction(t *testing.T) {
	runRuntimeErrorTest(t, `
fn fail() { throw "func error" }
fail()`)
}

// Test_CF_ThrowInLoop throw在循环内
func Test_CF_ThrowInLoop(t *testing.T) {
	runRuntimeErrorTest(t, `
for i := 0; i < 5; i = i + 1 {
    if i == 3 { throw "loop error" }
}`)
}

// Test_CF_ThrowInCondition throw在条件内
func Test_CF_ThrowInCondition(t *testing.T) {
	runRuntimeErrorTest(t, `if true { throw "cond error" }`)
}

// Test_CF_ThrowContainsMessage throw错误消息包含原始值
func Test_CF_ThrowContainsMessage(t *testing.T) {
	parser := NewParser()
	script, err := parser.Compile(`throw "my error"`)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}
	engine := NewEngine()
	ctx := NewContext()
	_, err = engine.Run(ctx, script)
	if err == nil {
		t.Fatal("throw应产生错误")
	}
	if !contains(err.Error(), "my error") {
		t.Errorf("错误消息应包含 'my error', 得到: %s", err.Error())
	}
}

// ---------- 混合控制流 ----------

// Test_CF_FunctionWithReturnAndIf 函数中的return和if/else
func Test_CF_FunctionWithReturnAndIf(t *testing.T) {
	runIntTest(t, `
fn classify(n) {
    if n > 0 { return 1 }
    if n < 0 { return -1 }
    return 0
}
classify(0)`, 0)
}

// Test_CF_FunctionReturnNeg 函数返回负值
func Test_CF_FunctionReturnNeg(t *testing.T) {
	runIntTest(t, `
fn classify(n) {
    if n > 0 { return 1 }
    if n < 0 { return -1 }
    return 0
}
classify(-42)`, -1)
}

// Test_CF_FunctionReturnPos 函数返回正值
func Test_CF_FunctionReturnPos(t *testing.T) {
	runIntTest(t, `
fn classify(n) {
    if n > 0 { return 1 }
    if n < 0 { return -1 }
    return 0
}
classify(99)`, 1)
}

// Test_CF_LoopWithFunctionCall 循环中的函数调用
func Test_CF_LoopWithFunctionCall(t *testing.T) {
	runIntTest(t, `
fn double(x) { x * 2 }
sum := 0
for i := 0; i < 5; i = i + 1 {
    sum = sum + double(i)
}
sum`, 20) // 0+2+4+6+8 = 20
}

// Test_CF_ConditionWithComplexExpr 条件中的复杂表达式
func Test_CF_ConditionWithComplexExpr(t *testing.T) {
	runIntTest(t, `
a := 5
b := 3
c := 2
total := (a + b) * c
r := 0
if total > 10 {
    r = 100
} else {
    r = 200
}
r`, 100) // (5+3)*2=16 > 10 => true
}

// Test_CF_ElseIfWithLogic else-if中使用逻辑运算
func Test_CF_ElseIfWithLogic(t *testing.T) {
	runIntTest(t, `
age := 25
hasLicense := true
r := 0
if age >= 18 && hasLicense {
    r = 1
} else if age >= 18 {
    r = 2
} else {
    r = 3
}
r`, 1)
}

// Test_CF_ForInRange 用range遍历查找
func Test_CF_ForInRange(t *testing.T) {
	runIntTest(t, `
arr := [3, 1, 4, 1, 5, 9, 2, 6]
target := 9
index := -1
for i, v := range arr {
    if v == target {
        index = i
        break
    }
}
index`, 5)
}

// Test_CF_AccumulatorPattern 循环累加器模式
func Test_CF_AccumulatorPattern(t *testing.T) {
	runIntTest(t, `
n := 5
fact := 1
for i := 1; i <= n; i = i + 1 {
    fact = fact * i
}
fact`, 120) // 5! = 120
}

// Test_CF_WhileSumPattern while求和模式
func Test_CF_WhileSumPattern(t *testing.T) {
	runIntTest(t, `
n := 100
sum := 0
i := 1
for i <= n {
    sum = sum + i
    i = i + 1
}
sum`, 5050)
}

// Test_CF_LoopAccumulateArray 循环构建数组并求和
func Test_CF_LoopAccumulateArray(t *testing.T) {
	runIntTest(t, `
arr := [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
evenSum := 0
for v := range arr {
    if v % 2 == 0 {
        evenSum = evenSum + v
    }
}
evenSum`, 30) // 2+4+6+8+10 = 30
}

// Test_CF_SwitchLikePattern 用if-else-if模拟switch
func Test_CF_SwitchLikePattern(t *testing.T) {
	tests := []struct {
		name     string
		day      int
		expected string
	}{
		{"周一", 1, "weekday"},
		{"周三", 3, "weekday"},
		{"周五", 5, "weekday"},
		{"周六", 6, "weekend"},
		{"周日", 7, "weekend"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := "day := " + strconv.Itoa(tt.day) + `
label := ""
if day <= 5 { label = "weekday" } else { label = "weekend" }
label`
			runStringTest(t, script, tt.expected)
		})
	}
}
