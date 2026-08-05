package script

import (
	"reflect"
	"testing"
)

// ========== 函数和递归高级模式测试 ==========
// 覆盖函数定义/调用、函数体语法、参数传递语义、递归、作用域、
// 返回值类型、控制流、高阶函数、外部函数(#fn)、实用模式

// ========== 1. 函数定义和调用 ==========

func Test_FuncAdvanced_DefineAndCall(t *testing.T) {
	t.Run("无参数函数", func(t *testing.T) {
		runIntTest(t, `fn f() { return 42 } f()`, 42)
	})

	t.Run("单参数函数", func(t *testing.T) {
		runIntTest(t, `fn f(x) { return x * 2 } f(21)`, 42)
	})

	t.Run("多参数函数", func(t *testing.T) {
		runIntTest(t, `fn f(a, b, c) { return a + b + c } f(10, 20, 30)`, 60)
	})

	t.Run("函数返回常量", func(t *testing.T) {
		runIntTest(t, `fn f() { return 100 } f()`, 100)
	})

	t.Run("函数返回表达式", func(t *testing.T) {
		runIntTest(t, `fn f(a, b) { return a * b + 10 } f(3, 5)`, 25)
	})

	t.Run("函数返回另一个函数调用的结果", func(t *testing.T) {
		runIntTest(t, `
fn base() { return 7 }
fn wrapper() { return base() }
wrapper()
`, 7)
	})
}

// ========== 2. 函数体语法 ==========

func Test_FuncAdvanced_BodySyntax(t *testing.T) {
	t.Run("箭头函数语法不支持_编译错误", func(t *testing.T) {
		// => 是返回类型注解, 不是箭头函数体
		runErrorTest(t, `fn f(x) => x + 1`)
	})

	t.Run("块函数多语句", func(t *testing.T) {
		runIntTest(t, `
fn f(a, b) {
	x := a + b
	y := x * 2
	return y
}
f(3, 4)
`, 14)
	})

	t.Run("块函数带条件返回", func(t *testing.T) {
		runIntTest(t, `
fn f(n) {
	if n > 0 { return 1 }
	if n < 0 { return -1 }
	return 0
}
f(5)
`, 1)
	})

	t.Run("空函数体返回nil", func(t *testing.T) {
		result := runScript(t, `fn f() {} f()`)
		assertNil(t, result)
	})

	t.Run("块函数隐式返回最后表达式", func(t *testing.T) {
		runIntTest(t, `fn f(a, b) { a + b } f(3, 4)`, 7)
	})
}

// ========== 3. 参数传递语义 ==========

func Test_FuncAdvanced_ParamSemantics(t *testing.T) {
	t.Run("值传递_int参数修改不影响外部", func(t *testing.T) {
		runIntTest(t, `
fn modify(n) { n = 999 return n }
x := 5
modify(x)
x
`, 5)
	})

	t.Run("数组参数引用语义_函数内修改影响外部", func(t *testing.T) {
		runIntTest(t, `
fn modify(arr) { arr[0] = 999 return arr[0] }
a := [1, 2, 3]
modify(a)
a[0]
`, 999)
	})

	t.Run("Map参数引用语义_函数内修改影响外部", func(t *testing.T) {
		runIntTest(t, `
fn modify(m) { m["x"] = 999 return m["x"] }
d := {"x": 1}
modify(d)
d["x"]
`, 999)
	})

	t.Run("参数是表达式", func(t *testing.T) {
		runIntTest(t, `fn f(x) { return x * 2 } f(3 + 4 * 2)`, 22)
	})

	t.Run("参数是函数调用", func(t *testing.T) {
		runIntTest(t, `
fn g(x) { return x * 3 }
fn f(x) { return x + 1 }
f(g(4))
`, 13)
	})

	t.Run("参数是复杂嵌套表达式", func(t *testing.T) {
		runIntTest(t, `
fn dbl(x) { return x * 2 }
fn add(a, b) { return a + b }
add(dbl(3), dbl(add(1, 2)))
`, 12)
	})
}

// ========== 4. 递归 ==========

