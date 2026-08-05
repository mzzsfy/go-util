package script

import "testing"

// ========== 循环模式和控制流完整测试 ==========
// 覆盖标准for、条件for、无限for、计数for、range遍历、嵌套循环、break/continue等

// ========== 标准 for 循环 ==========

// Test_LoopPattern_StandardFor_BasicAccumulate 基本累加
func Test_LoopPattern_StandardFor_BasicAccumulate(t *testing.T) {
	t.Run("0到9累加", func(t *testing.T) {
		runIntTest(t, `
sum := 0
for i := 0; i < 10; i = i + 1 {
    sum = sum + i
}
sum`, 45)
	})
}

// Test_LoopPattern_StandardFor_1To100 大循环求和
func Test_LoopPattern_StandardFor_1To100(t *testing.T) {
	t.Run("1到100累加", func(t *testing.T) {
		runIntTest(t, `
sum := 0
for i := 1; i <= 100; i = i + 1 {
    sum = sum + i
}
sum`, 5050)
	})
}

// Test_LoopPattern_StandardFor_PostDecrement post为减法
func Test_LoopPattern_StandardFor_PostDecrement(t *testing.T) {
	t.Run("倒数累加", func(t *testing.T) {
		runIntTest(t, `
sum := 0
for i := 5; i > 0; i = i - 1 {
    sum = sum + i
}
sum`, 15)
	})
}

// Test_LoopPattern_StandardFor_PostMultiply post为乘法
func Test_LoopPattern_StandardFor_PostMultiply(t *testing.T) {
	t.Run("指数增长计数", func(t *testing.T) {
		runIntTest(t, `
count := 0
for i := 1; i < 100; i = i * 2 {
    count = count + 1
}
count`, 7)
	})
}

// Test_LoopPattern_StandardFor_CondFalseAtStart 条件初始为false不执行
func Test_LoopPattern_StandardFor_CondFalseAtStart(t *testing.T) {
	t.Run("条件不满足时不执行", func(t *testing.T) {
		runIntTest(t, `
hit := 0
for i := 10; i < 5; i = i + 1 {
    hit = 1
}
hit`, 0)
	})
}

// Test_LoopPattern_StandardFor_InitZero 初始值为0的循环
func Test_LoopPattern_StandardFor_InitZero(t *testing.T) {
	t.Run("i从0开始", func(t *testing.T) {
		runIntTest(t, `
count := 0
for i := 0; i < 5; i = i + 1 {
    count = count + 1
}
count`, 5)
	})
}

// Test_LoopPattern_StandardFor_InitGreaterThanCond 初始值大于条件值
func Test_LoopPattern_StandardFor_InitGreaterThanCond(t *testing.T) {
	t.Run("初始值100条件值10", func(t *testing.T) {
		runIntTest(t, `
hit := 0
for i := 100; i < 10; i = i + 1 {
    hit = 1
}
hit`, 0)
	})
}

// Test_LoopPattern_StandardFor_Factorial 阶乘
func Test_LoopPattern_StandardFor_Factorial(t *testing.T) {
	t.Run("5的阶乘", func(t *testing.T) {
		runIntTest(t, `
n := 5
fact := 1
for i := 1; i <= n; i = i + 1 {
    fact = fact * i
}
fact`, 120)
	})
}

// ========== 条件 for 循环 ==========

// Test_LoopPattern_CondFor_WhileSemantic while语义
func Test_LoopPattern_CondFor_WhileSemantic(t *testing.T) {
	t.Run("while循环累加", func(t *testing.T) {
		runIntTest(t, `
i := 0
sum := 0
for i < 5 {
    sum = sum + i
    i = i + 1
}
sum`, 10)
	})
}

// Test_LoopPattern_CondFor_VarOutside 条件变量在循环外
func Test_LoopPattern_CondFor_VarOutside(t *testing.T) {
	t.Run("外部变量作为条件", func(t *testing.T) {
		runIntTest(t, `
n := 3
i := 0
for n > 0 {
    i = i + n
    n = n - 1
}
i`, 6)
	})
}

