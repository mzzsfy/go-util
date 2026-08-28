package script

import (
	"fmt"
	"testing"
)

// makeAdder 构造同代码指针、不同捕获变量的闭包
// 同一闭包字面量生成的实例共享代码指针, 触发以指针为 key 的缓存串调
func makeAdder(base int) func(int) int {
	return func(x int) int {
		return x + base
	}
}

// adderHost 方法值宿主, 不同实例的同名方法值共享代码指针
type adderHost struct {
	base int
}

func (h adderHost) add(x int) int {
	return x + h.base
}

// Test_external_ClosureDistinctInstances 不同闭包实例不得串调
// Given 两个共享代码指针、捕获不同变量的闭包绑定到不同名字
// When 脚本分别调用两个名字
// Then 各自按自己的捕获变量计算
func Test_external_ClosureDistinctInstances(t *testing.T) {
	parser := NewParser()
	compiled, err := parser.Compile("#fn f1(x)\n#fn f2(x)\nf1(1) + f2(1) * 100")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	ctx.BindFunc("f1", makeAdder(10))
	ctx.BindFunc("f2", makeAdder(20))

	engine := NewEngine()
	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	// 11 + 21*100 = 2111; 串调时 f2 命中 f1 的缓存, 结果变为 11 + 11*100 = 1111
	if result.Int() != 2111 {
		t.Fatalf("闭包串调, 期望 2111, 得到 %d", result.Int())
	}
}

// Test_external_MethodValueDistinctReceivers 不同接收者的方法值不得串调
func Test_external_MethodValueDistinctReceivers(t *testing.T) {
	parser := NewParser()
	compiled, err := parser.Compile("#fn m1(x)\n#fn m2(x)\nm1(1) * 100 + m2(1)")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	ctx.BindFunc("m1", adderHost{base: 10}.add)
	ctx.BindFunc("m2", adderHost{base: 20}.add)

	engine := NewEngine()
	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	// 11*100 + 21 = 1121; 串调时 m2 命中 m1 缓存, 结果变为 11*100 + 11 = 1111
	if result.Int() != 1121 {
		t.Fatalf("方法值串调, 期望 1121, 得到 %d", result.Int())
	}
}

// Test_external_SameNameReRegister 同名重复注册保持既有语义: 后注册覆盖前注册
func Test_external_SameNameReRegister(t *testing.T) {
	parser := NewParser()
	compiled, err := parser.Compile("#fn f(x)\nf(1)")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	ctx.BindFunc("f", func(x int) int { return x * 2 })
	ctx.BindFunc("f", func(x int) int { return x * 3 })

	engine := NewEngine()
	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	if result.Int() != 3 {
		t.Fatalf("同名重复注册应以后者为准, 期望 3, 得到 %d", result.Int())
	}
}

// Test_external_MixedArgTypes 混合类型参数调用回归
func Test_external_MixedArgTypes(t *testing.T) {
	fn := func(n int, s string, b bool, f float64) string {
		return fmt.Sprintf("%s:%d:%t:%d", s, n, b, int(f))
	}
	result, err := callExternalFunc(fn, []Value{
		NewValue(7), NewValue("v"), NewValue(true), NewValue(8.9),
	})
	if err != nil {
		t.Fatalf("调用失败: %v", err)
	}
	expected := "v:7:true:8"
	if result.String() != expected {
		t.Fatalf("期望 %q, 得到 %q", expected, result.String())
	}
}
