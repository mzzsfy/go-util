package script_test

import (
	"fmt"
	"log"

	"github.com/mzzsfy/go-util/script"
)

// ========== 包级别示例 ==========

func Example() {
	// 最简单的使用方式
	result, err := script.Eval("10 + 20")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Int())
	// Output: 30
}

// ========== 便捷函数示例 ==========

func ExampleEval() {
	// 简单表达式计算
	result, err := script.Eval("10 + 20 * 3")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Int())
	// Output: 70
}

func ExampleEvalWithBindings() {
	// 带外部变量的脚本
	result, err := script.EvalWithBindings(`
		x :=>int getBindValue("x")
		y :=>int getBindValue("y")
		x + y
	`, map[string]interface{}{"x": 10, "y": 20})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Int())
	// Output: 30
}

func ExampleMustEval() {
	// 用于初始化，保证不会失败
	result := script.MustEval("10 + 20")
	fmt.Println(result.Int())
	// Output: 30
}

func ExampleMustEvalWithBindings() {
	// 带绑定的Must版本
	result := script.MustEvalWithBindings(`
		x :=>int getBindValue("x")
		y :=>int getBindValue("y")
		x * y
	`, map[string]interface{}{"x": 5, "y": 6})
	fmt.Println(result.Int())
	// Output: 30
}

// ========== Parser示例 ==========

func ExampleNewParser() {
	// 创建Parser用于编译脚本
	parser := script.NewParser()

	// 编译一次，可多次执行
	compiled, err := parser.Compile(`
		n :=>int getBindValue("n")
		n * n
	`)
	if err != nil {
		log.Fatal(err)
	}

	// 多次执行同一脚本
	engine := script.NewEngine()
	for _, n := range []int{2, 3, 4} {
		ctx := script.NewContext()
		ctx.BindValue("n", n)
		result, _ := engine.Run(ctx, compiled)
		fmt.Println(result.Int())
	}
	// Output:
	// 4
	// 9
	// 16
}

func ExampleParser_Compile() {
	parser := script.NewParser()

	// 编译复杂脚本
	compiled, err := parser.Compile(`
		arr := [1, 2, 3, 4, 5]
		sum := 0
		for v := arr {
			sum = sum + v
		}
		sum
	`)
	if err != nil {
		log.Fatal(err)
	}

	engine := script.NewEngine()
	ctx := script.NewContext()
	result, _ := engine.Run(ctx, compiled)
	fmt.Println(result.Int())
	// Output: 15
}

func ExampleParser_Validate() {
	parser := script.NewParser()

	// 仅验证语法，不生成代码
	err := parser.Validate("x := 10 + 20")
	if err != nil {
		fmt.Println("Syntax error:", err)
	} else {
		fmt.Println("Syntax OK")
	}
	// Output: Syntax OK
}

// ========== Engine示例 ==========

func ExampleNewEngine() {
	// 创建执行引擎
	engine := script.NewEngine()

	// 编译脚本
	compiled, _ := script.NewParser().Compile("10 + 20")

	// 执行脚本
	ctx := script.NewContext()
	result, _ := engine.Run(ctx, compiled)
	fmt.Println(result.Int())
	// Output: 30
}

func ExampleEngine_Run() {
	engine := script.NewEngine()
	compiled, _ := script.NewParser().Compile(`
		arr :=>arr getBindValue("arr")
		sum := 0
		for v := arr {
			sum = sum + v
		}
		sum
	`)

	// 执行多次，不同输入
	for _, arr := range [][]int{{1, 2, 3}, {4, 5, 6}} {
		ctx := script.NewContext()
		ctx.BindValue("arr", arr)
		result, _ := engine.Run(ctx, compiled)
		fmt.Println(result.Int())
	}
	// Output:
	// 6
	// 15
}

// ========== Context示例 ==========

func ExampleNewContext() {
	// 创建上下文
	ctx := script.NewContext()

	// 绑定变量
	ctx.BindValue("name", "Alice")
	ctx.BindValue("age", 30)

	// 执行脚本
	result, _ := script.EvalWithBindings(`
		n :=>string getBindValue("name")
		a :=>int getBindValue("age")
		n + " is " + string(a)
	`, map[string]interface{}{
		"name": "Alice",
		"age":  30,
	})
	fmt.Println(result.String())
	// Output: Alice is 30
}