func Test_FuncAdvanced_Recursion(t *testing.T) {
	t.Run("阶乘_0的阶乘", func(t *testing.T) {
		runIntTest(t, `
fn fact(n) { if n <= 1 { return 1 } return n * fact(n - 1) }
fact(0)
`, 1)
	})

	t.Run("阶乘_5的阶乘", func(t *testing.T) {
		runIntTest(t, `
fn fact(n) { if n <= 1 { return 1 } return n * fact(n - 1) }
fact(5)
`, 120)
	})

	t.Run("阶乘_10的阶乘", func(t *testing.T) {
		runIntTest(t, `
fn fact(n) { if n <= 1 { return 1 } return n * fact(n - 1) }
fact(10)
`, 3628800)
	})

	t.Run("斐波那契fib_10", func(t *testing.T) {
		runIntTest(t, `
fn fib(n) { if n < 2 { return n } return fib(n - 1) + fib(n - 2) }
fib(10)
`, 55)
	})

	t.Run("斐波那契fib_15", func(t *testing.T) {
		runIntTest(t, `
fn fib(n) { if n < 2 { return n } return fib(n - 1) + fib(n - 2) }
fib(15)
`, 610)
	})

	t.Run("递归累加1到100", func(t *testing.T) {
		runIntTest(t, `
fn sum(n) { if n <= 0 { return 0 } return n + sum(n - 1) }
sum(100)
`, 5050)
	})

	t.Run("递归幂运算", func(t *testing.T) {
		runIntTest(t, `
fn power(base, exp) { if exp <= 0 { return 1 } return base * power(base, exp - 1) }
power(2, 10)
`, 1024)
	})

	t.Run("递归最大公约数", func(t *testing.T) {
		runIntTest(t, `
fn gcd(a, b) { if b == 0 { return a } return gcd(b, a - (a / b) * b) }
gcd(48, 18)
`, 6)
	})

	t.Run("深度递归_接近限制200层", func(t *testing.T) {
		runIntTest(t, `
fn depth(n) { if n <= 0 { return 0 } return 1 + depth(n - 1) }
depth(200)
`, 200)
	})

	t.Run("深度递归_超过限制报运行时错误", func(t *testing.T) {
		// 默认最大调用栈深度256, 500层递归应报错
		runRuntimeErrorTest(t, `
fn depth(n) { if n <= 0 { return 0 } return 1 + depth(n - 1) }
depth(500)
`)
	})

	t.Run("尾递归_引擎不优化但仍正确", func(t *testing.T) {
		// 尾递归形式, 引擎不做尾调用优化, 深度受调用栈限制
		runIntTest(t, `
fn helper(n, acc) { if n <= 0 { return acc } return helper(n - 1, acc + n) }
helper(100, 0)
`, 5050)
	})

	t.Run("递归条件终止_负数输入", func(t *testing.T) {
		runIntTest(t, `
fn f(n) { if n <= 0 { return 0 } return n + f(n - 2) }
f(9)
`, 25)
	})
}

// ========== 5. 函数作用域 ==========

func Test_FuncAdvanced_Scope(t *testing.T) {
	t.Run("函数内定义局部变量", func(t *testing.T) {
		runIntTest(t, `
fn f() {
	a := 10
	b := 20
	return a + b
}
f()
`, 30)
	})

	t.Run("函数不能访问外部变量_编译错误", func(t *testing.T) {
		// 无闭包: 函数不能引用外部作用域变量
		runErrorTest(t, `
x := 10
fn f() { return x }
`)
	})

	t.Run("参数遮蔽_同名参数在函数内独立", func(t *testing.T) {
		// 外部x和函数参数x互不影响
		runIntTest(t, `
fn f(x) { return x * 2 }
x := 100
f(5) + x
`, 110)
	})

	t.Run("嵌套函数调用中的变量隔离", func(t *testing.T) {
		// 函数a和函数b各有独立局部变量x
		runIntTest(t, `
fn a(n) { x := n + 1 return x }
fn b(n) { x := n + 2 return a(n) + x }
b(10)
`, 23)
	})

	t.Run("函数内修改参数不影响外部变量", func(t *testing.T) {
		runIntTest(t, `
fn reset(n) { n = 0 return n }
val := 77
reset(val)
val
`, 77)
	})

	t.Run("不同函数同名参数互不干扰", func(t *testing.T) {
		runIntTest(t, `
fn f1(x) { return x + 1 }
fn f2(x) { return x * 2 }
f1(10) + f2(10)
`, 31)
	})

	t.Run("函数内通过getBindValue访问绑定值", func(t *testing.T) {
		result := runScriptWithBinds(t, `
fn f() {
	v :=>int getBindValue("gval")
	return v * 2
}
f()
`, map[string]interface{}{"gval": 21})
		assertInt(t, result, 42)
	})
}

