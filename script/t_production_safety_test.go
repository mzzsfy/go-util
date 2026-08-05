package script

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// Test_PanicRecovery 外部函数panic被捕获转为RuntimeError
func Test_PanicRecovery(t *testing.T) {
	parser := NewParser()
	compiled, err := parser.Compile("#fn panicFn()\npanicFn()")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	ctx.BindFunc("panicFn", func() {
		panic("test panic from external func")
	})

	engine := NewEngine()
	_, err = engine.Run(ctx, compiled)
	if err == nil {
		t.Fatal("期望返回错误, 实际返回 nil")
	}

	re, ok := err.(*RuntimeError)
	if !ok {
		t.Fatalf("期望 *RuntimeError, 实际得到 %T", err)
	}

	if !strings.Contains(re.Message, "panic") {
		t.Errorf("错误消息应包含 'panic', 实际: %s", re.Message)
	}
}

// Test_MaxSteps 死循环被指令计数中断
func Test_MaxSteps(t *testing.T) {
	parser := NewParser()
	compiled, err := parser.Compile("for {}")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	engine := NewEngine(WithMaxSteps(1000))
	_, err = engine.Run(ctx, compiled)
	if err == nil {
		t.Fatal("期望返回错误, 实际返回 nil")
	}

	re, ok := err.(*RuntimeError)
	if !ok {
		t.Fatalf("期望 *RuntimeError, 实际得到 %T", err)
	}

	if !strings.Contains(re.Message, "指令数上限") {
		t.Errorf("错误消息应包含 '指令数上限', 实际: %s", re.Message)
	}
}

// Test_MaxSteps_Normal 正常脚本不受指令计数影响
func Test_MaxSteps_Normal(t *testing.T) {
	parser := NewParser()
	compiled, err := parser.Compile("1 + 2 + 3")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	engine := NewEngine(WithMaxSteps(100000))
	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("正常脚本不应出错: %v", err)
	}

	if result.Int() != 6 {
		t.Errorf("期望 6, 得到 %d", result.Int())
	}
}

// Test_Timeout 超时中断脚本, 返回 ErrTimeout 错误
func Test_Timeout(t *testing.T) {
	parser := NewParser()
	compiled, err := parser.Compile("for {}")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	engine := NewEngine(WithTimeout(10 * time.Millisecond))

	start := time.Now()
	_, err = engine.Run(ctx, compiled)
	elapsed := time.Since(start)

	// 超时应返回 ErrTimeout 错误
	if err == nil {
		t.Fatal("超时应返回错误, 实际返回 nil")
	}
	if !errors.Is(err, &RuntimeError{Code: ErrTimeout}) {
		t.Errorf("期望 errors.Is(err, &RuntimeError{Code: ErrTimeout}), 实际: %v", err)
	}

	// 验证在合理时间内返回而非死循环
	if elapsed > 2*time.Second {
		t.Errorf("超时机制未生效, 耗时 %v", elapsed)
	}
}

// Test_Timeout_Normal 正常脚本不受超时影响
func Test_Timeout_Normal(t *testing.T) {
	parser := NewParser()
	compiled, err := parser.Compile("1 + 2 + 3")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	ctx := NewContext()
	engine := NewEngine(WithTimeout(1 * time.Second))
	result, err := engine.Run(ctx, compiled)
	if err != nil {
		t.Fatalf("正常脚本不应出错: %v", err)
	}

	if result.Int() != 6 {
		t.Errorf("期望 6, 得到 %d", result.Int())
	}
}
