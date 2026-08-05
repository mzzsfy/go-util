package script

import (
	"testing"
)

// ========== 变量作用域与函数语义测试 ==========
// 参照 goja/otto(JavaScript引擎) 和 gopher-lua 的测试模式
// 覆盖函数返回值语义、参数传递、变量赋值、作用域隔离、递归、高阶函数

// ========== 1. 函数返回值语义 (参照 goja) ==========

// Test_Func_Semantics_NoReturnEmptyBody 空函数体返回nil
func Test_Func_Semantics_NoReturnEmptyBody(t *testing.T) {
	t.Run("空函数体返回nil", func(t *testing.T) {
		result := runScript(t, `fn f() {} f()`)
		assertNil(t, result)
	})
}

// Test_Func_Semantics_ReturnNoValue return无值返回nil
func Test_Func_Semantics_ReturnNoValue(t *testing.T) {
	t.Run("return无值返回nil", func(t *testing.T) {
		result := runScript(t, `
fn f() { return }
f()
`)
		assertNil(t, result)
	})
	t.Run("条件分支中return无值返回nil", func(t *testing.T) {
		result := runScript(t, `
fn f(n) {
	if n > 0 {
		return
	}
	return 1
}
f(1)
`)
		assertNil(t, result)
	})
}

// Test_Func_Semantics_ReturnZero return 0 返回int类型的0而非nil
func Test_Func_Semantics_ReturnZero(t *testing.T) {
	t.Run("return 0 是int类型非nil", func(t *testing.T) {
		result := runScript(t, `
fn f() { return 0 }
f()
`)
		if result.IsNil() {
			t.Errorf("return 0 不应是nil")
		}
		assertInt(t, result, 0)
	})
}

// Test_Func_Semantics_ReturnEmptyString return空字符串返回string类型
func Test_Func_Semantics_ReturnEmptyString(t *testing.T) {
	t.Run("return空字符串是string类型非nil", func(t *testing.T) {
		result := runScript(t, `
fn f() { return "" }
f()
`)
		if result.IsNil() {
			t.Errorf("return \"\" 不应是nil")
		}
		assertString(t, result, "")
	})
}