// ========== 6. 函数返回值类型 ==========

func Test_FuncAdvanced_ReturnTypes(t *testing.T) {
	t.Run("返回数组", func(t *testing.T) {
		result := runScript(t, `fn f() { return [1, 2, 3] } f()`)
		if result.Type != TypeArray {
			t.Fatalf("期望类型 TypeArray, 得到 %d", result.Type)
		}
		arr := result.Array()
		if len(arr.Elements) != 3 {
			t.Fatalf("期望长度 3, 得到 %d", len(arr.Elements))
		}
		assertInt(t, arr.Elements[0], 1)
		assertInt(t, arr.Elements[2], 3)
	})

	t.Run("返回Map", func(t *testing.T) {
		result := runScript(t, `fn f() { return {"a": 1, "b": 2} } f()`)
		if result.Type != TypeMap {
			t.Fatalf("期望类型 TypeMap, 得到 %d", result.Type)
		}
		m := result.Map()
		if len(m.Pairs) != 2 {
			t.Fatalf("期望2对键值, 得到 %d", len(m.Pairs))
		}
		assertInt(t, m.Pairs["a"], 1)
		assertInt(t, m.Pairs["b"], 2)
	})

	t.Run("返回布尔值", func(t *testing.T) {
		runBoolTest(t, `fn f(x) { return x > 0 } f(5)`, true)
	})

	t.Run("返回布尔值false", func(t *testing.T) {
		runBoolTest(t, `fn f(x) { return x > 0 } f(-5)`, false)
	})

	t.Run("返回字符串", func(t *testing.T) {
		runStringTest(t, `fn f() { return "hello" } f()`, "hello")
	})

	t.Run("返回nil_explicit", func(t *testing.T) {
		result := runScript(t, `fn f() { return nil } f()`)
		assertNil(t, result)
	})

	t.Run("无return语句隐式返回nil", func(t *testing.T) {
		result := runScript(t, `fn f() { x := 1 } f()`)
		assertNil(t, result)
	})

	t.Run("return无值返回nil", func(t *testing.T) {
		result := runScript(t, `fn f() { return } f()`)
		assertNil(t, result)
	})

	t.Run("多个return路径_正数", func(t *testing.T) {
		runStringTest(t, `
fn f(n) { if n > 0 { return "pos" } if n < 0 { return "neg" } return "zero" }
f(5)
`, "pos")
	})

	t.Run("多个return路径_负数", func(t *testing.T) {
		runStringTest(t, `
fn f(n) { if n > 0 { return "pos" } if n < 0 { return "neg" } return "zero" }
f(-3)
`, "neg")
	})

	t.Run("多个return路径_零", func(t *testing.T) {
		runStringTest(t, `
fn f(n) { if n > 0 { return "pos" } if n < 0 { return "neg" } return "zero" }
f(0)
`, "zero")
	})
}

// ========== 7. 函数控制流 ==========

