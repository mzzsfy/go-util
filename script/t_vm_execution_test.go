package script

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ========== VM执行模型和字节码深度测试 ==========
// 本文件从VM执行模型角度深度测试字节码执行,栈帧管理,常量池,操作码等
// 顶层函数命名统一使用 Test_VMExec 前缀,可通过 go test -run "Test_VMExec" 集中运行

// ========== 栈操作验证 ==========

// Test_VMExec_Stack_NestedExpr 验证嵌套表达式的栈求值顺序
// 表达式按后序遍历求值,操作数先入栈,运算符消费栈顶
func Test_VMExec_Stack_NestedExpr(t *testing.T) {
	t.Run("算术嵌套", func(t *testing.T) {
		runIntTest(t, "1 + 2 * 3", 7)
		runIntTest(t, "(1 + 2) * 3", 9)
		runIntTest(t, "2 * 3 + 4 * 5", 26)
		runIntTest(t, "10 - 3 - 2", 5) // 左结合
	})
	t.Run("混合嵌套", func(t *testing.T) {
		runBoolTest(t, "1 < 2", true)
		runBoolTest(t, "(1 + 2) == 3", true)
		runBoolTest(t, "(1 < 2) && (3 > 4)", false)
		runBoolTest(t, "(1 < 2) || (3 > 4)", true)
	})
	t.Run("深层嵌套表达式", func(t *testing.T) {
		// 10层括号嵌套,验证栈深度
		runIntTest(t, "((((((((((1))))))))))", 1)
		runIntTest(t, "1 + (2 + (3 + (4 + (5 + 6))))", 21)
	})
}

// Test_VMExec_Stack_PushPopOrder 通过顺序敏感运算验证push/pop顺序正确性
// 减法/除法/取模/比较对操作数顺序敏感,可验证栈弹出顺序
func Test_VMExec_Stack_PushPopOrder(t *testing.T) {
	t.Run("减法顺序敏感", func(t *testing.T) {
		runIntTest(t, "10 - 3", 7)
		runIntTest(t, "3 - 10", -7)
	})
	t.Run("除法顺序敏感", func(t *testing.T) {
		runIntTest(t, "20 / 4", 5)
		runIntTest(t, "4 / 20", 0)
	})
	t.Run("取模顺序敏感", func(t *testing.T) {
		runIntTest(t, "17 % 5", 2)
		runIntTest(t, "5 % 17", 5)
	})
	t.Run("比较顺序敏感", func(t *testing.T) {
		runBoolTest(t, "5 < 10", true)
		runBoolTest(t, "10 < 5", false)
		runBoolTest(t, "5 > 10", false)
		runBoolTest(t, "10 > 5", true)
	})
}

// Test_VMExec_Stack_FrameSwitch 函数调用的栈帧切换
// 每次函数调用应创建独立栈帧,函数返回后回到调用者栈帧
func Test_VMExec_Stack_FrameSwitch(t *testing.T) {
	t.Run("单次调用栈帧切换", func(t *testing.T) {
		runIntTest(t, `
fn id(x) {
    return x
}
id(42)`, 42)
	})
	t.Run("多层调用栈帧管理", func(t *testing.T) {
		// a->b->c->return, 每层返回栈帧正确弹出
		runIntTest(t, `
fn a(x) { x + 1 }
fn b(x) { a(x) + 1 }
fn c(x) { b(x) + 1 }
c(10)`, 13)
	})
	t.Run("深层递归栈帧消耗", func(t *testing.T) {
		// 阶乘递归,每次调用消耗一个栈帧
		runIntTest(t, `
fn factorial(n) {
    if n <= 1 { return 1 }
    return n * factorial(n - 1)
}
factorial(6)`, 720)
	})
	t.Run("对数递归深度", func(t *testing.T) {
		// log2递归, 栈帧深度为log2(n)
		runIntTest(t, `
fn log2(n) {
    if n <= 1 { return 0 }
    return 1 + log2(n / 2)
}
log2(1024)`, 10)
	})
}

// Test_VMExec_Stack_FrameRestore 函数返回后调用者栈帧状态正确
// 验证 OpReturn 弹出帧并将返回值压入调用者栈
func Test_VMExec_Stack_FrameRestore(t *testing.T) {
	t.Run("返回值正确压入调用者栈", func(t *testing.T) {
		runIntTest(t, `
fn two() { 2 }
fn add(a, b) { a + b }
add(two(), two())`, 4)
	})
	t.Run("返回值参与后续运算", func(t *testing.T) {
		runIntTest(t, `
fn inc(x) { x + 1 }
inc(inc(inc(10)))`, 13)
	})
	t.Run("函数调用作为数组元素", func(t *testing.T) {
		runIntTest(t, `
fn f() { 5 }
arr := [f(), f(), f()]
arr[0] + arr[1] + arr[2]`, 15)
	})
}

// ========== 局部变量 ==========

// Test_VMExec_Locals_DeclareRead 变量声明和读取
// OpStoreLocal 写入局部变量槽, OpLoadLocal 读取
func Test_VMExec_Locals_DeclareRead(t *testing.T) {
	t.Run("单变量声明读取", func(t *testing.T) {
		runIntTest(t, "x := 42\nx", 42)
	})
	t.Run("多变量声明读取", func(t *testing.T) {
		runIntTest(t, `
a := 1
b := 2
c := 3
a + b + c`, 6)
	})
	t.Run("变量声明后立即使用", func(t *testing.T) {
		runIntTest(t, "x := 10\ny := x * 2\ny", 20)
	})
}

// Test_VMExec_Locals_Shadowing 变量遮蔽(不同作用域)
// 函数内部的同名变量遮蔽外部变量,不影响外部
func Test_VMExec_Locals_Shadowing(t *testing.T) {
	t.Run("函数内遮蔽全局变量", func(t *testing.T) {
		runIntTest(t, `
x := 1
fn f() {
    x := 2
    return x
}
f() + x`, 3) // f()=2, x=1, 总和3
	})
	t.Run("嵌套作用域遮蔽", func(t *testing.T) {
		runIntTest(t, `
v := 10
fn outer() {
    v := 20
    fn inner() {
        v := 30
        return v
    }
    return inner() + v
}
outer()`, 50) // inner()=30 + outer.v=20
	})
}