// Test_LoopPattern_CondFor_ModifyInBody 条件变量在循环内修改
func Test_LoopPattern_CondFor_ModifyInBody(t *testing.T) {
	t.Run("循环内修改条件变量", func(t *testing.T) {
		runIntTest(t, `
i := 0
for i < 100 {
    i = i + 5
    if i >= 20 { break }
}
i`, 20)
	})
}

// Test_LoopPattern_CondFor_FuncCondition 条件使用函数返回值
func Test_LoopPattern_CondFor_FuncCondition(t *testing.T) {
	t.Run("函数作为循环条件", func(t *testing.T) {
		runIntTest(t, `
fn should_continue(n) { n < 5 }
i := 0
for should_continue(i) {
    i = i + 1
}
i`, 5)
	})
}

// Test_LoopPattern_CondFor_ArrayLenCondition 条件使用数组长度
func Test_LoopPattern_CondFor_ArrayLenCondition(t *testing.T) {
	t.Run("用len作为循环条件", func(t *testing.T) {
		runIntTest(t, `
arr := [10, 20, 30, 40, 50]
i := 0
sum := 0
for i < len(arr) {
    sum = sum + arr[i]
    i = i + 1
}
sum`, 150)
	})
}

// ========== 无限 for 循环 ==========

// Test_LoopPattern_InfiniteFor_BreakOnce 执行一次后break
func Test_LoopPattern_InfiniteFor_BreakOnce(t *testing.T) {
	t.Run("无限循环break执行一次", func(t *testing.T) {
		runIntTest(t, `
hit := 0
for {
    hit = hit + 1
    break
}
hit`, 1)
	})
}

// Test_LoopPattern_InfiniteFor_SimulateWhile 模拟while循环
func Test_LoopPattern_InfiniteFor_SimulateWhile(t *testing.T) {
	t.Run("for+if+break模拟while", func(t *testing.T) {
		runIntTest(t, `
i := 0
for {
    if i >= 10 { break }
    i = i + 1
}
i`, 10)
	})
}

// Test_LoopPattern_InfiniteFor_CondBreak 条件break
func Test_LoopPattern_InfiniteFor_CondBreak(t *testing.T) {
	t.Run("无限循环内条件break", func(t *testing.T) {
		runIntTest(t, `
i := 0
sum := 0
for {
    sum = sum + i
    i = i + 1
    if sum > 50 { break }
}
sum`, 55) // 0+1+2+...+10=55
	})
}

// ========== 计数 for 循环 ==========

// Test_LoopPattern_CountFor_FiveTimes 执行5次
func Test_LoopPattern_CountFor_FiveTimes(t *testing.T) {
	t.Run("for_i:=5循环5次", func(t *testing.T) {
		runIntTest(t, `
sum := 0
for i := 5 {
    sum = sum + 1
}
sum`, 5)
	})
}

// Test_LoopPattern_CountFor_ZeroTimes 不执行
func Test_LoopPattern_CountFor_ZeroTimes(t *testing.T) {
	t.Run("for_i:=0不执行", func(t *testing.T) {
		runIntTest(t, `
hit := 0
for i := 0 {
    hit = hit + 1
}
hit`, 0)
	})
}

// Test_LoopPattern_CountFor_Once 执行一次
func Test_LoopPattern_CountFor_Once(t *testing.T) {
	t.Run("for_i:=1循环1次", func(t *testing.T) {
		runIntTest(t, `
hit := 0
for i := 1 {
    hit = hit + 1
}
hit`, 1)
	})
}

// Test_LoopPattern_CountFor_Large 大循环
func Test_LoopPattern_CountFor_Large(t *testing.T) {
	t.Run("for_i:=100循环100次", func(t *testing.T) {
		runIntTest(t, `
sum := 0
for i := 100 {
    sum = sum + 1
}
sum`, 100)
	})
}

// Test_LoopPattern_CountFor_WithVar 使用变量
func Test_LoopPattern_CountFor_WithVar(t *testing.T) {
	t.Run("for_i:=n变量次数", func(t *testing.T) {
		runIntTest(t, `
n := 7
sum := 0
for i := n {
    sum = sum + 1
}
sum`, 7)
	})
}

// ========== range 数组 - 单变量 ==========

