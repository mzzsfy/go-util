package script

import (
	"testing"
)

// ========== 编译器完整测试 ==========
// 本文件测试脚本引擎编译器的各项功能

// Test_Compiler_Expressions 测试表达式编译功能
func Test_Compiler_Expressions(t *testing.T) {
	t.Run("一元表达式", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"负号", "-5"},
			{"逻辑非", "!true"},
			{"位取反", "^5"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				parser := NewParser()
				_, err := parser.Compile(tt.input)
				if err != nil {
					t.Errorf("一元表达式 %s 编译失败: %v", tt.name, err)
				}
			})
		}
	})

	t.Run("二元算术表达式", func(t *testing.T) {
		tests := []string{
			"1 + 2", "1 - 2", "1 * 2", "1 / 2", "1 % 2",
		}
		for _, tc := range tests {
			compileScript(t, tc)
		}
	})

	t.Run("二元比较表达式", func(t *testing.T) {
		tests := []string{
			"1 == 2", "1 != 2", "1 < 2", "1 > 2", "1 <= 2", "1 >= 2",
		}
		for _, tc := range tests {
			compileScript(t, tc)
		}
	})

	t.Run("二元逻辑表达式", func(t *testing.T) {
		tests := []string{
			"1 && 2", "1 || 2",
		}
		for _, tc := range tests {
			compileScript(t, tc)
		}
	})

	t.Run("二元位运算表达式", func(t *testing.T) {
		tests := []string{
			"1 & 2", "1 | 2", "1 ^ 2", "1 << 2", "1 >> 2",
		}
		for _, tc := range tests {
			compileScript(t, tc)
		}
	})
}

// Test_Compiler_Literals 测试字面量编译
func Test_Compiler_Literals(t *testing.T) {
	t.Run("数组字面量", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"空数组", "[]"},
			{"单元素数组", "[1]"},
			{"多元素数组", "[1, 2, 3]"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				parser := NewParser()
				_, err := parser.Compile(tt.input)
				if err != nil {
					t.Errorf("数组字面量 %s 编译失败: %v", tt.name, err)
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
					t.Errorf("Map字面量 %s 编译失败: %v", tt.name, err)
				}
			})
		}
	})

	t.Run("函数字面量", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"空函数", "fn empty() { }"},
			{"简单函数", "fn simple() { return 1 }"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				parser := NewParser()
				_, err := parser.Compile(tt.input)
				if err != nil {
					t.Errorf("函数字面量 %s 编译失败: %v", tt.name, err)
				}
			})
		}
	})
}

// Test_Compiler_ExternalFunctions 测试外部函数声明编译
func Test_Compiler_ExternalFunctions(t *testing.T) {
	t.Run("外部函数声明", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"无参数外部函数", "#fn externalFunc()=>int\n1 + 1"},
			{"单参数外部函数", "#fn externalFunc(x)=>int\n1 + 1"},
			{"多参数外部函数", "#fn externalFunc(x, y)=>int\n1 + 1"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				parser := NewParser()
				_, err := parser.Compile(tt.input)
				if err != nil {
					t.Errorf("外部函数 %s 编译失败: %v", tt.name, err)
				}
			})
		}
	})
}

// Test_Compiler_BuiltinFunctions 测试内置函数编译和执行
func Test_Compiler_BuiltinFunctions(t *testing.T) {
	t.Run("print函数", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"print单个参数", `print("hello")`},
			{"print多个参数", `print("a", "b", "c")`},
			{"print表达式", `print(1 + 2)`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				parser := NewParser()
				_, err := parser.Compile(tt.input)
				if err != nil {
					t.Errorf("print函数 %s 编译失败: %v", tt.name, err)
				}
			})
		}
	})

	t.Run("len函数", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected int
		}{
			{"len空数组", `len([])`, 0},
			{"len单元素", `len([1])`, 1},
			{"len多元素", `len([1, 2, 3, 4, 5])`, 5},
			{"len空字符串", `len("")`, 0},
			{"len单字符", `len("a")`, 1},
			{"len长字符串", `len("hello world")`, 11},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := Eval(tt.input)
				if err != nil {
					t.Errorf("len函数 %s 执行失败: %v", tt.name, err)
					return
				}
				if result.Int() != tt.expected {
					t.Errorf("%s: 期望 %d, 得到 %d", tt.name, tt.expected, result.Int())
				}
			})
		}
	})

	t.Run("类型转换函数", func(t *testing.T) {
		t.Run("int函数", func(t *testing.T) {
			tests := []struct {
				name     string
				input    string
				expected int
			}{
				{"int正数", `int("123")`, 123},
				{"int负数", `int("-456")`, -456},
				{"int零", `int("0")`, 0},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					result, err := Eval(tt.input)
					if err != nil {
						t.Errorf("int函数 %s 执行失败: %v", tt.name, err)
						return
					}
					if result.Int() != tt.expected {
						t.Errorf("%s: 期望 %d, 得到 %d", tt.name, tt.expected, result.Int())
					}
				})
			}
		})

		t.Run("float函数", func(t *testing.T) {
			tests := []struct {
				name  string
				input string
			}{
				{"float整数", `float("123")`},
				{"float小数", `float("3.14159")`},
				{"float科学计数", `float("1e10")`},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					result, err := Eval(tt.input)
					if err != nil {
						t.Errorf("float函数 %s 执行失败: %v", tt.name, err)
						return
					}
					if result.Float() == 0 {
						t.Errorf("%s: 期望非零浮点数, 得到 %f", tt.name, result.Float())
					}
				})
			}
		})

		t.Run("string函数", func(t *testing.T) {
			tests := []struct {
				name     string
				input    string
				expected string
			}{
				{"string整数", `string(123)`, "123"},
				{"string负数", `string(-456)`, "-456"},
				{"string字符串", `string("already")`, "already"},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					result, err := Eval(tt.input)
					if err != nil {
						t.Errorf("string函数 %s 执行失败: %v", tt.name, err)
						return
					}
					if result.String() != tt.expected {
						t.Errorf("%s: 期望 %s, 得到 %s", tt.name, tt.expected, result.String())
					}
				})
			}
		})
	})

	t.Run("内置函数组合使用", func(t *testing.T) {
		result, err := Eval(`
arr := [1, 2, 3, 4, 5]
s := 0
for v := range arr {
    s = s + v
}
len(arr) + s
`)
		if err != nil {
			t.Errorf("组合使用失败: %v", err)
			return
		}
		// len(arr) = 5, s = 15, 总计 = 20
		if result.Int() != 20 {
			t.Errorf("期望 20, 得到 %d", result.Int())
		}
	})
}

// Test_Compiler_BuiltinEdgeCases 测试内置函数边界情况
func Test_Compiler_BuiltinEdgeCases(t *testing.T) {
	t.Run("print各种类型", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"print整数", `print(1)`},
			{"print浮点", `print(1.5)`},
			{"print字符串", `print("str")`},
			{"print布尔", `print(true)`},
			{"print nil", `print(nil)`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				parser := NewParser()
				_, err := parser.Compile(tt.input)
				if err != nil {
					t.Errorf("print测试失败: %v", err)
				}
			})
		}
	})
}

// ========== 辅助函数 ==========
// compileScript 定义在 t_util_test.go 中