func Test_FuncAdvanced_ControlFlow(t *testing.T) {
	t.Run("函数内break_编译错误", func(t *testing.T) {
		// break必须在循环内部
		runErrorTest(t, `fn f() { break return 1 }`)
	})

	t.Run("函数内continue_编译错误", func(t *testing.T) {
		// continue必须在循环内部
		runErrorTest(t, `fn f() { continue return 1 }`)
	})

	t.Run("函数内throw_运行时错误", func(t *testing.T) {
		runRuntimeErrorTest(t, `fn f() { throw "error" } f()`)
	})

	t.Run("函数内无限循环带break", func(t *testing.T) {
		runIntTest(t, `
fn f() {
	for {
		break
	}
	return 42
}
f()
`, 42)
	})

	t.Run("函数内for循环正常退出", func(t *testing.T) {
		runIntTest(t, `
fn f() {
	sum := 0
	for i := 1; i <= 5; i = i + 1 {
		sum = sum + i
	}
	return sum
}
f()
`, 15)
	})

	t.Run("函数内条件throw", func(t *testing.T) {
		runRuntimeErrorTest(t, `
fn f(n) {
	if n < 0 { throw "negative" }
	return n
}
f(-1)
`)
	})
}

// ========== 8. 高阶函数模式 ==========

func Test_FuncAdvanced_HigherOrder(t *testing.T) {
	t.Run("函数作为参数传递", func(t *testing.T) {
		runIntTest(t, `
fn dbl(x) { return x * 2 }
fn apply(f, x) { return f(x) }
apply(dbl, 21)
`, 42)
	})

	t.Run("同一apply传入不同函数", func(t *testing.T) {
		runIntTest(t, `
fn inc(x) { return x + 1 }
fn dec(x) { return x - 1 }
fn apply(f, x) { return f(x) }
apply(inc, 10) + apply(dec, 10)
`, 20)
	})

	t.Run("函数组合f_g_x", func(t *testing.T) {
		runIntTest(t, `
fn g(x) { return x + 1 }
fn f(x) { return x * 2 }
f(g(5))
`, 12)
	})

	t.Run("多层嵌套调用f_g_h_x", func(t *testing.T) {
		runIntTest(t, `
fn h(x) { return x + 1 }
fn g(x) { return h(x) + 1 }
fn f(x) { return g(x) + 1 }
f(0)
`, 3)
	})

	t.Run("四层嵌套调用", func(t *testing.T) {
		runIntTest(t, `
fn d(x) { return x + 1 }
fn c(x) { return d(x) + 1 }
fn b(x) { return c(x) + 1 }
fn a(x) { return b(x) + 1 }
a(0)
`, 4)
	})

	t.Run("通过apply组合三个函数", func(t *testing.T) {
		runIntTest(t, `
fn add1(x) { return x + 1 }
fn mul2(x) { return x * 2 }
fn sub3(x) { return x - 3 }
fn apply(f, x) { return f(x) }
apply(sub3, apply(mul2, apply(add1, 10)))
`, 19)
	})

	t.Run("函数返回不同类型值_int", func(t *testing.T) {
		runIntTest(t, `fn f() { return 42 } f()`, 42)
	})

	t.Run("函数返回不同类型值_string", func(t *testing.T) {
		runStringTest(t, `fn f() { return "abc" } f()`, "abc")
	})
}

// ========== 9. 参数边界 ==========

func Test_FuncAdvanced_ParamBoundary(t *testing.T) {
	t.Run("零个参数", func(t *testing.T) {
		runIntTest(t, `fn f() { return 1 } f()`, 1)
	})

	t.Run("三个参数", func(t *testing.T) {
		runIntTest(t, `fn f(a, b, c) { return a * 100 + b * 10 + c } f(1, 2, 3)`, 123)
	})

	t.Run("四个参数", func(t *testing.T) {
		runIntTest(t, `fn f(a, b, c, d) { return a + b + c + d } f(10, 20, 30, 40)`, 100)
	})

	t.Run("参数是复杂表达式", func(t *testing.T) {
		runIntTest(t, `
fn f(a, b) { return a + b }
f((1 + 2) * 3, 4 * 5)
`, 29)
	})

	t.Run("参数数量少于定义_运行时错误", func(t *testing.T) {
		runRuntimeErrorTest(t, `
fn add(a, b) { return a + b }
add(1)
`)
	})

	t.Run("参数数量多于定义_运行时错误", func(t *testing.T) {
		runRuntimeErrorTest(t, `
fn f(a) { return a }
f(1, 2, 3)
`)
	})
}

