package script

import (
	"testing"
)

// Given 同一 engine 复用执行多次脚本
// When VM 从池中复用
// Then 每次执行结果一致, 步数计数不跨次残留导致误拦截
func Test_EngineReuse_StepCountNotLeak(t *testing.T) {
	t.Parallel()
	parser := NewParser()
	compiled, err := parser.Compile(`
sum := 0
for i := 1; i <= 100; i = i + 1 {
    sum = sum + i
}
sum`)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	engine := NewEngine(WithMaxSteps(10 * 1000))
	const want = 5050 // 1..100 累加
	for run := 0; run < 3; run++ {
		v, err := engine.Run(NewContext(), compiled)
		if err != nil {
			t.Fatalf("第 %d 次执行失败: %v", run+1, err)
		}
		if got := v.Int(); got != want {
			t.Fatalf("第 %d 次执行结果 = %d, want %d", run+1, got, want)
		}
	}
}

// Given 正常脚本指令数低于上限
// When WithMaxSteps 设置刚好充足的值
// Then 完整执行不被误拦截
func Test_MaxSteps_BoundaryAllowsNormalScript(t *testing.T) {
	t.Parallel()
	parser := NewParser()
	compiled, err := parser.Compile(`
sum := 0
for i := 0; i < 10; i = i + 1 {
    sum = sum + i
}
sum`)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	// 宽余量上限下正常脚本应成功
	engine := NewEngine(WithMaxSteps(100 * 1000))
	v, err := engine.Run(NewContext(), compiled)
	if err != nil {
		t.Fatalf("正常脚本不应被步数上限拦截: %v", err)
	}
	if got := v.Int(); got != 45 {
		t.Fatalf("sum = %d, want 45", got)
	}
}

// Given 默认上限
// When 无 WithMaxSteps 构造 engine 执行正常脚本
// Then 使用 DefaultMaxSteps 并成功
func Test_MaxSteps_DefaultApplies(t *testing.T) {
	t.Parallel()
	parser := NewParser()
	compiled, err := parser.Compile("1 + 2")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	engine := NewEngine()
	v, err := engine.Run(NewContext(), compiled)
	if err != nil {
		t.Fatalf("默认配置执行失败: %v", err)
	}
	if got := v.Int(); got != 3 {
		t.Fatalf("1+2 = %d, want 3", got)
	}
}
