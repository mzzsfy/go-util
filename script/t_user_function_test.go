package script

import (
	"testing"
)

// TestUserFunction 测试用户定义函数
func TestUserFunction(t *testing.T) {
	t.Run("简单函数定义和调用", func(t *testing.T) {
		code := `
			fn add(a, b) {
                a + b
            }
            add(1, 2)
        `
		result, err := Eval(code)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		if result.Int() != 3 {
			t.Errorf("期望 3, 实际: %d", result.Int())
		}
	})

	t.Run("函数返回语句", func(t *testing.T) {
		code := `
            fn multiply(a, b) {
                return a * b
            }
            multiply(3, 4)
        `
		result, err := Eval(code)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		if result.Int() != 12 {
			t.Errorf("期望 12, 实际: %d", result.Int())
		}
	})

	t.Run("递归函数", func(t *testing.T) {
		code := `
            fn factorial(n) {
                if n <= 1 {
                    return 1
                }
                return n * factorial(n - 1)
            }
            factorial(5)
        `
		result, err := Eval(code)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		if result.Int() != 120 {
			t.Errorf("期望 120, 实际: %d", result.Int())
		}
	})

	t.Run("函数作为参数", func(t *testing.T) {
		code := `
			fn double(x) {
				x * 2
			}
            fn apply(f, x) {
                f(x)
            }
            apply(double, 20)
        `
		result, err := Eval(code)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		if result.Int() != 40 {
			t.Errorf("期望 40, 实际: %d", result.Int())
		}
	})

	t.Run("空函数体", func(t *testing.T) {
		code := `
            fn empty() {}
            empty()
        `
		result, err := Eval(code)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		if !result.IsNil() {
			t.Errorf("期望 nil, 实际: %v", result)
		}
	})

	t.Run("函数无参数", func(t *testing.T) {
		code := `
            fn answer() {
                42
            }
            answer()
        `
		result, err := Eval(code)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		if result.Int() != 42 {
			t.Errorf("期望 42, 实际: %d", result.Int())
		}
	})

	t.Run("参数数量不匹配应报错", func(t *testing.T) {
		code := `
            fn add(a, b) {
                a + b
            }
            add(1)
        `
		_, err := Eval(code)
		if err == nil {
			t.Error("期望参数数量不匹配错误")
		}
	})

	t.Run("嵌套调用", func(t *testing.T) {
		code := `
            fn a(x) { x + 1 }
            fn b(x) { a(x) + 1 }
            fn c(x) { b(x) + 1 }
            c(0)
        `
		result, err := Eval(code)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		if result.Int() != 3 {
			t.Errorf("期望 3, 实际: %d", result.Int())
		}
	})
}
