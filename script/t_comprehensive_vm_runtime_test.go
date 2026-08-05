package script

import (
	"math"
	"testing"
)

// ========== VM执行和运行时综合测试 ==========

// ========== 算术运算测试 ==========

func Test_Comprehensive_Arithmetic_Int(t *testing.T) {
	t.Run("基本运算", func(t *testing.T) {
		tests := []struct {
			Input    string
			Expected int
		}{
			{"1 + 2", 3},
			{"10 - 3", 7},
			{"4 * 5", 20},
			{"20 / 4", 5},
			{"17 % 5", 2},
			{"0 + 0", 0},
			{"100 - 100", 0},
			{"1 * 0", 0},
		}
		RunIntTestsSimple(t, tests)
	})

	t.Run("复合表达式", func(t *testing.T) {
		tests := []struct {
			Input    string
			Expected int
		}{
			{"1 + 2 * 3", 7},
			{"(1 + 2) * 3", 9},
			{"10 - 2 * 3", 4},
			{"2 + 3 * 4 - 1", 13},
			{"(20 - 4) / 2", 8},
			{"(10 + 20) % 7", 2},
		}
		RunIntTestsSimple(t, tests)
	})

	t.Run("边界值", func(t *testing.T) {
		// 大数运算
		runIntTest(t, "1000000 * 1000000", 1000000000000)
		// 零的运算
		runIntTest(t, "0 + 5", 5)
		runIntTest(t, "5 - 0", 5)
		runIntTest(t, "0 * 100", 0)
		// 最大int边界
		runIntTest(t, "0x7FFFFFFFFFFFFFFF", math.MaxInt64)
	})
}

func Test_Comprehensive_Arithmetic_Float(t *testing.T) {
	t.Run("基本运算", func(t *testing.T) {
		tests := []struct {
			Input    string
			Expected float64
		}{
			{"1.5 + 2.5", 4.0},
			{"10.0 - 3.5", 6.5},
			{"2.5 * 4.0", 10.0},
			{"10.0 / 4.0", 2.5},
		}
		RunFloatTestsSimple(t, tests)
	})

	t.Run("浮点精度", func(t *testing.T) {
		runFloatTest(t, "0.1 + 0.2", 0.30000000000000004)
		runFloatTest(t, "1.0 / 3.0", 0.3333333333333333)
		runFloatTest(t, "2.0 / 3.0", 0.6666666666666666)
	})
}

func Test_Comprehensive_Arithmetic_Mixed(t *testing.T) {
	// int+float混合运算自动类型提升
	t.Run("int与float混合运算", func(t *testing.T) {
		runFloatTest(t, `1 + 2.5`, 3.5)
		runFloatTest(t, `10 - 0.5`, 9.5)
		runFloatTest(t, `3 * 1.5`, 4.5)
		runFloatTest(t, `10 / 4.0`, 2.5)
	})
	t.Run("显式转换后运算", func(t *testing.T) {
		runFloatTest(t, `float(1) + 2.5`, 3.5)
		runFloatTest(t, `float(10) - 0.5`, 9.5)
		runFloatTest(t, `float(3) * 1.5`, 4.5)
		runFloatTest(t, `float(10) / 4.0`, 2.5)
	})
	t.Run("纯float运算", func(t *testing.T) {
		runFloatTest(t, `1.0 + 2.5`, 3.5)
		runFloatTest(t, `10.0 - 0.5`, 9.5)
		runFloatTest(t, `3.0 * 1.5`, 4.5)
		runFloatTest(t, `10.0 / 4.0`, 2.5)
	})
}

func Test_Comprehensive_Arithmetic_Bitwise(t *testing.T) {
	t.Run("位运算", func(t *testing.T) {
		tests := []struct {
			Input    string
			Expected int
		}{
			{"0xFF & 0x0F", 0x0F},
			{"0xF0 | 0x0F", 0xFF},
			{"0xFF ^ 0x0F", 0xF0},
			{"1 << 4", 16},
			{"256 >> 2", 64},
			{"12 & 10", 8},
			{"12 | 10", 14},
			{"12 ^ 10", 6},
		}
		RunIntTestsSimple(t, tests)
	})

	t.Run("位取反", func(t *testing.T) {
		runIntTest(t, "^0", ^0)
		runIntTest(t, "^5", ^5)
		runIntTest(t, "^255", ^255)
	})
}