// Test_LoopPattern_RangeSingle_Traverse 遍历元素求和
func Test_LoopPattern_RangeSingle_Traverse(t *testing.T) {
	t.Run("range遍历求和", func(t *testing.T) {
		runIntTest(t, `
arr := [10, 20, 30, 40, 50]
sum := 0
for v := range arr {
    sum = sum + v
}
sum`, 150)
	})
}

// Test_LoopPattern_RangeSingle_EmptyArray 空数组不执行
func Test_LoopPattern_RangeSingle_EmptyArray(t *testing.T) {
	t.Run("空数组range", func(t *testing.T) {
		runIntTest(t, `
arr := []
hit := 0
for v := range arr {
    hit = hit + 1
}
hit`, 0)
	})
}

// Test_LoopPattern_RangeSingle_SingleElement 单元素数组
func Test_LoopPattern_RangeSingle_SingleElement(t *testing.T) {
	t.Run("单元素数组range", func(t *testing.T) {
		runIntTest(t, `
arr := [42]
sum := 0
for v := range arr {
    sum = sum + v
}
sum`, 42)
	})
}

// Test_LoopPattern_RangeSingle_Accumulate 遍历中累加
func Test_LoopPattern_RangeSingle_Accumulate(t *testing.T) {
	t.Run("range累加奇数", func(t *testing.T) {
		runIntTest(t, `
arr := [1, 2, 3, 4, 5, 6, 7, 8, 9]
oddSum := 0
for v := range arr {
    if v % 2 != 0 {
        oddSum = oddSum + v
    }
}
oddSum`, 25)
	})
}

// ========== range 数组 - 双变量 ==========

// Test_LoopPattern_RangeDouble_IndexValue 索引+值
func Test_LoopPattern_RangeDouble_IndexValue(t *testing.T) {
	t.Run("双变量range求和", func(t *testing.T) {
		runIntTest(t, `
arr := [10, 20, 30]
sum := 0
for i, v := range arr {
    sum = sum + i * 100 + v
}
sum`, 360)
	})
}

// Test_LoopPattern_RangeDouble_UseIndexAndValue 使用索引和值
func Test_LoopPattern_RangeDouble_UseIndexAndValue(t *testing.T) {
	t.Run("索引查找特定值", func(t *testing.T) {
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
	})
}

// Test_LoopPattern_RangeDouble_EmptyArray 空数组
func Test_LoopPattern_RangeDouble_EmptyArray(t *testing.T) {
	t.Run("空数组双变量range", func(t *testing.T) {
		runIntTest(t, `
arr := []
hit := 0
for i, v := range arr {
    hit = hit + 1
}
hit`, 0)
	})
}

// Test_LoopPattern_RangeDouble_SingleElement 单元素
func Test_LoopPattern_RangeDouble_SingleElement(t *testing.T) {
	t.Run("单元素双变量range", func(t *testing.T) {
		runIntTest(t, `
arr := [99]
idx := -1
val := -1
for i, v := range arr {
    idx = i
    val = v
}
idx * 1000 + val`, 99)
	})
}

// ========== range Map ==========

// Test_LoopPattern_RangeMap_Key 遍历key
func Test_LoopPattern_RangeMap_Key(t *testing.T) {
	t.Run("range_map遍历计数", func(t *testing.T) {
		runIntTest(t, `
m := {"a": 1, "b": 2, "c": 3}
count := 0
for k := range m {
    count = count + 1
}
count`, 3)
	})
}

// Test_LoopPattern_RangeMap_KeyValue 遍历键值
func Test_LoopPattern_RangeMap_KeyValue(t *testing.T) {
	t.Run("range_map双变量求和", func(t *testing.T) {
		runIntTest(t, `
m := {"a": 10, "b": 20, "c": 30}
sum := 0
for k, v := range m {
    sum = sum + v
}
sum`, 60)
	})
}

// Test_LoopPattern_RangeMap_Empty 空 Map
func Test_LoopPattern_RangeMap_Empty(t *testing.T) {
	t.Run("空Map遍历", func(t *testing.T) {
		runIntTest(t, `
m := {}
hit := 0
for k := range m {
    hit = 1
}
hit`, 0)
	})
}