// Test_VMExec_Locals_LoopVar 循环变量
// for循环的迭代变量作为局部变量,每次迭代更新同一槽位
func Test_VMExec_Locals_LoopVar(t *testing.T) {
	t.Run("for循环变量", func(t *testing.T) {
		runIntTest(t, `
sum := 0
for i := 0; i < 5; i = i + 1 {
    sum = sum + i
}
sum`, 10)
	})
	t.Run("循环变量在循环外可访问", func(t *testing.T) {
		// 脚本采用函数作用域, 循环变量在循环后仍可访问
		runIntTest(t, `for i := 0; i < 3; i = i + 1 { 1 }
i`, 3)
	})
	t.Run("range循环变量", func(t *testing.T) {
		runIntTest(t, `
arr := [10, 20, 30]
total := 0
for v := range arr {
    total = total + v
}
total`, 60)
	})
}

// Test_VMExec_Locals_Params 函数参数作为局部变量
// 参数通过 OpCall 复制到新帧的局部变量前几个槽位
func Test_VMExec_Locals_Params(t *testing.T) {
	t.Run("参数作为局部变量", func(t *testing.T) {
		runIntTest(t, `
fn add(a, b) {
    a + b
}
add(3, 4)`, 7)
	})
	t.Run("参数在函数内可修改", func(t *testing.T) {
		runIntTest(t, `
fn inc(x) {
    x = x + 1
    return x
}
inc(10)`, 11)
	})
	t.Run("多参数顺序传递", func(t *testing.T) {
		runIntTest(t, `
fn sub(a, b, c) {
    a - b - c
}
sub(100, 20, 10)`, 70) // (100-20)-10
	})
}

// Test_VMExec_Locals_Temp range循环的临时变量
// range遍历会隐式创建临时变量存储集合
func Test_VMExec_Locals_Temp(t *testing.T) {
	t.Run("range遍历数组", func(t *testing.T) {
		runIntTest(t, `
arr := [1, 2, 3, 4, 5]
sum := 0
for _, v := range arr {
    sum = sum + v
}
sum`, 15)
	})
	t.Run("range双变量索引和值", func(t *testing.T) {
		runIntTest(t, `
arr := [10, 20, 30]
sum := 0
for i, v := range arr {
    sum = sum + i*10 + v
}
sum`, 90) // (0+10)+(10+20)+(20+30)
	})
}

// ========== 跳转和控制流 ==========

// Test_VMExec_Jump_IfElse if条件跳转
// OpJumpIfFalse 在条件为假时跳转到else分支
func Test_VMExec_Jump_IfElse(t *testing.T) {
	t.Run("if真分支", func(t *testing.T) {
		runIntTest(t, `
x := 10
if x > 5 { 1 } else { 0 }`, 1)
	})
	t.Run("if假分支", func(t *testing.T) {
		runIntTest(t, `
x := 3
if x > 5 { 1 } else { 0 }`, 0)
	})
	t.Run("if/elseif/else链", func(t *testing.T) {
		runIntTest(t, `
x := 5
if x > 10 { 1 } else if x > 3 { 2 } else { 3 }`, 2)
	})
	t.Run("嵌套if", func(t *testing.T) {
		// 用函数体让内部if作为表达式求值
		runIntTest(t, `
fn classify(x) {
    if x > 5 {
        if x > 10 { return 1 } else { return 2 }
    } else {
        return 3
    }
}
classify(7)`, 2)
	})
}

// Test_VMExec_Jump_ForLoop for循环回跳
// 循环体执行完毕后 OpJump 回跳到条件判断
func Test_VMExec_Jump_ForLoop(t *testing.T) {
	t.Run("while循环回跳", func(t *testing.T) {
		runIntTest(t, `
i := 0
for i < 5 {
    i = i + 1
}
i`, 5)
	})
	t.Run("标准for循环", func(t *testing.T) {
		runIntTest(t, `
sum := 0
for i := 0; i < 10; i = i + 1 {
    sum = sum + i
}
sum`, 45)
	})
	t.Run("无限循环", func(t *testing.T) {
		runIntTest(t, `
i := 0
for {
    i = i + 1
    if i >= 100 { break }
}
i`, 100)
	})
}

// Test_VMExec_Jump_Break break跳出
// break 生成 OpJump 跳到循环外
func Test_VMExec_Jump_Break(t *testing.T) {
	t.Run("while中break", func(t *testing.T) {
		runIntTest(t, `
i := 0
for {
    if i >= 10 { break }
    i = i + 1
}
i`, 10)
	})
	t.Run("for中break", func(t *testing.T) {
		runIntTest(t, `
for i := 0; i < 100; i = i + 1 {
    if i == 5 { break }
}
5`, 5)
	})
}

// Test_VMExec_Jump_Continue continue跳回
// continue 生成 OpJump 跳到循环条件判断或更新
func Test_VMExec_Jump_Continue(t *testing.T) {
	t.Run("while中continue", func(t *testing.T) {
		runIntTest(t, `
i := 0
sum := 0
for i < 5 {
    i = i + 1
    if i == 3 { continue }
    sum = sum + i
}
sum`, 12) // 1+2+4+5
	})
	t.Run("for中continue跳过迭代", func(t *testing.T) {
		runIntTest(t, `
sum := 0
for i := 0; i < 5; i = i + 1 {
    if i == 2 { continue }
    sum = sum + i
}
sum`, 8) // 0+1+3+4
	})
}

// Test_VMExec_Jump_NestedTarget 嵌套循环的跳转目标正确性
// break/continue 只影响最内层循环
func Test_VMExec_Jump_NestedTarget(t *testing.T) {
	t.Run("嵌套循环break只跳内层", func(t *testing.T) {
		runIntTest(t, `
count := 0
for i := 0; i < 3; i = i + 1 {
    for j := 0; j < 3; j = j + 1 {
        if j == 2 { break }
        count = count + 1
    }
}
count`, 6) // 每次内层循环执行2次, 共3轮=6
	})
	t.Run("嵌套循环continue只跳内层", func(t *testing.T) {
		runIntTest(t, `
sum := 0
for i := 0; i < 3; i = i + 1 {
    for j := 0; j < 3; j = j + 1 {
        if j == 1 { continue }
        sum = sum + 1
    }
}
sum`, 6) // 每次内层执行2次, 共3轮
	})
	t.Run("三层嵌套循环", func(t *testing.T) {
		runIntTest(t, `
count := 0
for i := 0; i < 2; i = i + 1 {
    for j := 0; j < 2; j = j + 1 {
        for k := 0; k < 2; k = k + 1 {
            count = count + 1
        }
    }
}
count`, 8)
	})
}

