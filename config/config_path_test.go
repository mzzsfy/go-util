package config

import (
	"reflect"
	"testing"
)

// pathSample 构造 map/slice 混合嵌套样例
func pathSample() map[string]any {
	return map[string]any{
		"a": map[string]any{
			"b": map[string]any{"c": 1},
		},
		"arr": []any{10, 20, 30},
		"m": []any{
			map[string]any{"k": "v"},
		},
		"grid": []any{[]any{1, 2}, []any{3, 4}},
	}
}

// Test_GetByPath_Nested 深嵌套与数组路径
func Test_GetByPath_Nested(t *testing.T) {
	t.Parallel()
	m := pathSample()
	cases := []struct {
		path string
		want any
	}{
		{"a.b.c", 1},
		{"arr.0", 10},
		{"arr.2", 30},
		{"m.0.k", "v"},
		{"grid[0][1]", 2},
		{"grid.1.0", 3},
	}
	for _, c := range cases {
		if got := GetByPath(m, c.path); got != c.want {
			t.Errorf("GetByPath(%q) = %v, 期望 %v", c.path, got, c.want)
		}
	}
}

// Test_GetByPath_Miss 未命中路径返回 nil 而非 panic
func Test_GetByPath_Miss(t *testing.T) {
	t.Parallel()
	m := pathSample()
	cases := []struct {
		name string
		path string
	}{
		{"键不存在", "no.such.key"},
		{"下标越界", "arr.99"},
		{"负数下标", "arr.-1"},
		{"非数字下标", "arr.abc"},
		{"切片上取字段", "arr.0.k"},
		{"空段路径", "a..b"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := GetByPath(m, c.path); got != nil {
				t.Errorf("GetByPath(%q) = %v, 期望 nil", c.path, got)
			}
		})
	}
}

// Test_GetByPathSlice 直接以切片为根取值
func Test_GetByPathSlice(t *testing.T) {
	t.Parallel()
	s := []any{1, map[string]any{"k": "v"}, []any{"x", "y"}}
	cases := []struct {
		path string
		want any
	}{
		{"0", 1},
		{"1.k", "v"},
		{"2.1", "y"},
		{"[2][0]", "x"},
	}
	for _, c := range cases {
		if got := GetByPathSlice(s, c.path); got != c.want {
			t.Errorf("GetByPathSlice(%q) = %v, 期望 %v", c.path, got, c.want)
		}
	}
	cases = []struct {
		path string
		want any
	}{{"9", nil}, {"abc", nil}}
	for _, c := range cases {
		if got := GetByPathSlice(s, c.path); got != nil {
			t.Errorf("GetByPathSlice(%q) = %v, 期望 nil", c.path, got)
		}
	}
	// 空路径返回整个切片, 与 GetByPathAny 空路径语义对齐
	if got := GetByPathSlice(s, ""); !reflect.DeepEqual(got, s) {
		t.Errorf("GetByPathSlice(\"\") 应返回整个切片, 实际 %v", got)
	}
}

// Test_GetByPath_EmptyPath 空路径返回整个对象, 三个入口语义一致
func Test_GetByPath_EmptyPath(t *testing.T) {
	t.Parallel()
	m := pathSample()
	if got := GetByPath(m, ""); !reflect.DeepEqual(got, m) {
		t.Errorf("GetByPath(\"\") 应返回整个对象, 实际 %v", got)
	}
	if got := GetByPathAny(m, ""); got == nil {
		t.Error("GetByPathAny(\"\") 应返回对象本身")
	}
}

// Test_GetByPathAny_PathValidation 合法性校验分支与空路径语义
func Test_GetByPathAny_PathValidation(t *testing.T) {
	t.Parallel()
	m := pathSample()
	// 空路径返回对象本身
	if got := GetByPathAny(m, ""); got == nil {
		t.Error("空路径应返回对象本身")
	}
	// 非法路径段返回 nil
	for _, p := range []string{".a", "a.", "a..b"} {
		if got := GetByPathAny(m, p); got != nil {
			t.Errorf("GetByPathAny(%q) = %v, 期望 nil", p, got)
		}
	}
	// 合法路径正常取值
	if got := GetByPathAny(m, "a.b.c"); got != 1 {
		t.Errorf("GetByPathAny(a.b.c) = %v, 期望 1", got)
	}
}

// Test_FindFirstKey_Composite 组合索引路径的分段
func Test_FindFirstKey_Composite(t *testing.T) {
	t.Parallel()
	cases := []struct {
		path    string
		key     string
		nextKey string
	}{
		{"grid[0][1]", "grid", "0[1]"},
		{"a[0]", "a", "0"},
		{"a.b.c", "a", "b.c"},
		{"", "", ""},
	}
	for _, c := range cases {
		key, nextKey := FindFirstKey(c.path)
		if key != c.key || nextKey != c.nextKey {
			t.Errorf("FindFirstKey(%q) = (%q, %q), 期望 (%q, %q)", c.path, key, nextKey, c.key, c.nextKey)
		}
	}
}
