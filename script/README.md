# Script - Go 嵌入式脚本语言

Script 是为 Go 应用程序设计的嵌入式脚本语言，提供动态脚本执行、变量绑定、外部函数调用和完整的控制流语句。

## 安装

```bash
go get github.com/mzzsfy/go-util/script
```

## 快速开始

<details>
<summary>简单示例</summary>

```go
package main

import (
    "fmt"
    "github.com/mzzsfy/go-util/script"
)

func main() {
    // 单行脚本执行
    result, err := script.Eval("10 + 20")
    if err != nil {
        panic(err)
    }
    fmt.Println(result.Int()) // 输出: 30
}
```

</details>

## 目录

- [概述](#概述)
- [类型系统](#类型系统)
- [基础语法](#基础语法)
- [运算符](#运算符)
- [数据结构](#数据结构)
- [控制流](#控制流)
- [函数](#函数)
- [内置函数](#内置函数)
- [Go 绑定](#go-绑定)
- [便捷函数](#便捷函数)
- [标准用法](#标准用法)
- [高级示例](#高级示例)
- [API 参考](#api-参考)
- [最佳实践](#最佳实践)

---

## 概述

### 设计目标

1. **简洁性**：语法清晰，易于学习和使用
2. **安全性**：强类型检查，详细的运行时错误提示
3. **高性能**：编译为字节码，由虚拟机执行
4. **可扩展性**：支持外部函数绑定，易于集成
5. **并发安全**：线程安全的上下文和执行引擎

<details>
<summary>架构</summary>

```
源代码 → 词法分析 → 语法分析 → 编译器 → 字节码
                                              ↓
                                      虚拟机执行 ← 上下文
```

**核心组件**：
- **解析器 (Parser)**：将源代码编译为字节码
- **编译器 (Compiler)**：生成优化的字节码
- **虚拟机 (VM)**：执行字节码指令
- **上下文 (Context)**：管理变量和函数绑定
- **引擎 (Engine)**：协调虚拟机和上下文

</details>

---

## 类型系统

Script 是一种动态类型语言，具有运行时类型检查：

| 类型 | 别名 | 描述 | 字面量示例 |
|------|------|-------------|----------|
| `int` | `i` | 整数 | `42`, `-10`, `0xFF`, `10_000` |
| `float` | `f` | 浮点数 | `3.14`, `-1.5`, `0.0` |
| `string` | `str`, `s` | 字符串 | `"hello"`, `"world"` |
| `bool` | `b` | 布尔值 | `true`, `false` |
| `nil` | - | 空值 | `nil` |
| `array` | `arr` | 数组 | `[1, 2, 3]` |
| `map` | - | 映射 | `{"key": "value"}` |
| `any` | - | 任意类型 | 动态类型值 |

### 类型注意事项

- **数字支持下划线**：`10_000`, `1_000_000`, `0xFF_FF_FF`
- **十六进制支持**：`0xFF` (255), `0x1F` (31)
- **科学计数法**：字面量不支持 `1e10`（会被解析为变量），但 `float("1e10")` 可正确解析

---

## 基础语法

<details>
<summary>注释与标识符</summary>

### 注释
```go
// 单行注释

/*
  多行注释
  支持多行文本
*/
```

### 标识符
标识符由字母、数字和下划线组成，不能以数字开头。

```
有效: myVar, _temp, value1, userName
无效: 1value, my-var, @var
```

</details>

<details>
<summary>变量声明</summary>

### 隐式声明 (`:=`)
类型可被后续赋值覆盖，编译时推断类型。

```go
x := 10              // int
y := 3.14            // float
name := "Alice"      // string
flag := true         // bool
arr := [1, 2, 3]     // array
m := {"a": 1}        // map
```

### 显式类型声明 (`:=>`)
类型固定，编译时检查。支持类型别名。

```go
// 基本类型
x :=>int 10              // 明确指定 int 类型
y :=>float 3.14          // 明确指定 float 类型
name :=>str "Alice"      // 使用别名 str
flag :=>bool true        // 布尔类型

// 数组类型
arr :=>arr [1, 2, "3"]       // 数组，不指定内部类型
arr :=>arr[int] [1, 2, 3]    // 指定元素类型为 int

// 映射类型
m :=>map {"a": 1, "b": "ok"}     // map，value 可混写类型
m :=>map[int] {"a": 1}           // 指定 value 类型为 int

// 任意类型
val :=>any getBindValue("x")     // 可接受任何类型
```

### 变量赋值
使用 `=` 对已声明的变量重新赋值。

```go
x := 10
x = 20              // 重新赋值
```

</details>

---

## 运算符

<details>
<summary>算术运算符</summary>

```go
a := 10 + 5     // 加法: 15
b := 10 - 5     // 减法: 5
c := 10 * 5     // 乘法: 50
d := 10 / 5     // 除法: 2
e := 10 % 3     // 取模: 1
```

</details>

<details>
<summary>比较运算符</summary>

```go
a := 10 == 10   // 等于: true
b := 10 != 5    // 不等于: true
c := 10 < 15    // 小于: true
d := 10 <= 10   // 小于等于: true
e := 10 > 5     // 大于: true
f := 10 >= 10   // 大于等于: true
```

</details>

<details>
<summary>逻辑运算符</summary>

```go
a := true && false   // 与: false
b := true || false   // 或: true
c := !true           // 非: false
```

</details>

<details>
<summary>位运算符</summary>

```go
a := 5 & 3      // 按位与: 1
b := 5 | 3      // 按位或: 7
c := 5 ^ 3      // 按位异或: 6
d := 5 << 1     // 左移: 10
e := 5 >> 1     // 右移: 2
```

</details>

---

## 数据结构

<details>
<summary>数组</summary>

```go
// 创建数组
arr := [1, 2, 3, 4, 5]

// 访问元素
first := arr[0]         // 1
last := arr[4]          // 5

// 修改元素
arr[0] = 10

// 切片
sub := arr[1:4]         // [2, 3, 4]
prefix := arr[:3]       // [1, 2, 3]
suffix := arr[2:]       // [3, 4, 5]

// 获取长度
len := len(arr)         // 5
```

</details>

<details>
<summary>映射</summary>

```go
// 创建映射
m := {
    "name": "Alice",
    "age": 30,
    "city": "Beijing"
}

// 访问值
name := m["name"]       // "Alice"

// 修改值
m["age"] = 31

// 添加新键
m["country"] = "China"

// 获取长度
len := len(m)           // 4
```

</details>

---

## 控制流

<details>
<summary>If 语句</summary>

```go
// 基本if
if x > 10 { "large" }

// 可选括号
if (x > 10) { "large" }

// if-else
if x > 10 { "large" } else { "small" }

// if-else if-else
if x > 15 {
    "very large"
} else if x > 10 {
    "large"
} else {
    "small"
}
```

**注意**：if 语句中的括号是可选的，两种形式都有效。

</details>

<details>
<summary>For 循环</summary>

Script 支持 5 种 for 循环，具有灵活的语法选项：

#### 1. 无限循环
```go
// for { }: 无限循环（必须使用 break 或 return 退出）
i := 0
for {
    print(i)
    i = i + 1
    if i >= 5 { break }  // 退出条件
}

// 可选括号
for ( ) {
    print("infinite")
}
```

#### 2. 条件循环（while）
```go
// for condition: 当条件为真时继续
i := 0
for i < 5 {
    print(i)
    i = i + 1
}

// 可选括号
i := 0
for (i < 5) {
    print(i)
    i = i + 1
}
```

#### 3. 计数器循环
```go
// for n: 循环 n 次（无变量）
for 5 {
    print("A")    // 输出: A, A, A, A, A
}

// for i := n: 循环 n 次，i 从 1 到 n
for i := 5 {
    print(i)    // 输出: 1, 2, 3, 4, 5
}

// 可选括号
for (i := 5) {
    print(i)
}
```

**注意**：
- `for n` 和 `for i := n` 都循环 n 次
- 如果 n 是数组或映射，自动变成范围循环
- 如需从 0 开始的计数器，使用标准 for 循环：`for i := 0; i < 5; i = i + 1 { }`

#### 4. 标准 For 循环
```go
// 完整形式: for init; cond; post
for i := 0; i < 10; i = i + 1 {
    print(i)
}

// 可选括号
for (i := 0; i < 10; i = i + 1) {
    print(i)
}

// 省略部分 - 支持所有变体:
for i := 0; ; i = i + 1 {     // 无条件（无限）
    if i >= 10 { break }
}

for ; i < 10; i = i + 1 {     // 无初始化
    print(i)
}

for i := 0; i < 10; {         // 无后置
    print(i)
    i = i + 1
}

for ; ; {                     // 全部省略（无限）
    print("infinite")
}

for ( ; ; ) {                 // 带括号（无限）
    print("infinite")
}
```

#### 5. 范围循环
```go
// 隐式范围（自动检测数组/映射）
arr := [10, 20, 30]
for v := arr {
    print(v)    // 输出: 10, 20, 30
}

// 显式 range 关键字（推荐，更清晰）
arr := [10, 20, 30]
for v := range arr {
    print(v)    // 输出: 10, 20, 30
}

// 遍历映射
m := {"a": 1, "b": 2}
for k := m {
    print(k)    // 输出: "a", "b"
}

// 显式 range 遍历映射
m := {"a": 1, "b": 2}
for k := range m {
    print(k)    // 输出: "a", "b"
}

// 可选括号
for (v := range arr) {
    print(v)
}
```

**建议**：遍历集合时显式使用 `range` 关键字以提高代码清晰度。

#### 总结表

| 循环类型 | 语法 | 可选括号 | 用途 |
|-----------|--------|---------------------|----------|
| 无限循环 | `for { }` | `for ( ) { }` | 事件循环、服务器 |
| While循环 | `for cond { }` | `for (cond) { }` | 未知迭代次数 |
| 计数器循环 | `for i := n { }` | `for (i := n) { }` | 固定迭代次数 |
| 标准循环 | `for init; cond; post { }` | `for (init; cond; post) { }` | 复杂逻辑 |
| 范围循环 | `for k := range arr { }` | `for (k := range arr) { }` | 集合遍历 |

</details>

<details>
<summary>Break、Continue 和 Return</summary>

```go
// break: 提前退出循环
for i := 10 {
    if i == 5 { break }  // 当 i=5 时退出
    print(i)
}

// break 在无限循环中
i := 0
for {
    print(i)
    i = i + 1
    if i >= 10 { break }  // 必须使用 break 退出无限循环
}

// continue: 跳过当前迭代
for i := 10 {
    if i % 2 == 0 { continue }  // 跳过偶数
    print(i)  // 只打印奇数
}

// return: 函数返回
fn findFirst(arr, target) {
    for i := len(arr) {
        if arr[i] == target {
            return i    // 找到后立即返回
        }
    }
    return -1            // 未找到返回 -1
}
```

</details>

---

## 函数

<details>
<summary>函数定义</summary>

```go
// 无参数
fn greet() {
    "Hello"
}

// 带参数
fn add(a, b) {
    a + b
}

// 带返回语句
fn factorial(n) {
    if n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}
```

### 函数调用
```go
result := add(10, 20)       // 30
fact := factorial(5)        // 120
```

</details>

---

## 内置函数

<details>
<summary>类型转换</summary>

```go
int("123")           // 123
int(3.14)            // 3
float("3.14")        // 3.14
float(42)            // 42.0
string(123)          // "123"
```

</details>

<details>
<summary>类型检查</summary>

```go
typeof(123)          // "int"
typeof("hello")      // "string"
typeof(true)         // "bool"
typeof(3.14)         // "float"
typeof(nil)          // "nil"
typeof([1, 2, 3])    // "array"
typeof({"a": 1})     // "map"
```

</details>

<details>
<summary>长度与输出</summary>

```go
// 长度
len([1,2,3])         // 3
len("hello")         // 5
len({"a": 1})        // 1

// 输出
print("hello")
println("world")

// 删除键
m := {"a": 1, "b": 2}
delete(m, "a")       // 从 m 中删除键 "a"
// m 现在为 {"b": 2}

// 向数组追加元素
arr := [1, 2, 3]
result := push(arr, 4)  // [1, 2, 3, 4]
// 注意: push 返回新数组，不修改原数组
```

</details>

---

## Go 绑定

脚本可以通过 `Context` 绑定访问 Go 代码中的值和函数。

### 绑定值 (BindValue)

使用 `getBindValue()` 访问，必须**显式指定类型**，否则编译报错。

```go
// Go 端
ctx := script.NewContext()
ctx.BindValue("name", "Alice")
ctx.BindValue("age", 30)
ctx.BindValue("items", []int{1, 2, 3})

// 脚本端 - 使用 getBindValue() 访问，必须指定类型
name :=>str getBindValue("name")    // "Alice"
age :=>int getBindValue("age")      // 30
items :=>arr getBindValue("items")  // [1, 2, 3]
```

### 绑定函数 (BindFunc)

脚本中需要使用 `#fn` 进行预定义，否则编译报错。

```go
// Go 端 - 绑定函数
ctx.BindFunc("double", func(x int) int {
    return x * 2
})

ctx.BindFunc("greet", func(name string) string {
    return "Hello, " + name
})

// 脚本端 - 调用绑定的函数
// 需要使用 #fn 进行预定义
#fn double(int)=>int      // 预定义签名
#fn greet(s)=>s           // 使用类型别名

result := double(5)       // 10
greeting := greet("Bob")  // "Hello, Bob"
```

**签名格式**：`#fn 函数名(参数类型)=>返回类型`
- 参数类型和返回类型支持别名（如 `s` = `string`, `i` = `int`）
- 多参数：`#fn add(int, int)=>int`

---

## 便捷函数

Script 为常见用例提供便捷函数。

### 核心函数

```go
// Eval - 快速脚本执行
result, err := script.Eval("10 + 20")

// EvalWithBindings - 使用变量绑定执行
bindings := map[string]interface{}{"x": 10, "y": 20}
result, err := script.EvalWithBindings(`
    vx := getBindValue("x")
    vy := getBindValue("y")
    vx + vy
`, bindings)

// MustEval - 执行脚本，出错时 panic（用于初始化）
config := script.MustEval(`m := {"timeout": 30}; m`)
```

### 使用场景

| 场景 | 推荐函数 |
|----------|---------------------|
| 一次性执行 | `Eval` |
| 带外部数据 | `EvalWithBindings` |
| 初始化（不会失败） | `MustEval` |
| 复用脚本 | [标准方式](#标准用法) |
| 绑定外部函数 | [标准方式](#标准用法) |

<details>
<summary>完整便捷函数列表</summary>

```go
// Eval - 快速脚本执行
result, err := script.Eval(source string) (Value, error)

// EvalWithBindings - 使用变量绑定执行
result, err := script.EvalWithBindings(source string, bindings map[string]interface{}) (Value, error)

// MustEval - 执行脚本，出错时 panic
result := script.MustEval(source string) Value

// MustEvalWithBindings - 使用绑定执行，出错时 panic
result := script.MustEvalWithBindings(source string, bindings map[string]interface{}) Value
```

**参数**：
- `source`: 脚本源代码字符串
- `bindings`: 变量绑定映射，键为变量名（字符串），值为任意 Go 值

**返回值**：
- `Value`: 脚本执行结果
- `error`: 编译或运行时错误

</details>

---

## 标准用法

使用 Parser、Engine 和 Context 进行更精细的控制：

```go
parser := script.NewParser()
engine := script.NewEngine()
ctx := script.NewContext()

// 绑定值
ctx.BindValue("x", 10)
ctx.BindValue("y", 20)

// 绑定外部函数
ctx.BindFunc("double", func(x int) int {
    return x * 2
})

// 一次编译，多次执行
// ⚠️ 重要：外部变量需要预定义或使用 any 类型
script, _ := parser.Compile(`
    #fn double(int)=>int    // 预定义外部函数签名
    vx :=>int getBindValue("x")  // 显式类型声明
    vy :=>int getBindValue("y")  // 显式类型声明
    double(vx + vy)
`)

for i := 0; i < 10; i++ {
    result, _ := engine.Run(ctx, script)
    fmt.Println(result.Int())
}
```

---

## 高级示例

<details>
<summary>示例 1: 配置处理器</summary>

```go
package main

import (
    "fmt"
    "github.com/mzzsfy/go-util/script"
)

func main() {
    config := map[string]interface{}{
        "database": map[string]interface{}{
            "host": "localhost",
            "port": 5432,
            "name": "mydb",
        },
        "features": map[string]interface{}{
            "cache":   true,
            "logging": false,
        },
    }

    result, _ := script.EvalWithBindings(`
        cfg := getBindValue("config")
        db := cfg["database"]
        connStr := db["host"] + ":" + string(db["port"]) + "/" + db["name"]
        connStr
    `, map[string]interface{}{"config": config})

    fmt.Printf("连接字符串: %s\n", result.String())
    // 输出: 连接字符串: localhost:5432/mydb
}
```

</details>

<details>
<summary>示例 2: 数据处理管道</summary>

```go
package main

import (
    "fmt"
    "github.com/mzzsfy/go-util/script"
)

func main() {
    ctx := script.NewContext()

    // 绑定外部函数
    ctx.BindFunc("filter", func(arr []int, predicate func(int) bool) []int {
        result := []int{}
        for _, v := range arr {
            if predicate(v) {
                result = append(result, v)
            }
        }
        return result
    })

    ctx.BindFunc("map", func(arr []int, mapper func(int) int) []int {
        result := []int{}
        for _, v := range arr {
            result = append(result, mapper(v))
        }
        return result
    })

    data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    ctx.BindValue("data", data)

    parser := script.NewParser()
    engine := script.NewEngine()

    script, _ := parser.Compile(`
        arr := getBindValue("data")

        fn isEven(n) { return n % 2 == 0 }
        fn square(n) { return n * n }

        filtered := filter(arr, isEven)
        mapped := map(filtered, square)

        sum := 0
        for v := mapped {
            sum = sum + v
        }
        sum
    `)

    result, _ := engine.Run(ctx, script)
    fmt.Printf("偶数平方和: %d\n", result.Int())
    // 计算: 2^2 + 4^2 + 6^2 + 8^2 + 10^2 = 220
}
```

</details>

<details>
<summary>示例 3: 递归算法 - 快速排序</summary>

```go
package main

import (
    "fmt"
    "github.com/mzzsfy/go-util/script"
)

func main() {
    result, _ := script.Eval(`
        fn quickSort(arr) {
            if len(arr) <= 1 { return arr }

            pivot := arr[0]
            less := []
            equal := []
            greater := []

            for v := arr {
                if v < pivot {
                    less = push(less, v)
                } else if v == pivot {
                    equal = push(equal, v)
                } else {
                    greater = push(greater, v)
                }
            }

            sortedLess := quickSort(less)
            sortedGreater := quickSort(greater)

            result := []
            for v := sortedLess { result = push(result, v) }
            for v := equal { result = push(result, v) }
            for v := sortedGreater { result = push(result, v) }

            return result
        }

        arr := [64, 34, 25, 12, 22, 11, 90]
        quickSort(arr)
    `)

    sorted := result.Array()
    fmt.Print("排序结果: ")
    for i, elem := range sorted.Elements {
        if i > 0 { fmt.Print(", ") }
        fmt.Print(elem.Int())
    }
    fmt.Println()
    // 输出: 排序结果: 11, 12, 22, 25, 34, 64, 90
}
```

</details>

---

## API 参考

### 便捷函数

```go
// Eval - 快速脚本执行
result, err := script.Eval(source string) (Value, error)

// EvalWithBindings - 使用变量绑定执行
result, err := script.EvalWithBindings(source string, bindings map[string]interface{}) (Value, error)

// MustEval - 执行脚本，出错时 panic
result := script.MustEval(source string) Value
```

<details>
<summary>Parser API</summary>

```go
// 创建解析器
parser := script.NewParser()

// 编译脚本
compiled, err := parser.Compile(source string) (*CompiledScript, error)
```

</details>

<details>
<summary>Context API</summary>

```go
// 创建上下文
ctx := script.NewContext()

// 绑定值（自动类型转换）
ctx.BindValue(name string, value any)

// 绑定函数
ctx.BindFunc(name string, fn interface{})

// 获取绑定的值
val, ok := ctx.GetBindValue(name string) (Value, bool)

// 获取绑定的函数
fn, ok := ctx.GetBindFunc(name string) (interface{}, bool)

// 克隆上下文
newCtx := ctx.Clone() *Context

// 设置输出
ctx.SetOutput(w io.Writer)

// 获取执行统计
stats := ctx.GetStats() ExecStats
```

</details>

<details>
<summary>Engine API</summary>

```go
// 创建引擎
engine := script.NewEngine()

// 使用选项创建引擎
engine := script.NewEngine(
    script.WithMaxCallDepth(1000),
)

// 执行脚本
result, err := engine.Run(ctx *Context, script *CompiledScript) (Value, error)
```

</details>

<details>
<summary>Value API</summary>

```go
// 创建值
val := script.NewValue(data any) Value

// 类型检查
val.Type == script.TypeInt
val.Type == script.TypeString
val.IsNil() bool

// 类型转换
val.Int() int
val.Float() float64
val.String() string
val.Bool() bool
val.Array() *ArrayValue
val.Map() *MapValue
val.Function() *FunctionValue
```

</details>

---

## 最佳实践

### 核心原则

<details>
<summary>1. 变量命名</summary>

```go
// 推荐: 有意义的名称
userAge := 25
userName := "Alice"

// 不推荐: 单字母或无意义的名称
x := 25
y := "Alice"
```

</details>

<details>
<summary>2. 全局变量访问（重要）</summary>

**绑定的变量必须通过 getBindValue() 访问后才能使用**

```go
// ✅ 正确
bindings := map[string]interface{}{"arr": []int{1, 2, 3}}
result, _ := script.EvalWithBindings(`
    localArr := getBindValue("arr")
    localArr[0] = 100
    localArr
`, bindings)

// ❌ 错误 - arr 未定义
result, _ := script.EvalWithBindings(`
    arr[0] = 100
`, bindings)
```

**原因**：脚本编译会检查变量是否已定义。全局绑定的变量不在脚本作用域内。

</details>

<details>
<summary>3. 错误处理</summary>

```go
// 推荐: 始终检查错误
script, err := parser.Compile(source)
if err != nil {
    return fmt.Errorf("编译失败: %w", err)
}

result, err := engine.Run(ctx, script)
if err != nil {
    return fmt.Errorf("执行失败: %w", err)
}
```

</details>

<details>
<summary>4. 性能优化</summary>

```go
// 推荐: 编译一次，多次执行
script, _ := parser.Compile(complexScript)

for i := 0; i < 1000; i++ {
    ctx := script.NewContext()
    ctx.BindValue("input", data[i])
    result, _ := engine.Run(ctx, script)
}

// 不推荐: 每次重新编译
for i := 0; i < 1000; i++ {
    script, _ := parser.Compile(complexScript)  // 浪费性能
    engine.Run(ctx, script)
}
```

</details>

<details>
<summary>5. 上下文复用</summary>

```go
// 推荐: 使用 Clone 共享基础配置
baseCtx := script.NewContext()
baseCtx.BindValue("config", config)

for _, user := range users {
    ctx := baseCtx.Clone()
    ctx.BindValue("user", user)
    engine.Run(ctx, script)
}
```

</details>

<details>
<summary>6. 函数设计</summary>

```go
// 推荐: 单一职责
fn validateEmail(email) {
    return contains(email, "@")
}

fn validateAge(age) {
    return age >= 0 && age <= 150
}

// 不推荐: 混合职责
fn validate(email, age) {
    // 验证不同类型
}
```

</details>

<details>
<summary>7. 选择执行方式</summary>

```go
// 一次性执行 → 使用便捷函数
result, _ := script.Eval("10 + 20")

// 带数据 → 使用绑定函数
bindings := map[string]interface{}{"x": 10}
result, _ := script.EvalWithBindings(`
    vx := getBindValue("x")
    vx * 2
`, bindings)

// 需要复用 → 使用标准方式
parser := script.NewParser()
engine := script.NewEngine()
script, _ := parser.Compile("复杂计算...")
for i := 0; i < 1000; i++ {
    ctx := script.NewContext()
    ctx.BindValue("input", data[i])
    engine.Run(ctx, script)
}

// 需要外部函数 → 使用标准方式
ctx := script.NewContext()
ctx.BindFunc("customFunc", myGoFunc)
engine.Run(ctx, script)
```

</details>

<details>
<summary>8. 外部变量类型声明（重要）</summary>

**从 Go 绑定的变量必须显式声明类型**：

```go
// ❌ 错误 - 未指定类型
result, _ := script.EvalWithBindings(`
    x := getBindValue("x")  // 编译错误！
    x + 1
`, bindings)

// ✅ 正确 - 指定类型
result, _ := script.EvalWithBindings(`
    x :=>int getBindValue("x")  // 正确
    x + 1
`, bindings)

// ✅ 正确 - 使用 any 类型
result, _ := script.EvalWithBindings(`
    x :=>any getBindValue("x")  // 接受任何类型
    x
`, bindings)
```

**原因**：编译器需要知道变量的类型信息才能进行类型检查。

</details>

<details>
<summary>9. 外部函数预定义（重要）</summary>

**绑定的外部函数必须在脚本中预定义签名**：

```go
// Go 端绑定函数
ctx.BindFunc("double", func(x int) int {
    return x * 2
})

ctx.BindFunc("greet", func(name string) string {
    return "Hello, " + name
})
```

```go
// ❌ 错误 - 未预定义函数签名
result, _ := parser.Compile(`
    double(5)  // 编译错误：函数未定义！
`)

// ✅ 正确 - 预定义函数签名
result, _ := parser.Compile(`
    #fn double(int)=>int      // 预定义签名
    #fn greet(s)=>s           // 使用类型别名

    double(5)  // 正确
`)
```

**签名格式**：`#fn 函数名(参数类型)=>返回类型`
- 参数类型和返回类型支持别名（如 `s` = `string`, `i` = `int`）
- 多参数：`#fn add(int, int)=>int`

</details>

---

## 注意事项

<details>
<summary>类型系统</summary>

- **动态类型**：类型在运行时检查
- **自动转换**：某些类型会自动转换（如 int -> float）
- **强类型检查**：无隐式类型转换（如 string + int）

</details>

<details>
<summary>性能考虑</summary>

- **编译开销**：首次编译有性能成本，应缓存编译结果
- **虚拟机执行**：字节码执行比原生 Go 代码慢
- **并发安全**：Context 是线程安全的，可以并发使用
- **便捷函数**：每次调用都会重新编译，适合一次性执行

</details>

<details>
<summary>限制</summary>

**不支持**：
- 类和接口
- 泛型
- 并发（goroutine/channel）

**限制**：
- 递归深度：默认最大调用深度 256
- 内存：受 Go 运行时限制

</details>

<details>
<summary>已知限制</summary>

| 限制                     | 状态    | 替代方案                   |
|------------------------|-------|------------------------|
| 科学计数法 (`1e10`)         | 不支持   | 使用完整数字或 `float("1e10")` |
| 类和接口                   | 不支持   | -                      |
| 并发 (goroutine/channel) | 不支持   | -                      |
| 最大递归深度                 | 默认256 | 使用 WithMaxCallDepth 配置 |

</details>

---

## 许可证

MIT License