// Test_Func_Semantics_ImplicitReturn 函数隐式返回最后一个表达式的值
func Test_Func_Semantics_ImplicitReturn(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"最后表达式为隐式返回", `fn f(a, b) { a + b } f(3, 4)`, 7},
		{"多条语句隐式返回最后值", `fn f() { 1 2 3 } f()`, 3},
		{"赋值后隐式返回", `fn f() { x := 10 x * 2 } f()`, 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// Test_Func_Semantics_MultiReturnPaths 多return路径不同条件触发不同返回值
func Test_Func_Semantics_MultiReturnPaths(t *testing.T) {
	t.Run("分数等级多路径返回", func(t *testing.T) {
		result := runScript(t, `
fn grade(score) {
	if score >= 90 {
		return 1
	}
	if score >= 80 {
		return 2
	}
	return 3
}
grade(85)
`)
		assertInt(t, result, 2)
	})
	t.Run("不同分数不同等级", func(t *testing.T) {
		scripts := []struct {
			input    string
			expected int
		}{
			{`fn grade(score) { if score >= 90 { return 1 } if score >= 80 { return 2 } return 3 } grade(95)`, 1},
			{`fn grade(score) { if score >= 90 { return 1 } if score >= 80 { return 2 } return 3 } grade(85)`, 2},
			{`fn grade(score) { if score >= 90 { return 1 } if score >= 80 { return 2 } return 3 } grade(50)`, 3},
		}
		for _, tt := range scripts {
			runIntTest(t, tt.input, tt.expected)
		}
	})
}

// Test_Func_Semantics_ReturnPropagation 嵌套函数调用链中的return传播
func Test_Func_Semantics_ReturnPropagation(t *testing.T) {
	t.Run("三层调用链return传播", func(t *testing.T) {
		result := runScript(t, `
fn base() { return 42 }
fn mid() { return base() }
fn top() { return mid() }
top()
`)
		assertInt(t, result, 42)
	})
	t.Run("调用链中计算传播", func(t *testing.T) {
		result := runScript(t, `
fn inc(x) { return x + 1 }
fn dbl(x) { return x * 2 }
inc(dbl(inc(3)))
`)
		assertInt(t, result, 9)
	})
}

// ========== 2. 函数参数传递 (参照 gopher-lua) ==========

// Test_Func_Semantics_ParamByValue 按值传递: int参数修改不影响外部
func Test_Func_Semantics_ParamByValue(t *testing.T) {
	t.Run("int参数修改不影响外部变量", func(t *testing.T) {
		result := runScript(t, `
fn modify(n) {
	n = 999
	return n
}
x := 5
modify(x)
x
`)
		assertInt(t, result, 5)
	})
	t.Run("参数在函数内重新赋值后返回新值", func(t *testing.T) {
		result := runScript(t, `
fn reassign(a) {
	a = a + 10
	return a
}
reassign(5)
`)
		assertInt(t, result, 15)
	})
}

// Test_Func_Semantics_MultiParamOrder 多参数传递顺序正确
func Test_Func_Semantics_MultiParamOrder(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"两参数顺序", `fn sub(a, b) { return a - b } sub(10, 3)`, 7},
		{"三参数顺序", `fn f(a, b, c) { return a * 100 + b * 10 + c } f(1, 2, 3)`, 123},
		{"四参数顺序", `fn f(a, b, c, d) { return a + b + c + d } f(10, 20, 30, 40)`, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// Test_Func_Semantics_ParamNameNoConflict 参数名不与外部变量冲突(无闭包)
func Test_Func_Semantics_ParamNameNoConflict(t *testing.T) {
	t.Run("参数名与外部变量同名不冲突", func(t *testing.T) {
		// 无闭包: 函数内x是局部参数, 不引用外部x
		result := runScript(t, `
fn f(x) {
	return x * 2
}
x := 100
f(5) + x
`)
		assertInt(t, result, 110)
	})
	t.Run("不同函数同名参数互不干扰", func(t *testing.T) {
		result := runScript(t, `
fn f1(x) { return x + 1 }
fn f2(x) { return x * 2 }
f1(10) + f2(10)
`)
		assertInt(t, result, 31)
	})
}

// Test_Func_Semantics_FewerArgsError 参数少于定义时报错
func Test_Func_Semantics_FewerArgsError(t *testing.T) {
	t.Run("调用时少传参数报错", func(t *testing.T) {
		runRuntimeErrorTest(t, `
fn add(a, b) { return a + b }
add(1)
`)
	})
}

// Test_Func_Semantics_MoreArgsError 参数多于定义时报错
func Test_Func_Semantics_MoreArgsError(t *testing.T) {
	t.Run("调用时多传参数报错", func(t *testing.T) {
		runRuntimeErrorTest(t, `
fn f(a) { return a }
f(1, 2, 3)
`)
	})
}

// Test_Func_Semantics_FuncCallFuncParam 函数调用函数传参
func Test_Func_Semantics_FuncCallFuncParam(t *testing.T) {
	t.Run("函数返回值作为另一函数参数", func(t *testing.T) {
		result := runScript(t, `
fn dbl(x) { return x * 2 }
fn add(a, b) { return a + b }
add(dbl(3), dbl(4))
`)
		assertInt(t, result, 14)
	})
	t.Run("多层函数调用作为参数", func(t *testing.T) {
		result := runScript(t, `
fn inc(x) { return x + 1 }
fn add(a, b) { return a + b }
add(inc(inc(1)), inc(inc(2)))
`)
		assertInt(t, result, 7)
	})
}

// Test_Func_Semantics_StringParam 字符串参数传递
func Test_Func_Semantics_StringParam(t *testing.T) {
	t.Run("字符串参数正确传递", func(t *testing.T) {
		result := runScript(t, `
fn greet(name) {
	return name
}
greet("world")
`)
		assertString(t, result, "world")
	})
}

// ========== 3. 变量赋值与重新赋值 (参照 goja) ==========

// Test_Var_Reassign 同一变量多次赋值
func Test_Var_Reassign(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"两次赋值取最后值", `x := 1 x = 2 x = 3 x`, 3},
		{"赋值后参与运算再赋值", `x := 10 x = x + 5 x = x * 2 x`, 30},
		{"多次赋值链式", `a := 1 b := 2 a = b b = a a`, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// Test_Var_AssignAndUse 赋值后立即使用
func Test_Var_AssignAndUse(t *testing.T) {
	t.Run("赋值后立即参与运算", func(t *testing.T) {
		result := runScript(t, `
x := 5
y := x + 3
z := y * 2
z
`)
		assertInt(t, result, 16)
	})
	t.Run("连续依赖赋值", func(t *testing.T) {
		result := runScript(t, `
a := 1
b := a + 1
c := b + 1
d := c + 1
d
`)
		assertInt(t, result, 4)
	})
}

// Test_Var_Swap 通过临时变量交换两个变量
func Test_Var_Swap(t *testing.T) {
	t.Run("temp交换后a持有原b值", func(t *testing.T) {
		result := runScript(t, `
a := 1
b := 2
temp := a
a = b
b = temp
a
`)
		assertInt(t, result, 2)
	})
	t.Run("temp交换后b持有原a值", func(t *testing.T) {
		result := runScript(t, `
a := 11
b := 22
temp := a
a = b
b = temp
b
`)
		assertInt(t, result, 11)
	})
	t.Run("交换字符串", func(t *testing.T) {
		result := runScript(t, `
a := "hello"
b := "world"
temp := a
a = b
b = temp
a
`)
		assertString(t, result, "world")
	})
}

// Test_Var_DifferentTypeAssign 变量赋值不同类型值
func Test_Var_DifferentTypeAssign(t *testing.T) {
	t.Run("int赋值为string", func(t *testing.T) {
		result := runScript(t, `x := 42 x = "hello" x`)
		assertString(t, result, "hello")
	})
	t.Run("string赋值为int", func(t *testing.T) {
		result := runScript(t, `x := "abc" x = 99 x`)
		assertInt(t, result, 99)
	})
	t.Run("int赋值为bool", func(t *testing.T) {
		result := runScript(t, `x := 1 x = true x`)
		assertBool(t, result, true)
	})
}

// ========== 4. 作用域隔离 (参照 gopher-lua) ==========

// Test_Scope_FuncLocalIsolation 函数内声明的变量不影响外部
func Test_Scope_FuncLocalIsolation(t *testing.T) {
	t.Run("函数内局部变量不泄漏到外部", func(t *testing.T) {
		result := runScript(t, `
x := 10
fn f() {
	x := 20
	return x
}
f()
x
`)
		assertInt(t, result, 10)
	})
	t.Run("函数内声明的变量对外部不可见", func(t *testing.T) {
		result := runScript(t, `
fn f() {
	inner := 42
	return inner
}
f()
`)
		assertInt(t, result, 42)
	})
}

// Test_Scope_ForBlockVar for循环体内声明的变量
func Test_Scope_ForBlockVar(t *testing.T) {
	t.Run("for循环变量在循环体内可用", func(t *testing.T) {
		result := runScript(t, `
sum := 0
for i := 0; i < 5; i = i + 1 {
	sum = sum + i
}
sum
`)
		assertInt(t, result, 10)
	})
	t.Run("for循环条件变量累加", func(t *testing.T) {
		result := runScript(t, `
count := 0
for i := 1; i <= 3; i = i + 1 {
	count = count + 1
}
count
`)
		assertInt(t, result, 3)
	})
}

// Test_Scope_IfBlockVar if块内声明的变量
func Test_Scope_IfBlockVar(t *testing.T) {
	t.Run("if块内声明并使用变量", func(t *testing.T) {
		result := runScript(t, `
x := 5
if x > 0 {
	y := x * 2
	y
} else {
	0
}
`)
		assertInt(t, result, 10)
	})
	t.Run("if-else不同分支声明不同变量", func(t *testing.T) {
		result := runScript(t, `
x := 3
if x > 5 {
	big := 100
	big
} else {
	small := 10
	small
}
`)
		assertInt(t, result, 10)
	})
}

// Test_Scope_SameNameDiffFunc 同名变量在不同函数中互不干扰
func Test_Scope_SameNameDiffFunc(t *testing.T) {
	t.Run("两个函数使用同名局部变量", func(t *testing.T) {
		result := runScript(t, `
fn f1() {
	x := 100
	return x
}
fn f2() {
	x := 200
	return x
}
f1() + f2()
`)
		assertInt(t, result, 300)
	})
	t.Run("三个函数同名变量独立", func(t *testing.T) {
		result := runScript(t, `
fn fa() { v := 1 return v }
fn fb() { v := 2 return v }
fn fc() { v := 3 return v }
fa() * 100 + fb() * 10 + fc()
`)
		assertInt(t, result, 123)
	})
}

// Test_Scope_NoClosure 无闭包: 函数不能访问外部作用域变量
func Test_Scope_NoClosure(t *testing.T) {
	t.Run("函数不能引用外部变量", func(t *testing.T) {
		// 编译错误: 未定义的变量
		runErrorTest(t, `
x := 10
fn f() {
	return x
}
`)
	})
	t.Run("函数只能访问自身参数和局部变量", func(t *testing.T) {
		result := runScript(t, `
fn f(x) {
	y := x * 2
	return y
}
f(21)
`)
		assertInt(t, result, 42)
	})
}

// Test_Scope_ParamIsolation 函数参数修改不影响调用方变量
func Test_Scope_ParamIsolation(t *testing.T) {
	t.Run("参数在函数内修改不传播", func(t *testing.T) {
		result := runScript(t, `
fn reset(n) {
	n = 0
	return n
}
val := 77
reset(val)
val
`)
		assertInt(t, result, 77)
	})
	t.Run("多个参数各自独立修改", func(t *testing.T) {
		result := runScript(t, `
fn swap_local(a, b) {
	tmp := a
	a = b
	b = tmp
	return a * 100 + b
}
swap_local(1, 2)
`)
		assertInt(t, result, 201)
	})
}

// ========== 5. 递归模式 (参照主流引擎) ==========

// Test_Recursion_Fibonacci 斐波那契递归
func Test_Recursion_Fibonacci(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"fib(0)=0", `fn fib(n) { if n < 2 { return n } return fib(n - 1) + fib(n - 2) } fib(0)`, 0},
		{"fib(1)=1", `fn fib(n) { if n < 2 { return n } return fib(n - 1) + fib(n - 2) } fib(1)`, 1},
		{"fib(5)=5", `fn fib(n) { if n < 2 { return n } return fib(n - 1) + fib(n - 2) } fib(5)`, 5},
		{"fib(10)=55", `fn fib(n) { if n < 2 { return n } return fib(n - 1) + fib(n - 2) } fib(10)`, 55},
		{"fib(15)=610", `fn fib(n) { if n < 2 { return n } return fib(n - 1) + fib(n - 2) } fib(15)`, 610},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// Test_Recursion_Factorial 阶乘递归
func Test_Recursion_Factorial(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"0!=1", `fn fact(n) { if n <= 1 { return 1 } return n * fact(n - 1) } fact(0)`, 1},
		{"1!=1", `fn fact(n) { if n <= 1 { return 1 } return n * fact(n - 1) } fact(1)`, 1},
		{"5!=120", `fn fact(n) { if n <= 1 { return 1 } return n * fact(n - 1) } fact(5)`, 120},
		{"10!=3628800", `fn fact(n) { if n <= 1 { return 1 } return n * fact(n - 1) } fact(10)`, 3628800},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runIntTest(t, tt.input, tt.expected)
		})
	}
}

