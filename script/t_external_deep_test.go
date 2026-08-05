package script

import (
	"errors"
	"strings"
	"testing"
)

// ========== 绑定值与外部函数深度交互测试 ==========
// 测试Go与脚本引擎之间的值传递: 绑定值类型转换、外部函数参数/返回值转换、
// 复杂调用场景、错误处理、修改语义和多函数协作

// runExpectRuntimeError 编译并执行脚本, 期望运行时返回错误
func runExpectRuntimeError(t *testing.T, input string, ctx *Context) {
	t.Helper()
	parser := NewParser()
	compiled, err := parser.Compile(input)
	if err != nil {
		t.Fatalf("不期望编译失败: %v", err)
	}
	engine := NewEngine()
	_, err = engine.Run(ctx, compiled)
	if err == nil {
		t.Error("期望运行时错误, 但执行成功")
	}
}

// ========== 绑定值类型转换测试 ==========

func Test_ExternalDeep_BindValue_TypeConversion(t *testing.T) {
	t.Run("Go int 转脚本 int", func(t *testing.T) {
		result := runScriptWithBinds(t,
			`x :=>int getBindValue("x")
x`,
			map[string]interface{}{"x": 42},
		)
		assertInt(t, result, 42)
	})

	t.Run("Go int64 转脚本 int", func(t *testing.T) {
		result := runScriptWithBinds(t,
			`x :=>int getBindValue("x")
x`,
			map[string]interface{}{"x": int64(100)},
		)
		assertInt(t, result, 100)
	})

	t.Run("Go float64 转脚本 float", func(t *testing.T) {
		result := runScriptWithBinds(t,
			`x :=>float getBindValue("x")
x`,
			map[string]interface{}{"x": 3.14},
		)
		assertFloat(t, result, 3.14)
	})

	t.Run("Go string 转脚本 string", func(t *testing.T) {
		result := runScriptWithBinds(t,
			`x :=>string getBindValue("x")
x`,
			map[string]interface{}{"x": "hello"},
		)
		assertString(t, result, "hello")
	})

	t.Run("Go bool 转脚本 bool", func(t *testing.T) {
		result := runScriptWithBinds(t,
			`x :=>bool getBindValue("x")
x`,
			map[string]interface{}{"x": true},
		)
		assertBool(t, result, true)
	})

	t.Run("Go nil 转脚本 nil", func(t *testing.T) {
		result := runScriptWithBinds(t,
			`x :=>any getBindValue("x")
x`,
			map[string]interface{}{"x": nil},
		)
		assertNil(t, result)
	})

	t.Run("Go []int 转脚本 array", func(t *testing.T) {
		result := runScriptWithBinds(t,
			`x :=>arr getBindValue("x")
x[0] + x[1] + x[2]`,
			map[string]interface{}{"x": []int{1, 2, 3}},
		)
		assertInt(t, result, 6)
	})

	t.Run("Go []string 转脚本 array", func(t *testing.T) {
		result := runScriptWithBinds(t,
			`x :=>arr getBindValue("x")
x[0] + x[1]`,
			map[string]interface{}{"x": []string{"foo", "bar"}},
		)
		assertString(t, result, "foobar")
	})

	t.Run("Go []interface{} 转脚本 array", func(t *testing.T) {
		result := runScriptWithBinds(t,
			`x :=>arr getBindValue("x")
x[0]`,
			map[string]interface{}{"x": []interface{}{42, "a", true}},
		)
		assertInt(t, result, 42)
	})

	t.Run("Go map[string]int 转脚本 map", func(t *testing.T) {
		result := runScriptWithBinds(t,
			`m :=>any getBindValue("m")
m["a"] + m["b"]`,
			map[string]interface{}{"m": map[string]int{"a": 10, "b": 20}},
		)
		assertInt(t, result, 30)
	})

	t.Run("Go map[string]interface{} 转脚本 map", func(t *testing.T) {
		result := runScriptWithBinds(t,
			`m :=>any getBindValue("m")
m["name"]`,
			map[string]interface{}{"m": map[string]interface{}{"name": "Alice", "age": 30}},
		)
		assertString(t, result, "Alice")
	})

	t.Run("Go struct 转脚本 nil", func(t *testing.T) {
		// struct类型没有注册转换器, 引擎返回nil
		type MyStruct struct{ X int }
		result := runScriptWithBinds(t,
			`x :=>any getBindValue("x")
x`,
			map[string]interface{}{"x": MyStruct{X: 42}},
		)
		assertNil(t, result)
	})
}