// ========== 10. 外部函数(#fn指令) ==========

func Test_FuncAdvanced_ExternalFn(t *testing.T) {
	t.Run("调用简单的绑定函数", func(t *testing.T) {
		result := runScriptWithFunc(t, `
#fn extAdd(int, int) => int
extAdd(3, 4)
`, "extAdd", func(a, b int) int { return a + b })
		assertInt(t, result, 7)
	})

	t.Run("绑定函数返回字符串", func(t *testing.T) {
		result := runScriptWithFunc(t, `
#fn extGreet(string) => string
extGreet("world")
`, "extGreet", func(s string) string { return "hello " + s })
		assertString(t, result, "hello world")
	})

	t.Run("绑定函数返回布尔值", func(t *testing.T) {
		result := runScriptWithFunc(t, `
#fn extBool(int) => bool
extBool(5)
`, "extBool", func(n int) bool { return n > 0 })
		assertBool(t, result, true)
	})

	t.Run("绑定函数接收多个参数", func(t *testing.T) {
		result := runScriptWithFunc(t, `
#fn extConcat(string, string, string) => string
extConcat("a", "b", "c")
`, "extConcat", func(a, b, c string) string { return a + b + c })
		assertString(t, result, "abc")
	})

	t.Run("绑定函数无参数", func(t *testing.T) {
		result := runScriptWithFunc(t, `
#fn extNoParam() => int
extNoParam()
`, "extNoParam", func() int { return 99 })
		assertInt(t, result, 99)
	})

	t.Run("绑定函数无返回值返回nil", func(t *testing.T) {
		result := runScriptWithFunc(t, `
#fn extVoid(int)
extVoid(5)
`, "extVoid", func(n int) {})
		assertNil(t, result)
	})

	t.Run("绑定函数接收数组参数", func(t *testing.T) {
		result := runScriptWithFunc(t, `
#fn extSum(array) => int
extSum([1, 2, 3])
`, "extSum", func(arr []int) int {
			sum := 0
			for _, v := range arr {
				sum += v
			}
			return sum
		})
		assertInt(t, result, 6)
	})

	t.Run("绑定函数返回数组", func(t *testing.T) {
		result := runScriptWithFunc(t, `
#fn extGenArr(int) => array
extGenArr(5)
`, "extGenArr", func(n int) []int { return []int{n, n + 1, n + 2} })
		if result.Type != TypeArray {
			t.Fatalf("期望 TypeArray, 得到 %d", result.Type)
		}
		arr := result.Array()
		if len(arr.Elements) != 3 {
			t.Fatalf("期望长度 3, 得到 %d", len(arr.Elements))
		}
		assertInt(t, arr.Elements[0], 5)
		assertInt(t, arr.Elements[2], 7)
	})

	t.Run("绑定函数在脚本函数内调用", func(t *testing.T) {
		result := runScriptWithFunc(t, `
#fn extDbl(int) => int
fn f(x) { return extDbl(x) + 1 }
f(10)
`, "extDbl", func(n int) int { return n * 2 })
		assertInt(t, result, 21)
	})
}

// ========== 11. 实用模式 ==========