// Test_Recursion_Deep 深层递归测试深度
func Test_Recursion_Deep(t *testing.T) {
	t.Run("递归深度100", func(t *testing.T) {
		result := runScript(t, `
fn count(n) {
	if n <= 0 { return 0 }
	return 1 + count(n - 1)
}
count(100)
`)
		assertInt(t, result, 100)
	})
	t.Run("递归累加1到50", func(t *testing.T) {
		result := runScript(t, `
fn sum(n) {
	if n <= 0 { return 0 }
	return n + sum(n - 1)
}
sum(50)
`)
		assertInt(t, result, 1275)
	})
	t.Run("递归深度200", func(t *testing.T) {
		result := runScript(t, `
fn depth(n) {
	if n <= 0 { return 0 }
	return 1 + depth(n - 1)
}
depth(200)
`)
		assertInt(t, result, 200)
	})
}

// Test_Recursion_ReturnPropagation 递归中的return值传播
func Test_Recursion_ReturnPropagation(t *testing.T) {
	t.Run("递归乘法传播", func(t *testing.T) {
		result := runScript(t, `
fn power(base, exp) {
	if exp <= 0 { return 1 }
	return base * power(base, exp - 1)
}
power(2, 10)
`)
		assertInt(t, result, 1024)
	})
	t.Run("递归求最大公约数", func(t *testing.T) {
		result := runScript(t, `
fn gcd(a, b) {
	if b == 0 { return a }
	return gcd(b, a - (a / b) * b)
}
gcd(48, 18)
`)
		assertInt(t, result, 6)
	})
}

