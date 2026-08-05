package script

import (
	"testing"
)

// ========== VM完整测试 ==========
// 本文件测试脚本引擎VM的各项功能

// Test_VM_Constants 测试常量处理
func Test_VM_Constants(t *testing.T) {
	t.Run("基本类型常量", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected interface{}
		}{
			{"整数常量", `42`, 42},
			{"浮点常量", `3.14`, 3.14},
			{"字符串常量", `"hello"`, "hello"},
			{"布尔true", `true`, true},
			{"布尔false", `false`, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := Eval(tt.input)
				if err != nil {
					t.Errorf("常量加载失败: %v", err)
					return
				}

				switch expected := tt.expected.(type) {
				case int:
					if result.Int() != expected {
						t.Errorf("期望 %v, 得到 %v", expected, result.Int())
					}
				case float64:
					if result.Float() != expected {
						t.Errorf("期望 %v, 得到 %v", expected, result.Float())
					}
				case string:
					if result.String() != expected {
						t.Errorf("期望 %v, 得到 %v", expected, result.String())
					}
				case bool:
					if result.Bool() != expected {
						t.Errorf("期望 %v, 得到 %v", expected, result.Bool())
					}
				}
			})
		}
	})

	t.Run("nil常量", func(t *testing.T) {
		result, err := Eval(`nil`)
		if err != nil {
			t.Errorf("nil常量加载失败: %v", err)
			return
		}
		if !result.IsNil() {
			t.Errorf("期望 nil, 得到 %v", result)
		}
	})
}

// Test_VM_Variables 测试变量操作
func Test_VM_Variables(t *testing.T) {
	t.Run("局部变量", func(t *testing.T) {
		result, err := Eval(`
x := 10
y := x
y
`)
		if err != nil {
			t.Errorf("局部变量操作失败: %v", err)
			return
		}
		if result.Int() != 10 {
			t.Errorf("期望 10, 得到 %d", result.Int())
		}
	})

	t.Run("变量交换", func(t *testing.T) {
		result, err := Eval(`
a := 1
b := 2
c := 3
temp := a
a = b
b = c
c = temp
a + b + c
`)
		if err != nil {
			t.Errorf("变量交换失败: %v", err)
			return
		}
		if result.Int() != 6 {
			t.Errorf("期望 6, 得到 %d", result.Int())
		}
	})
}

// Test_VM_TypeConversion 测试类型转换
func Test_VM_TypeConversion(t *testing.T) {
	t.Run("布尔值转换", func(t *testing.T) {
		t.Run("布尔参数传递", func(t *testing.T) {
			ctx := NewContext()
			ctx.BindFunc("isPositive", func(val int) bool {
				return val > 0
			})

			parser := NewParser()
			script, err := parser.Compile(`#fn isPositive(val int)=>bool
isPositive(5)`)
			if err != nil {
				t.Fatalf("编译失败: %v", err)
			}

			engine := NewEngine()
			result, err := engine.Run(ctx, script)
			if err != nil {
				t.Fatalf("执行失败: %v", err)
			}

			if result.Bool() != true {
				t.Errorf("期望 true, 得到 %v", result.Bool())
			}
		})

		t.Run("布尔返回值", func(t *testing.T) {
			ctx := NewContext()
			ctx.BindFunc("and", func(a, b bool) bool {
				return a && b
			})

			parser := NewParser()
			script, err := parser.Compile(`#fn and(a bool, b bool)=>bool
and(true, false)`)
			if err != nil {
				t.Fatalf("编译失败: %v", err)
			}

			engine := NewEngine()
			result, err := engine.Run(ctx, script)
			if err != nil {
				t.Fatalf("执行失败: %v", err)
			}

			if result.Bool() != false {
				t.Errorf("期望 false, 得到 %v", result.Bool())
			}
		})

		t.Run("布尔值在条件中使用", func(t *testing.T) {
			result, err := Eval(`
flag := true
result := 0
if flag {
    result = 1
} else {
    result = 0
}
result
`)
			if err != nil {
				t.Errorf("布尔值条件执行失败: %v", err)
				return
			}

			if result.Int() != 1 {
				t.Errorf("期望 1, 得到 %d", result.Int())
			}
		})
	})

	t.Run("数组访问", func(t *testing.T) {
		result, err := Eval(`
arr := [[1, 2], [3, 4]]
arr[0][1]
`)
		if err != nil {
			t.Errorf("多维数组访问失败: %v", err)
			return
		}
		if result.Int() != 2 {
			t.Errorf("期望 2, 得到 %d", result.Int())
		}
	})
}

// Test_VM_Context 测试Context功能
func Test_VM_Context(t *testing.T) {
	t.Run("Context克隆", func(t *testing.T) {
		t.Run("Clone空context", func(t *testing.T) {
			ctx := NewContext()
			cloned := ctx.Clone()
			if cloned == nil {
				t.Error("Clone不应该返回nil")
			}
		})

		t.Run("Clone带绑定值的context", func(t *testing.T) {
			ctx := NewContext()
			ctx.BindValue("x", 10)
			ctx.BindValue("y", 20)

			cloned := ctx.Clone()

			val, ok := cloned.GetBindValue("x")
			if !ok || val.Int() != 10 {
				t.Error("Clone的context应该有相同的绑定值")
			}

			val, ok = cloned.GetBindValue("y")
			if !ok || val.Int() != 20 {
				t.Error("Clone的context应该有相同的绑定值")
			}
		})

		t.Run("Clone带绑定函数的context", func(t *testing.T) {
			ctx := NewContext()
			ctx.BindFunc("double", func(x int) int {
				return x * 2
			})

			cloned := ctx.Clone()

			fn, ok := cloned.GetBindFunc("double")
			if !ok || fn == nil {
				t.Error("Clone的context应该有相同的绑定函数")
			}
		})
	})

	t.Run("外部函数缓存", func(t *testing.T) {
		ctx := NewContext()
		ctx.BindFunc("test", func(x int) int {
			return x * 2
		})

		parser := NewParser()
		script, err := parser.Compile(`#fn test(x int)=>int
test(5)`)
		if err != nil {
			t.Fatalf("编译失败: %v", err)
		}

		engine := NewEngine()
		result, err := engine.Run(ctx, script)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}

		if result.Int() != 10 {
			t.Errorf("期望 10, 得到 %d", result.Int())
		}
	})
}

// Test_VM_Engine 测试Engine功能
func Test_VM_Engine(t *testing.T) {
	t.Run("Engine停止", func(t *testing.T) {
		t.Run("Stop空引擎", func(t *testing.T) {
			engine := NewEngine()
			engine.Stop()
		})

		t.Run("Stop有VM的引擎", func(t *testing.T) {
			engine := NewEngine()
			ctx := NewContext()
			parser := NewParser()

			script, err := parser.Compile(`1 + 1`)
			if err != nil {
				t.Fatalf("编译失败: %v", err)
			}

			_, err = engine.Run(ctx, script)
			if err != nil {
				t.Fatalf("运行失败: %v", err)
			}

			engine.Stop()
		})
	})
}