func Test_Comprehensive_Arithmetic_Unary(t *testing.T) {
	t.Run("取负", func(t *testing.T) {
		runIntTest(t, "-5", -5)
		runIntTest(t, "-0", 0)
		runIntTest(t, "-(-10)", 10)
		runIntTest(t, "-(-(-5))", -5)
		runFloatTest(t, "-3.14", -3.14)
		runFloatTest(t, "-(-2.5)", 2.5)
	})

	t.Run("逻辑非", func(t *testing.T) {
		runBoolTest(t, "!false", true)
		runBoolTest(t, "!true", false)
		runBoolTest(t, "!!true", true)
		runBoolTest(t, "!!false", false)
	})
}

// ========== 字符串操作测试 ==========

func Test_Comprehensive_String(t *testing.T) {
	t.Run("拼接", func(t *testing.T) {
		runStringTest(t, `"a" + "b"`, "ab")
		runStringTest(t, `"hello" + " " + "world"`, "hello world")
		runStringTest(t, `"" + ""`, "")
		runStringTest(t, `"num:" + 42`, "num:42")
		runStringTest(t, `42 + "!"`, "42!")
	})

	t.Run("索引", func(t *testing.T) {
		runStringTest(t, `"hello"[0]`, "h")
		runStringTest(t, `"hello"[4]`, "o")
	})

	t.Run("切片", func(t *testing.T) {
		runStringTest(t, `"hello"[1:4]`, "ell")
		runStringTest(t, `"hello"[:3]`, "hel")
		runStringTest(t, `"hello"[2:]`, "llo")
		runStringTest(t, `"hello"[0:5]`, "hello")
	})

	t.Run("长度", func(t *testing.T) {
		runIntTest(t, `len("hello")`, 5)
		runIntTest(t, `len("")`, 0)
		runIntTest(t, `len("a")`, 1)
	})
}

// ========== 数组操作测试 ==========

func Test_Comprehensive_Array(t *testing.T) {
	t.Run("创建和访问", func(t *testing.T) {
		runIntTest(t, `[1, 2, 3][0]`, 1)
		runIntTest(t, `[1, 2, 3][1]`, 2)
		runIntTest(t, `[1, 2, 3][2]`, 3)
	})

	t.Run("修改", func(t *testing.T) {
		runIntTest(t, `arr := [1, 2, 3]
arr[0] = 10
arr[0]`, 10)
		runIntTest(t, `arr := [1, 2, 3]
arr[2] = 30
arr[2]`, 30)
	})

	t.Run("切片", func(t *testing.T) {
		result := runScript(t, `[1, 2, 3, 4, 5][1:3]`)
		arr := result.Array()
		if len(arr.Elements) != 2 || arr.Elements[0].Int() != 2 || arr.Elements[1].Int() != 3 {
			t.Errorf("数组切片结果错误: %v", arr.Elements)
		}
	})

	t.Run("追加", func(t *testing.T) {
		runIntTest(t, `arr := [1, 2, 3]
arr2 := push(arr, 4)
arr2[3]`, 4)
		// 原数组不变
		runIntTest(t, `arr := [1, 2, 3]
arr2 := push(arr, 4)
len(arr)`, 3)
	})

	t.Run("range单变量遍历", func(t *testing.T) {
		runIntTest(t, `
arr := [10, 20, 30]
sum := 0
for v := range arr {
    sum = sum + v
}
sum`, 60)
	})

	t.Run("range双变量遍历", func(t *testing.T) {
		runIntTest(t, `
arr := [10, 20, 30]
sum := 0
for i, v := range arr {
    sum = sum + i*10 + v
}
sum`, 90) // (0+10)+(10+20)+(20+30) = 90
	})

	t.Run("嵌套数组", func(t *testing.T) {
		runIntTest(t, `[[1, 2], [3, 4]][0][0]`, 1)
		runIntTest(t, `[[1, 2], [3, 4]][1][0]`, 3)
		runIntTest(t, `[[1, 2], [3, 4]][1][1]`, 4)
		runIntTest(t, `arr := [[1, 2], [3, 4]]
arr[0][1]`, 2)
	})
}

// ========== Map操作测试 ==========

