package script

import (
	"testing"
)

// TestIfExpression 测试 if 表达式返回值
// if 表达式出现在表达式位置（如变量声明的右边），返回值
// if 语句出现在语句位置，不返回值
func TestIfExpression(t *testing.T) {
	t.Run("基本if表达式赋值", func(t *testing.T) {
		code := `
			result := if 10 > 5 { "large" } else { "small" }
			result
		`
		result, err := Eval(code)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		if result.String() != "large" {
			t.Errorf("期望 \"large\", 实际: %s", result.String())
		}
	})

	t.Run("if表达式条件为假", func(t *testing.T) {
		code := `
			result := if 3 > 5 { "large" } else { "small" }
			result
		`
		result, err := Eval(code)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		if result.String() != "small" {
			t.Errorf("期望 \"small\", 实际: %s", result.String())
		}
	})

	t.Run("if表达式无else返回nil", func(t *testing.T) {
		code := `
			result := if 3 > 5 { "large" }
			result
		`
		result, err := Eval(code)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		// 没有else分支且条件为假，应返回nil
		if !result.IsNil() {
			t.Errorf("期望 nil, 实际: %v", result)
		}
	})

	t.Run("if表达式嵌套使用", func(t *testing.T) {
		code := `
			x := 15
			category := if x > 20 {
				"huge"
			} else if x > 10 {
				"medium"
			} else {
				"small"
			}
			category
		`
		result, err := Eval(code)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		if result.String() != "medium" {
			t.Errorf("期望 \"medium\", 实际: %s", result.String())
		}
	})

	t.Run("if表达式返回数字", func(t *testing.T) {
		code := `
			x := 5
			result := if x > 3 { x * 2 } else { x + 1 }
			result
		`
		result, err := Eval(code)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		if result.Int() != 10 {
			t.Errorf("期望 10, 实际: %d", result.Int())
		}
	})

	t.Run("if表达式作为函数参数", func(t *testing.T) {
		code := `
			#fn len(s: string)=>int
			result := len(if true { "hello" } else { "world" })
			result
		`
		ctx := NewContext()
		ctx.BindValue("len", func(args []Value) (Value, error) {
			s := args[0].String()
			return NewValue(len(s)), nil
		})
		parser := NewParser()
		script, err := parser.Compile(code)
		if err != nil {
			t.Fatalf("编译失败: %v", err)
		}
		engine := NewEngine()
		result, err := engine.Run(ctx, script)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		if result.Int() != 5 {
			t.Errorf("期望 5, 实际: %d", result.Int())
		}
	})

	t.Run("if表达式在数组中", func(t *testing.T) {
		code := `
			arr := [if true { 1 } else { 2 }, if false { 3 } else { 4 }]
			arr
		`
		result, err := Eval(code)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		// 数组被存储为 ArrayValue
		arr := result.Array()
		if len(arr.Elements) != 2 {
			t.Fatalf("期望长度 2, 实际: %d", len(arr.Elements))
		}
		if arr.Elements[0].Int() != 1 {
			t.Errorf("期望 arr[0] = 1, 实际: %d", arr.Elements[0].Int())
		}
		if arr.Elements[1].Int() != 4 {
			t.Errorf("期望 arr[1] = 4, 实际: %d", arr.Elements[1].Int())
		}
	})
}
