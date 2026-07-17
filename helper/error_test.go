package helper

import (
	"testing"
)

// TestStringError_Error 测试 Error 方法
func TestStringError_Error(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		err     StringError
		wantMsg string
	}{
		{"普通错误消息", "something went wrong", "something went wrong"},
		{"空字符串错误", "", ""},
		{"中文错误消息", "发生错误", "发生错误"},
		{"特殊字符", "error: [test]", "error: [test]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("StringError.Error() = %q, 期望 %q", got, tt.wantMsg)
			}
		})
	}
}

// TestStringError_String 测试 String 方法
func TestStringError_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		err     StringError
		wantMsg string
	}{
		{"普通消息", "test error", "test error"},
		{"空字符串", "", ""},
		{"Unicode 消息", "测试错误", "测试错误"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.String(); got != tt.wantMsg {
				t.Errorf("StringError.String() = %q, 期望 %q", got, tt.wantMsg)
			}
		})
	}
}

// TestStringError_ErrorAndStringEquivalent 测试 Error 和 String 返回相同值
func TestStringError_ErrorAndStringEquivalent(t *testing.T) {
	t.Parallel()
	err := StringError("test message")
	if err.Error() != err.String() {
		t.Errorf("Error() = %q, String() = %q, 应该相等", err.Error(), err.String())
	}
}

// TestNewError 测试 NewError 函数
func TestNewError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		msg     string
		wantMsg string
	}{
		{"普通错误", "test error", "test error"},
		{"空字符串", "", ""},
		{"中文错误", "测试", "测试"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewError(tt.msg)
			if err == nil {
				t.Error("NewError 不应返回 nil")
				return
			}
			if err.Error() != tt.wantMsg {
				t.Errorf("NewError(%q).Error() = %q, 期望 %q", tt.msg, err.Error(), tt.wantMsg)
			}
		})
	}
}

// TestNewError_ReturnsStringError 测试 NewError 返回 StringError 类型
func TestNewError_ReturnsStringError(t *testing.T) {
	t.Parallel()
	err := NewError("test")
	// 验证类型转换
	if _, ok := err.(StringError); !ok {
		t.Error("NewError 应该返回 StringError 类型")
	}
}