// ========== 6. 高阶函数模式 (参照 goja) ==========

// Test_HOF_FuncAsParam 函数作为参数传递
func Test_HOF_FuncAsParam(t *testing.T) {
	t.Run("传入不同函数得到不同结果", func(t *testing.T) {
		result := runScript(t, `
fn dbl(x) { return x * 2 }
fn apply(f, x) { return f(x) }
apply(dbl, 21)
`)
		assertInt(t, result, 42)
	})
	t.Run("同一apply函数传入不同逻辑", func(t *testing.T) {
		result := runScript(t, `
fn inc(x) { return x + 1 }
fn dec(x) { return x - 1 }
fn apply(f, x) { return f(x) }
apply(inc, 10) + apply(dec, 10)
`)
		assertInt(t, result, 20)
	})
}

// Test_HOF_FuncReturnAsArg 函数返回值作为另一个函数的参数
func Test_HOF_FuncReturnAsArg(t *testing.T) {
	t.Run("函数返回值直接传递", func(t *testing.T) {
		result := runScript(t, `
fn gen() { return 5 }
fn dbl(x) { return x * 2 }
dbl(gen())
`)
		assertInt(t, result, 10)
	})
	t.Run("多层嵌套函数返回值传递", func(t *testing.T) {
		result := runScript(t, `
fn a() { return 3 }
fn b(x) { return x + 1 }
fn c(x) { return x * 2 }
c(b(a()))
`)
		assertInt(t, result, 8)
	})
}