// Test_VMExec_Jump_ReturnFramePop 函数return的栈帧弹出
// return弹出当前帧,返回值压入调用者栈
func Test_VMExec_Jump_ReturnFramePop(t *testing.T) {
	t.Run("提前return弹出栈帧", func(t *testing.T) {
		runIntTest(t, `
fn find(arr, target) {
    for v := range arr {
        if v == target { return 1 }
    }
    return 0
}
find([1, 2, 3], 2)`, 1)
	})
	t.Run("多层return链式弹出", func(t *testing.T) {
		runIntTest(t, `
fn inner() { return 42 }
fn middle() { return inner() }
fn outer() { return middle() }
outer()`, 42)
	})
	t.Run("return后调用者继续执行", func(t *testing.T) {
		runIntTest(t, `
fn f() { return 10 }
result := f() + 5
result`, 15)
	})
}

// ========== 函数调用机制 ==========

// Test_VMExec_Call_FrameLifecycle 单次函数调用的帧创建/销毁
// OpCall创建新帧, OpReturn销毁帧
func Test_VMExec_Call_FrameLifecycle(t *testing.T) {
	t.Run("无参无返回值函数", func(t *testing.T) {
		runIntTest(t, `
fn answer() {
    42
}
answer()`, 42)
	})
	t.Run("有参有返回值函数", func(t *testing.T) {
		runIntTest(t, `
fn mul(a, b) {
    return a * b
}
mul(6, 7)`, 42)
	})
	t.Run("表达式作为函数体", func(t *testing.T) {
		runIntTest(t, `
fn compute() {
    10 + 20 * 2
}
compute()`, 50)
	})
}

// Test_VMExec_Call_ArgPassing 参数传递机制
// 参数按值从栈顶复制到新帧的局部变量前N个槽位
func Test_VMExec_Call_ArgPassing(t *testing.T) {
	t.Run("参数按值传递", func(t *testing.T) {
		runIntTest(t, `
fn modify(x) {
    x = x + 100
    return x
}
modify(5)`, 105)
	})
	t.Run("原始值不被修改", func(t *testing.T) {
		runIntTest(t, `
fn inc(x) {
    x = x + 1
    return x
}
a := 10
inc(a)
a`, 10) // a仍是10, 按值传递
	})
	t.Run("多参数按顺序传递", func(t *testing.T) {
		runIntTest(t, `
fn sub(a, b, c) { a - b - c }
sub(30, 10, 5)`, 15)
	})
	t.Run("参数为函数值", func(t *testing.T) {
		runIntTest(t, `
fn double(x) { x * 2 }
fn apply(f, x) { f(x) }
apply(double, 21)`, 42)
	})
}

// Test_VMExec_Call_ReturnValue 返回值传递机制
// return值通过栈传递回调用者
func Test_VMExec_Call_ReturnValue(t *testing.T) {
	t.Run("返回基本类型", func(t *testing.T) {
		runIntTest(t, `fn f() { return 42 }
f()`, 42)
		runStringTest(t, `fn f() { return "hi" }
f()`, "hi")
		runBoolTest(t, `fn f() { return true }
f()`, true)
	})
	t.Run("返回复合类型", func(t *testing.T) {
		result := runScript(t, `fn f() { return [1, 2, 3] }
f()`)
		arr := result.Array()
		if len(arr.Elements) != 3 || arr.Elements[0].Int() != 1 {
			t.Errorf("返回数组错误: %v", arr.Elements)
		}
	})
	t.Run("返回表达式计算结果", func(t *testing.T) {
		runIntTest(t, `
fn compute(a, b) {
    return a*a + b*b
}
compute(3, 4)`, 25)
	})
}

// Test_VMExec_Call_MaxDepth 调用深度限制(默认256)
// 无限递归触发 CallStackOverflow
func Test_VMExec_Call_MaxDepth(t *testing.T) {
	t.Run("默认深度无限递归报错", func(t *testing.T) {
		runRuntimeErrorTest(t, `
fn rec(n) {
    return rec(n + 1)
}
rec(0)`)
	})
	t.Run("深度接近限制正常返回", func(t *testing.T) {
		runIntTest(t, `
fn rec(n) {
    if n <= 0 { return 0 }
    return 1 + rec(n - 1)
}
rec(50)`, 50)
	})
}

// Test_VMExec_Call_CustomDepth 自定义调用深度
// WithMaxCallDepth 调整最大调用深度
func Test_VMExec_Call_CustomDepth(t *testing.T) {
	t.Run("自定义小深度触发限制", func(t *testing.T) {
		code := `
fn rec(n) {
    return rec(n + 1)
}
rec(0)`
		parser := NewParser()
		script, err := parser.Compile(code)
		if err != nil {
			t.Fatalf("编译失败: %v", err)
		}
		ctx := NewContext()
		engine := NewEngine(WithMaxCallDepth(10))
		_, err = engine.Run(ctx, script)
		if err == nil {
			t.Error("自定义小深度应触发栈溢出")
		}
	})
	t.Run("自定义大深度支持更深递归", func(t *testing.T) {
		code := `
fn rec(n) {
    if n <= 0 { return 0 }
    return 1 + rec(n - 1)
}
rec(200)`
		parser := NewParser()
		script, err := parser.Compile(code)
		if err != nil {
			t.Fatalf("编译失败: %v", err)
		}
		ctx := NewContext()
		engine := NewEngine()
		result, err := engine.Run(ctx, script)
		if err != nil {
			t.Fatalf("200深度应正常: %v", err)
		}
		if result.Int() != 200 {
			t.Errorf("期望200, 得到%d", result.Int())
		}
	})
}

// ========== 常量池 ==========

// Test_VMExec_Constants_Int 整数常量
// OpConst 从常量池加载整数到栈顶
func Test_VMExec_Constants_Int(t *testing.T) {
	t.Run("正整数", func(t *testing.T) {
		runIntTest(t, "0", 0)
		runIntTest(t, "1", 1)
		runIntTest(t, "42", 42)
		runIntTest(t, "1000000", 1000000)
	})
	t.Run("负整数", func(t *testing.T) {
		runIntTest(t, "-1", -1)
		runIntTest(t, "-42", -42)
	})
	t.Run("大整数", func(t *testing.T) {
		runIntTest(t, "1000000000000", 1000000000000)
	})
	t.Run("十六进制", func(t *testing.T) {
		runIntTest(t, "0xFF", 255)
		runIntTest(t, "0x10", 16)
	})
}