func ExampleContext_BindValue() {
	ctx := script.NewContext()

	// 绑定各种类型的值
	ctx.BindValue("int", 42)
	ctx.BindValue("float", 3.14)
	ctx.BindValue("string", "hello")
	ctx.BindValue("bool", true)
	ctx.BindValue("array", []int{1, 2, 3})

	// 使用绑定值
	compiled, _ := script.NewParser().Compile(`
		arr :=>arr getBindValue("array")
		len(arr)
	`)
	result, _ := script.NewEngine().Run(ctx, compiled)
	fmt.Println(result.Int())
	// Output: 3
}

func ExampleContext_BindFunc() {
	ctx := script.NewContext()

	// 绑定Go函数到脚本
	ctx.BindFunc("double", func(n int) int {
		return n * 2
	})

	ctx.BindFunc("greet", func(name string) string {
		return "Hello, " + name
	})

	// 在脚本中调用绑定的函数
	compiled, _ := script.NewParser().Compile(`
		#fn double(int)=>int
		#fn greet(string)=>string
		double(21)
	`)

	result, _ := script.NewEngine().Run(ctx, compiled)
	fmt.Println(result.Int())
	// Output: 42
}

func ExampleContext_GetBindValue() {
	ctx := script.NewContext()
	ctx.BindValue("config", map[string]interface{}{
		"host": "localhost",
		"port": 8080,
	})

	// 获取绑定的值
	val, ok := ctx.GetBindValue("config")
	if ok {
		fmt.Println(val.Map().Pairs["host"].String())
	}
	// Output: localhost
}

func ExampleContext_Clone() {
	// 创建基础上下文
	base := script.NewContext()
	base.BindValue("prefix", "Hello")

	// 编译脚本
	compiled, _ := script.NewParser().Compile(`
		p :=>string getBindValue("prefix")
		n :=>string getBindValue("name")
		p + ", " + n + "!"
	`)

	engine := script.NewEngine()

	// 克隆上下文，每个用户独立的name
	for _, name := range []string{"Alice", "Bob", "Charlie"} {
		ctx := base.Clone()
		ctx.BindValue("name", name)
		result, _ := engine.Run(ctx, compiled)
		fmt.Println(result.String())
	}
	// Output:
	// Hello, Alice!
	// Hello, Bob!
	// Hello, Charlie!
}

// ========== Value示例 ==========

func ExampleValue() {
	// 从脚本获取数组结果
	result, _ := script.Eval(`[1, 2, 3, 4, 5]`)

	// 访问数组值
	arr := result.Array()
	fmt.Println(arr.Elements[0].Int())
	fmt.Println(len(arr.Elements))
	// Output:
	// 1
	// 5
}

func ExampleValue_Int() {
	result, _ := script.Eval("42")
	fmt.Println(result.Int())
	// Output: 42
}

func ExampleValue_Float() {
	result, _ := script.Eval("3.14")
	fmt.Println(result.Float())
	// Output: 3.14
}

func ExampleValue_String() {
	result, _ := script.Eval(`"hello world"`)
	fmt.Println(result.String())
	// Output: hello world
}

func ExampleValue_Bool() {
	result, _ := script.Eval("true")
	fmt.Println(result.Bool())
	// Output: true
}

func ExampleValue_Array() {
	result, _ := script.Eval("[1, 2, 3, 4, 5]")
	arr := result.Array()
	for i, elem := range arr.Elements {
		fmt.Printf("arr[%d] = %d\n", i, elem.Int())
	}
	// Output:
	// arr[0] = 1
	// arr[1] = 2
	// arr[2] = 3
	// arr[3] = 4
	// arr[4] = 5
}

func ExampleValue_Map() {
	// 通过绑定传递Map
	m := map[string]interface{}{"a": 1, "b": 2}
	result, _ := script.EvalWithBindings(`
		m :=>any getBindValue("m")
		m["a"]
	`, map[string]interface{}{"m": m})

	fmt.Println(result.Int())
	// Output: 1
}