// Test_HOF_NestedCallChain 函数调用链 f(g(h(x)))
func Test_HOF_NestedCallChain(t *testing.T) {
	t.Run("三层调用链", func(t *testing.T) {
		result := runScript(t, `
fn h(x) { return x + 1 }
fn g(x) { return h(x) + 1 }
fn f(x) { return g(x) + 1 }
f(0)
`)
		assertInt(t, result, 3)
	})
	t.Run("四层调用链", func(t *testing.T) {
		result := runScript(t, `
fn d(x) { return x + 1 }
fn c(x) { return d(x) + 1 }
fn b(x) { return c(x) + 1 }
fn a(x) { return b(x) + 1 }
a(0)
`)
		assertInt(t, result, 4)
	})
}

// Test_HOF_Composition 函数组合模式
func Test_HOF_Composition(t *testing.T) {
	t.Run("通过apply组合两个函数", func(t *testing.T) {
		result := runScript(t, `
fn dbl(x) { return x * 2 }
fn inc(x) { return x + 1 }
fn apply(f, x) { return f(x) }
apply(dbl, apply(inc, 5))
`)
		assertInt(t, result, 12)
	})
	t.Run("三函数组合", func(t *testing.T) {
		result := runScript(t, `
fn add1(x) { return x + 1 }
fn mul2(x) { return x * 2 }
fn sub3(x) { return x - 3 }
fn apply(f, x) { return f(x) }
apply(sub3, apply(mul2, apply(add1, 10)))
`)
		assertInt(t, result, 19)
	})
}

// Test_HOF_Pipeline 通过循环实现管道式处理
func Test_HOF_Pipeline(t *testing.T) {
	t.Run("循环中调用函数处理累积值", func(t *testing.T) {
		result := runScript(t, `
fn step(x) { return x * 3 + 1 }
val := 1
for i := 0; i < 4; i = i + 1 {
	val = step(val)
}
val
`)
		// val: 1 -> 4 -> 13 -> 40 -> 121
		assertInt(t, result, 121)
	})
}

// Test_HOF_Accumulator 函数作为累加器
func Test_HOF_Accumulator(t *testing.T) {
	t.Run("函数在循环中作为映射使用", func(t *testing.T) {
		result := runScript(t, `
fn transform(x) { return x * x }
sum := 0
for i := 1; i <= 4; i = i + 1 {
	sum = sum + transform(i)
}
sum
`)
		// 1 + 4 + 9 + 16 = 30
		assertInt(t, result, 30)
	})
}