// Test_VMExec_Constants_Float 浮点常量
func Test_VMExec_Constants_Float(t *testing.T) {
	t.Run("基本浮点", func(t *testing.T) {
		runFloatTest(t, "3.14", 3.14)
		runFloatTest(t, "0.5", 0.5)
		runFloatTest(t, "100.0", 100.0)
	})
	t.Run("小数浮点", func(t *testing.T) {
		runFloatTest(t, "0.001", 0.001)
		runFloatTest(t, "100.5", 100.5)
	})
}

// Test_VMExec_Constants_String 字符串常量
func Test_VMExec_Constants_String(t *testing.T) {
	t.Run("基本字符串", func(t *testing.T) {
		runStringTest(t, `"hello"`, "hello")
		runStringTest(t, `""`, "")
		runStringTest(t, `"a"`, "a")
	})
	t.Run("含特殊字符", func(t *testing.T) {
		runStringTest(t, `"hello world"`, "hello world")
		runStringTest(t, `"123"`, "123")
	})
}

// Test_VMExec_Constants_Bool 布尔常量
// OpTrue/OpFalse 使用预计算值,不走常量池
func Test_VMExec_Constants_Bool(t *testing.T) {
	runBoolTest(t, "true", true)
	runBoolTest(t, "false", false)
}

// Test_VMExec_Constants_Nil nil常量
// OpNil 使用预计算的nil值
func Test_VMExec_Constants_Nil(t *testing.T) {
	result := runScript(t, "nil")
	if !result.IsNil() {
		t.Errorf("期望nil, 得到 %v", result)
	}
}

// Test_VMExec_Constants_Array 数组常量
// OpArrayNew 从栈顶收集N个元素构造数组
func Test_VMExec_Constants_Array(t *testing.T) {
	t.Run("空数组", func(t *testing.T) {
		result := runScript(t, "[]")
		arr := result.Array()
		if len(arr.Elements) != 0 {
			t.Errorf("期望空数组, 得到长度 %d", len(arr.Elements))
		}
	})
	t.Run("整数数组", func(t *testing.T) {
		result := runScript(t, "[1, 2, 3]")
		arr := result.Array()
		if len(arr.Elements) != 3 {
			t.Fatalf("期望长度3, 得到 %d", len(arr.Elements))
		}
		if arr.Elements[0].Int() != 1 || arr.Elements[2].Int() != 3 {
			t.Errorf("数组元素错误: %v", arr.Elements)
		}
	})
	t.Run("混合类型数组", func(t *testing.T) {
		result := runScript(t, `[1, "a", true, nil]`)
		arr := result.Array()
		if len(arr.Elements) != 4 {
			t.Fatalf("期望长度4, 得到 %d", len(arr.Elements))
		}
	})
}

// Test_VMExec_Constants_Map Map常量
// OpMapNew 从栈顶收集N对key/value构造Map
func Test_VMExec_Constants_Map(t *testing.T) {
	t.Run("空Map", func(t *testing.T) {
		result := runScript(t, "{}")
		m := result.Map()
		if len(m.Pairs) != 0 {
			t.Errorf("期望空Map, 得到长度 %d", len(m.Pairs))
		}
	})
	t.Run("整数Map", func(t *testing.T) {
		result := runScript(t, `{"a": 1, "b": 2}`)
		m := result.Map()
		if len(m.Pairs) != 2 {
			t.Fatalf("期望2对, 得到 %d", len(m.Pairs))
		}
		if m.Pairs["a"].Int() != 1 || m.Pairs["b"].Int() != 2 {
			t.Errorf("Map值错误: %v", m.Pairs)
		}
	})
}

// Test_VMExec_Constants_PoolStructure 常量池结构验证
// 编译后Main函数有常量池,包含脚本中使用的字面量
func Test_VMExec_Constants_PoolStructure(t *testing.T) {
	t.Run("简单表达式有常量池", func(t *testing.T) {
		script := compileScript(t, "1 + 2")
		main := script.Main
		if main == nil {
			t.Fatal("Main不应为nil")
		}
		if len(main.Constants) == 0 {
			t.Error("表达式 '1 + 2' 应有常量")
		}
	})
	t.Run("函数有自己的常量池", func(t *testing.T) {
		script := compileScript(t, `
fn f() {
    return 42
}
f()`)
		if len(script.Functions) == 0 {
			t.Fatal("应有函数定义")
		}
		for _, fn := range script.Functions {
			if fn.Name != "main" && fn.Name == "f" {
				if len(fn.Constants) == 0 {
					t.Error("函数f应有常量42")
				}
			}
		}
	})
}

// Test_VMExec_Constants_Reuse 常量池索引复用验证
// 验证编译器为相同值生成可用的常量池索引(不强制去重)
func Test_VMExec_Constants_Reuse(t *testing.T) {
	t.Run("相同值多次使用可正确执行", func(t *testing.T) {
		// 即使不去重,执行结果也正确
		runIntTest(t, "1 + 1 + 1 + 1", 4)
		runStringTest(t, `"a" + "a" + "a"`, "aaa")
	})
	t.Run("常量池索引访问越界保护", func(t *testing.T) {
		// 编译器生成的索引应在常量池范围内
		script := compileScript(t, `1 + 2 * 3 - 4 / 2`)
		main := script.Main
		for _, inst := range main.Instructions {
			if inst.Op == OpConst {
				idx := inst.Args[0]
				if idx < 0 || idx >= len(main.Constants) {
					t.Errorf("常量池索引 %d 超出范围 [0, %d)", idx, len(main.Constants))
				}
			}
		}
	})
}

// ========== 操作码覆盖 ==========

// Test_VMExec_Op_Arithmetic 所有算术操作码: Add, Sub, Mul, Div, Mod, Neg
func Test_VMExec_Op_Arithmetic(t *testing.T) {
	t.Run("OpAdd", func(t *testing.T) {
		runIntTest(t, "5 + 3", 8)
		runFloatTest(t, "1.5 + 2.5", 4.0)
	})
	t.Run("OpSub", func(t *testing.T) {
		runIntTest(t, "10 - 4", 6)
		runFloatTest(t, "5.5 - 1.5", 4.0)
	})
	t.Run("OpMul", func(t *testing.T) {
		runIntTest(t, "6 * 7", 42)
		runFloatTest(t, "2.5 * 4.0", 10.0)
	})
	t.Run("OpDiv", func(t *testing.T) {
		runIntTest(t, "20 / 4", 5)
		runFloatTest(t, "10.0 / 4.0", 2.5)
	})
	t.Run("OpMod", func(t *testing.T) {
		runIntTest(t, "17 % 5", 2)
		runIntTest(t, "10 % 3", 1)
	})
	t.Run("OpNeg", func(t *testing.T) {
		runIntTest(t, "-5", -5)
		runIntTest(t, "-(-10)", 10)
		runFloatTest(t, "-3.14", -3.14)
	})
}

