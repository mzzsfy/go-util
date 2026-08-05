package script

import (
	"errors"
	"testing"
	"time"
)

// Test_ErrCode_Timeout 超时返回 ErrTimeout 错误码
func Test_ErrCode_Timeout(t *testing.T) {
	parser := NewParser()
	compiled, err := parser.Compile("for {}")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}
	ctx := NewContext()
	engine := NewEngine(WithTimeout(10 * time.Millisecond))
	_, err = engine.Run(ctx, compiled)
	if err == nil {
		t.Fatal("期望超时错误")
	}
	if !errors.Is(err, &RuntimeError{Code: ErrTimeout}) {
		t.Errorf("errors.Is(err, &RuntimeError{Code: ErrTimeout}) 失败, 实际: %v", err)
	}
}

// Test_ErrCode_StepLimit 指令数超限返回 ErrStepLimit 错误码
func Test_ErrCode_StepLimit(t *testing.T) {
	parser := NewParser()
	compiled, err := parser.Compile("for {}")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}
	ctx := NewContext()
	engine := NewEngine(WithMaxSteps(100))
	_, err = engine.Run(ctx, compiled)
	if err == nil {
		t.Fatal("期望指令数超限错误")
	}
	if !errors.Is(err, &RuntimeError{Code: ErrStepLimit}) {
		t.Errorf("errors.Is(err, &RuntimeError{Code: ErrStepLimit}) 失败, 实际: %v", err)
	}
}

// Test_ErrCode_TypeMismatch 类型不匹配返回 ErrTypeMismatch 错误码
func Test_ErrCode_TypeMismatch(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"加法类型不匹配", `true + 5`},
		{"比较类型不匹配", `nil < 5`},
		{"数组索引非整数", `[1,2,3]["a"]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Eval(tt.code)
			if err == nil {
				t.Fatal("期望类型不匹配错误")
			}
			if !errors.Is(err, &RuntimeError{Code: ErrTypeMismatch}) {
				t.Errorf("errors.Is(err, &RuntimeError{Code: ErrTypeMismatch}) 失败, 实际: %v", err)
			}
		})
	}
}

// Test_ErrCode_IndexOutOfBounds 索引越界返回 ErrIndexOutOfBounds 错误码
func Test_ErrCode_IndexOutOfBounds(t *testing.T) {
	_, err := Eval("arr := [1,2,3]\narr[5] = 0")
	if err == nil {
		t.Fatal("期望索引越界错误")
	}
	if !errors.Is(err, &RuntimeError{Code: ErrIndexOutOfBounds}) {
		t.Errorf("errors.Is(err, &RuntimeError{Code: ErrIndexOutOfBounds}) 失败, 实际: %v", err)
	}
}

// Test_ErrCode_CallStackOverflow 调用栈溢出返回 ErrCallStackOverflow 错误码
func Test_ErrCode_CallStackOverflow(t *testing.T) {
	parser := NewParser()
	compiled, err := parser.Compile(`
		fn rec() {
			return rec()
		}
		rec()
	`)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}
	ctx := NewContext()
	engine := NewEngine(WithMaxCallDepth(10))
	_, err = engine.Run(ctx, compiled)
	if err == nil {
		t.Fatal("期望调用栈溢出错误")
	}
	if !errors.Is(err, &RuntimeError{Code: ErrCallStackOverflow}) {
		t.Errorf("errors.Is(err, &RuntimeError{Code: ErrCallStackOverflow}) 失败, 实际: %v", err)
	}
}

// Test_ErrCode_DivisionByZero 除零返回 ErrDivisionByZero 错误码
func Test_ErrCode_DivisionByZero(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"整数除零", `1 / 0`},
		{"浮点除零", `1.0 / 0.0`},
		{"整数模零", `5 % 0`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Eval(tt.code)
			if err == nil {
				t.Fatal("期望除零错误")
			}
			if !errors.Is(err, &RuntimeError{Code: ErrDivisionByZero}) {
				t.Errorf("errors.Is(err, &RuntimeError{Code: ErrDivisionByZero}) 失败, 实际: %v", err)
			}
		})
	}
}

// Test_ErrCode_UnsupportedOp 不支持的操作返回 ErrUnsupportedOp 错误码
func Test_ErrCode_UnsupportedOp(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"整数索引访问", `42[0]`},
		{"整数切片", `42[0:2]`},
		{"nil索引赋值", "x := 1\nx[0] = 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Eval(tt.code)
			if err == nil {
				t.Fatal("期望不支持的操作错误")
			}
			if !errors.Is(err, &RuntimeError{Code: ErrUnsupportedOp}) {
				t.Errorf("errors.Is(err, &RuntimeError{Code: ErrUnsupportedOp}) 失败, 实际: %v", err)
			}
		})
	}
}

// Test_ErrCode_Panic panic恢复返回 ErrPanic 错误码
func Test_ErrCode_Panic(t *testing.T) {
	parser := NewParser()
	compiled, err := parser.Compile("#fn panicFn()\npanicFn()")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}
	ctx := NewContext()
	ctx.BindFunc("panicFn", func() {
		panic("test panic")
	})
	engine := NewEngine()
	_, err = engine.Run(ctx, compiled)
	if err == nil {
		t.Fatal("期望 panic 恢复错误")
	}
	if !errors.Is(err, &RuntimeError{Code: ErrPanic}) {
		t.Errorf("errors.Is(err, &RuntimeError{Code: ErrPanic}) 失败, 实际: %v", err)
	}
}

// Test_ErrCode_Asr 使用 errors.As 提取 *RuntimeError
func Test_ErrCode_Asr(t *testing.T) {
	_, err := Eval(`1 / 0`)
	if err == nil {
		t.Fatal("期望错误")
	}
	var re *RuntimeError
	if !errors.As(err, &re) {
		t.Fatalf("errors.As(err, &re) 失败, 实际类型: %T", err)
	}
	if re.Code != ErrDivisionByZero {
		t.Errorf("期望 Code=%s, 实际 Code=%s", ErrDivisionByZero, re.Code)
	}
}

// Test_ErrCode_Is_EmptyCode 空Code的RuntimeError不匹配任何错误
func Test_ErrCode_Is_EmptyCode(t *testing.T) {
	err := &RuntimeError{Message: "test"}
	// 空 Code 不应匹配
	if errors.Is(err, &RuntimeError{Code: ErrTimeout}) {
		t.Error("空 Code 的 RuntimeError 不应匹配 ErrTimeout")
	}
	// 空 Code target 也不应匹配
	if errors.Is(err, &RuntimeError{}) {
		t.Error("空 Code target 不应匹配")
	}
}