// ========== 数据结构示例 ==========

func Example_arrayOperations() {
	// 数组访问
	result, _ := script.Eval(`
		arr := [1, 2, 3, 4, 5]
		arr[2]
	`)
	fmt.Println(result.Int())
	// Output: 3
}

func Example_mapOperations() {
	// 通过绑定传递Map
	m := map[string]interface{}{"a": 1, "b": 2, "c": 3}
	result, _ := script.EvalWithBindings(`
		m :=>any getBindValue("m")
		m["b"]
	`, map[string]interface{}{"m": m})
	fmt.Println(result.Int())
	// Output: 2
}

// ========== 控制流示例 ==========

func Example_ifStatement() {
	result, _ := script.Eval(`
		x := 15
		result := ""
		if x > 10 {
			result = "large"
		} else {
			result = "small"
		}
		result
	`)
	fmt.Println(result.String())
	// Output: large
}

func Example_forLoop() {
	result, _ := script.Eval(`
		sum := 0
		for i := 5 {
			sum = sum + i
		}
		sum
	`)
	// i 从 1 到 5：1 + 2 + 3 + 4 + 5 = 15
	fmt.Println(result.Int())
	// Output: 15
}

func Example_rangeLoop() {
	result, _ := script.Eval(`
		arr := [10, 20, 30]
		sum := 0
		for v := range arr {
			sum = sum + v
		}
		sum
	`)
	fmt.Println(result.Int())
	// Output: 60
}

// ========== 外部函数示例 ==========

func Example_externalFunction() {
	// 使用外部绑定函数（推荐方式）
	ctx := script.NewContext()
	ctx.BindFunc("multiply", func(a, b int) int {
		return a * b
	})

	compiled, _ := script.NewParser().Compile(`
		#fn multiply(int, int)=>int
		multiply(6, 7)
	`)

	result, _ := script.NewEngine().Run(ctx, compiled)
	fmt.Println(result.Int())
	// Output: 42
}

func Example_externalFunctionMultiple() {
	// 多个外部函数协同工作
	ctx := script.NewContext()
	ctx.BindFunc("double", func(n int) int {
		return n * 2
	})
	ctx.BindFunc("add", func(a, b int) int {
		return a + b
	})

	compiled, _ := script.NewParser().Compile(`
		#fn double(int)=>int
		#fn add(int, int)=>int
		x := double(15)
		y := add(x, 12)
		y
	`)

	result, _ := script.NewEngine().Run(ctx, compiled)
	fmt.Println(result.Int())
	// Output: 42
}

// ========== 错误处理示例 ==========

func Example_errorHandling() {
	// 编译错误
	_, err := script.Eval("x := ")
	if err != nil {
		fmt.Println("Compilation error")
	}

	// 运行时错误
	_, err = script.Eval("1 / 0")
	if err != nil {
		fmt.Println("Runtime error")
	}
	// Output:
	// Compilation error
	// Runtime error
}

// ========== 实际应用示例 ==========

func Example_configProcessing() {
	// 处理配置文件
	config := map[string]interface{}{
		"database": map[string]interface{}{
			"host": "localhost",
			"port": 5432,
		},
	}

	result, _ := script.EvalWithBindings(`
		cfg :=>any getBindValue("config")
		db := cfg["database"]
		db["host"] + ":" + string(db["port"])
	`, map[string]interface{}{"config": config})

	fmt.Println(result.String())
	// Output: localhost:5432
}

func Example_dataTransformation() {
	// 数据转换：计算数组元素之和
	data := []int{1, 2, 3, 4, 5}
	ctx := script.NewContext()
	ctx.BindValue("data", data)

	compiled, _ := script.NewParser().Compile(`
		arr :=>arr getBindValue("data")
		sum := 0
		for v := arr {
			sum = sum + v
		}
		sum
	`)

	result, _ := script.NewEngine().Run(ctx, compiled)
	fmt.Println(result.Int())
	// Output: 15
}