// ========== 绑定值在脚本中的使用测试 ==========

func Test_ExternalDeep_BindValue_Usage(t *testing.T) {
	t.Run("绑定 int 参与算术", func(t *testing.T) {
		result := runScriptWithBinds(t,
			`x :=>int getBindValue("x")
x + 1`,
			map[string]interface{}{"x": 41},
		)
		assertInt(t, result, 42)
	})

	t.Run("绑定 string 参与拼接", func(t *testing.T) {
		result := runScriptWithBinds(t,
			`s :=>string getBindValue("s")
s + " world"`,
			map[string]interface{}{"s": "hello"},
		)
		assertString(t, result, "hello world")
	})

	t.Run("绑定数组索引访问", func(t *testing.T) {
		result := runScriptWithBinds(t,
			`arr :=>arr getBindValue("arr")
arr[2]`,
			map[string]interface{}{"arr": []int{10, 20, 30}},
		)
		assertInt(t, result, 30)
	})

	t.Run("绑定数组遍历求和", func(t *testing.T) {
		result := runScriptWithBinds(t,
			`arr :=>arr getBindValue("arr")
sum := 0
for v := range arr {
sum = sum + v
}
sum`,
			map[string]interface{}{"arr": []int{1, 2, 3, 4, 5}},
		)
		assertInt(t, result, 15)
	})

	t.Run("绑定 Map 访问", func(t *testing.T) {
		result := runScriptWithBinds(t,
			`m :=>any getBindValue("m")
m["key"]`,
			map[string]interface{}{"m": map[string]int{"key": 99}},
		)
		assertInt(t, result, 99)
	})

	t.Run("绑定 Map 遍历计数", func(t *testing.T) {
		result := runScriptWithBinds(t,
			`m :=>any getBindValue("m")
count := 0
for k := range m {
count = count + 1
}
count`,
			map[string]interface{}{"m": map[string]int{"a": 1, "b": 2, "c": 3}},
		)
		assertInt(t, result, 3)
	})

	t.Run("绑定 bool 在条件中使用", func(t *testing.T) {
		result := runScriptWithBinds(t,
			`b :=>bool getBindValue("b")
if b { 1 } else { 0 }`,
			map[string]interface{}{"b": true},
		)
		assertInt(t, result, 1)
	})
}

// ========== 外部函数 - 基本类型测试 ==========