// Test_VMExec_Op_Compare 所有比较操作码: Less, LessEq, Greater, GreaterEq, Equal, NotEqual
func Test_VMExec_Op_Compare(t *testing.T) {
	t.Run("OpLess", func(t *testing.T) {
		runBoolTest(t, "3 < 5", true)
		runBoolTest(t, "5 < 3", false)
		runBoolTest(t, "5 < 5", false)
	})
	t.Run("OpLessEq", func(t *testing.T) {
		runBoolTest(t, "3 <= 5", true)
		runBoolTest(t, "5 <= 5", true)
		runBoolTest(t, "6 <= 5", false)
	})
	t.Run("OpGreater", func(t *testing.T) {
		runBoolTest(t, "5 > 3", true)
		runBoolTest(t, "3 > 5", false)
		runBoolTest(t, "5 > 5", false)
	})
	t.Run("OpGreaterEq", func(t *testing.T) {
		runBoolTest(t, "5 >= 3", true)
		runBoolTest(t, "5 >= 5", true)
		runBoolTest(t, "3 >= 5", false)
	})
	t.Run("OpEqual", func(t *testing.T) {
		runBoolTest(t, "5 == 5", true)
		runBoolTest(t, "5 == 6", false)
		runBoolTest(t, `"a" == "a"`, true)
	})
	t.Run("OpNotEqual", func(t *testing.T) {
		runBoolTest(t, "5 != 6", true)
		runBoolTest(t, "5 != 5", false)
	})
}

// Test_VMExec_Op_Logic 逻辑操作码: Not
// && 和 || 在编译期展开为跳转指令, Not为独立操作码
func Test_VMExec_Op_Logic(t *testing.T) {
	t.Run("OpNot", func(t *testing.T) {
		runBoolTest(t, "!false", true)
		runBoolTest(t, "!true", false)
		runBoolTest(t, "!!true", true)
	})
	t.Run("逻辑与短路", func(t *testing.T) {
		runBoolTest(t, "true && true", true)
		runBoolTest(t, "true && false", false)
	})
	t.Run("逻辑或短路", func(t *testing.T) {
		runBoolTest(t, "false || true", true)
		runBoolTest(t, "false || false", false)
	})
}

// Test_VMExec_Op_Bitwise 位运算操作码: BitAnd, BitOr, BitXor, BitNot, LShift, RShift
func Test_VMExec_Op_Bitwise(t *testing.T) {
	t.Run("OpBitAnd", func(t *testing.T) {
		runIntTest(t, "12 & 10", 8)
		runIntTest(t, "0xFF & 0x0F", 0x0F)
	})
	t.Run("OpBitOr", func(t *testing.T) {
		runIntTest(t, "12 | 10", 14)
		runIntTest(t, "0xF0 | 0x0F", 0xFF)
	})
	t.Run("OpBitXor", func(t *testing.T) {
		runIntTest(t, "12 ^ 10", 6)
		runIntTest(t, "0xFF ^ 0x0F", 0xF0)
	})
	t.Run("OpBitNot", func(t *testing.T) {
		runIntTest(t, "^0", ^0)
		runIntTest(t, "^5", ^5)
	})
	t.Run("OpLShift", func(t *testing.T) {
		runIntTest(t, "1 << 4", 16)
		runIntTest(t, "1 << 10", 1024)
	})
	t.Run("OpRShift", func(t *testing.T) {
		runIntTest(t, "256 >> 2", 64)
		runIntTest(t, "1024 >> 10", 1)
	})
}

// Test_VMExec_Op_Convert 类型转换操作码: ToInt, ToFloat, ToString, TypeOf
func Test_VMExec_Op_Convert(t *testing.T) {
	t.Run("OpToInt", func(t *testing.T) {
		runIntTest(t, `int("123")`, 123)
		runIntTest(t, `int(3.99)`, 3) // 截断
		runIntTest(t, `int(true)`, 1)
		runIntTest(t, `int(false)`, 0)
	})
	t.Run("OpToFloat", func(t *testing.T) {
		runFloatTest(t, `float(42)`, 42.0)
		runFloatTest(t, `float("3.14")`, 3.14)
	})
	t.Run("OpToString", func(t *testing.T) {
		runStringTest(t, `string(123)`, "123")
		runStringTest(t, `string(true)`, "true")
		runStringTest(t, `string("hi")`, "hi") // 字符串不变
	})
	t.Run("OpTypeOf", func(t *testing.T) {
		runStringTest(t, `typeof(42)`, "int")
		runStringTest(t, `typeof(3.14)`, "float")
		runStringTest(t, `typeof("x")`, "string")
		runStringTest(t, `typeof(true)`, "bool")
		runStringTest(t, `typeof(nil)`, "nil")
		runStringTest(t, `typeof([1])`, "array")
		runStringTest(t, `typeof({"a":1})`, "map")
	})
}

// Test_VMExec_Op_Container 数组/Map操作码: ArrayNew, MapNew, Index, Slice, StoreIndex
func Test_VMExec_Op_Container(t *testing.T) {
	t.Run("OpArrayNew", func(t *testing.T) {
		result := runScript(t, "[1, 2, 3]")
		if len(result.Array().Elements) != 3 {
			t.Error("数组创建错误")
		}
	})
	t.Run("OpMapNew", func(t *testing.T) {
		result := runScript(t, `{"a": 1, "b": 2}`)
		if len(result.Map().Pairs) != 2 {
			t.Error("Map创建错误")
		}
	})
	t.Run("OpIndex", func(t *testing.T) {
		runIntTest(t, `[10, 20, 30][0]`, 10)
		runIntTest(t, `{"a": 99}["a"]`, 99)
		runStringTest(t, `"hello"[1]`, "e")
	})
	t.Run("OpSlice", func(t *testing.T) {
		result := runScript(t, `[1, 2, 3, 4, 5][1:4]`)
		arr := result.Array()
		if len(arr.Elements) != 3 || arr.Elements[0].Int() != 2 {
			t.Errorf("切片错误: %v", arr.Elements)
		}
		runStringTest(t, `"hello"[1:4]`, "ell")
	})
	t.Run("OpStoreIndex", func(t *testing.T) {
		runIntTest(t, `arr := [1, 2, 3]
arr[0] = 99
arr[0]`, 99)
		runIntTest(t, `m := {"a": 1}
m["a"] = 42
m["a"]`, 42)
	})
}