func Test_Comprehensive_Map(t *testing.T) {
	t.Run("创建和访问", func(t *testing.T) {
		runIntTest(t, `{"a": 1, "b": 2}["a"]`, 1)
		runIntTest(t, `{"a": 1, "b": 2}["b"]`, 2)
	})

	t.Run("修改", func(t *testing.T) {
		runIntTest(t, `m := {"a": 1, "b": 2}
m["a"] = 10
m["a"]`, 10)
	})

	t.Run("添加新key", func(t *testing.T) {
		runIntTest(t, `m := {"a": 1}
m["new"] = 3
m["new"]`, 3)
	})

	t.Run("删除", func(t *testing.T) {
		runIntTest(t, `m := {"a": 1, "b": 2, "c": 3}
delete(m, "a")
len(m)`, 2)
		// 验证删除后访问返回nil
		result := runScript(t, `m := {"a": 1, "b": 2}
delete(m, "a")
m["a"]`)
		if !result.IsNil() {
			t.Errorf("删除后访问应返回nil, 得到 %v", result)
		}
	})

	t.Run("长度", func(t *testing.T) {
		runIntTest(t, `len({"a": 1, "b": 2, "c": 3})`, 3)
		runIntTest(t, `len({})`, 0)
	})

	t.Run("不存在的key返回nil", func(t *testing.T) {
		result := runScript(t, `m := {"a": 1}
m["nonexistent"]`)
		if !result.IsNil() {
			t.Errorf("不存在的key应返回nil, 得到 %v", result)
		}
	})
}

// ========== 控制流测试 ==========

func Test_Comprehensive_ControlFlow_IfElse(t *testing.T) {
	t.Run("if分支", func(t *testing.T) {
		runIntTest(t, `x := 10
if x > 5 { 1 } else { 0 }`, 1)
		runIntTest(t, `x := 3
if x > 5 { 1 } else { 0 }`, 0)
	})

	t.Run("if/else if/else", func(t *testing.T) {
		runIntTest(t, `
x := 5
if x > 10 {
    1
} else if x > 3 {
    2
} else {
    3
}`, 2)
		runIntTest(t, `
x := 2
if x > 10 {
    1
} else if x > 3 {
    2
} else {
    3
}`, 3)
		runIntTest(t, `
x := 20
if x > 10 {
    1
} else if x > 3 {
    2
} else {
    3
}`, 1)
	})
}