func Test_ExternalDeep_Func_BasicTypes(t *testing.T) {
	t.Run("加倍函数", func(t *testing.T) {
		result := runScriptWithFunc(t,
			`#fn double(int)=>int
double(21)`,
			"double", func(x int) int { return x * 2 },
		)
		assertInt(t, result, 42)
	})

	t.Run("平方函数", func(t *testing.T) {
		result := runScriptWithFunc(t,
			`#fn square(int)=>int
square(9)`,
			"square", func(x int) int { return x * x },
		)
		assertInt(t, result, 81)
	})

	t.Run("转大写", func(t *testing.T) {
		result := runScriptWithFunc(t,
			`#fn toUpper(string)=>string
toUpper("hello")`,
			"toUpper", func(s string) string { return strings.ToUpper(s) },
		)
		assertString(t, result, "HELLO")
	})

	t.Run("字符串拼接", func(t *testing.T) {
		result := runScriptWithFunc(t,
			`#fn concat(string, string)=>string
concat("foo", "bar")`,
			"concat", func(a, b string) string { return a + b },
		)
		assertString(t, result, "foobar")
	})

	t.Run("加法", func(t *testing.T) {
		result := runScriptWithFunc(t,
			`#fn add(int, int)=>int
add(10, 20)`,
			"add", func(a, b int) int { return a + b },
		)
		assertInt(t, result, 30)
	})

	t.Run("减法", func(t *testing.T) {
		result := runScriptWithFunc(t,
			`#fn sub(int, int)=>int
sub(50, 20)`,
			"sub", func(a, b int) int { return a - b },
		)
		assertInt(t, result, 30)
	})

	t.Run("最大值", func(t *testing.T) {
		result := runScriptWithFunc(t,
			`#fn max(int, int)=>int
max(3, 7)`,
			"max", func(a, b int) int {
				if a > b {
					return a
				}
				return b
			},
		)
		assertInt(t, result, 7)
	})

	t.Run("无参数返回常量", func(t *testing.T) {
		result := runScriptWithFunc(t,
			`#fn answer()=>int
answer()`,
			"answer", func() int { return 42 },
		)
		assertInt(t, result, 42)
	})

	t.Run("无参数返回字符串", func(t *testing.T) {
		result := runScriptWithFunc(t,
			`#fn greet()=>string
greet()`,
			"greet", func() string { return "hello" },
		)
		assertString(t, result, "hello")
	})
}

// ========== 外部函数 - 返回值模式测试 ==========

func Test_ExternalDeep_Func_ReturnPatterns(t *testing.T) {
	t.Run("返回 int", func(t *testing.T) {
		result := runScriptWithFunc(t,
			`#fn val()=>int
val()`,
			"val", func() int { return 100 },
		)
		assertInt(t, result, 100)
	})

	t.Run("返回 string", func(t *testing.T) {
		result := runScriptWithFunc(t,
			`#fn msg()=>string
msg()`,
			"msg", func() string { return "test" },
		)
		assertString(t, result, "test")
	})

	t.Run("返回 bool", func(t *testing.T) {
		result := runScriptWithFunc(t,
			`#fn flag()=>bool
flag()`,
			"flag", func() bool { return true },
		)
		assertBool(t, result, true)
	})

	t.Run("返回 nil 无返回值", func(t *testing.T) {
		// Go函数无返回值时, 脚本侧返回nil
		result := runScriptWithFunc(t,
			`#fn noop(int)
noop(42)`,
			"noop", func(x int) {},
		)
		assertNil(t, result)
	})

	t.Run("返回 []int", func(t *testing.T) {
		result := runScriptWithFunc(t,
			`#fn genArr()=>arr
r := genArr()
r[0] + r[1] + r[2]`,
			"genArr", func() []int { return []int{10, 20, 30} },
		)
		assertInt(t, result, 60)
	})

	t.Run("返回 map[string]int", func(t *testing.T) {
		result := runScriptWithFunc(t,
			`#fn genMap()=>any
m := genMap()
m["x"] + m["y"]`,
			"genMap", func() map[string]int { return map[string]int{"x": 5, "y": 15} },
		)
		assertInt(t, result, 20)
	})

	t.Run("返回 (int, error) 正常路径", func(t *testing.T) {
		result := runScriptWithFunc(t,
			`#fn safeDiv(int, int)=>int
safeDiv(10, 2)`,
			"safeDiv", func(a, b int) (int, error) {
				if b == 0 {
					return 0, errors.New("div by zero")
				}
				return a / b, nil
			},
		)
		assertInt(t, result, 5)
	})

	t.Run("返回 (int, error) 错误路径", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("safeDiv", func(a, b int) (int, error) {
			if b == 0 {
				return 0, errors.New("div by zero")
			}
			return a / b, nil
		})
		runExpectRuntimeError(t,
			`#fn safeDiv(int, int)=>int
safeDiv(10, 0)`,
			ctx,
		)
	})

	t.Run("返回 (string, error)", func(t *testing.T) {
		result := runScriptWithFunc(t,
			`#fn safeGet(int)=>string
safeGet(0)`,
			"safeGet", func(i int) (string, error) {
				arr := []string{"a", "b", "c"}
				if i < 0 || i >= len(arr) {
					return "", errors.New("out of range")
				}
				return arr[i], nil
			},
		)
		assertString(t, result, "a")
	})

	t.Run("多个返回值不支持", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("three", func() (int, int, int) { return 1, 2, 3 })
		runExpectRuntimeError(t,
			`#fn three()=>int
three()`,
			ctx,
		)
	})
}