// Test_LoopPattern_RangeMap_Multiple 多元素 Map
func Test_LoopPattern_RangeMap_Multiple(t *testing.T) {
	t.Run("多元素Map统计", func(t *testing.T) {
		runIntTest(t, `
scores := {"a": 85, "b": 60, "c": 95, "d": 40}
pass := 0
for k, v := range scores {
    if v >= 60 {
        pass = pass + 1
    }
}
pass`, 3)
	})
}

// ========== break ==========

// Test_LoopPattern_Break_Basic 基本break
func Test_LoopPattern_Break_Basic(t *testing.T) {
	t.Run("break跳出循环", func(t *testing.T) {
		runIntTest(t, `
result := 0
for i := 0; i < 100; i = i + 1 {
    if i == 10 { break }
    result = i
}
result`, 9)
	})
}

// Test_LoopPattern_Break_Conditional 条件break
func Test_LoopPattern_Break_Conditional(t *testing.T) {
	t.Run("条件break", func(t *testing.T) {
		runIntTest(t, `
arr := [1, 2, 3, 4, 5, 6, 7, 8]
result := 0
for v := range arr {
    if v > 5 { break }
    result = result + v
}
result`, 15)
	})
}

// Test_LoopPattern_Break_OnlyInnermost 多层嵌套只跳出最内层
func Test_LoopPattern_Break_OnlyInnermost(t *testing.T) {
	t.Run("break只跳出内层", func(t *testing.T) {
		runIntTest(t, `
sum := 0
for i := 0; i < 3; i = i + 1 {
    for j := 0; j < 5; j = j + 1 {
        if j == 2 { break }
        sum = sum + 1
    }
}
sum`, 6)
	})
}

// Test_LoopPattern_Break_InIf break在嵌套if中
func Test_LoopPattern_Break_InIf(t *testing.T) {
	t.Run("嵌套if中的break", func(t *testing.T) {
		runIntTest(t, `
result := 0
for i := 0; i < 20; i = i + 1 {
    if i > 5 {
        if i > 8 {
            result = i
            break
        }
    }
    result = i
}
result`, 9)
	})
}

// Test_LoopPattern_Break_InRange range循环中的break
func Test_LoopPattern_Break_InRange(t *testing.T) {
	t.Run("range循环break", func(t *testing.T) {
		runIntTest(t, `
arr := [10, 20, 30, 40, 50]
found := 0
for v := range arr {
    if v == 30 {
        found = v
        break
    }
}
found`, 30)
	})
}

// ========== continue ==========

// Test_LoopPattern_Continue_Basic 基本continue
func Test_LoopPattern_Continue_Basic(t *testing.T) {
	t.Run("continue跳过当前迭代", func(t *testing.T) {
		runIntTest(t, `
sum := 0
for i := 0; i < 10; i = i + 1 {
    if i == 3 { continue }
    if i == 5 { continue }
    sum = sum + i
}
sum`, 37)
	})
}

// Test_LoopPattern_Continue_Conditional 条件continue
func Test_LoopPattern_Continue_Conditional(t *testing.T) {
	t.Run("条件continue跳过偶数", func(t *testing.T) {
		runIntTest(t, `
arr := [1, 2, 3, 4, 5, 6]
result := 0
for v := range arr {
    if v % 2 == 0 { continue }
    result = result + v
}
result`, 9)
	})
}

// Test_LoopPattern_Continue_OnlyInnermost 嵌套中只影响最内层
func Test_LoopPattern_Continue_OnlyInnermost(t *testing.T) {
	t.Run("continue只影响内层", func(t *testing.T) {
		runIntTest(t, `
count := 0
for i := 0; i < 3; i = i + 1 {
    for j := 0; j < 5; j = j + 1 {
        if j == 2 { continue }
        count = count + 1
    }
}
count`, 12)
	})
}

// Test_LoopPattern_Continue_SkipElements continue跳过某些元素
func Test_LoopPattern_Continue_SkipElements(t *testing.T) {
	t.Run("跳过小于3的元素", func(t *testing.T) {
		runIntTest(t, `
arr := [1, 2, 3, 4, 5]
sum := 0
for v := range arr {
    if v < 3 { continue }
    sum = sum + v
}
sum`, 12)
	})
}

// ========== 嵌套循环 ==========

