package script

import (
	"testing"
)

// ========== Parser完整测试 ==========
// 本文件测试脚本引擎Parser的各项功能

// Test_Parser_Statements 测试语句解析
func Test_Parser_Statements(t *testing.T) {
	t.Run("赋值语句", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"变量声明", "x := 10"},
			{"变量赋值", "x = 10"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				parser := NewParser()
				_, err := parser.Compile(tt.input)
				if err != nil {
					t.Errorf("赋值语句 %s 解析失败: %v", tt.name, err)
				}
			})
		}
	})

	t.Run("函数声明", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"简单函数", "fn add(a, b) { return a + b }"},
			{"无参数函数", "fn get() { return 1 }"},
			{"单参数函数", "fn double(x) { return x + x }"},
			{"多参数函数", "fn sum(a, b, c) { return a + b + c }"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				parser := NewParser()
				_, err := parser.Compile(tt.input)
				if err != nil {
					t.Errorf("函数声明 %s 解析失败: %v", tt.name, err)
				}
			})
		}
	})

	t.Run("if语句", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"if-else", "if 1 { 2 } else { 3 }"},
			{"简单if", "if true { 1 }"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				parser := NewParser()
				_, err := parser.Compile(tt.input)
				if err != nil {
					t.Errorf("if语句 %s 解析失败: %v", tt.name, err)
				}
			})
		}
	})
}

// Test_Parser_Comparisons 测试比较表达式
func Test_Parser_Comparisons(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"小于", "1 < 2"},
		{"大于", "1 > 2"},
		{"小于等于", "1 <= 2"},
		{"大于等于", "1 >= 2"},
		{"相等", "1 == 2"},
		{"不等", "1 != 2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			_, err := parser.Compile(tt.input)
			if err != nil {
				t.Errorf("比较表达式 %s 解析失败: %v", tt.name, err)
			}
		})
	}
}

// Test_Parser_Literals 测试字面量解析
func Test_Parser_Literals(t *testing.T) {
	t.Run("数组字面量", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"空数组", "[]"},
			{"多元素数组", "[1, 2, 3]"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				parser := NewParser()
				_, err := parser.Compile(tt.input)
				if err != nil {
					t.Errorf("数组字面量 %s 解析失败: %v", tt.name, err)
				}
			})
		}
	})

	t.Run("Map字面量", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"空Map", "{}"},
			{"单元素Map", `{"a": 1}`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				parser := NewParser()
				_, err := parser.Compile(tt.input)
				if err != nil {
					t.Errorf("Map字面量 %s 解析失败: %v", tt.name, err)
				}
			})
		}
	})

	t.Run("基本类型字面量", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"布尔真", "true"},
			{"布尔假", "false"},
			{"nil", "nil"},
			{"字符串", `"hello"`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				parser := NewParser()
				_, err := parser.Compile(tt.input)
				if err != nil {
					t.Errorf("字面量 %s 解析失败: %v", tt.name, err)
				}
			})
		}
	})

	t.Run("括号表达式", func(t *testing.T) {
		parser := NewParser()
		_, err := parser.Compile("(1 + 2)")
		if err != nil {
			t.Errorf("括号表达式解析失败: %v", err)
		}
	})
}

// Test_Parser_Errors 测试解析错误
func Test_Parser_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"不完整表达式", "1 +"},
		{"不完整if语句", "if { }"},
		{"不完整数组", "[1, 2"},
		{"缺少右括号", "fn test( { return 1 }"},
		{"缺少左大括号", "if true print(1) }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			_, err := parser.Compile(tt.input)
			if err == nil {
				t.Errorf("错误语法 %s 应该返回错误", tt.name)
			}
		})
	}
}

// Test_Parser_ExpectToken 测试token期望
func Test_Parser_ExpectToken(t *testing.T) {
	t.Run("正确的语法", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"if语句", `if true { print(1) }`},
			{"for循环", `for i := 0; i < 5; i = i + 1 { print(i) }`},
			{"函数定义", `fn test() { return 1 }`},
			{"数组定义", `arr := [1, 2, 3]`},
			{"括号表达式", `x := (1 + 2)`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				parser := NewParser()
				_, err := parser.Compile(tt.input)
				if err != nil {
					t.Errorf("正确语法 %s 应该通过: %v", tt.name, err)
				}
			})
		}
	})

	t.Run("错误的语法", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"缺少左大括号", `if true print(1) }`},
			{"缺少右小括号", `fn test( { return 1 }`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				parser := NewParser()
				_, err := parser.Compile(tt.input)
				if err == nil {
					t.Errorf("错误语法 %s 应该返回错误", tt.name)
				}
			})
		}
	})
}

// Test_Parser_TypeAnnotations 测试类型注解
func Test_Parser_TypeAnnotations(t *testing.T) {
	t.Run("函数参数类型注解", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"无参数", `fn get() { return 1 }`},
			{"单参数", `fn double(x) { return x + x }`},
			{"双参数", `fn add(a, b) { return a + b }`},
			{"三参数", `fn sum(a, b, c) { return a + b + c }`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				parser := NewParser()
				_, err := parser.Compile(tt.input)
				if err != nil {
					t.Errorf("函数类型注解 %s 解析失败: %v", tt.name, err)
				}
			})
		}
	})
}

// Test_Parser_Expressions 测试表达式解析
func Test_Parser_Expressions(t *testing.T) {
	t.Run("算术表达式", func(t *testing.T) {
		tests := []string{
			"1 + 2",
			"10 - 5",
			"3 * 4",
			"10 / 2",
			"10 % 3",
		}

		for _, input := range tests {
			compileScript(t, input)
		}
	})

	t.Run("复杂表达式", func(t *testing.T) {
		tests := []string{
			"1 + 2 * 3",
			"(1 + 2) * 3",
			"1 + 2 + 3 + 4",
		}

		for _, input := range tests {
			compileScript(t, input)
		}
	})
}