// ========== 外部函数 - 参数转换测试 ==========

func Test_ExternalDeep_Func_ParamConversion(t *testing.T) {
	t.Run("脚本 int -> Go int", func(t *testing.T) {
		fn := func(x int) int { return x + 1 }
		result, err := callExternalFunc(fn, []Value{NewValue(41)})
		assertNoError(t, err)
		assertInt(t, result, 42)
	})

	t.Run("脚本 int -> Go int64", func(t *testing.T) {
		fn := func(x int64) int64 { return x * 2 }
		result, err := callExternalFunc(fn, []Value{NewValue(50)})
		assertNoError(t, err)
		assertInt(t, result, 100)
	})

	t.Run("脚本 int -> Go float64", func(t *testing.T) {
		// convertNumericValue对TypeInt值调用Float(), 引擎实际行为
		fn := func(x float64) float64 { return x }
		result, err := callExternalFunc(fn, []Value{NewValue(42)})
		assertNoError(t, err)
		// 引擎实际行为: int到float64转换
		if result.Type != TypeFloat {
			t.Errorf("期望 float 类型, 得到 %v", result.Type)
		}
	})

	t.Run("脚本 string -> Go string", func(t *testing.T) {
		fn := func(s string) string { return s + "!" }
		result, err := callExternalFunc(fn, []Value{NewValue("hi")})
		assertNoError(t, err)
		assertString(t, result, "hi!")
	})

	t.Run("脚本 bool -> Go bool", func(t *testing.T) {
		fn := func(b bool) bool { return !b }
		result, err := callExternalFunc(fn, []Value{NewValue(true)})
		assertNoError(t, err)
		assertBool(t, result, false)
	})

	t.Run("脚本 array -> Go []int", func(t *testing.T) {
		fn := func(arr []int) int {
			sum := 0
			for _, v := range arr {
				sum += v
			}
			return sum
		}
		result, err := callExternalFunc(fn, []Value{NewValue([]Value{NewValue(1), NewValue(2), NewValue(3)})})
		assertNoError(t, err)
		assertInt(t, result, 6)
	})

	t.Run("脚本 array -> Go []interface{}", func(t *testing.T) {
		// 引擎限制: convertNumericValue不支持reflect.Interface目标类型
		// 向[]interface{}参数传数组时, 元素转换失败
		fn := func(arr []interface{}) int { return len(arr) }
		_, err := callExternalFunc(fn, []Value{NewValue([]Value{NewValue(1), NewValue("a")})})
		assertError(t, err)
	})

	t.Run("脚本 map -> Go map[string]int", func(t *testing.T) {
		fn := func(m map[string]int) int { return m["a"] + m["b"] }
		scriptMap := &MapValue{
			Pairs:   map[string]Value{"a": NewValue(10), "b": NewValue(20)},
			KeyType: TypeString,
		}
		result, err := callExternalFunc(fn, []Value{NewValue(scriptMap)})
		assertNoError(t, err)
		assertInt(t, result, 30)
	})

	t.Run("脚本 nil -> Go interface{}", func(t *testing.T) {
		fn := func(x interface{}) bool { return x == nil }
		result, err := callExternalFunc(fn, []Value{NewValue(nil)})
		assertNoError(t, err)
		assertBool(t, result, true)
	})
}