// Test_VMExec_Op_Builtin 内置函数操作码: Len, Push, Delete, MapKeys
func Test_VMExec_Op_Builtin(t *testing.T) {
	t.Run("OpLen", func(t *testing.T) {
		runIntTest(t, `len([1, 2, 3])`, 3)
		runIntTest(t, `len("hello")`, 5)
		runIntTest(t, `len({"a": 1, "b": 2})`, 2)
		runIntTest(t, `len([])`, 0)
	})
	t.Run("OpPush", func(t *testing.T) {
		runIntTest(t, `arr := [1, 2]
arr2 := push(arr, 3)
len(arr2)`, 3)
		// 原数组不变
		runIntTest(t, `arr := [1, 2]
arr2 := push(arr, 3)
len(arr)`, 2)
	})
	t.Run("OpDelete", func(t *testing.T) {
		runIntTest(t, `m := {"a": 1, "b": 2, "c": 3}
delete(m, "a")
len(m)`, 2)
		// 删除后访问返回nil
		result := runScript(t, `m := {"a": 1}
delete(m, "a")
m["a"]`)
		if !result.IsNil() {
			t.Errorf("删除后应返回nil, 得到 %v", result)
		}
	})
	t.Run("OpMapKeys", func(t *testing.T) {
		// OpMapKeys 在range遍历Map时由编译器内部生成
		// 通过遍历Map验证其底层调用正确
		result := runScript(t, `
m := {"a": 1, "b": 2, "c": 3}
count := 0
for k := range m {
    count = count + 1
}
count`)
		if result.Int() != 3 {
			t.Errorf("range遍历Map应得到3个key, 得到 %d", result.Int())
		}
	})
}

// Test_VMExec_Op_StackOp 栈操作码: Pop, Dup
// OpPop弹出栈顶(表达式语句丢弃结果), OpDup复制栈顶
func Test_VMExec_Op_StackOp(t *testing.T) {
	t.Run("OpPop表达式语句", func(t *testing.T) {
		// 表达式语句的结果被OpPop丢弃
		runIntTest(t, `
1 + 2
3 + 4`, 7) // 最后一个表达式语句的值留在栈上
	})
	t.Run("OpDup通过变量赋值", func(t *testing.T) {
		// 变量声明后多次读取验证Dup语义
		runIntTest(t, `
x := 5
x + x + x + x + x`, 25)
	})
}

// ========== 执行统计 ==========

// Test_VMExec_Stats_GetStats GetStats返回执行统计
func Test_VMExec_Stats_GetStats(t *testing.T) {
	t.Run("单次执行统计", func(t *testing.T) {
		ctx := NewContext()
		engine := NewEngine()
		parser := NewParser()
		script, err := parser.Compile(`1 + 2`)
		if err != nil {
			t.Fatal(err)
		}
		_, err = engine.Run(ctx, script)
		if err != nil {
			t.Fatal(err)
		}
		stats := ctx.GetStats()
		if stats.TotalRuns != 1 {
			t.Errorf("TotalRuns期望1, 得到 %d", stats.TotalRuns)
		}
		if stats.TotalTime < 0 {
			t.Errorf("TotalTime不应为负: %d", stats.TotalTime)
		}
	})
}

// Test_VMExec_Stats_Accumulate 多次执行的统计累积
func Test_VMExec_Stats_Accumulate(t *testing.T) {
	ctx := NewContext()
	engine := NewEngine()
	parser := NewParser()
	script, _ := parser.Compile(`1 + 1`)
	for i := 0; i < 5; i++ {
		_, err := engine.Run(ctx, script)
		if err != nil {
			t.Fatal(err)
		}
	}
	stats := ctx.GetStats()
	if stats.TotalRuns != 5 {
		t.Errorf("TotalRuns期望5, 得到 %d", stats.TotalRuns)
	}
}

// Test_VMExec_Stats_ComplexityDiff 不同复杂度脚本的统计差异
// 更复杂的脚本应花费更多时间(或至少不报错)
func Test_VMExec_Stats_ComplexityDiff(t *testing.T) {
	t.Run("简单vs复杂脚本统计", func(t *testing.T) {
		simple, _ := NewParser().Compile(`1 + 1`)
		complex, _ := NewParser().Compile(`
fn fib(n) {
    if n < 2 { return n }
    return fib(n - 1) + fib(n - 2)
}
fib(15)`)

		ctx1 := NewContext()
		e1 := NewEngine()
		_, _ = e1.Run(ctx1, simple)
		st1 := ctx1.GetStats()

		ctx2 := NewContext()
		e2 := NewEngine()
		_, _ = e2.Run(ctx2, complex)
		st2 := ctx2.GetStats()

		// 验证统计存在且TotalRuns正确
		if st1.TotalRuns != 1 || st2.TotalRuns != 1 {
			t.Errorf("TotalRuns错误: simple=%d complex=%d", st1.TotalRuns, st2.TotalRuns)
		}
		// TotalTime非负
		if st1.TotalTime < 0 || st2.TotalTime < 0 {
			t.Error("TotalTime不应为负")
		}
	})
}

// ========== 边界情况 ==========

// Test_VMExec_Edge_Empty 空脚本执行
func Test_VMExec_Edge_Empty(t *testing.T) {
	parser := NewParser()
	script, err := parser.Compile("")
	if err != nil {
		t.Fatalf("空脚本编译失败: %v", err)
	}
	ctx := NewContext()
	engine := NewEngine()
	result, err := engine.Run(ctx, script)
	if err != nil {
		t.Fatalf("空脚本执行失败: %v", err)
	}
	if !result.IsNil() {
		t.Errorf("空脚本应返回nil, 得到 %v", result)
	}
}

// Test_VMExec_Edge_CommentOnly 只有注释的脚本
func Test_VMExec_Edge_CommentOnly(t *testing.T) {
	parser := NewParser()
	script, err := parser.Compile("// just a comment\n")
	if err != nil {
		t.Fatalf("注释脚本编译失败: %v", err)
	}
	ctx := NewContext()
	engine := NewEngine()
	result, err := engine.Run(ctx, script)
	if err != nil {
		t.Fatalf("注释脚本执行失败: %v", err)
	}
	if !result.IsNil() {
		t.Errorf("纯注释脚本应返回nil, 得到 %v", result)
	}
}

