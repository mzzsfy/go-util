package script

import (
	"testing"
)

// TestBuiltinDelete 测试 delete 内置函数
func TestBuiltinDelete(t *testing.T) {
	t.Run("从map中删除存在的key", func(t *testing.T) {
		code := `
			m := {"a": 1, "b": 2, "c": 3}
			delete(m, "b")
			len(m)
		`
		result, err := Eval(code)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		// 删除后长度应为 2
		if result.Int() != 2 {
			t.Errorf("期望 len(m) == 2, 实际: %d", result.Int())
		}
	})

	t.Run("从map中删除不存在的key", func(t *testing.T) {
		code := `
			m := {"a": 1, "b": 2}
			delete(m, "c")
			len(m)
		`
		result, err := Eval(code)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		// 删除不存在的key，长度应不变
		if result.Int() != 2 {
			t.Errorf("期望 len(m) == 2, 实际: %d", result.Int())
		}
	})

	t.Run("delete返回nil", func(t *testing.T) {
		code := `
			m := {"a": 1}
			delete(m, "a")
		`
		result, err := Eval(code)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		if !result.IsNil() {
			t.Errorf("期望 delete 返回 nil, 实际: %v", result)
		}
	})

	t.Run("验证key已被删除", func(t *testing.T) {
		code := `
			m := {"a": 1, "b": 2}
			delete(m, "a")
			m["a"]
		`
		result, err := Eval(code)
		if err != nil {
			t.Fatalf("执行失败: %v", err)
		}
		// 访问已删除的key应返回nil
		if !result.IsNil() {
			t.Errorf("期望 m[\"a\"] == nil, 实际: %v", result)
		}
	})
}