// Test_LoopPattern_Nested_DoubleSum 双层嵌套求和
func Test_LoopPattern_Nested_DoubleSum(t *testing.T) {
	t.Run("双层嵌套计数", func(t *testing.T) {
		runIntTest(t, `
sum := 0
for i := 0; i < 4; i = i + 1 {
    for j := 0; j < 4; j = j + 1 {
        sum = sum + 1
    }
}
sum`, 16)
	})
}

// Test_LoopPattern_Nested_Matrix 矩阵乘法
func Test_LoopPattern_Nested_Matrix(t *testing.T) {
	t.Run("2x2矩阵乘法", func(t *testing.T) {
		runIntTest(t, `
a := [[1, 2], [3, 4]]
b := [[5, 6], [7, 8]]
result := []
for i := 0; i < 2; i = i + 1 {
    row := []
    for j := 0; j < 2; j = j + 1 {
        s := 0
        for k := 0; k < 2; k = k + 1 {
            s = s + a[i][k] * b[k][j]
        }
        row = push(row, s)
    }
    result = push(result, row)
}
result[0][0] + result[1][1]`, 69) // 19 + 50 = 69
	})
}

// Test_LoopPattern_Nested_Triple 三层嵌套
func Test_LoopPattern_Nested_Triple(t *testing.T) {
	t.Run("三层嵌套计数", func(t *testing.T) {
		runIntTest(t, `
count := 0
for i := 0; i < 3; i = i + 1 {
    for j := 0; j < 3; j = j + 1 {
        for k := 0; k < 3; k = k + 1 {
            count = count + 1
        }
    }
}
count`, 27)
	})
}

// Test_LoopPattern_Nested_OuterBreakWithFlag 外层break需要条件标志
func Test_LoopPattern_Nested_OuterBreakWithFlag(t *testing.T) {
	t.Run("标志变量控制外层break", func(t *testing.T) {
		runIntTest(t, `
result := 0
found := false
for i := 0; i < 5; i = i + 1 {
    for j := 0; j < 5; j = j + 1 {
        if i == 2 && j == 3 {
            found = true
            result = i * 100 + j
            break
        }
    }
    if found { break }
}
result`, 203)
	})
}

// Test_LoopPattern_Nested_InnerBreakNoOuter 内层break不影响外层
func Test_LoopPattern_Nested_InnerBreakNoOuter(t *testing.T) {
	t.Run("内层break不影响外层", func(t *testing.T) {
		runIntTest(t, `
count := 0
for i := 0; i < 3; i = i + 1 {
    for j := 0; j < 10; j = j + 1 {
        if j == 2 { break }
    }
    count = count + 1
}
count`, 3)
	})
}

// ========== 循环内控制流组合 ==========

// Test_LoopPattern_IfBreak if+break组合
func Test_LoopPattern_IfBreak(t *testing.T) {
	t.Run("if+break查找", func(t *testing.T) {
		runIntTest(t, `
found := 0
for i := 0; i < 100; i = i + 1 {
    if i == 7 {
        found = i
        break
    }
}
found`, 7)
	})
}

// Test_LoopPattern_IfContinue if+continue组合
func Test_LoopPattern_IfContinue(t *testing.T) {
	t.Run("if+continue跳过", func(t *testing.T) {
		runIntTest(t, `
sum := 0
for i := 0; i < 10; i = i + 1 {
    if i % 3 == 0 { continue }
    sum = sum + i
}
sum`, 27) // 1+2+4+5+7+8=27
	})
}

// Test_LoopPattern_MixedBreakContinue if+break+continue混合
func Test_LoopPattern_MixedBreakContinue(t *testing.T) {
	t.Run("break和continue混合", func(t *testing.T) {
		runIntTest(t, `
sum := 0
for i := 0; i < 20; i = i + 1 {
    if i == 15 { break }
    if i < 5 { continue }
    sum = sum + i
}
sum`, 95) // 5+6+7+8+9+10+11+12+13+14=95
	})
}

// Test_LoopPattern_NestedIfInLoop 循环中嵌套if
func Test_LoopPattern_NestedIfInLoop(t *testing.T) {
	t.Run("循环中多层嵌套if", func(t *testing.T) {
		runIntTest(t, `
result := 0
for i := 0; i < 10; i = i + 1 {
    if i > 2 {
        if i < 7 {
            if i == 5 {
                result = result + 100
            }
            result = result + 1
        }
    }
}
result`, 104) // i=3,4,5,6 各+1, i=5额外+100 => 4+100=104
	})
}