func Test_FuncAdvanced_PracticalPatterns(t *testing.T) {
	t.Run("计数器函数", func(t *testing.T) {
		runIntTest(t, `
fn countDown(n) {
	count := 0
	for i := n; i > 0; i = i - 1 {
		count = count + 1
	}
	return count
}
countDown(5)
`, 5)
	})

	t.Run("累加器函数_数组求和", func(t *testing.T) {
		runIntTest(t, `
fn acc(arr) {
	sum := 0
	for i := 0; i < len(arr); i = i + 1 {
		sum = sum + arr[i]
	}
	return sum
}
acc([1, 2, 3, 4, 5])
`, 15)
	})

	t.Run("查找函数_找到目标", func(t *testing.T) {
		runIntTest(t, `
fn find(arr, target) {
	for i := 0; i < len(arr); i = i + 1 {
		if arr[i] == target { return i }
	}
	return -1
}
find([10, 20, 30], 20)
`, 1)
	})

	t.Run("查找函数_未找到返回负一", func(t *testing.T) {
		runIntTest(t, `
fn find(arr, target) {
	for i := 0; i < len(arr); i = i + 1 {
		if arr[i] == target { return i }
	}
	return -1
}
find([10, 20, 30], 99)
`, -1)
	})

	t.Run("过滤函数模式", func(t *testing.T) {
		result := runScript(t, `
fn filterPos(arr) {
	result := []
	for i := 0; i < len(arr); i = i + 1 {
		if arr[i] > 0 { result = push(result, arr[i]) }
	}
	return result
}
filterPos([-1, 2, -3, 4])
`)
		if result.Type != TypeArray {
			t.Fatalf("期望 TypeArray, 得到 %d", result.Type)
		}
		arr := result.Array()
		if len(arr.Elements) != 2 {
			t.Fatalf("期望长度 2, 得到 %d", len(arr.Elements))
		}
		assertInt(t, arr.Elements[0], 2)
		assertInt(t, arr.Elements[1], 4)
	})

	t.Run("映射函数模式", func(t *testing.T) {
		result := runScript(t, `
fn mapDouble(arr) {
	result := []
	for i := 0; i < len(arr); i = i + 1 {
		result = push(result, arr[i] * 2)
	}
	return result
}
mapDouble([1, 2, 3])
`)
		if result.Type != TypeArray {
			t.Fatalf("期望 TypeArray, 得到 %d", result.Type)
		}
		arr := result.Array()
		if len(arr.Elements) != 3 {
			t.Fatalf("期望长度 3, 得到 %d", len(arr.Elements))
		}
		assertInt(t, arr.Elements[0], 2)
		assertInt(t, arr.Elements[1], 4)
		assertInt(t, arr.Elements[2], 6)
	})

	t.Run("函数内修改Map并返回", func(t *testing.T) {
		result := runScript(t, `
fn addEntry(m, key, val) {
	m[key] = val
	return m
}
d := {"a": 1}
addEntry(d, "b", 2)
d["b"]
`)
		assertInt(t, result, 2)
	})

	t.Run("多层函数调用链计算", func(t *testing.T) {
		runIntTest(t, `
fn step(x) { return x * 3 + 1 }
val := 1
for i := 0; i < 4; i = i + 1 {
	val = step(val)
}
val
`, 121)
	})
}

// ========== 12. 函数值类型与深度相等 ==========

func Test_FuncAdvanced_DeepEqual(t *testing.T) {
	t.Run("函数返回数组深度比较", func(t *testing.T) {
		result := runScript(t, `
fn makeArr() { return [1, 2, 3] }
makeArr()
`)
		expected := Value{Type: TypeArray, Data: &ArrayValue{Elements: []Value{
			intVal(1), intVal(2), intVal(3),
		}}}
		if !result.Equal(expected) {
			t.Errorf("期望 %v, 得到 %v", expected, result)
		}
	})

	t.Run("函数返回Map深度比较", func(t *testing.T) {
		result := runScript(t, `
fn makeMap() { return {"x": 1, "y": 2} }
makeMap()
`)
		expected := Value{Type: TypeMap, Data: &MapValue{Pairs: map[string]Value{
			"x": intVal(1), "y": intVal(2),
		}}}
		if !result.Equal(expected) {
			t.Errorf("期望 %v, 得到 %v", expected, result)
		}
	})

	t.Run("函数返回数组类型断言", func(t *testing.T) {
		result := runScript(t, `fn f() { return [10, 20] } f()`)
		arr := result.Array()
		if arr == nil {
			t.Fatal("期望非nil数组")
		}
		actual := make([]int, len(arr.Elements))
		for i, e := range arr.Elements {
			actual[i] = e.Int()
		}
		expected := []int{10, 20}
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("期望 %v, 得到 %v", expected, actual)
		}
	})
}
