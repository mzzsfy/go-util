package script

import (
	"strings"
	"testing"
)

// Test_FriendlyErrorMessages 测试友好的错误消息
func Test_FriendlyErrorMessages(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		shouldContain    string
		shouldNotContain string
	}{
		{
			name:             "缺少右括号",
			input:            `if true { print(1 }`,
			shouldContain:    ")",
			shouldNotContain: "47",
		},
		{
			name:             "缺少左大括号",
			input:            `if true print(1) }`,
			shouldContain:    "{",
			shouldNotContain: "期望 53",
		},
		{
			name:             "缺少赋值运算符",
			input:            `for i + 0 { }`,
			shouldContain:    ":=",
			shouldNotContain: "得到 54",
		},
		{
			name:             "缺少分号",
			input:            `for i := 0 i < 10 { }`,
			shouldContain:    ";",
			shouldNotContain: "期望 68",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			_, err := parser.Compile(tt.input)
			if err == nil {
				t.Error("期望返回错误，但没有")
				return
			}

			errMsg := err.Error()

			// 检查错误消息包含期望的友好描述
			if tt.shouldContain != "" && !strings.Contains(errMsg, tt.shouldContain) {
				t.Errorf("错误消息应该包含 '%s'\n实际错误: %s", tt.shouldContain, errMsg)
			}

			// 检查错误消息不包含数字token ID
			if tt.shouldNotContain != "" && strings.Contains(errMsg, tt.shouldNotContain) {
				t.Errorf("错误消息不应该包含 '%s'\n实际错误: %s", tt.shouldNotContain, errMsg)
			}

			t.Logf("错误消息: %s", errMsg)
		})
	}
}