// Test_LoopPattern_ThrowInLoop 循环内throw
func Test_LoopPattern_ThrowInLoop(t *testing.T) {
	t.Run("循环内throw触发错误", func(t *testing.T) {
		runRuntimeErrorTest(t, `
for i := 0; i < 5; i = i + 1 {
    if i == 3 { throw "loop error" }
}`)
	})
}

// ========== 循环和函数 ==========

// Test_LoopPattern_LoopInsideFunc 函数内循环
func Test_LoopPattern_LoopInsideFunc(t *testing.T) {
	t.Run("函数内for求和", func(t *testing.T) {
		runIntTest(t, `
fn sum_range(a, b) {
    s := 0
    for i := a; i <= b; i = i + 1 {
        s = s + i
    }
    return s
}
sum_range(1, 10)`, 55)
	})
}

// Test_LoopPattern_FuncCallInLoop 循环内调用函数
func Test_LoopPattern_FuncCallInLoop(t *testing.T) {
	t.Run("循环内函数调用累加", func(t *testing.T) {
		runIntTest(t, `
fn double(x) { x * 2 }
sum := 0
for i := 0; i < 5; i = i + 1 {
    sum = sum + double(i)
}
sum`, 20)
	})
}

// Test_LoopPattern_RecursionInLoop 循环中的递归
func Test_LoopPattern_RecursionInLoop(t *testing.T) {
	t.Run("循环内递归fib", func(t *testing.T) {
		runIntTest(t, `
fn fib(n) {
    if n < 2 { return n }
    return fib(n - 1) + fib(n - 2)
}
sum := 0
for i := 0; i < 5; i = i + 1 {
    sum = sum + fib(i)
}
sum`, 7) // 0+1+1+2+3=7
	})
}

// Test_LoopPattern_FuncReturnAsCond 函数返回值作为循环条件
func Test_LoopPattern_FuncReturnAsCond(t *testing.T) {
	t.Run("函数返回值控制循环", func(t *testing.T) {
		runIntTest(t, `
fn limit(n) { n < 8 }
i := 0
for limit(i) {
    i = i + 2
}
i`, 8)
	})
}

// Test_LoopPattern_FuncReturnAsBodyOp 函数返回值作为循环体操作
func Test_LoopPattern_FuncReturnAsBodyOp(t *testing.T) {
	t.Run("循环内函数返回值参与运算", func(t *testing.T) {
		runIntTest(t, `
fn sq(x) { x * x }
sum := 0
for i := 1; i <= 5; i = i + 1 {
    sum = sum + sq(i)
}
sum`, 55) // 1+4+9+16+25=55
	})
}

// ========== 循环和数组操作 ==========

// Test_LoopPattern_BuildArray 循环构建数组
func Test_LoopPattern_BuildArray(t *testing.T) {
	t.Run("push模式构建数组", func(t *testing.T) {
		runIntTest(t, `
arr := []
for i := 1; i <= 5; i = i + 1 {
    arr = push(arr, i * i)
}
arr[4]`, 25)
	})
}

// Test_LoopPattern_ModifyArray 循环修改数组元素
func Test_LoopPattern_ModifyArray(t *testing.T) {
	t.Run("循环翻倍数组元素", func(t *testing.T) {
		runIntTest(t, `
arr := [1, 2, 3, 4, 5]
for i := 0; i < len(arr); i = i + 1 {
    arr[i] = arr[i] * 2
}
arr[3]`, 8)
	})
}

// Test_LoopPattern_FindInArray 循环查找数组元素
func Test_LoopPattern_FindInArray(t *testing.T) {
	t.Run("线性查找", func(t *testing.T) {
		runIntTest(t, `
arr := [3, 1, 4, 1, 5, 9, 2, 6]
target := 9
index := -1
for i := 0; i < len(arr); i = i + 1 {
    if arr[i] == target {
        index = i
        break
    }
}
index`, 5)
	})
}