// ========== 外部函数 - 复杂场景测试 ==========

func Test_ExternalDeep_Func_Complex(t *testing.T) {
	t.Run("脚本函数调用外部函数", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("double", func(x int) int { return x * 2 })
		result := runScriptWithContext(t,
			`#fn double(int)=>int
fn process(n) {
double(n) + 1
}
process(10)`,
			ctx,
		)
		assertInt(t, result, 21)
	})

	t.Run("外部函数调用链串联", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("add1", func(x int) int { return x + 1 })
		ctx.BindFunc("mul2", func(x int) int { return x * 2 })
		result := runScriptWithContext(t,
			`#fn add1(int)=>int
#fn mul2(int)=>int
mul2(add1(mul2(5)))`,
			ctx,
		)
		// mul2(5)=10 -> add1(10)=11 -> mul2(11)=22
		assertInt(t, result, 22)
	})

	t.Run("外部函数在循环中调用", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("square", func(x int) int { return x * x })
		result := runScriptWithContext(t,
			`#fn square(int)=>int
sum := 0
for i := 1; i <= 5; i = i + 1 {
sum = sum + square(i)
}
sum`,
			ctx,
		)
		// 1+4+9+16+25=55
		assertInt(t, result, 55)
	})

	t.Run("外部函数在条件中调用", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("isEven", func(n int) bool { return n%2 == 0 })
		result := runScriptWithContext(t,
			`#fn isEven(int)=>bool
if isEven(8) { 1 } else { 0 }`,
			ctx,
		)
		assertInt(t, result, 1)
	})

	t.Run("外部函数返回数组后遍历", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("range1toN", func(n int) []int {
			result := make([]int, n)
			for i := 0; i < n; i++ {
				result[i] = i + 1
			}
			return result
		})
		result := runScriptWithContext(t,
			`#fn range1toN(int)=>arr
arr := range1toN(4)
sum := 0
for v := range arr {
sum = sum + v
}
sum`,
			ctx,
		)
		// 1+2+3+4=10
		assertInt(t, result, 10)
	})

	t.Run("外部函数返回 Map 后访问", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("userInfo", func() map[string]interface{} {
			return map[string]interface{}{
				"name": "Alice",
				"age":  30,
			}
		})
		result := runScriptWithContext(t,
			`#fn userInfo()=>any
m := userInfo()
m["name"]`,
			ctx,
		)
		assertString(t, result, "Alice")
	})

	t.Run("外部函数作为累加器状态", func(t *testing.T) {
		ctx := NewContext()
		total := 0
		ctx.BindFunc("accumulate", func(x int) int {
			total += x
			return total
		})
		result := runScriptWithContext(t,
			`#fn accumulate(int)=>int
accumulate(10)
accumulate(20)
accumulate(30)`,
			ctx,
		)
		// 10 -> 30 -> 60
		assertInt(t, result, 60)
	})
}

// ========== 外部函数 - 错误场景测试 ==========

func Test_ExternalDeep_Func_Errors(t *testing.T) {
	t.Run("参数类型不匹配", func(t *testing.T) {
		// 向int参数传string, convertStringValue忽略目标类型直接返回string
		// reflect.Call会因为类型不匹配而panic
		fn := func(x int) int { return x }
		defer func() {
			if r := recover(); r == nil {
				t.Error("期望类型不匹配产生panic")
			}
		}()
		callExternalFunc(fn, []Value{NewValue("not an int")})
	})

	t.Run("参数数量不足", func(t *testing.T) {
		fn := func(a, b int) int { return a + b }
		_, err := callExternalFunc(fn, []Value{NewValue(1)})
		assertError(t, err)
	})

	t.Run("参数数量过多", func(t *testing.T) {
		fn := func(a int) int { return a }
		_, err := callExternalFunc(fn, []Value{NewValue(1), NewValue(2)})
		assertError(t, err)
	})

	t.Run("函数内部 panic", func(t *testing.T) {
		fn := func(x int) int { panic("boom") }
		defer func() {
			if r := recover(); r == nil {
				t.Error("期望panic传播")
			}
		}()
		callExternalFunc(fn, []Value{NewValue(1)})
	})

	t.Run("返回 error 通过引擎", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("fail", func() (int, error) {
			return 0, errors.New("test error")
		})
		runExpectRuntimeError(t,
			`#fn fail()=>int
fail()`,
			ctx,
		)
	})

	t.Run("绑定非函数值调用", func(t *testing.T) {
		// BindValue绑定的是值, #fn查找bindFuncs找不到
		ctx := NewContext()
		ctx.BindValue("notfunc", 42)
		runExpectRuntimeError(t,
			`#fn notfunc()=>int
notfunc()`,
			ctx,
		)
	})
}