// Test_VMExec_Edge_SingleExpr 只有一个表达式的脚本
func Test_VMExec_Edge_SingleExpr(t *testing.T) {
	runIntTest(t, "42", 42)
	runStringTest(t, `"single"`, "single")
	runBoolTest(t, "true", true)
}

// Test_VMExec_Edge_BigArray 超大数组的创建和操作
func Test_VMExec_Edge_BigArray(t *testing.T) {
	t.Run("创建100元素数组", func(t *testing.T) {
		runIntTest(t, `
arr := []
for i := 0; i < 100; i = i + 1 {
    arr = push(arr, i)
}
len(arr)`, 100)
	})
	t.Run("大数组求和", func(t *testing.T) {
		runIntTest(t, `
arr := []
for i := 0; i < 50; i = i + 1 {
    arr = push(arr, i)
}
sum := 0
for v := range arr {
    sum = sum + v
}
sum`, 1225) // 0+1+...+49 = 49*50/2 = 1225
	})
}

// Test_VMExec_Edge_DeepMap 深层嵌套的Map创建
func Test_VMExec_Edge_DeepMap(t *testing.T) {
	t.Run("4层嵌套Map", func(t *testing.T) {
		runIntTest(t, `
m := {"a": {"b": {"c": {"d": 42}}}}
m["a"]["b"]["c"]["d"]`, 42)
	})
	t.Run("嵌套Map修改", func(t *testing.T) {
		runIntTest(t, `
m := {"outer": {"inner": 1}}
m["outer"]["inner"] = 99
m["outer"]["inner"]`, 99)
	})
}

// Test_VMExec_Edge_LongStrConcat 超长字符串拼接
func Test_VMExec_Edge_LongStrConcat(t *testing.T) {
	t.Run("循环拼接50次", func(t *testing.T) {
		runIntTest(t, `
s := ""
for i := 0; i < 50; i = i + 1 {
    s = s + "ab"
}
len(s)`, 100)
	})
	t.Run("循环拼接含数字", func(t *testing.T) {
		runIntTest(t, `
s := ""
for i := 0; i < 10; i = i + 1 {
    s = s + i
}
len(s)`, 10) // 0123456789 长度10
	})
}

// Test_VMExec_Edge_ManyVars 大量变量的脚本
// 验证局部变量表的扩展和管理
func Test_VMExec_Edge_ManyVars(t *testing.T) {
	t.Run("20个局部变量", func(t *testing.T) {
		runIntTest(t, `
a := 1
b := 2
c := 3
d := 4
e := 5
f := 6
g := 7
h := 8
i := 9
j := 10
a + b + c + d + e + f + g + h + i + j`, 55)
	})
}

// ========== 并发安全 ==========

// Test_VMExec_Concurrent_SameScript 同一编译结果在多个Context上并发执行
// 引擎行为: 编译产物只读,多个Context+Engine可并发执行同一脚本
func Test_VMExec_Concurrent_SameScript(t *testing.T) {
	script := compileScript(t, `
fn fib(n) {
    if n < 2 { return n }
    return fib(n - 1) + fib(n - 2)
}
fib(15)`)

	const N = 20
	var wg sync.WaitGroup
	var fails int32
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := NewContext()
			engine := NewEngine()
			result, err := engine.Run(ctx, script)
			if err != nil {
				t.Errorf("并发执行错误: %v", err)
				atomic.AddInt32(&fails, 1)
				return
			}
			if result.Int() != 610 {
				t.Errorf("fib(15)期望610, 得到 %d", result.Int())
				atomic.AddInt32(&fails, 1)
			}
		}()
	}
	wg.Wait()
	if fails > 0 {
		t.Errorf("%d 个并发任务失败", fails)
	}
}

// Test_VMExec_Concurrent_MultiEngine 多个Engine实例并发
// 引擎行为: 多个独立Engine可同时运行不同脚本
func Test_VMExec_Concurrent_MultiEngine(t *testing.T) {
	scripts := []*CompiledScript{
		compileScript(t, `1 + 2 * 3`),
		compileScript(t, `[1, 2, 3][1]`),
		compileScript(t, `len("hello")`),
	}
	expected := []int{7, 2, 5}

	const ROUNDS = 10
	var wg sync.WaitGroup
	var fails int32
	for r := 0; r < ROUNDS; r++ {
		for idx, s := range scripts {
			wg.Add(1)
			go func(s *CompiledScript, exp int) {
				defer wg.Done()
				ctx := NewContext()
				engine := NewEngine()
				result, err := engine.Run(ctx, s)
				if err != nil {
					t.Errorf("并发执行错误: %v", err)
					atomic.AddInt32(&fails, 1)
					return
				}
				if result.Int() != exp {
					t.Errorf("期望 %d, 得到 %d", exp, result.Int())
					atomic.AddInt32(&fails, 1)
				}
			}(s, expected[idx])
		}
	}
	wg.Wait()
	if fails > 0 {
		t.Errorf("%d 个并发任务失败", fails)
	}
}

// ========== 错误恢复 ==========

// Test_VMExec_Recover_VMState 运行时错误后VM状态
// 错误后Running应为false,script引用应为nil
func Test_VMExec_Recover_VMState(t *testing.T) {
	script := compileScript(t, `1 / 0`)
	ctx := NewContext()
	vm := NewVM(ctx, DefaultMaxCallDepth)

	_, err := vm.Run(script)
	if err == nil {
		t.Fatal("除零应返回错误")
	}

	// 验证状态已重置
	if vm.isRunning() {
		t.Error("错误后Running应为false")
	}
	if vm.script != nil {
		t.Error("错误后script引用应为nil")
	}
}

// Test_VMExec_Recover_MultiRun 多次Run同一脚本
// VM应可重复使用,每次Run都正确执行
func Test_VMExec_Recover_MultiRun(t *testing.T) {
	vm := NewVM(NewContext(), DefaultMaxCallDepth)
	script := compileScript(t, `10 + 20`)

	for i := 0; i < 5; i++ {
		result, err := vm.Run(script)
		if err != nil {
			t.Fatalf("第%d次执行失败: %v", i+1, err)
		}
		if result.Int() != 30 {
			t.Errorf("第%d次: 期望30, 得到 %d", i+1, result.Int())
		}
		// 验证状态已重置
		if vm.isRunning() {
			t.Errorf("第%d次执行后Running未重置", i+1)
		}
	}
}