// Test_LoopPattern_FilterArray 循环过滤数组
func Test_LoopPattern_FilterArray(t *testing.T) {
	t.Run("过滤偶数", func(t *testing.T) {
		runIntTest(t, `
arr := [1, 2, 3, 4, 5, 6, 7, 8]
result := []
for v := range arr {
    if v % 2 == 0 {
        result = push(result, v)
    }
}
len(result)`, 4)
	})
}

// Test_LoopPattern_MapArray 循环映射数组
func Test_LoopPattern_MapArray(t *testing.T) {
	t.Run("映射为平方", func(t *testing.T) {
		runIntTest(t, `
arr := [1, 2, 3, 4]
result := []
for v := range arr {
    result = push(result, v * v)
}
result[2]`, 9)
	})
}

// ========== 循环和 Map 操作 ==========

// Test_LoopPattern_BuildMap 循环构建Map
func Test_LoopPattern_BuildMap(t *testing.T) {
	t.Run("循环构建Map", func(t *testing.T) {
		runIntTest(t, `
m := {}
for i := 1; i <= 3; i = i + 1 {
    m[i] = i * i
}
m[3]`, 9)
	})
}

// Test_LoopPattern_ModifyMapValue 循环修改Map值
func Test_LoopPattern_ModifyMapValue(t *testing.T) {
	t.Run("range遍历修改Map值", func(t *testing.T) {
		runIntTest(t, `
m := {"a": 1, "b": 2, "c": 3}
m2 := {}
for k, v := range m {
    m2[k] = v * 10
}
m2["c"]`, 30)
	})
}

// Test_LoopPattern_CountMap 循环统计Map
func Test_LoopPattern_CountMap(t *testing.T) {
	t.Run("统计Map中满足条件的数量", func(t *testing.T) {
		runIntTest(t, `
m := {"a": 10, "b": 20, "c": 30, "d": 5}
count := 0
for k, v := range m {
    if v >= 15 {
        count = count + 1
    }
}
count`, 2) // 20和30
	})
}

// ========== 循环边界 ==========

// Test_LoopPattern_LargeLoop 大循环1000次
func Test_LoopPattern_LargeLoop(t *testing.T) {
	t.Run("1000次循环计数", func(t *testing.T) {
		runIntTest(t, `
sum := 0
for i := 1; i <= 1000; i = i + 1 {
    sum = sum + 1
}
sum`, 1000)
	})
}

// Test_LoopPattern_CondNeverSatisfied 循环条件永不满足
func Test_LoopPattern_CondNeverSatisfied(t *testing.T) {
	t.Run("条件恒为false", func(t *testing.T) {
		runIntTest(t, `
hit := 0
for i := 0; 1 < 0; i = i + 1 {
    hit = 1
}
hit`, 0)
	})
}

// Test_LoopPattern_EmptyBody 循环体为空
func Test_LoopPattern_EmptyBody(t *testing.T) {
	t.Run("空循环体执行后变量值", func(t *testing.T) {
		runIntTest(t, `
for i := 0; i < 10; i = i + 1 {}
0`, 0)
	})
}

// Test_LoopPattern_BreakPreservesVar break后循环变量保持当前值
func Test_LoopPattern_BreakPreservesVar(t *testing.T) {
	t.Run("break后变量值", func(t *testing.T) {
		runIntTest(t, `
i := 0
for i = 0; i < 100; i = i + 1 {
    if i == 42 { break }
}
i`, 42)
	})
}

// ========== continue + break 组合模式 ==========

// Test_LoopPattern_SkipEvenBreakAtValue 跳过偶数到某值停止
func Test_LoopPattern_SkipEvenBreakAtValue(t *testing.T) {
	t.Run("跳过偶数遇11停止", func(t *testing.T) {
		runIntTest(t, `
sum := 0
for i := 0; i < 100; i = i + 1 {
    if i % 2 == 0 { continue }
    if i > 10 { break }
    sum = sum + i
}
sum`, 25) // 1+3+5+7+9=25
	})
}

// Test_LoopPattern_CollectFiltered 收集满足条件的元素
func Test_LoopPattern_CollectFiltered(t *testing.T) {
	t.Run("收集大于3小于8的元素", func(t *testing.T) {
		runIntTest(t, `
arr := [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
result := []
for v := range arr {
    if v > 3 {
        if v < 8 {
            result = push(result, v)
        }
    }
}
len(result)`, 4)
	})
}