// ========== 绑定值修改语义测试 ==========

func Test_ExternalDeep_BindValue_ModifySemantics(t *testing.T) {
	t.Run("修改绑定数组不影响Go端", func(t *testing.T) {
		data := []int{1, 2, 3}
		result := runScriptWithBinds(t,
			`arr :=>arr getBindValue("data")
arr[0] = 100
arr[0]`,
			map[string]interface{}{"data": data},
		)
		assertInt(t, result, 100)
		// Go端数据不受影响: 转换时创建了新的[]Value
		if data[0] != 1 {
			t.Errorf("Go端数据不应被修改, 得到 data[0]=%d", data[0])
		}
	})

	t.Run("修改绑定 Map 不影响Go端", func(t *testing.T) {
		data := map[string]int{"a": 1, "b": 2}
		result := runScriptWithBinds(t,
			`m :=>any getBindValue("data")
m["a"] = 100
m["a"]`,
			map[string]interface{}{"data": data},
		)
		assertInt(t, result, 100)
		// Go端数据不受影响: 转换时创建了新的map[string]Value
		if data["a"] != 1 {
			t.Errorf("Go端数据不应被修改, 得到 data[a]=%d", data["a"])
		}
	})

	t.Run("push 绑定数组不影响原数组", func(t *testing.T) {
		data := []int{1, 2, 3}
		result := runScriptWithBinds(t,
			`arr :=>arr getBindValue("data")
newArr := push(arr, 4)
len(arr)`,
			map[string]interface{}{"data": data},
		)
		// push返回新数组, 原数组长度不变
		assertInt(t, result, 3)
		if len(data) != 3 {
			t.Errorf("Go端数组长度不应变化, 得到 len=%d", len(data))
		}
	})
}

// ========== 多函数协作测试 ==========

func Test_ExternalDeep_MultiFunc_Collaboration(t *testing.T) {
	t.Run("绑定多个函数协同工作", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("double", func(x int) int { return x * 2 })
		ctx.BindFunc("add", func(a, b int) int { return a + b })
		result := runScriptWithContext(t,
			`#fn double(int)=>int
#fn add(int, int)=>int
add(double(5), double(3))`,
			ctx,
		)
		// 10 + 6 = 16
		assertInt(t, result, 16)
	})

	t.Run("函数间数据传递", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("toStr", func(n int) string { return string(rune('0' + n)) })
		ctx.BindFunc("concat", func(a, b string) string { return a + b })
		result := runScriptWithContext(t,
			`#fn toStr(int)=>string
#fn concat(string, string)=>string
concat(toStr(1), toStr(2))`,
			ctx,
		)
		assertString(t, result, "12")
	})

	t.Run("脚本函数和外部函数混合调用", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("external", func(x int) int { return x * 10 })
		result := runScriptWithContext(t,
			`#fn external(int)=>int
fn compute(a, b) {
external(a) + external(b)
}
compute(3, 4)`,
			ctx,
		)
		// 30 + 40 = 70
		assertInt(t, result, 70)
	})
}