func Test_Comprehensive_ControlFlow_For(t *testing.T) {
	t.Run("while循环", func(t *testing.T) {
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
for i := 0; i < 5; i = i + 1 {
    sum = sum + i
}
sum`, 10)
	})

	t.Run("无限循环加break", func(t *testing.T) {
		runIntTest(t, `
i := 0
for {
    i = i + 1
    if i >= 10 { break }
}
i`, 10)
	})

	t.Run("range循环", func(t *testing.T) {
		runIntTest(t, `
arr := [1, 2, 3, 4, 5]
sum := 0
for v := range arr {
    sum = sum + v
}
sum`, 15)
	})

	t.Run("range双变量", func(t *testing.T) {
		runIntTest(t, `
arr := [1, 2, 3]
sum := 0
for i, v := range arr {
    sum = sum + i + v
}
sum`, 9)
	})
}

func Test_Comprehensive_ControlFlow_NestedLoop(t *testing.T) {
	t.Run("嵌套循环break", func(t *testing.T) {
		runIntTest(t, `
count := 0
for i := 0; i < 3; i = i + 1 {
    for j := 0; j < 3; j = j + 1 {
        if j == 2 { break }
        count = count + 1
    }
}
count`, 6)
	})

	t.Run("嵌套循环continue", func(t *testing.T) {
		runIntTest(t, `
sum := 0
for i := 0; i < 3; i = i + 1 {
    for j := 0; j < 3; j = j + 1 {
        if j == 1 { continue }
        sum = sum + 1
    }
}
sum`, 6)
	})
}

func Test_Comprehensive_ControlFlow_EarlyReturn(t *testing.T) {
	t.Run("函数提前return", func(t *testing.T) {
		runIntTest(t, `
fn find(arr, target) {
    for v := range arr {
        if v == target {
            return 1
        }
    }
    return 0
}
find([1, 2, 3], 2)`, 1)
	})

	t.Run("未找到返回0", func(t *testing.T) {
		runIntTest(t, `
fn find(arr, target) {
    for v := range arr {
        if v == target {
            return 1
        }
    }
    return 0
}
find([1, 2, 3], 99)`, 0)
	})
}

// ========== 函数测试 ==========

func Test_Comprehensive_Function(t *testing.T) {
	t.Run("无参函数", func(t *testing.T) {
		runIntTest(t, `
fn answer() {
    42
}
answer()`, 42)
	})

	t.Run("有参函数", func(t *testing.T) {
		runIntTest(t, `
fn add(a, b) {
    a + b
}
add(3, 4)`, 7)
	})

	t.Run("隐式返回", func(t *testing.T) {
		runIntTest(t, `
fn double(x) {
    x * 2
}
double(21)`, 42)
	})

	t.Run("return语句", func(t *testing.T) {
		runIntTest(t, `
fn multiply(a, b) {
    return a * b
}
multiply(6, 7)`, 42)
	})
}

func Test_Comprehensive_Function_Recursion(t *testing.T) {
	t.Run("阶乘", func(t *testing.T) {
		runIntTest(t, `
fn factorial(n) {
    if n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}
factorial(5)`, 120)
	})

	t.Run("斐波那契", func(t *testing.T) {
		runIntTest(t, `
fn fib(n) {
    if n < 2 {
        return n
    }
    return fib(n - 1) + fib(n - 2)
}
fib(10)`, 55)
	})
}

func Test_Comprehensive_Function_HigherOrder(t *testing.T) {
	t.Run("函数作为参数", func(t *testing.T) {
		runIntTest(t, `
fn double(x) {
    x * 2
}
fn apply(f, x) {
    f(x)
}
apply(double, 21)`, 42)
	})

	t.Run("嵌套调用", func(t *testing.T) {
		runIntTest(t, `
fn a(x) { x + 1 }
fn b(x) { a(x) + 1 }
fn c(x) { b(x) + 1 }
c(0)`, 3)
	})

	t.Run("高阶函数组合", func(t *testing.T) {
		runIntTest(t, `
fn inc(x) { x + 1 }
fn double(x) { x * 2 }
fn apply(f, x) { f(x) }
apply(inc, apply(double, 5))`, 11)
	})
}

// ========== 类型转换测试 ==========

func Test_Comprehensive_TypeConversion(t *testing.T) {
	t.Run("int转换", func(t *testing.T) {
		runIntTest(t, `int("123")`, 123)
		runIntTest(t, `int("-456")`, -456)
		runIntTest(t, `int(3.14)`, 3)
		runIntTest(t, `int(3.99)`, 3)
		runIntTest(t, `int(true)`, 1)
		runIntTest(t, `int(false)`, 0)
	})

	t.Run("float转换", func(t *testing.T) {
		runFloatTest(t, `float("3.14")`, 3.14)
		runFloatTest(t, `float(42)`, 42.0)
		runFloatTest(t, `float("2.5")`, 2.5)
	})

	t.Run("string转换", func(t *testing.T) {
		runStringTest(t, `string(123)`, "123")
		runStringTest(t, `string(-456)`, "-456")
		runStringTest(t, `string("already")`, "already")
		runStringTest(t, `string(true)`, "true")
		runStringTest(t, `string(false)`, "false")
	})

	t.Run("typeof", func(t *testing.T) {
		runStringTest(t, `typeof(123)`, "int")
		runStringTest(t, `typeof(3.14)`, "float")
		runStringTest(t, `typeof("hello")`, "string")
		runStringTest(t, `typeof(true)`, "bool")
		runStringTest(t, `typeof(nil)`, "nil")
		runStringTest(t, `typeof([1, 2])`, "array")
		runStringTest(t, `typeof({"a": 1})`, "map")
		// fn是语句不是表达式, 通过变量获取类型
		runStringTest(t, `fn f(){1}
typeof(f)`, "function")
	})
}

// ========== 错误处理测试 ==========

func Test_Comprehensive_Error_DivisionByZero(t *testing.T) {
	t.Run("int除零", func(t *testing.T) {
		runRuntimeErrorTest(t, "1 / 0")
	})
	t.Run("int模零", func(t *testing.T) {
		runRuntimeErrorTest(t, "5 % 0")
	})
}

func Test_Comprehensive_Error_TypeMismatch(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"字符串减整数", `"hello" - 5`},
		{"布尔加整数", `true + 5`},
		{"数组乘整数", `[1] * 2`},
		{"字符串取负", `-"hello"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runRuntimeErrorTest(t, tt.code)
		})
	}
}

func Test_Comprehensive_Error_ArrayOutOfBounds(t *testing.T) {
	// 越界访问返回nil
	result := runScript(t, `arr := [1, 2, 3]
arr[100]`)
	if !result.IsNil() {
		t.Errorf("越界访问期望nil, 得到 %v", result)
	}
}

func Test_Comprehensive_Error_UndefinedVariable(t *testing.T) {
	// 未定义变量是编译错误
	runErrorTest(t, `x = undefined_var`)
}

func Test_Comprehensive_Error_InvalidTypeConversion(t *testing.T) {
	t.Run("int转换数组", func(t *testing.T) {
		runRuntimeErrorTest(t, `int([1, 2])`)
	})
	t.Run("int转换Map", func(t *testing.T) {
		runRuntimeErrorTest(t, `int({"a": 1})`)
	})
	t.Run("int转换无效字符串", func(t *testing.T) {
		runRuntimeErrorTest(t, `int("abc")`)
	})
	t.Run("float转换无效字符串", func(t *testing.T) {
		runRuntimeErrorTest(t, `float("xyz")`)
	})
}