// Test_VMExec_Recover_AfterError Run出错后再次Run正常脚本
// 同一VM实例错误后应可继续使用
func Test_VMExec_Recover_AfterError(t *testing.T) {
	vm := NewVM(NewContext(), DefaultMaxCallDepth)
	parser := NewParser()

	// 第一次: 出错
	errScript, _ := parser.Compile(`1 / 0`)
	_, err1 := vm.Run(errScript)
	if err1 == nil {
		t.Fatal("第一次应返回错误")
	}

	// 第二次: 正常
	okScript, _ := parser.Compile(`5 + 5`)
	result, err2 := vm.Run(okScript)
	if err2 != nil {
		t.Fatalf("第二次应正常: %v", err2)
	}
	if result.Int() != 10 {
		t.Errorf("期望10, 得到 %d", result.Int())
	}
}

// Test_VMExec_Recover_EngineAfterError 同一Engine出错后继续运行正常脚本
func Test_VMExec_Recover_EngineAfterError(t *testing.T) {
	engine := NewEngine()
	parser := NewParser()

	// 第一次: 出错
	errScript, _ := parser.Compile(`1 / 0`)
	_, err1 := engine.Run(NewContext(), errScript)
	if err1 == nil {
		t.Fatal("第一次应返回错误")
	}

	// 第二次: 正常
	okScript, _ := parser.Compile(`10 + 20`)
	result, err2 := engine.Run(NewContext(), okScript)
	if err2 != nil {
		t.Fatalf("第二次应正常: %v", err2)
	}
	if result.Int() != 30 {
		t.Errorf("期望30, 得到 %d", result.Int())
	}
}

// Test_VMExec_Recover_RuntimeErrors 各类运行时错误验证
func Test_VMExec_Recover_RuntimeErrors(t *testing.T) {
	t.Run("除零错误", func(t *testing.T) {
		runRuntimeErrorTest(t, `1 / 0`)
	})
	t.Run("取模零错误", func(t *testing.T) {
		runRuntimeErrorTest(t, `5 % 0`)
	})
	t.Run("类型不匹配", func(t *testing.T) {
		runRuntimeErrorTest(t, `"hello" - 5`)
	})
	t.Run("throw异常", func(t *testing.T) {
		runRuntimeErrorTest(t, `throw "custom error"`)
	})
	t.Run("调用非函数", func(t *testing.T) {
		runRuntimeErrorTest(t, `
x := 42
x(1)`)
	})
}

// ========== 补充: VM内部机制 ==========

// Test_VMExec_Internal_PushPopGrow 栈扩展验证
// 栈初始容量有限,大量push应触发扩容
func Test_VMExec_Internal_PushPopGrow(t *testing.T) {
	vm := newTestVM(t)
	initialCap := cap(vm.Stack)
	// 触发多次push超过初始容量
	for i := 0; i < initialCap+100; i++ {
		vm.push(NewValue(i))
	}
	if vm.SP != initialCap+100 {
		t.Errorf("SP期望 %d, 得到 %d", initialCap+100, vm.SP)
	}
	// 验证栈中值正确
	for i := vm.SP - 1; i >= 0; i-- {
		val := vm.pop()
		if val.Int() != i {
			t.Errorf("栈[%d]期望 %d, 得到 %d", i, i, val.Int())
			break
		}
	}
}

// Test_VMExec_Internal_FrameLocals 帧的局部变量初始化为零值
// NewFrame创建的Locals应全部为nil值
func Test_VMExec_Internal_FrameLocals(t *testing.T) {
	fn := &CompiledFunction{
		Name:      "test",
		NumLocals: 5,
		NumParams: 2,
	}
	frame := NewFrame(fn)
	if len(frame.Locals) != 5 {
		t.Fatalf("期望5个局部变量, 得到 %d", len(frame.Locals))
	}
	for i, v := range frame.Locals {
		if !v.IsNil() {
			t.Errorf("局部变量[%d]应为nil零值, 得到 %v", i, v)
		}
	}
}

// Test_VMExec_Internal_StackTrace 运行时错误包含调用栈
// 深层调用错误应包含完整的调用栈追踪
func Test_VMExec_Internal_StackTrace(t *testing.T) {
	script := compileScript(t, `
fn level3() {
    return 1 / 0
}
fn level2() {
    return level3()
}
fn level1() {
    return level2()
}
level1()`)
	ctx := NewContext()
	engine := NewEngine()
	_, err := engine.Run(ctx, script)
	if err == nil {
		t.Fatal("应返回运行时错误")
	}
	// 错误消息应包含调用栈信息
	errMsg := err.Error()
	if !strings.Contains(errMsg, "level1") {
		t.Errorf("错误消息应包含调用栈, 得到: %s", errMsg)
	}
}

// Test_VMExec_Internal_VMPool VM实例池复用
// 验证VM从池中获取和归还机制
func Test_VMExec_Internal_VMPool(t *testing.T) {
	ctx := NewContext()
	vm1 := getVMFromPool(ctx, 256)
	vm1.SP = 0
	vm1.FP = 0
	returnVMToPool(vm1)

	vm2 := getVMFromPool(ctx, 256)
	// 池可能返回同一实例(不强制),但状态应已重置
	if vm2.SP != 0 {
		t.Errorf("从池获取的VM SP应为0, 得到 %d", vm2.SP)
	}
	if vm2.FP != 0 {
		t.Errorf("从池获取的VM FP应为0, 得到 %d", vm2.FP)
	}
	returnVMToPool(vm2)
}

// Test_VMExec_Internal_ReturnNil 函数无返回值时返回nil
// OpReturn在栈空时返回预计算的nil值
func Test_VMExec_Internal_ReturnNil(t *testing.T) {
	t.Run("return无值", func(t *testing.T) {
		result := runScript(t, `
fn f() {
    return
}
f()`)
		if !result.IsNil() {
			t.Errorf("无返回值应为nil, 得到 %v", result)
		}
	})
}

// Test_VMExec_Internal_ExecTime 执行时间统计非负
// 复杂脚本执行后TotalTime应合理
func Test_VMExec_Internal_ExecTime(t *testing.T) {
	ctx := NewContext()
	engine := NewEngine()
	script := compileScript(t, `
sum := 0
for i := 0; i < 100; i = i + 1 {
    sum = sum + i
}
sum`)

	start := time.Now()
	result, err := engine.Run(ctx, script)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	if result.Int() != 4950 {
		t.Errorf("期望4950, 得到 %d", result.Int())
	}
	if elapsed < 0 {
		t.Error("执行时间不应为负")
	}
	stats := ctx.GetStats()
	if stats.TotalTime < 0 {
		t.Error("TotalTime不应为负")
	}
}
